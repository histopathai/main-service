package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase/validator"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageUseCase struct {
	repo             port.ImageRepository
	uow              port.UnitOfWorkFactory
	imageValidator   *validator.ImageValidator
	originStorage    port.Storage // Origin Image Storage
	processedStorage port.Storage // Processed Image Storage
}

func NewImageUseCase(repo port.ImageRepository, uow port.UnitOfWorkFactory, originStorage port.Storage, processedStorage port.Storage) *ImageUseCase {
	return &ImageUseCase{
		repo:             repo,
		uow:              uow,
		imageValidator:   validator.NewImageValidator(repo, uow),
		originStorage:    originStorage,
		processedStorage: processedStorage,
	}
}

func (uc *ImageUseCase) Upload(ctx context.Context, cmd command.UploadImageCommand) ([]port.PresignedURLPayload, error) {

	image, err := cmd.ToEntity()
	if err != nil {
		return nil, err
	}

	image.SetID(uuid.New().String())

	image.Processing = &vobj.ProcessingInfo{
		Status:          vobj.StatusPending,
		Version:         vobj.ProcessingV2,
		RetryCount:      0,
		LastProcessedAt: time.Now(),
	}

	var createdImage *model.Image
	uowerr := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.imageValidator.ValidateCreate(txCtx, image); err != nil {
			return err
		}

		createdEntity, err := uc.repo.Create(txCtx, image)
		if err != nil {
			return errors.NewInternalError("failed to create image", err)
		}
		if createdEntity == nil {
			return errors.NewInternalError("failed to create image", nil)
		}
		createdImage = createdEntity

		return nil
	})

	if uowerr != nil {
		return nil, uowerr
	}

	if createdImage == nil {
		return nil, errors.NewInternalError("failed to create image", nil)
	}

	presignedURLs, err := uc.generatePresignedURLS(ctx, cmd, createdImage.ID)
	if err != nil {
		return nil, err
	}

	return presignedURLs, nil

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

func (uc *ImageUseCase) Transfer(ctx context.Context, cmd command.TransferCommand) error {
	err := uc.uow.WithTx(ctx, func(txCtx context.Context) error {
		if err := uc.imageValidator.ValidateTransfer(txCtx, cmd); err != nil {
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

func (uc *ImageUseCase) TransferMany(ctx context.Context, cmd command.TransferManyCommand) error {

	ch := make(chan error)
	for _, id := range cmd.GetIDs() {
		go func(id string) {
			ch <- uc.Transfer(ctx, command.TransferCommand{
				ID:         id,
				NewParent:  cmd.GetNewParent(),
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

// Helper functions

func (uc *ImageUseCase) generatePresignedURLS(ctx context.Context, cmd command.UploadImageCommand, imageID string) ([]port.PresignedURLPayload, error) {
	var presignedURLs []port.PresignedURLPayload

	if cmd.Contents == nil {
		return nil, errors.NewInternalError("no contents provided", nil)
	}

	for _, partialContent := range cmd.Contents {

		contentTypeStr := partialContent.ContentType

		currentContentType, _ := vobj.NewContentTypeFromString(contentTypeStr)

		contentID := uuid.New().String()
		content := &model.Content{
			Entity: vobj.Entity{
				ID:         contentID,
				EntityType: vobj.EntityTypeContent,
				Parent: vobj.ParentRef{
					ID:   imageID,
					Type: vobj.ParentTypeImage,
				},
				CreatorID: cmd.CreatorID,
				Name:      partialContent.Name,
			},
			Provider:      uc.originStorage.Provider(),
			ContentType:   currentContentType,
			Size:          partialContent.Size,
			Path:          fmt.Sprintf("%s/%s", contentID, partialContent.Name),
			UploadPending: true,
		}

		if currentContentType.IsOriginImage() {
			presignedURLPayload, err := uc.originStorage.GenerateSignedURL(ctx, port.MethodPut, *content, time.Duration(1*time.Hour))
			if err != nil {
				return nil, errors.NewInternalError("failed to generate presigned url", err)
			}
			presignedURLs = append(presignedURLs, *presignedURLPayload)
		} else {
			presignedURLPayload, err := uc.processedStorage.GenerateSignedURL(ctx, port.MethodPut, *content, time.Duration(1*time.Hour))
			if err != nil {
				return nil, errors.NewInternalError("failed to generate presigned url", err)
			}
			presignedURLs = append(presignedURLs, *presignedURLPayload)
		}

	}

	return presignedURLs, nil

}
