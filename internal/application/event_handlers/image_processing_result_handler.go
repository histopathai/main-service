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

// ImageProcessingResultHandler handles processing completion events
type ImageProcessingResultHandler struct {
	*BaseEventHandler
	imageRepo port.ImageRepository
	logger    *slog.Logger
}

func NewImageProcessingResultHandler(
	imageRepo port.ImageRepository,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *ImageProcessingResultHandler {
	return &ImageProcessingResultHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageRepo: imageRepo,
		logger:    logger,
	}
}

func (h *ImageProcessingResultHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageProcessingResultHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.ImageProcessingCompletedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing image processing completed event",
		slog.String("image_id", event.ImageID),
		slog.String("processed_path", event.ProcessedPath),
		slog.Int("width", event.Width),
		slog.Int("height", event.Height),
		slog.Int64("size", event.Size))

	// Update image in database
	now := time.Now()
	updates := map[string]interface{}{
		"Status":          model.StatusProcessed,
		"ProcessedPath":   event.ProcessedPath,
		"Width":           event.Width,
		"Height":          event.Height,
		"Size":            event.Size,
		"LastProcessedAt": &now,
		"UpdatedAt":       now,
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image after successful processing: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", event.ImageID).
			WithContext("processed_path", event.ProcessedPath)
	}

	h.logger.Info("Successfully updated image status to PROCESSED",
		slog.String("image_id", event.ImageID))

	return nil
}

type ImageProcessingFailureHandler struct {
	*BaseEventHandler
	imageRepo port.ImageRepository
	publisher port.ImageEventPublisher
	logger    *slog.Logger
}

// NewImageProcessingFailureHandler creates a new handler
func NewImageProcessingFailureHandler(
	imageRepo port.ImageRepository,
	publisher port.ImageEventPublisher,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *ImageProcessingFailureHandler {
	return &ImageProcessingFailureHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageRepo: imageRepo,
		publisher: publisher,
		logger:    logger,
	}
}

// Handle processes image processing failed events
func (h *ImageProcessingFailureHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageProcessingFailureHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.ImageProcessingFailedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Error("Image processing failed",
		slog.String("image_id", event.ImageID),
		slog.String("failure_reason", event.FailureReason),
		slog.Bool("retryable", event.Retryable))

	// Read current image state
	image, err := h.imageRepo.Read(ctx, event.ImageID)
	if err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to read image: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", event.ImageID)
	}

	// Check if retryable and under max retry limit
	maxRetries := 3 // Configure as needed
	if event.Retryable && image.RetryCount < maxRetries {
		h.logger.Info("Scheduling retry for failed image processing",
			slog.String("image_id", event.ImageID),
			slog.Int("retry_count", image.RetryCount+1),
			slog.Int("max_retries", maxRetries))

		// Mark for retry
		image.MarkForRetry()

		now := time.Now()
		updates := map[string]interface{}{
			"Status":          image.Status,
			"RetryCount":      image.RetryCount,
			"FailureReason":   event.FailureReason,
			"LastProcessedAt": &now,
			"UpdatedAt":       now,
		}

		if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
			return NewRetryableError(
				fmt.Errorf("failed to update image for retry: %w", err),
				events.CategoryDatabase,
				events.SeverityHigh,
			).WithContext("image_id", event.ImageID)
		}

		// Re-publish processing request
		retryEvent := events.NewImageProcessingRequestedEvent(
			image.ID,
			image.OriginPath,
		)

		if err := h.publisher.PublishImageProcessingRequested(ctx, &retryEvent); err != nil {
			return NewRetryableError(
				fmt.Errorf("failed to publish retry event: %w", err),
				events.CategoryProcessing,
				events.SeverityHigh,
			).WithContext("image_id", event.ImageID)
		}

		h.logger.Info("Successfully scheduled retry",
			slog.String("image_id", event.ImageID))

		return nil
	}

	// Non-retryable or max retries exceeded - mark as permanently failed
	h.logger.Error("Image processing permanently failed",
		slog.String("image_id", event.ImageID),
		slog.Int("retry_count", image.RetryCount),
		slog.Bool("retryable", event.Retryable))

	now := time.Now()
	updates := map[string]interface{}{
		"Status":          model.StatusFailed,
		"FailureReason":   event.FailureReason,
		"RetryCount":      image.RetryCount + 1,
		"LastProcessedAt": &now,
		"UpdatedAt":       now,
	}

	if err := h.imageRepo.Update(ctx, event.ImageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image permanent failure: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", event.ImageID)
	}

	return nil
}
