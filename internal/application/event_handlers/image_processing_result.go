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

type ImageProcessingResult struct {
	events.BaseEvent
	ImageID       string  `json:"image-id"`
	Success       bool    `json:"success"`
	ProcessedPath *string `json:"processed-path,omitempty"`
	Width         *int    `json:"width,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Size          *int64  `json:"size,omitempty"`
	FailureReason *string `json:"failure-reason,omitempty"`
	Retryable     *bool   `json:"retryable,omitempty"`
}

// ImageProcessingResultHandler handles both success and failure processing results
type ImageProcessingResultHandler struct {
	*BaseEventHandler
	imageRepo      port.ImageRepository
	imagePublisher port.ImageEventPublisher
	logger         *slog.Logger
}

func NewImageProcessingResultHandler(
	imageRepo port.ImageRepository,
	imagePublisher port.ImageEventPublisher,
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
		imageRepo:      imageRepo,
		imagePublisher: imagePublisher,
		logger:         logger,
	}
}

func (h *ImageProcessingResultHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageProcessingResultHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event - can be either success or failure
	var resultEvent ImageProcessingResult

	if err := h.DeserializeEvent(data, &resultEvent); err != nil {
		return err
	}

	// Handle based on success flag
	if resultEvent.Success {
		return h.handleSuccess(ctx, &resultEvent)
	}
	return h.handleFailure(ctx, &resultEvent)
}

func (h *ImageProcessingResultHandler) handleSuccess(ctx context.Context, result *ImageProcessingResult) error {
	h.logger.Info("Processing successful image processing result",
		slog.String("image_id", result.ImageID),
		slog.String("processed_path", *result.ProcessedPath),
		slog.Int("width", *result.Width),
		slog.Int("height", *result.Height),
		slog.Int64("size", *result.Size))

	// Validate required fields for success
	if result.ProcessedPath == nil || result.Width == nil || result.Height == nil || result.Size == nil {
		return NewNonRetryableError(
			fmt.Errorf("missing required fields for successful processing"),
			events.CategoryValidation,
			events.SeverityHigh,
		).WithContext("image_id", result.ImageID)
	}

	// Update image in database
	now := time.Now()
	updates := map[string]interface{}{
		"Status":          model.StatusProcessed,
		"ProcessedPath":   *result.ProcessedPath,
		"Width":           *result.Width,
		"Height":          *result.Height,
		"Size":            *result.Size,
		"LastProcessedAt": &now,
		"UpdatedAt":       now,
	}

	if err := h.imageRepo.Update(ctx, result.ImageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image after successful processing: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", result.ImageID).
			WithContext("processed_path", *result.ProcessedPath)
	}

	h.logger.Info("Successfully updated image status to PROCESSED",
		slog.String("image_id", result.ImageID))

	return nil
}

func (h *ImageProcessingResultHandler) handleFailure(ctx context.Context, result *ImageProcessingResult) error {
	failureReason := "Unknown error"
	if result.FailureReason != nil {
		failureReason = *result.FailureReason
	}

	retryable := true
	if result.Retryable != nil {
		retryable = *result.Retryable
	}

	h.logger.Error("Image processing failed",
		slog.String("image_id", result.ImageID),
		slog.String("failure_reason", failureReason),
		slog.Bool("retryable", retryable))

	// Read current image state
	image, err := h.imageRepo.Read(ctx, result.ImageID)
	if err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to read image: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", result.ImageID)
	}

	// Check if retryable and under max retry limit
	maxRetries := 3
	if retryable && image.RetryCount < maxRetries {
		h.logger.Info("Scheduling retry for failed image processing",
			slog.String("image_id", result.ImageID),
			slog.Int("retry_count", image.RetryCount+1),
			slog.Int("max_retries", maxRetries))

		// Mark for retry
		image.MarkForRetry()

		now := time.Now()
		updates := map[string]interface{}{
			"Status":          image.Status,
			"RetryCount":      image.RetryCount,
			"FailureReason":   failureReason,
			"LastProcessedAt": &now,
			"UpdatedAt":       now,
		}

		if err := h.imageRepo.Update(ctx, result.ImageID, updates); err != nil {
			return NewRetryableError(
				fmt.Errorf("failed to update image for retry: %w", err),
				events.CategoryDatabase,
				events.SeverityHigh,
			).WithContext("image_id", result.ImageID)
		}

		// Re-publish processing request
		retryEvent := events.NewImageProcessingRequestedEvent(
			image.ID,
			image.OriginPath,
		)

		if err := h.imagePublisher.PublishImageProcessingRequested(ctx, &retryEvent); err != nil {
			return NewRetryableError(
				fmt.Errorf("failed to publish retry event: %w", err),
				events.CategoryProcessing,
				events.SeverityHigh,
			).WithContext("image_id", result.ImageID)
		}

		h.logger.Info("Successfully scheduled retry",
			slog.String("image_id", result.ImageID))

		return nil
	}

	// Non-retryable or max retries exceeded - mark as permanently failed
	h.logger.Error("Image processing permanently failed",
		slog.String("image_id", result.ImageID),
		slog.Int("retry_count", image.RetryCount),
		slog.Bool("retryable", retryable))

	now := time.Now()
	updates := map[string]interface{}{
		"Status":          model.StatusFailed,
		"FailureReason":   failureReason,
		"RetryCount":      image.RetryCount + 1,
		"LastProcessedAt": &now,
		"UpdatedAt":       now,
	}

	if err := h.imageRepo.Update(ctx, result.ImageID, updates); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to update image permanent failure: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", result.ImageID)
	}

	return nil
}
