package entityspecific

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/constants"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type CreateaAnnotationUseCase struct {
	uowFactory port.UnitOfWorkFactory
}

func NewCreateAnnotationUseCase(uowFactory port.UnitOfWorkFactory) *CreateaAnnotationUseCase {
	return &CreateaAnnotationUseCase{uowFactory: uowFactory}
}

func (uc *CreateaAnnotationUseCase) Execute(ctx context.Context, entity *model.Annotation) (*model.Annotation, error) {
	createdEntity := &model.Annotation{}
	uowerr := uc.uowFactory.WithTx(ctx, func(txCtx context.Context, repos *port.Repositories) error {

		//check parent existence
		parentID := entity.GetParent().GetID()

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

type UpdateAnnotationUseCase struct {
	uowFactory port.UnitOfWorkFactory
}

func NewUpdateAnnotationUseCase(uowFactory port.UnitOfWorkFactory) *UpdateAnnotationUseCase {
	return &UpdateAnnotationUseCase{uowFactory: uowFactory}
}

func (uc *UpdateAnnotationUseCase) Execute(ctx context.Context, id string, updates map[string]any) (*model.Annotation, error) {
	// Will be improve later
	repos, err := uc.uowFactory.WithoutTx(ctx)
	if err != nil {
		return nil, err
	}

	err = repos.AnnotationRepo.Update(ctx, id, updates)
	if err != nil {
		return nil, err
	}

	updatedEntity, err := repos.AnnotationRepo.Read(ctx, id)
	if err != nil {
		return nil, err
	}
	return updatedEntity, nil
}
