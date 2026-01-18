package usecase

import (
	"context"
	"fmt"
	"slices"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
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
		// 1. Check parent exists
		if err := CheckParentExists(txCtx, &entity.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent image not found", map[string]interface{}{
				"parent_id": entity.GetParent().ID,
				"error":     err.Error(),
			})
		}

		// 2. Check if AnnotationType with this name exists
		annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()

		annotationTypeFilters := []query.Filter{
			{
				Field:    constants.NameField,
				Operator: query.OpEqual,
				Value:    entity.Name,
			},
			{
				Field:    constants.DeletedField,
				Operator: query.OpEqual,
				Value:    false,
			},
		}

		annotationTypeResult, err := annotationTypeRepo.FindByFilters(txCtx, annotationTypeFilters, &query.Pagination{Limit: 1})
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
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {

		//Read current annotation
		currentAnnotation, err := uc.repo.Read(txCtx, annotationID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation", err)
		}

		// If value is being updated, validate it
		if value, ok := updates["tag_value"]; ok {
			// Find current annotation type
			annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()

			annotationTypeFilters := []query.Filter{
				{
					Field:    constants.NameField,
					Operator: query.OpEqual,
					Value:    currentAnnotation.Name,
				},
				{
					Field:    constants.DeletedField,
					Operator: query.OpEqual,
					Value:    false,
				},
			}

			annotationTypeResult, err := annotationTypeRepo.FindByFilters(txCtx, annotationTypeFilters, &query.Pagination{Limit: 1})
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

		// Polygon can be updated (free, no validation)
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
