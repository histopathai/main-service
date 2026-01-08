package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
	"github.com/histopathai/main-service/internal/shared/query"
)

// Interface'leri implement ettiğini garanti et
var _ CreateExecutor[model.Patient] = (*CreatePatientUseCase)(nil)
var _ UpdateExecutor[model.Patient] = (*UpdatePatientUseCase)(nil)

type CreatePatientUseCase struct {
	uowFactory port.UnitOfWorkFactory
}

func NewCreatePatientUseCase(uowFactory port.UnitOfWorkFactory) *CreatePatientUseCase {
	return &CreatePatientUseCase{uowFactory: uowFactory}
}

func (uc *CreatePatientUseCase) Execute(ctx context.Context, entity *model.Patient) (*model.Patient, error) {
	createdEntity := &model.Patient{}
	uowerr := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		// Check parent ID existence
		parentID := entity.GetParent().GetID()

		parentEntity, err := repos.WorkspaceRepo.Read(txCtx, parentID)
		if err != nil {
			return err
		}

		if parentEntity == nil {
			details := map[string]any{
				"where":     "CreatePatientUseCase.Execute",
				"type":      "parent not found",
				"parent_id": parentID,
			}
			return errors.NewValidationError("parent workspace not found for patient", details)
		}

		if parentEntity.Deleted {
			details := map[string]any{
				"where":     "CreatePatientUseCase.Execute",
				"type":      "parent deleted",
				"parent_id": parentID,
			}
			return errors.NewValidationError("cannot add patient to deleted workspace", details)
		}

		if parentEntity.GetParent().GetID() == "" {
			details := map[string]any{
				"where":     "CreatePatientUseCase.Execute",
				"type":      "invalid parent",
				"parent_id": parentID,
			}
			return errors.NewValidationError("Workspace has no valid annotation type", details)
		}

		// Check uniqueness constraints in parent scope if any (e.g., name uniqueness within workspace)
		filter := []query.Filter{
			{
				Field:    constants.NameField,
				Operator: query.OpEqual,
				Value:    entity.Name,
			},
			{
				Field:    constants.ParentTypeField,
				Operator: query.OpEqual,
				Value:    vobj.EntityTypeWorkspace,
			},
			{
				Field:    constants.ParentIDField,
				Operator: query.OpEqual,
				Value:    parentID,
			},
		}

		count, err := repos.PatientRepo.Count(txCtx, filter)
		if err != nil {
			return err
		}
		if count > 0 {
			details := map[string]any{
				"where":     "CreatePatientUseCase.Execute",
				"type":      "uniqueness violation",
				"name":      entity.Name,
				"parent_id": parentID,
			}
			return errors.NewConflictError("patient with the given name already exists in the workspace", details)
		}
		created, err := repos.PatientRepo.Create(txCtx, entity)
		if err != nil {
			return err
		}
		createdEntity = created

		// Update parent child count
		err = repos.WorkspaceRepo.Update(txCtx, parentID, map[string]any{
			constants.ChildCountField: parentEntity.GetChildCount() + 1,
		})
		if err != nil {
			return err
		}

		return nil
	})
	if uowerr != nil {
		return nil, uowerr
	}
	return createdEntity, nil
}

func (uc *CreatePatientUseCase) ExecuteMany(ctx context.Context, entities []model.Patient) ([]model.Patient, error) {
	created := []model.Patient{}

	// Create patients one by one to ensure validations and UoW are applied
	for _, entity := range entities {
		createdEntity, err := uc.Execute(ctx, &entity)
		if err != nil {
			return nil, err
		}
		created = append(created, *createdEntity)
	}
	return created, nil
}

type UpdatePatientUseCase struct {
	uowFactory port.UnitOfWorkFactory
}

func NewUpdatePatientUseCase(uowFactory port.UnitOfWorkFactory) *UpdatePatientUseCase {
	return &UpdatePatientUseCase{uowFactory: uowFactory}
}

func (uc *UpdatePatientUseCase) Execute(ctx context.Context, id string, updates map[string]any) (*model.Patient, error) {

	uowerr := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {
		if name, ok := updates[constants.NameField]; ok {
			// Check name uniqueness within parent scope
			filters := []query.Filter{
				{
					Field:    constants.NameField,
					Operator: query.OpEqual,
					Value:    name,
				},
				{
					Field:    constants.ParentTypeField,
					Operator: query.OpEqual,
					Value:    vobj.EntityTypeWorkspace,
				},
			}
			count, err := repos.PatientRepo.Count(txCtx, filters)
			if err != nil {
				return err
			}
			if count > 0 {
				details := map[string]any{
					"where": "UpdatePatientUseCase.Execute",
					"type":  "uniqueness violation",
					"name":  name,
				}
				return errors.NewConflictError("patient with the given name already exists", details)
			}
		}
		if _, ok := updates[constants.ParentIDField]; ok {
			// Transfer ops are not allowed in update use case
			details := map[string]any{
				"where": "UpdatePatientUseCase.Execute",
				"type":  "invalid update field",
			}
			return errors.NewValidationError("updating parent ID is not allowed in update use case", details)
		}

		// Perform the update

		return repos.PatientRepo.Update(txCtx, id, updates)
	})

	if uowerr != nil {
		return nil, uowerr
	}

	repos, err := uc.uowFactory.WithoutTx(ctx)
	if err != nil {
		return nil, err
	}
	updatedEntity, err := repos.PatientRepo.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}

func (uc *UpdatePatientUseCase) ExecuteMany(ctx context.Context, ids []string, updates map[string]any) error {

	// Perform updates one by one to ensure validations and UoW are applied
	for _, id := range ids {
		if _, err := uc.Execute(ctx, id, updates); err != nil {
			return err
		}
	}
	return nil
}
