package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

type AnnotationTypeUseCase struct {
	repo port.Repository[*model.AnnotationType]
	uow  port.UnitOfWorkFactory
}

func NewAnnotationTypeUseCase(repo port.Repository[*model.AnnotationType], uow port.UnitOfWorkFactory) *AnnotationTypeUseCase {
	return &AnnotationTypeUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *AnnotationTypeUseCase) Create(ctx context.Context, entity *model.AnnotationType) (*model.AnnotationType, error) {
	var createdAnnotationType *model.AnnotationType

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Validate parent type is None (AnnotationType has no parent)
		if entity.Parent.Type != vobj.ParentTypeNone || entity.Parent.ID != "" {
			return errors.NewValidationError("annotation type cannot have a parent", map[string]interface{}{
				"parent_type": entity.Parent.Type,
				"parent_id":   entity.Parent.ID,
			})
		}

		// Check name uniqueness
		isUnique, err := CheckNameUniqueInCollection(txCtx, uc.repo, entity.Name)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("annotation type name already exists", map[string]interface{}{
				"name": entity.Name,
			})
		}

		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create annotation type", err)
		}

		createdAnnotationType = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdAnnotationType, nil
}

func (uc *AnnotationTypeUseCase) Update(ctx context.Context, annotationTypeID string, updates map[string]interface{}) error {
	// AnnotationType has restricted updates
	// Immutable fields: tag_type, is_global, is_required
	immutableFields := []string{constants.TagTypeField, constants.TagGlobalField, constants.TagRequiredField}
	for _, field := range immutableFields {
		if _, ok := updates[field]; ok {
			return errors.NewValidationError("field is immutable and cannot be updated", map[string]interface{}{
				"field":   field,
				"message": "this field cannot be changed after creation",
			})
		}
	}

	currentAnnotationType, err := uc.repo.Read(ctx, annotationTypeID)
	if err != nil {
		return errors.NewInternalError("failed to read annotation type", err)
	}

	// Check if options are being updated
	if options, ok := updates[constants.TagOptionsField]; ok {
		// Options can only be updated for SELECT and MULTI_SELECT types
		if currentAnnotationType.TagType != vobj.MultiSelectTag && currentAnnotationType.TagType != vobj.SelectTag {
			return errors.NewValidationError("options can only be updated for SELECT or MULTI_SELECT types", map[string]interface{}{
				"annotation_type_id": annotationTypeID,
				"current_tag_type":   currentAnnotationType.TagType,
			})
		}

		newOptions, ok := options.([]string)
		if !ok {
			return errors.NewValidationError("invalid options format", map[string]interface{}{
				"annotation_type_id": annotationTypeID,
			})
		}

		// Check if any annotation uses this annotation type
		inUse, err := uc.isAnnotationTypeInUseByAnyAnnotation(ctx, currentAnnotationType.Name)
		if err != nil {
			return errors.NewInternalError("failed to check if annotation type is in use", err)
		}

		if !inUse {
			// No annotations using this type, safe to update options
			goto UPDATE
		}

		// Annotations exist, check if removed options are in use
		inUse, usedOption, err := uc.checkRemovedOptionsInUse(ctx, currentAnnotationType, newOptions)
		if err != nil {
			return errors.NewInternalError("failed to check if options are in use", err)
		}
		if inUse {
			return errors.NewConflictError("cannot remove option that is in use by annotations", map[string]interface{}{
				"annotation_type_id": annotationTypeID,
				"used_option":        usedOption,
			})
		}
	}

UPDATE:
	err = uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Name update requires updating all annotations
		if name, ok := updates[constants.NameField]; ok {
			newName := name.(string)

			// Check uniqueness
			isUnique, err := CheckNameUniqueInCollection(txCtx, uc.repo, newName, annotationTypeID)
			if err != nil {
				return errors.NewInternalError("failed to check name uniqueness", err)
			}
			if !isUnique {
				return errors.NewConflictError("annotation type name already exists", map[string]interface{}{
					"name": newName,
				})
			}

			// Update all annotations with this name
			err = uc.updateAnnotationNames(txCtx, currentAnnotationType.Name, newName)
			if err != nil {
				return err
			}
		}

		// Update annotation type
		err := uc.repo.Update(txCtx, annotationTypeID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update annotation type", err)
		}

		return nil
	})

	return err
}

// updateAnnotationNames updates all annotations that reference the old annotation type name
func (uc *AnnotationTypeUseCase) updateAnnotationNames(ctx context.Context, oldName, newName string) error {
	annotationRepo := uc.uow.GetAnnotationRepo()

	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    oldName,
		},
		{
			Field:    constants.DeletedField,
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	// Use pagination to handle large datasets
	const limit = 100
	offset := 0

	for {
		result, err := annotationRepo.FindByFilters(ctx, filters, &query.Pagination{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return errors.NewInternalError("failed to fetch annotations", err)
		}

		if len(result.Data) == 0 {
			break
		}

		// Update batch
		for _, annotation := range result.Data {
			err := annotationRepo.Update(ctx, annotation.GetID(), map[string]interface{}{
				constants.NameField: newName,
			})
			if err != nil {
				return errors.NewInternalError("failed to update annotation name", err)
			}
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return nil
}

func (uc *AnnotationTypeUseCase) checkRemovedOptionsInUse(ctx context.Context, annotationType *model.AnnotationType, newOptions []string) (bool, string, error) {
	// Identify removed options
	newOptionsMap := make(map[string]bool)
	for _, opt := range newOptions {
		newOptionsMap[opt] = true
	}

	var removedOptions []string
	for _, currentOpt := range annotationType.Options {
		if !newOptionsMap[currentOpt] {
			removedOptions = append(removedOptions, currentOpt)
		}
	}

	if len(removedOptions) == 0 {
		return false, "", nil
	}

	// Fetch annotations using this annotation type
	annotationRepo := uc.uow.GetAnnotationRepo()

	filters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    annotationType.Name,
		},
		{
			Field:    constants.DeletedField,
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	removedOptionsMap := make(map[string]bool)
	for _, opt := range removedOptions {
		removedOptionsMap[opt] = true
	}

	const limit = 1000
	offset := 0

	for {
		result, err := annotationRepo.FindByFilters(ctx, filters, &query.Pagination{
			Limit:  limit,
			Offset: offset,
		})
		if err != nil {
			return false, "", err
		}

		if len(result.Data) == 0 {
			break
		}

		// Check each annotation's value
		for _, annotation := range result.Data {
			// Handle SELECT type (string value)
			if valueStr, ok := annotation.Value.(string); ok {
				if removedOptionsMap[valueStr] {
					return true, valueStr, nil
				}
			}

			// Handle MULTI_SELECT type ([]string value)
			if valueSlice, ok := annotation.Value.([]string); ok {
				for _, v := range valueSlice {
					if removedOptionsMap[v] {
						return true, v, nil
					}
				}
			}

			// Handle MULTI_SELECT from JSON unmarshalling ([]interface{})
			if valueInterface, ok := annotation.Value.([]interface{}); ok {
				for _, item := range valueInterface {
					if str, ok := item.(string); ok {
						if removedOptionsMap[str] {
							return true, str, nil
						}
					}
				}
			}
		}

		if !result.HasMore {
			break
		}

		offset += limit
	}

	return false, "", nil
}

// isAnnotationTypeInUseByAnyAnnotation checks if any annotation uses this annotation type
func (uc *AnnotationTypeUseCase) isAnnotationTypeInUseByAnyAnnotation(ctx context.Context, annotationTypeName string) (bool, error) {
	annotationRepo := uc.uow.GetAnnotationRepo()

	count, err := annotationRepo.Count(ctx, []query.Filter{
		{Field: constants.NameField, Operator: query.OpEqual, Value: annotationTypeName},
		{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}
