package eventhandlers

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
)

// ImageProcessingRequestHandler handles image processing requests
type ImageProcessingRequestHandler struct {
	*BaseEventHandler
	imageRepo  port.ImageRepository
	worker     port.ImageProcessingWorker
	bucketName string
	logger     *slog.Logger
}

// NewImageProcessingRequestHandler creates a new handler
func NewImageProcessingRequestHandler(
	imageRepo port.ImageRepository,
	worker port.ImageProcessingWorker,
	bucketName string,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *ImageProcessingRequestHandler {
	return &ImageProcessingRequestHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageRepo:  imageRepo,
		worker:     worker,
		bucketName: bucketName,
		logger:     logger,
	}
}

// Handle processes image processing request events
func (h *ImageProcessingRequestHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageProcessingRequestHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.ImageProcessingRequestedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing image processing request",
		slog.String("image_id", event.ImageID),
		slog.String("origin_path", event.OriginPath))

	// Update image status to PROCESSING
	if err := h.updateImageStatus(ctx, event.ImageID, model.StatusProcessing); err != nil {
		return err
	}

	// Trigger worker (Cloud Run Job)
	workerInput := &port.ProcessingInput{
		ImageID:    event.ImageID,
		OriginPath: event.OriginPath,
		BucketName: h.bucketName,
	}

	if err := h.worker.ProcessImage(ctx, workerInput); err != nil {
		h.logger.Error("Worker failed to process image",
			slog.String("image_id", event.ImageID),
			slog.String("error", err.Error()))

		// Update image status to FAILED
		if updateErr := h.updateImageStatusWithFailure(ctx, event.ImageID, err.Error()); updateErr != nil {
			h.logger.Error("Failed to update image status after worker failure",
				slog.String("error", updateErr.Error()))
		}

		return err
	}

	h.logger.Info("Successfully triggered worker for image processing",
		slog.String("image_id", event.ImageID))

	return nil
}

func (h *ImageProcessingRequestHandler) updateImageStatus(ctx context.Context, imageID string, status model.ImageStatus) error {
	updates := map[string]interface{}{
		"Status":    status,
		"UpdatedAt": time.Now(),
	}

	if err := h.imageRepo.Update(ctx, imageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image status: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", imageID)
	}

	return nil
}

func (h *ImageProcessingRequestHandler) updateImageStatusWithFailure(ctx context.Context, imageID string, failureReason string) error {
	// Read current image to check retry count
	image, err := h.imageRepo.Read(ctx, imageID)
	if err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to read image: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		)
	}

	now := time.Now()
	updates := map[string]interface{}{
		"Status":          model.StatusFailed,
		"FailureReason":   failureReason,
		"RetryCount":      image.RetryCount + 1,
		"LastProcessedAt": &now,
		"UpdatedAt":       now,
	}

	if err := h.imageRepo.Update(ctx, imageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image failure status: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", imageID)
	}

	return nil
}
