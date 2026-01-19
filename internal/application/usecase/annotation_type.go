package usecase

import (
	"context"
	"fmt"

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
	currentAnnotationType, err := uc.repo.Read(ctx, annotationTypeID)
	if err != nil {
		return errors.NewInternalError("failed to read annotation type", err)
	}

	if options, ok := updates["options"]; ok {
		newOptions, ok := options.([]string)
		if !ok {
			return errors.NewValidationError("invalid options format", map[string]interface{}{
				"options": options,
			})
		}

		currentOptionsMap := make(map[string]bool)
		for _, opt := range currentAnnotationType.Options {
			currentOptionsMap[opt] = true
		}

		newOptionsMap := make(map[string]bool)
		for _, opt := range newOptions {
			newOptionsMap[opt] = true
		}

		var removedOptions []string
		for _, currentOpt := range currentAnnotationType.Options {
			if !newOptionsMap[currentOpt] {
				removedOptions = append(removedOptions, currentOpt)
			}
		}

		if len(removedOptions) > 0 {
			inUse, usedOption, err := uc.checkOptionsInUse(ctx, currentAnnotationType.Name, removedOptions)
			if err != nil {
				return errors.NewInternalError("failed to check options usage", err)
			}
			if inUse {
				return errors.NewConflictError("cannot remove option that is in use", map[string]interface{}{
					"option":             usedOption,
					"annotation_type_id": annotationTypeID,
				})
			}
		}
	}

	err = uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if name, ok := updates["name"]; ok {
			newName := name.(string)

			isUnique, err := CheckNameUniqueInCollection(txCtx, uc.repo, newName, annotationTypeID)
			if err != nil {
				return errors.NewInternalError("failed to check name uniqueness", err)
			}
			if !isUnique {
				return errors.NewConflictError("annotation type name already exists", map[string]interface{}{
					"name": newName,
				})
			}

			annotationRepo := uc.uow.GetAnnotationRepo()

			annotationFilters := []query.Filter{
				{
					Field:    constants.NameField,
					Operator: query.OpEqual,
					Value:    currentAnnotationType.Name,
				},
				{
					Field:    constants.DeletedField,
					Operator: query.OpEqual,
					Value:    false,
				},
			}

			annotationResult, err := annotationRepo.FindByFilters(txCtx, annotationFilters, &query.Pagination{Limit: 10000})
			if err != nil {
				return errors.NewInternalError("failed to fetch annotations", err)
			}

			for _, annotation := range annotationResult.Data {
				err := annotationRepo.Update(txCtx, annotation.GetID(), map[string]interface{}{
					constants.NameField: newName,
				})
				if err != nil {
					return errors.NewInternalError(fmt.Sprintf("failed to update annotation %s", annotation.GetID()), err)
				}
			}
		}

		err := uc.repo.Update(txCtx, annotationTypeID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update annotation type", err)
		}

		return nil
	})

	return err
}

func (uc *AnnotationTypeUseCase) checkOptionsInUse(ctx context.Context, annotationTypeName string, removedOptions []string) (bool, string, error) {
	annotationRepo := uc.uow.GetAnnotationRepo()

	annotationFilters := []query.Filter{
		{
			Field:    constants.NameField,
			Operator: query.OpEqual,
			Value:    annotationTypeName,
		},
		{
			Field:    constants.DeletedField,
			Operator: query.OpEqual,
			Value:    false,
		},
	}

	annotationResult, err := annotationRepo.FindByFilters(ctx, annotationFilters, &query.Pagination{Limit: 10000})
	if err != nil {
		return false, "", fmt.Errorf("failed to fetch annotations: %w", err)
	}

	removedOptionsMap := make(map[string]bool)
	for _, opt := range removedOptions {
		removedOptionsMap[opt] = true
	}

	for _, annotation := range annotationResult.Data {
		if valueStr, ok := annotation.Value.(string); ok {
			if removedOptionsMap[valueStr] {
				return true, valueStr, nil
			}
		}

		if valueSlice, ok := annotation.Value.([]interface{}); ok {
			for _, v := range valueSlice {
				if vStr, ok := v.(string); ok {
					if removedOptionsMap[vStr] {
						return true, vStr, nil
					}
				}
			}
		}
	}

	return false, "", nil
}
