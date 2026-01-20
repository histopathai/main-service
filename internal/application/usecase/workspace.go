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

type WorkspaceUseCase struct {
	repo port.Repository[*model.Workspace]
	uow  port.UnitOfWorkFactory
}

func NewWorkspaceUseCase(repo port.Repository[*model.Workspace], uow port.UnitOfWorkFactory) *WorkspaceUseCase {
	return &WorkspaceUseCase{
		repo: repo,
		uow:  uow,
	}
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

		// STEP 1: Check if workspace has any patients
		hasPatients, err := uc.hasAnyPatients(ctx, id)
		if err != nil {
			return errors.NewInternalError("failed to check patients existence", err)
		}

		if !hasPatients {
			// No patients -> Safe to update any types
			goto UPDATE
		}

		// STEP 2: Workspace has patients, check if annotation types are being removed
		currentEntity, err := uc.repo.Read(ctx, id)
		if err != nil {
			return errors.NewInternalError("failed to read current workspace", err)
		}

		removedATIDs := uc.getRemovedAnnotationTypes(currentEntity.AnnotationTypes, newATList)

		if len(removedATIDs) == 0 {
			// Only additions → Safe
			goto UPDATE
		}

		// STEP 3: Check if workspace has any annotations
		hasAnnotations, err := uc.hasAnyAnnotations(ctx, id)
		if err != nil {
			return errors.NewInternalError("failed to check annotations existence", err)
		}

		if !hasAnnotations {
			// No annotations yet → Safe to remove any types
			goto UPDATE
		}

		// STEP 4: Annotations exist, validate removed types are not in use
		err = uc.validateRemovedAnnotationTypesNotInUse(ctx, id, removedATIDs)
		if err != nil {
			return err
		}
	}

UPDATE:
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

func (uc *WorkspaceUseCase) hasAnyPatients(ctx context.Context, workspaceID string) (bool, error) {
	patientRepo := uc.uow.GetPatientRepo()

	count, err := patientRepo.Count(ctx, []query.Filter{
		{Field: constants.ParentIDField, Operator: query.OpEqual, Value: workspaceID},
		{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (uc *WorkspaceUseCase) hasAnyAnnotations(ctx context.Context, workspaceID string) (bool, error) {
	annotationRepo := uc.uow.GetAnnotationRepo()

	count, err := annotationRepo.Count(ctx, []query.Filter{
		{Field: constants.WsIDField, Operator: query.OpEqual, Value: workspaceID},
		{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
	})

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (uc *WorkspaceUseCase) getRemovedAnnotationTypes(current, new []string) []string {
	newMap := make(map[string]bool)
	for _, atID := range new {
		newMap[atID] = true
	}

	var removed []string
	for _, currentATID := range current {
		if !newMap[currentATID] {
			removed = append(removed, currentATID)
		}
	}

	return removed
}

func (uc *WorkspaceUseCase) validateRemovedAnnotationTypesNotInUse(ctx context.Context, workspaceID string, removedATIDs []string) error {
	annotationTypeRepo := uc.uow.GetAnnotationTypeRepo()
	annotationRepo := uc.uow.GetAnnotationRepo()

	// Check each removed annotation type
	for _, atID := range removedATIDs {
		// Get annotation type to access its name
		annotationType, err := annotationTypeRepo.Read(ctx, atID)
		if err != nil {
			return errors.NewInternalError("failed to read annotation type", err)
		}

		// Count annotations using this annotation type name in this workspace
		count, err := annotationRepo.Count(ctx, []query.Filter{
			{Field: constants.WsIDField, Operator: query.OpEqual, Value: workspaceID},
			{Field: constants.NameField, Operator: query.OpEqual, Value: annotationType.Name},
			{Field: constants.DeletedField, Operator: query.OpEqual, Value: false},
		})

		if err != nil {
			return errors.NewInternalError("failed to check annotation usage", err)
		}

		if count > 0 {
			return errors.NewConflictError("cannot remove annotation type that is in use", map[string]interface{}{
				"annotation_type_id":   atID,
				"annotation_type_name": annotationType.Name,
				"usage_count":          count,
			})
		}
	}

	return nil
}
