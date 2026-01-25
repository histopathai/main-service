package usecase

import (
	"context"
	"fmt"
	"slices"

	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationUseCase struct {
	repo port.Repository[*model.Annotation]
	uow  port.UnitOfWorkFactory
}

func NewAnnotationUseCase(repo port.Repository[*model.Annotation], uow port.UnitOfWorkFactory) *AnnotationUseCase {
	return &AnnotationUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *AnnotationUseCase) Create(ctx context.Context, entity *model.Annotation) (*model.Annotation, error) {
	var createdAnnotation *model.Annotation

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// 1. Check parent exists (must be Image)
		if err := CheckParentExists(txCtx, &entity.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent image not found", map[string]interface{}{
				"parent_id":   entity.GetParent().ID,
				"parent_type": entity.GetParent().Type,
				"error":       err.Error(),
			})
		}

		// 2. Validate parent type is Image
		if entity.Parent.Type != vobj.ParentTypeImage {
			return errors.NewValidationError("annotation parent must be an image", map[string]interface{}{
				"parent_type": entity.Parent.Type,
				"expected":    vobj.ParentTypeImage,
			})
		}

		// 2. Check if AnnotationType with this name exists
		annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()

		annotationTypeFilters := query.NewBuilder()
		annotationTypeFilters.Where(fields.EntityName.DomainName(), query.OpEqual, entity.Name)
		annotationTypeFilters.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
		annotationTypeFilters.Limit(1)

		annotationTypeResult, err := annotationTypeRepo.Find(txCtx, annotationTypeFilters.Build())
		if err != nil {
			return errors.NewInternalError("failed to fetch annotation type", err)
		}

		if len(annotationTypeResult.Data) == 0 {
			return errors.NewValidationError("annotation type not found", map[string]interface{}{
				"name": entity.Name,
			})
		}

		annotationType := annotationTypeResult.Data[0]

		// 3. Check tag type matches
		if entity.TagType != annotationType.TagType {
			return errors.NewValidationError("tag type mismatch", map[string]interface{}{
				"expected": annotationType.TagType,
				"got":      entity.TagType,
			})
		}

		// 4. Check value is valid according to annotation type
		if err := uc.validateAnnotationValue(entity.Value, annotationType); err != nil {
			return err
		}

		if entity.WsID == "" {
			// Fetch parent image to get workspace ID
			imageRepo := uc.uow.GetImageRepo()
			parentImage, err := imageRepo.Read(txCtx, entity.Parent.ID)
			if err != nil {
				return errors.NewInternalError("failed to read parent image", err)
			}
			entity.WsID = parentImage.WsID
		}

		// Create annotation
		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create annotation", err)
		}

		createdAnnotation = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdAnnotation, nil
}

func (uc *AnnotationUseCase) Update(ctx context.Context, annotationID string, updates map[string]interface{}) error {
	// Annotation name cannot be updated (it references annotation type)
	if _, ok := updates[fields.EntityName.DomainName()]; ok {
		return errors.NewValidationError("annotation name cannot be updated", map[string]interface{}{
			"field":   "name",
			"message": "name field references annotation type and is immutable",
		})
	}

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Read current annotation
		currentAnnotation, err := uc.repo.Read(txCtx, annotationID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation", err)
		}

		// If value is being updated, validate it
		if value, ok := updates[fields.AnnotationTagValue.DomainName()]; ok {
			// Find annotation type
			annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()

			annotationTypeFilters := query.NewBuilder()
			annotationTypeFilters.Where(fields.EntityName.DomainName(), query.OpEqual, currentAnnotation.Name)
			annotationTypeFilters.Where(fields.EntityIsDeleted.DomainName(), query.OpEqual, false)
			annotationTypeFilters.Limit(1)

			annotationTypeResult, err := annotationTypeRepo.Find(txCtx, annotationTypeFilters.Build())
			if err != nil {
				return errors.NewInternalError("failed to fetch annotation type", err)
			}

			if len(annotationTypeResult.Data) == 0 {
				return errors.NewValidationError("annotation type not found", map[string]interface{}{
					"name": currentAnnotation.Name,
				})
			}

			annotationType := annotationTypeResult.Data[0]

			// Validate new value
			if err := uc.validateAnnotationValue(value, annotationType); err != nil {
				return err
			}
		}

		// Polygon can be updated (free, no validation beyond basic structure)
		// Color can be updated (free, no validation)

		// Perform update
		err = uc.repo.Update(txCtx, annotationID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update annotation", err)
		}

		return nil
	})

	return err
}

func (uc *AnnotationUseCase) validateAnnotationValue(value interface{}, annotationType *model.AnnotationType) error {
	switch annotationType.TagType {
	case vobj.NumberTag:
		// Number validation
		var numValue float64

		switch v := value.(type) {
		case float64:
			numValue = v
		case float32:
			numValue = float64(v)
		case int:
			numValue = float64(v)
		case int64:
			numValue = float64(v)
		default:
			return errors.NewValidationError("value must be a number", map[string]interface{}{
				"value": value,
				"type":  fmt.Sprintf("%T", value),
			})
		}

		// Min control
		if annotationType.Min != nil && numValue < *annotationType.Min {
			return errors.NewValidationError("value is below minimum", map[string]interface{}{
				"value": numValue,
				"min":   *annotationType.Min,
			})
		}

		// Max control
		if annotationType.Max != nil && numValue > *annotationType.Max {
			return errors.NewValidationError("value is above maximum", map[string]interface{}{
				"value": numValue,
				"max":   *annotationType.Max,
			})
		}

	case vobj.TextTag:
		// Text validation - must be a string
		_, ok := value.(string)
		if !ok {
			return errors.NewValidationError("value must be a string for text tag", map[string]interface{}{
				"value": value,
				"type":  fmt.Sprintf("%T", value),
			})
		}

	case vobj.BooleanTag:
		// Boolean validation
		_, ok := value.(bool)
		if !ok {
			return errors.NewValidationError("value must be a boolean", map[string]interface{}{
				"value": value,
				"type":  fmt.Sprintf("%T", value),
			})
		}

	case vobj.SelectTag:
		// Single option validation
		valueStr, ok := value.(string)
		if !ok {
			return errors.NewValidationError("value must be a string for select", map[string]interface{}{
				"value": value,
				"type":  fmt.Sprintf("%T", value),
			})
		}

		// Is value in options?
		if len(annotationType.Options) == 0 {
			return errors.NewValidationError("annotation type has no options defined", map[string]interface{}{
				"annotation_type": annotationType.Name,
			})
		}

		found := slices.Contains(annotationType.Options, valueStr)

		if !found {
			return errors.NewValidationError("value is not in allowed options", map[string]interface{}{
				"value":   valueStr,
				"options": annotationType.Options,
			})
		}

	case vobj.MultiSelectTag:
		// Multiple options validation
		var valueSlice []string

		switch v := value.(type) {
		case []string:
			valueSlice = v
		case []interface{}:
			valueSlice = make([]string, len(v))
			for i, item := range v {
				str, ok := item.(string)
				if !ok {
					return errors.NewValidationError("all values must be strings", map[string]interface{}{
						"value": value,
					})
				}
				valueSlice[i] = str
			}
		default:
			return errors.NewValidationError("value must be an array for multi-select", map[string]interface{}{
				"value": value,
				"type":  fmt.Sprintf("%T", value),
			})
		}

		// Are options defined?
		if len(annotationType.Options) == 0 {
			return errors.NewValidationError("annotation type has no options defined", map[string]interface{}{
				"annotation_type": annotationType.Name,
			})
		}

		// At least one option must be selected
		if len(valueSlice) == 0 {
			return errors.NewValidationError("at least one option must be selected", map[string]interface{}{
				"annotation_type": annotationType.Name,
			})
		}

		// Is each value in options?
		optionsMap := make(map[string]bool)
		for _, option := range annotationType.Options {
			optionsMap[option] = true
		}

		for _, val := range valueSlice {
			if !optionsMap[val] {
				return errors.NewValidationError("value is not in allowed options", map[string]interface{}{
					"value":   val,
					"options": annotationType.Options,
				})
			}
		}

	default:
		return errors.NewValidationError("unsupported tag type", map[string]interface{}{
			"tag_type": annotationType.TagType,
		})
	}

	return nil
}
