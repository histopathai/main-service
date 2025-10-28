package eventhandlers

import (
	"context"
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
}

func NewProcessResultHandler(
	imageRepo repository.ImageRepository,
	serializer events.EventSerializer,
	publisher service.EventPublisher,
) *ProcessResultHandler {
	return &ProcessResultHandler{
		imageRepo:  imageRepo,
		serializer: serializer,
		publisher:  publisher,
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
		return errors.NewInternalError("Failed to deserialize event: %v", err)
	}

	updates := map[string]interface{}{
		constants.ImageStatusField:        model.StatusProcessed,
		constants.ImageProcessedPathField: event.ProcessedPath,
		constants.ImageWidthField:         event.Width,
		constants.ImageHeightField:        event.Height,
		constants.ImageSizeField:          event.Size,
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		return errors.NewInternalError("Failed to update image: %v", err)
	}

	return nil
}

func (h *ProcessResultHandler) handleFailed(ctx context.Context, data []byte, attributes map[string]string) error {
	var event events.ImageProcessingFailedEvent
	if err := h.serializer.Deserialize(data, &event); err != nil {
		return errors.NewInternalError("Failed to deserialize event: %v", err)
	}

	image, err := h.imageRepo.Read(ctx, event.ImageID)
	if err != nil {
		return errors.NewInternalError("Failed to fetch image: %v", err)
	}
	var status model.ImageStatus

	if image.RetryCount >= MaxRetries {
		status = model.StatusFailed
	} else {
		status = model.StatusProcessing
	}

	updates := map[string]interface{}{
		constants.ImageStatusField:          status,
		constants.ImageFailureReasonField:   event.FailureReason,
		constants.ImageRetryCountField:      image.RetryCount + 1,
		constants.ImageLastProcessedAtField: time.Now(),
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		return errors.NewInternalError("Failed to update image: %v", err)
	}

	if status == model.StatusProcessing {
		republishEvent := events.NewImageProcessingRequestedEvent(
			event.ImageID,
			image.OriginPath,
		)
		h.publisher.PublishImageProcessingRequested(ctx, &republishEvent)
	}

	return nil
}
