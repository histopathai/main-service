package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageService struct {
	*Service[*model.Image]
	imageUseCase  *usecase.ImageUseCase
	objectStorage port.Storage
	logger        slog.Logger
}

func NewImageService(
	repo port.Repository[*model.Image],
	uowFactory port.UnitOfWorkFactory,
	originStorage port.Storage,
	logger slog.Logger,
) *ImageService {
	return &ImageService{
		Service: &Service[*model.Image]{
			repo:       repo,
			uowFactory: uowFactory,
		},
		imageUseCase:  usecase.NewImageUseCase(repo, uowFactory),
		objectStorage: originStorage,
		logger:        logger,
	}
}

func (s *ImageService) Upload(ctx context.Context, cmd any) (*port.PresignedURLPayload, error) {

	// Type assertion
	uploadCmd, ok := cmd.(command.CreateImageCommand)
	if !ok {
		return nil, errors.NewInternalError("invalid command type for uploading image", nil)
	}

	entity, err := uploadCmd.ToEntity()
	if err != nil {
		return nil, err
	}

	contentEntity := uploadCmd.GetContent()

	if contentEntity == nil {
		return nil, errors.NewValidationError("content is required for image upload", nil)
	}

	if contentEntity.ContentType.IsThumbnail() {
		return nil, errors.NewValidationError("thumbnail content is not allowed for original image upload", nil)
	}

	newImageIdStr := uuid.New().String()
	newContentIdStr := uuid.New().String()

	entity.SetID(newImageIdStr)
	entity.OriginContentID = &newContentIdStr

	createdImageEntity, err := s.imageUseCase.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	contentEntity.Parent.ID = createdImageEntity.ID
	contentEntity.Path = fmt.Sprintf("%s/%s", createdImageEntity.ID, contentEntity.Name)
	contentEntity.Provider = s.objectStorage.Provider()
	contentEntity.SetID(newContentIdStr)

	presignedURLPayload, err := s.objectStorage.GenerateSignedURL(ctx, port.MethodPut, *contentEntity, UPLOADEXPIRY_DURATION)
	if err != nil {

		// Rollback image creation
		//TODO: Will need  Hard delete after
		err = s.repo.SoftDelete(ctx, createdImageEntity.ID)
		if err != nil {
			s.logger.Error("failed to rollback image creation after signed URL generation failure", "image_id", createdImageEntity.ID, "error", err)
		}

		return nil, errors.NewInternalError("failed to generate signed URL for image upload", err)
	}

	return presignedURLPayload, nil

}

func (s *ImageService) Update(ctx context.Context, cmd any) error {

	// Type assertion
	updateCmd, ok := cmd.(command.UpdateImageCommand)
	if !ok {
		return errors.NewInternalError("invalid command type for updating image", nil)
	}

	update := updateCmd.GetUpdates()

	if update == nil {
		return errors.NewValidationError("no updates provided for image update", nil)
	}

	err := s.repo.Update(ctx, updateCmd.GetID(), update)
	if err != nil {
		return errors.NewInternalError("failed to update image", err)
	}
	return nil
}
