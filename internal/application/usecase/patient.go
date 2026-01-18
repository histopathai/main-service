package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type PatientUseCase struct {
	repo port.Repository[*model.Patient]
	uow  port.UnitOfWorkFactory
}

func NewPatientUseCase(repo port.Repository[*model.Patient], uow port.UnitOfWorkFactory) *PatientUseCase {
	return &PatientUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *PatientUseCase) Create(ctx context.Context, entity *model.Patient) (*model.Patient, error) {
	var createdPatient *model.Patient

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if err := CheckParentExists(txCtx, &entity.Parent, uc.uow); err != nil {
			return errors.NewValidationError("parent validation failed", map[string]interface{}{
				"parent_type": entity.GetParent().Type,
				"parent_id":   entity.GetParent().ID,
				"error":       err.Error(),
			})
		}

		parentWorkspace, err := uc.uow.GetWorkspaceRepo().Read(txCtx, entity.Parent.ID)
		if err != nil {
			return errors.NewInternalError("failed to read parent workspace", err)
		}

		if len(parentWorkspace.AnnotationTypes) == 0 {
			return errors.NewValidationError("parent workspace has no annotation types defined", map[string]interface{}{
				"parent_id": entity.Parent.ID,
			})
		}

		isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, entity.Name, entity.Parent.ID)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("patient name already exists", map[string]interface{}{
				"name":      entity.Name,
				"parent_id": entity.Parent.ID,
			})
		}

		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return errors.NewInternalError("failed to create patient", err)
		}

		createdPatient = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdPatient, nil
}

func (uc *PatientUseCase) Update(ctx context.Context, patientID string, updates map[string]interface{}) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		if name, ok := updates[constants.NameField]; ok {
			currentPatient, err := uc.repo.Read(txCtx, patientID)
			if err != nil {
				return errors.NewInternalError("failed to read patient", err)
			}

			isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, name.(string), currentPatient.Parent.ID, patientID)
			if err != nil {
				return errors.NewInternalError("failed to check name uniqueness", err)
			}
			if !isUnique {
				return errors.NewConflictError("patient name already exists", map[string]interface{}{
					"name":      name,
					"parent_id": currentPatient.Parent.ID,
				})
			}
		}

		err := uc.repo.Update(txCtx, patientID, updates)
		if err != nil {
			return errors.NewInternalError("failed to update patient", err)
		}

		return nil
	})

	return err
}

func (uc *PatientUseCase) Delete(ctx context.Context, patientID string) error {
	// Use soft delete for now
	return uc.repo.SoftDelete(ctx, patientID)
}

func (uc *PatientUseCase) TransferPatient(ctx context.Context, patientID string, newParentID string) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		currentPatient, err := uc.repo.Read(txCtx, patientID)
		if err != nil {
			return errors.NewInternalError("failed to read patient", err)
		}

		newParentWorkspace, err := uc.uow.GetWorkspaceRepo().Read(txCtx, newParentID)
		if err != nil {
			return errors.NewValidationError("new parent workspace does not exist", map[string]interface{}{
				"parent_id": newParentID,
				"error":     err.Error(),
			})
		}

		if len(newParentWorkspace.AnnotationTypes) == 0 {
			return errors.NewValidationError("new parent workspace has no annotation types defined", map[string]interface{}{
				"parent_id": newParentID,
			})
		}

		isUnique, err := CheckNameUniqueUnderParent(txCtx, uc.repo, currentPatient.Name, newParentID)
		if err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		}
		if !isUnique {
			return errors.NewConflictError("patient with same name already exists in new parent", map[string]interface{}{
				"name":      currentPatient.Name,
				"parent_id": newParentID,
			})
		}

		updates := map[string]interface{}{
			constants.ParentIDField: newParentID,
		}

		err = uc.repo.Update(txCtx, patientID, updates)
		if err != nil {
			return errors.NewInternalError("failed to transfer patient", err)
		}

		return nil
	})

	return err
}
