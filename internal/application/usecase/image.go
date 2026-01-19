package usecase

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

type ImageUseCase struct {
	repo port.Repository[*model.Image]
	uow  port.UnitOfWorkFactory
}

func NewImageUseCase(repo port.Repository[*model.Image], uow port.UnitOfWorkFactory) *ImageUseCase {
	return &ImageUseCase{
		repo: repo,
		uow:  uow,
	}
}

func (uc *ImageUseCase) Create(ctx context.Context, entity *model.Image) (*model.Image, error) {
	var createdImage *model.Image
	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {
		// Implement any necessary validation here

		created, err := uc.repo.Create(txCtx, entity)
		if err != nil {
			return err
		}
		createdImage = created
		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdImage, nil
}

func (uc *ImageUseCase) Update(ctx context.Context, imageID string, updates map[string]interface{}) error {

	// Implement the logic to update an image

	err := uc.repo.Update(ctx, imageID, updates)
	if err != nil {
		return err
	}
	return nil
}

func (uc *ImageUseCase) Transfer(ctx context.Context, imageID string, newParent vobj.ParentRef) error {

	err := uc.uow.WithTx(ctx, func(txCtx context.Context, repos map[vobj.EntityType]any) error {

		// Check if new parent exists
		if err := CheckParentExists(txCtx, &newParent, uc.uow); err != nil {
			return err
		}

		err := uc.repo.Transfer(txCtx, imageID, newParent.ID)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
