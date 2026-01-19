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

type WorkspaceUseCase struct {
	repo port.Repository[*model.Workspace]
	uow  port.UnitOfWorkFactory
}

func (uc *WorkspaceUseCase) Create(ctx context.Context, entity *model.Workspace) (*model.Workspace, error) {
	var createdWorkspace *model.Workspace

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {

		isUnique, err := CheckNameUniqueInCollection(txCtx, uc.repo, entity.Name)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("workspace name already exists", map[string]interface{}{
				"name": entity.Name,
			})
		}

		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create workspace", err)
		}

		createdWorkspace = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdWorkspace, nil
}

func (uc *WorkspaceUseCase) Update(ctx context.Context, id string, updates map[string]interface{}) error {
	if annotationTypes, ok := updates["annotation_types"]; ok {
		newATList, ok := annotationTypes.([]string)
		if !ok {
			return errors.NewValidationError("invalid annotation_types format", map[string]interface{}{
				"annotation_types": annotationTypes,
			})
		}

		currentEntity, err := uc.repo.Read(ctx, id)
		if err != nil {
			return errors.NewInternalError("failed to read current workspace", err)
		}

		newATMap := make(map[string]bool)
		for _, atID := range newATList {
			newATMap[atID] = true
		}

		annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()

		for _, currentATID := range currentEntity.AnnotationTypes {
			if !newATMap[currentATID] {
				annotationType, err := annotationTypeRepo.Read(ctx, currentATID)
				if err != nil {
					return errors.NewInternalError("failed to read annotation type", err)
				}

				inUse, err := uc.CheckAnnotationTypeInUse(ctx, id, annotationType.Name)
				if err != nil {
					return errors.NewInternalError("failed to check annotation type usage", err)
				}

				if inUse {
					return errors.NewConflictError("cannot remove annotation type that is in use", map[string]interface{}{
						"annotation_type_id":   currentATID,
						"annotation_type_name": annotationType.Name,
					})
				}
			}
		}
	}

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if name, ok := updates[constants.NameField]; ok {
			isUnique, err := CheckNameUniqueInCollection(txCtx, uc.repo, name.(string), id)
			if err != nil {
				return errors.NewInternalError("failed to check name uniqueness", err)
			}
			if !isUnique {
				return errors.NewConflictError("workspace name already exists", map[string]interface{}{
					"name": name,
				})
			}
		}

		err := uc.repo.Update(txCtx, id, updates)
		if err != nil {
			return errors.NewInternalError("failed to update workspace", err)
		}

		return nil
	})

	return err
}

func (uc *WorkspaceUseCase) CheckAnnotationTypeInUse(ctx context.Context, workspaceID string, annotationTypeName string) (bool, error) {
	patientRepo := uc.uow.GetPatientRepo()
	imageRepo := uc.uow.GetImageRepo()
	annotationRepo := uc.uow.GetAnnotationRepo()

	patientFilters := []query.Filter{
		{Field: constants.ParentIDField, Operator: query.OpEqual, Value: workspaceID},
		{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
	}

	patientResult, err := patientRepo.FindByFilters(ctx, patientFilters, &query.Pagination{Limit: 1000})
	if err != nil {
		return false, fmt.Errorf("failed to fetch patients: %w", err)
	}

	if len(patientResult.Data) == 0 {
		return false, nil
	}

	patientIDs := make([]string, len(patientResult.Data))
	for i, patient := range patientResult.Data {
		patientIDs[i] = patient.GetID()
	}

	const batchSize = 30
	var allImageIDs []string

	for i := 0; i < len(patientIDs); i += batchSize {
		end := i + batchSize
		if end > len(patientIDs) {
			end = len(patientIDs)
		}
		batch := patientIDs[i:end]

		imageFilters := []query.Filter{
			{Field: constants.ParentIDField, Operator: query.OpIn, Value: batch},
			{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
		}

		imageResult, err := imageRepo.FindByFilters(ctx, imageFilters, &query.Pagination{Limit: 1000})
		if err != nil {
			return false, fmt.Errorf("failed to fetch images batch %d: %w", i/batchSize, err)
		}

		for _, image := range imageResult.Data {
			allImageIDs = append(allImageIDs, image.GetID())
		}
	}

	if len(allImageIDs) == 0 {
		return false, nil
	}

	for i := 0; i < len(allImageIDs); i += batchSize {
		end := i + batchSize
		if end > len(allImageIDs) {
			end = len(allImageIDs)
		}
		batch := allImageIDs[i:end]

		annotationFilters := []query.Filter{
			{Field: constants.ParentIDField, Operator: query.OpIn, Value: batch},
			{Field: constants.NameField, Operator: query.OpEqual, Value: annotationTypeName},
			{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
		}

		count, err := annotationRepo.Count(ctx, annotationFilters)
		if err != nil {
			return false, fmt.Errorf("failed to check annotations batch %d: %w", i/batchSize, err)
		}

		if count > 0 {
			return true, nil
		}
	}

	return false, nil
}
