package eventhandlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
)

// UploadStatusHandler handles image upload status events.
type UploadStatusHandler struct {
	*BaseEventHandler
	imageService port.IImageService
	logger       *slog.Logger
}

func NewUploadStatusHandler(
	imageService port.IImageService,
	logger *slog.Logger,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
) *UploadStatusHandler {
	return &UploadStatusHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageService: imageService,
		logger:       logger,
	}
}

func (h *UploadStatusHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *UploadStatusHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.ImageUploadedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing upload status notification",
		slog.String("file_name", event.Name),
		slog.String("bucket", event.Bucket))

	width := intPointerToInt(event.Metadata.Width)
	height := intPointerToInt(event.Metadata.Height)
	size := int64PointerToInt64(event.Metadata.Size)

	status := model.ImageStatus(event.Metadata.Status)

	confirm := port.ConfirmUploadInput{
		ImageID:   event.Metadata.ImageID,
		PatientID: event.Metadata.PatientID,
		CreatorID: event.Metadata.CreatorID,
		Name:      event.Metadata.Name,
		Format:    event.Metadata.Format,
		Width:     width,
		Height:    height,
		Size:      size,
		Status:    status,
	}

	if err := h.imageService.ConfirmUpload(ctx, &confirm); err != nil {
		h.logger.Error("Failed to confirm upload status",
			slog.String("image_id", event.Metadata.ImageID),
			slog.Any("error", err))
		return fmt.Errorf("failed to confirm upload status: %w", err)
	}

	h.logger.Info("Successfully processed upload status notification",
		slog.String("image_id", event.Metadata.ImageID),
		slog.String("status", event.Metadata.Status))

	return nil
}

func intPointerToInt(ptr *int) *int {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}

func int64PointerToInt64(ptr *int64) *int64 {
	if ptr == nil {
		return nil
	}
	val := *ptr
	return &val
}
