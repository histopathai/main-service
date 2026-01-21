package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/histopathai/main-service/internal/application/command"
	"github.com/histopathai/main-service/internal/application/usecase"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	portevent "github.com/histopathai/main-service/internal/port/event"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageService struct {
	*Service[*model.Image]
	usecase          *usecase.ImageUseCase
	originStorage    port.Storage
	processedStorage port.Storage
	eventPublisher   portevent.EventPublisher
}

func NewImageService(
	repo port.Repository[*model.Image],
	uowFactory port.UnitOfWorkFactory,
	originStorage port.Storage,
	processedStorage port.Storage,
	eventPublisher portevent.EventPublisher,
) *ImageService {
	return &ImageService{
		Service: &Service[*model.Image]{
			repo:       repo,
			uowFactory: uowFactory,
		},
		usecase:          usecase.NewImageUseCase(repo, uowFactory),
		originStorage:    originStorage,
		processedStorage: processedStorage,
		eventPublisher:   eventPublisher,
	}
}

func (s *ImageService) UploadPresignedURL(ctx context.Context, cmd any) (*port.StoragePayload, error) {
	// Type assertion
	uploadCmd, ok := cmd.(command.CreateImageCommand)

	if !ok {
		return nil, errors.NewInternalError("invalid command type for getting presigned URL", nil)
	}

	entity, err := uploadCmd.ToEntity()
	if err != nil {
		return nil, err
	}

	// Generate unique ID for the image
	entity.ID = uuid.New().String()

	entity.OriginContent.Provider = s.originStorage.Provider()
	entity.OriginContent.Path = fmt.Sprintf("%s-%s", entity.ID, entity.Name)

	entity.Processing = &vobj.ProcessingInfo{
		Status:          vobj.StatusPending,
		Version:         vobj.ProcessingV2,
		RetryCount:      0,
		LastProcessedAt: time.Now(),
	}

	createdEntity, err := s.usecase.Create(ctx, entity)
	if err != nil {
		return nil, err
	}

	// Metadata  is being set for UploadedEvent
	metadata := map[string]string{
		"origin-provider": createdEntity.OriginContent.Provider.String(),
		"image-id":        createdEntity.ID,
		"origin-path":     createdEntity.OriginContent.Path,
		"size":            fmt.Sprintf("%d", createdEntity.OriginContent.Size),
		"content-type":    createdEntity.OriginContent.ContentType.String(),
	}

	// Generate presigned URL for uploading the origin content
	payload, err := s.originStorage.GenerateSignedURL(
		ctx,
		createdEntity.OriginContent.Path,
		port.MethodPut,
		createdEntity.OriginContent.ContentType.String(),
		metadata,
		time.Hour,
	)

	if err != nil {
		return nil, err
	}

	return payload, nil
}

// Update handles image updates
func (s *ImageService) Update(ctx context.Context, imageID string, cmd any) error {
	updateCmd, ok := cmd.(command.UpdateImageCommand)
	if !ok {
		return errors.NewInternalError("invalid command type for updating image", nil)
	}

	updates := updateCmd.GetUpdates()
	if updates == nil || len(updates) == 0 {
		return errors.NewValidationError("no updates provided", nil)
	}

	return s.usecase.Update(ctx, imageID, updates)
}

// RetryProcessing retries failed image processing
func (s *ImageService) RetryProcessing(ctx context.Context, imageID string, maxRetries int) error {
	return s.usecase.RetryProcessing(ctx, imageID, maxRetries)
}
