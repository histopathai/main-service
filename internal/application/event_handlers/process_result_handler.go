package eventhandlers

import (
	"context"
	"log/slog"
	"time"

	"github.com/histopathai/main-service-refactor/internal/domain/events"
	"github.com/histopathai/main-service-refactor/internal/domain/model"
	"github.com/histopathai/main-service-refactor/internal/domain/repository"
	"github.com/histopathai/main-service-refactor/internal/service"
	"github.com/histopathai/main-service-refactor/internal/shared/constants"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
)

const (
	MaxRetries = 3
)

type ProcessResultHandler struct {
	imageRepo  repository.ImageRepository
	serializer events.EventSerializer
	publisher  service.EventPublisher
	logger     *slog.Logger
}

func NewProcessResultHandler(
	imageRepo repository.ImageRepository,
	serializer events.EventSerializer,
	publisher service.EventPublisher,
	logger *slog.Logger,
) *ProcessResultHandler {
	return &ProcessResultHandler{
		imageRepo:  imageRepo,
		serializer: serializer,
		publisher:  publisher,
		logger:     logger,
	}
}

func (h *ProcessResultHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	eventType := events.EventType(attributes["event_type"])

	switch eventType {
	case events.EventTypeImageProcessingCompleted:
		return h.handleCompleted(ctx, data, attributes)
	case events.EventTypeImageProcessingFailed:
		return h.handleFailed(ctx, data, attributes)
	default:
		h.logger.Warn("Unknown event type received", "eventType", eventType)
		return errors.NewValidationError("unknown event type", nil)
	}
}

type ImageProcessingCompletedEvent struct {
	events.BaseEvent
	ImageID       string `json:"image-id"`
	ProcessedPath string `json:"processed-path"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Size          int64  `json:"size"`
}

func (h *ProcessResultHandler) handleCompleted(ctx context.Context, data []byte, attributes map[string]string) error {
	var event ImageProcessingCompletedEvent
	if err := h.serializer.Deserialize(data, &event); err != nil {
		h.logger.Error("Failed to deserialize ImageProcessingCompletedEvent", slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to deserialize event: %v", err)
	}

	h.logger.Info("Processing ImageProcessingCompletedEvent", "ImageID", event.ImageID)

	updates := map[string]interface{}{
		constants.ImageStatusField:        model.StatusProcessed,
		constants.ImageProcessedPathField: event.ProcessedPath,
		constants.ImageWidthField:         event.Width,
		constants.ImageHeightField:        event.Height,
		constants.ImageSizeField:          event.Size,
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		h.logger.Error("Failed to update image status to PROCESSED", slog.String("ImageID", event.ImageID), slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to update image: %v", err)
	}

	h.logger.Info("Successfully processed ImageProcessingCompletedEvent", "ImageID", event.ImageID)
	return nil
}

func (h *ProcessResultHandler) handleFailed(ctx context.Context, data []byte, attributes map[string]string) error {
	var event events.ImageProcessingFailedEvent
	if err := h.serializer.Deserialize(data, &event); err != nil {
		h.logger.Error("Failed to deserialize ImageProcessingFailedEvent", slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to deserialize event: %v", err)
	}

	h.logger.Warn("Processing failed event",
		"imageID", event.ImageID,
		"reason", event.FailureReason)

	image, err := h.imageRepo.GetByID(ctx, event.ImageID)
	if err != nil {
		h.logger.Error("Failed to fetch image for failed processing", slog.String("ImageID", event.ImageID), slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to fetch image: %v", err)
	}
	var status model.ImageStatus

	if image.RetryCount >= MaxRetries {
		status = model.StatusFailed
		h.logger.Warn("Max retries reached for image processing", slog.String("ImageID", event.ImageID))
	} else {
		status = model.StatusProcessing
		h.logger.Info("Retrying image processing", slog.String("ImageID", event.ImageID), slog.Int("CurrentRetryCount", image.RetryCount))
	}

	updates := map[string]interface{}{
		constants.ImageStatusField:          status,
		constants.ImageFailureReasonField:   event.FailureReason,
		constants.ImageRetryCountField:      image.RetryCount + 1,
		constants.ImageLastProcessedAtField: time.Now(),
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		h.logger.Error("Failed to update image status to FAILED/RETRYING", slog.String("ImageID", event.ImageID), slog.String("error", err.Error()))
		return errors.NewInternalError("Failed to update image: %v", err)
	}

	if status == model.StatusProcessing {
		republishEvent := events.NewImageProcessingRequestedEvent(
			event.ImageID,
			image.OriginPath,
		)
		h.publisher.PublishImageProcessingRequested(ctx, &republishEvent)
		h.logger.Info("Republished image processing request", slog.String("ImageID", event.ImageID))
	}

	h.logger.Info("Successfully processed ImageProcessingFailedEvent", "ImageID", event.ImageID)
	return nil
}
