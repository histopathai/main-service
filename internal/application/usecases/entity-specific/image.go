package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateImageUseCase struct {
	uowFactory repository.UnitOfWorkFactory
}

func NewCreateImageUseCase(uowFactory repository.UnitOfWorkFactory) *CreateImageUseCase {
	return &CreateImageUseCase{uowFactory: uowFactory}
}

func (uc *CreateImageUseCase) Execute(ctx context.Context, entity *model.Image) (*model.Image, error) {
	var createdEntity *model.Image

	uowerr := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {

		parentID := entity.GetParent().GetID()

		parentEntity, err := repos.PatientRepo.Read(txCtx, parentID)
		if err != nil {
			return err
		}

		if parentEntity == nil {
			details := map[string]any{
				"where":     "CreateImageUseCase.Execute",
				"type":      "parent not found",
				"parent_id": parentID,
			}
			return errors.NewValidationError("parent patient not found for image", details)
		}

		if parentEntity.Deleted {
			details := map[string]any{
				"where":     "CreateImageUseCase.Execute",
				"type":      "parent deleted",
				"parent_id": parentID,
			}
			return errors.NewValidationError("cannot add image to deleted patient", details)
		}

		created, err := repos.ImageRepo.Create(txCtx, entity)
		if err != nil {
			return err
		}

		//Update parent child count
		err = repos.PatientRepo.Update(txCtx, parentID, map[string]any{
			constants.ChildCountField: parentEntity.GetChildCount() + 1,
		})
		if err != nil {
			return err
		}

		createdEntity = created

		return nil
	})

	if uowerr != nil {
		return nil, uowerr
	}
	return createdEntity, nil
}

func (uc *CreateImageUseCase) ExecuteMany(ctx context.Context, entities []model.Image) ([]model.Image, error) {
	created := make([]model.Image, 0, len(entities))
	// Bulk operations do not applied here, Later we can improve it if needed
	for _, entity := range entities {
		createdEntity, err := uc.Execute(ctx, &entity)
		if err != nil {
			return nil, err
		}
		created = append(created, *createdEntity)
	}
	return created, nil
}
