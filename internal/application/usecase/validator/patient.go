package validator

import (
	"context"

	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/helper"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type PatientValidator struct {
	repo port.PatientRepository
	uow  port.UnitOfWorkFactory
}

func NewPatientValidator(repo port.PatientRepository, uow port.UnitOfWorkFactory) *PatientValidator {
	return &PatientValidator{repo: repo, uow: uow}
}

func (v *PatientValidator) ValidateCreate(ctx context.Context, patient *model.Patient) error {
	// Check if parent exists
	if err := helper.CheckParentExists(ctx, &patient.Parent, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}

	if isUnique, err := helper.CheckNameUniqueUnderParent(ctx, v.repo, patient.Name, patient.Parent.ID); err != nil {
		return errors.NewInternalError("failed to check name uniqueness", err)
	} else if !isUnique {
		return errors.NewConflictError("patient name already exists", map[string]interface{}{
			"name":      patient.Name,
			"parent_id": patient.Parent.ID,
		})
	}
	return nil
}

func (v *PatientValidator) ValidateUpdate(ctx context.Context, id string, updates map[string]interface{}) error {

	if parentID, ok := updates[fields.EntityParentID.DomainName()]; ok {
		if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: parentID.(string), Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
			return errors.NewInternalError("failed to check parent exists", err)
		}
	}

	if name, ok := updates[fields.EntityName.DomainName()]; ok {
		if isUnique, err := helper.CheckNameUniqueUnderParent(ctx, v.repo, name.(string), id); err != nil {
			return errors.NewInternalError("failed to check name uniqueness", err)
		} else if !isUnique {
			return errors.NewConflictError("patient name already exists", map[string]interface{}{
				"name":      name,
				"parent_id": id,
			})
		}
	}
	return nil
}

func (v *PatientValidator) ValidateTransfer(ctx context.Context, command *command.TransferCommand) error {
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: command.GetNewParent(), Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}
	if err := helper.CheckParentExists(ctx, &vobj.ParentRef{ID: command.GetOldParent(), Type: vobj.ParentTypeWorkspace}, v.uow); err != nil {
		return errors.NewInternalError("failed to check parent exists", err)
	}

	if isUnique, err := helper.CheckNameUniqueUnderParent(ctx, v.repo, command.GetNewParent(), command.GetOldParent()); err != nil {
		return errors.NewInternalError("failed to check name uniqueness", err)
	} else if !isUnique {
		return errors.NewConflictError("patient name already exists", map[string]interface{}{
			"name":      command.GetNewParent(),
			"parent_id": command.GetOldParent(),
		})
	}

	return nil
}
