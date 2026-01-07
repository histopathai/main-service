package composite

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type TransferUseCase struct {
	uowFactory repository.UnitOfWorkFactory
}

func NewTransferUseCase(uowFactory repository.UnitOfWorkFactory) *TransferUseCase {
	return &TransferUseCase{
		uowFactory: uowFactory,
	}
}

func (uc *TransferUseCase) Execute(ctx context.Context, id, newParentID string, entityType vobj.EntityType) error {
	switch entityType {
	case vobj.EntityTypePatient:
		return uc.transferPatient(ctx, id, newParentID)
	case vobj.EntityTypeImage:
		return uc.transferImage(ctx, id, newParentID)
	default:
		return errors.NewValidationError("unsupported entity type for transfer", nil)
	}
}

func (uc *TransferUseCase) ExecuteMany(ctx context.Context, ids []string, newParentID string, entityType vobj.EntityType) error {
	//Will be optimized later with batch transfer if needed
	for _, id := range ids {
		if err := uc.Execute(ctx, id, newParentID, entityType); err != nil {
			return err
		}
	}
	return nil
}

func (uc *TransferUseCase) transferPatient(ctx context.Context, id, parentId string) error {
	return uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
		patient, err := repos.PatientRepo.Read(txCtx, id)
		if err != nil {
			return err
		}

		oldWorkspaceID := (*patient).GetParent().GetID()

		if oldWorkspaceID == "" {
			return errors.NewValidationError("patient has no workspace", nil)
		}

		oldWorkspace, err := repos.WorkspaceRepo.Read(txCtx, oldWorkspaceID)
		if err != nil {
			return err
		}

		newWorkspace, err := repos.WorkspaceRepo.Read(txCtx, parentId)
		if err != nil {
			return err
		}

		// Transfer the patient
		if err := repos.PatientRepo.Transfer(txCtx, id, parentId, constants.ParentIDField); err != nil {
			return err
		}

		// Update child counts
		if err := repos.WorkspaceRepo.Update(txCtx, oldWorkspace.ID, map[string]any{
			constants.ChildCountField: oldWorkspace.GetChildCount() - 1,
		}); err != nil {
			return err
		}

		if err := repos.WorkspaceRepo.Update(txCtx, newWorkspace.ID, map[string]any{
			constants.ChildCountField: newWorkspace.GetChildCount() + 1,
		}); err != nil {
			return err
		}

		return nil
	})
}

func (uc *TransferUseCase) transferImage(ctx context.Context, id, parentId string) error {
	return uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {
		image, err := repos.ImageRepo.Read(txCtx, id)
		if err != nil {
			return err
		}

		oldPatientID := image.GetParent().GetID()

		if oldPatientID == "" {
			return errors.NewValidationError("image has no patient", nil)
		}

		oldPatient, err := repos.PatientRepo.Read(txCtx, oldPatientID)
		if err != nil {
			return err
		}

		newPatient, err := repos.PatientRepo.Read(txCtx, parentId)
		if err != nil {
			return err
		}

		// Transfer the image
		if err := repos.ImageRepo.Transfer(txCtx, id, parentId, constants.ParentIDField); err != nil {
			return err
		}

		// Update child counts
		if err := repos.PatientRepo.Update(txCtx, oldPatient.ID, map[string]any{
			constants.ChildCountField: oldPatient.GetChildCount() - 1,
		}); err != nil {
			return err
		}

		if err := repos.PatientRepo.Update(txCtx, newPatient.ID, map[string]any{
			constants.ChildCountField: newPatient.GetChildCount() + 1,
		}); err != nil {
			return err
		}

		return nil
	})
}
