package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateaAnnotationUseCase struct {
	uowFactory repository.UnitOfWorkFactory
}

func NewCreateAnnotationUseCase(uowFactory repository.UnitOfWorkFactory) *CreateaAnnotationUseCase {
	return &CreateaAnnotationUseCase{uowFactory: uowFactory}
}

func (uc *CreateaAnnotationUseCase) Execute(ctx context.Context, entity *model.Annotation) (*model.Annotation, error) {
	createdEntity := &model.Annotation{}
	uowerr := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *repository.Repositories) error {

		//check parent existence
		parentID := entity.GetParentID()

		parentEntity, err := repos.ImageRepo.Read(txCtx, parentID)
		if err != nil {
			return err
		}

		if parentEntity == nil {
			details := map[string]any{
				"where":     "CreateAnnotationUseCase.Execute",
				"type":      "parent not found",
				"parent_id": parentID,
			}
			return errors.NewValidationError("parent image not found for annotation", details)
		}

		if parentEntity.IsDeleted() {
			details := map[string]any{
				"where":     "CreateAnnotationUseCase.Execute",
				"type":      "parent deleted",
				"parent_id": parentID,
			}
			return errors.NewValidationError("cannot add annotation to deleted image", details)
		}

		created, err := repos.AnnotationRepo.Create(txCtx, entity)
		if err != nil {
			return err
		}

		//update parent child count
		err = repos.ImageRepo.Update(txCtx, parentID, map[string]any{
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
