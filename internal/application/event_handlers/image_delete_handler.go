package eventhandlers

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/repository"
	"github.com/histopathai/main-service/internal/domain/storage"
	"github.com/histopathai/main-service/pkg/config"
)

type ImageDeleteHandler struct {
	imageRepo     repository.ImageRepository
	serializer    events.EventSerializer
	publisher     events.ImageEventPublisher
	objectStorage storage.ObjectStorage
	logger        *slog.Logger
	cfg           *config.GCPConfig
}

func NewImageDeleteHandler(
	imageRepo repository.ImageRepository,
	serializer events.EventSerializer,
	publisher events.ImageEventPublisher,
	objectStorage storage.ObjectStorage,
	logger *slog.Logger,
	cfg *config.GCPConfig,
) *ImageDeleteHandler {
	return &ImageDeleteHandler{
		imageRepo:     imageRepo,
		serializer:    serializer,
		publisher:     publisher,
		logger:        logger,
		objectStorage: objectStorage,
		cfg:           cfg,
	}
}

func (h *ImageDeleteHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	h.logger.Info("Processing image deletion event",
		"EventID", attributes["event_id"],
	)

	var event events.ImageDeletionRequestedEvent
	if err := h.serializer.Deserialize(data, &event); err != nil {
		h.logger.Error("Failed to deserialize image deletion event, dropping message", slog.String("error", err.Error()), "data", string(data))
		return nil
	}

	image, err := h.imageRepo.Read(ctx, event.ImageID)
	if err != nil {
		h.logger.Error("Failed to fetch image record from database", "ImageID", event.ImageID, "Error", err.Error())

		return err
	}
	if image == nil {
		h.logger.Warn("Image record not found in database, skipping deletion", "ImageID", event.ImageID)
		return nil
	}

	if err := h.objectStorage.DeleteObject(ctx, h.cfg.OriginalBucketName, image.OriginPath); err != nil {
		h.logger.Error("Failed to delete image from storage", "ImageID", event.ImageID, "OriginPath", image.OriginPath, "Error", err.Error())
		return err
	}

	if image.ProcessedPath != nil {
		if err := h.objectStorage.DeleteByPrefix(ctx, h.cfg.ProcessedBucketName, *image.ProcessedPath); err != nil {
			h.logger.Error("Failed to delete processed images from storage", "ImageID", event.ImageID, "ProcessedPath", image.ProcessedPath, "Error", err.Error())
			return err
		}
	}

	if err := h.imageRepo.Delete(ctx, event.ImageID); err != nil {
		h.logger.Error("Failed to delete image record from database", "ImageID", event.ImageID, "Error", err.Error())
		return err
	}

	h.logger.Info("Successfully deleted image and its records", "ImageID", event.ImageID)
	return nil
}
