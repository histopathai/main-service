package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageUseCase struct {
	repo      port.ImageRepository
	uow       port.UnitOfWorkFactory
	validator *validator.ImageValidator
	storage   port.Storage // Origin Image Storage
}

func NewImageUseCase(repo port.ImageRepository, uow port.UnitOfWorkFactory, storage port.Storage) *ImageUseCase {
	return &ImageUseCase{
		repo:      repo,
		uow:       uow,
		validator: validator.NewImageValidator(repo, uow),
		storage:   storage,
	}
}

func (uc *ImageUseCase) Upload(ctx context.Context, cmd command.UploadImageCommand) (*port.PresignedURLPayload, error) {

	image, err := cmd.ToEntity()
	if err != nil {
		return nil, err
	}

	image.SetID(uuid.New().String())

	content := cmd.GetContent()

	if content == nil {
		return nil, errors.NewValidationError("content is required", map[string]interface{}{
			"Upload":  "Origin Image upload",
			"content": "Content is required",
		})
	}

	if !content.ContentType.IsOriginImage() {
		return nil, errors.NewValidationError("content type is not origin image", map[string]interface{}{
			"Upload":  "Origin Image upload",
			"content": "Content type is not origin image",
		})
	}

	content.SetParent(&vobj.ParentRef{
		ID:   image.GetID(),
		Type: vobj.ParentTypeImage,
	})

	content.Provider = uc.storage.Provider()
	content.SetID(uuid.New().String())

	content.Path = fmt.Sprintf("%s/%s", content.GetID(), content.Name)

	presignedURLPayload, err := uc.storage.GenerateSignedURL(ctx, port.MethodPut, *content, time.Duration(1*time.Hour))
	if err != nil {
		return nil, errors.NewInternalError("failed to generate presigned url", err)
	}

	image.OriginContentID = &content.ID

	image.Processing = &vobj.ProcessingInfo{
		Status:          vobj.StatusPending,
		Version:         vobj.ProcessingV2,
		RetryCount:      0,
		LastProcessedAt: time.Now(),
	}

	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.validator.ValidateCreate(txCtx, image); err != nil {
			return err
		}

		createdEntity, err := uc.repo.Create(txCtx, image)
		if err != nil {
			return errors.NewInternalError("failed to create image", err)
		}
		if createdEntity == nil {
			return errors.NewInternalError("failed to create image", nil)
		}

		return nil
	})

	if uowerr != nil {
		return nil, uowerr
	}

	return presignedURLPayload, nil

}

func (uc *ImageUseCase) Update(ctx context.Context, cmd command.UpdateImageCommand) error {
	updates := cmd.GetUpdates()
	if updates == nil {
		return errors.NewInternalError("no updates provided", nil)
	}

	id := cmd.GetID()

	if err := uc.repo.Update(ctx, id, updates); err != nil {
		return errors.NewInternalError("failed to update image", err)
	}

	return nil
}

func (uc *ImageUseCase) Transfer(ctx context.Context, cmd *command.TransferCommand) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.validator.ValidateTransfer(txCtx, cmd); err != nil {
			return err
		}

		id := cmd.GetID()

		if err := uc.repo.Transfer(txCtx, id, cmd.GetNewParent()); err != nil {
			return errors.NewInternalError("failed to transfer image", err)
		}

		return nil
	})

	return err
}

func (uc *ImageUseCase) TransferMany(ctx context.Context, cmd *command.TransferManyCommand) error {

	ch := make(chan error)
	for _, id := range cmd.GetIDs() {
		go func(id string) {
			ch <- uc.Transfer(ctx, &command.TransferCommand{
				ID:         id,
				NewParent:  cmd.GetNewParent(),
				OldParent:  cmd.GetOldParent(),
				ParentType: vobj.EntityTypeImage.String(),
			})
		}(id)
	}

	for range cmd.GetIDs() {
		if err := <-ch; err != nil {
			return err
		}
	}

	return nil

}
