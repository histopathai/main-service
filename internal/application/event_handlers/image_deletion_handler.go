package eventhandlers

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

// ImageDeletionHandler handles image deletion requests
type ImageDeletionHandler struct {
	*BaseEventHandler
	imageRepo  port.ImageRepository
	storage    port.ObjectStorage
	bucketName string
	logger     *slog.Logger
}

// NewImageDeletionHandler creates a new handler
func NewImageDeletionHandler(
	imageRepo port.ImageRepository,
	storage port.ObjectStorage,
	bucketName string,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *ImageDeletionHandler {
	return &ImageDeletionHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageRepo:  imageRepo,
		storage:    storage,
		bucketName: bucketName,
		logger:     logger,
	}
}

// Handle processes image deletion request events
func (h *ImageDeletionHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageDeletionHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	// Deserialize event
	var event events.ImageDeletionRequestedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	h.logger.Info("Processing image deletion request",
		slog.String("image_id", event.ImageID))

	// Read image from database
	image, err := h.imageRepo.Read(ctx, event.ImageID)
	if err != nil {
		// If image not found, consider it already deleted
		if isNotFoundError(err) {
			h.logger.Warn("Image not found in database, assuming already deleted",
				slog.String("image_id", event.ImageID))
			return nil
		}

		return NewRetryableError(
			fmt.Errorf("failed to read image: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", event.ImageID)
	}

	// Delete files from GCS
	if err := h.deleteImageFiles(ctx, image); err != nil {
		return err
	}

	// Delete image record from database
	if err := h.imageRepo.Delete(ctx, event.ImageID); err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to delete image from database: %w", err),
			events.CategoryDatabase,
			events.SeverityHigh,
		).WithContext("image_id", event.ImageID)
	}

	h.logger.Info("Successfully deleted image",
		slog.String("image_id", event.ImageID))

	return nil
}

func (h *ImageDeletionHandler) deleteImageFiles(ctx context.Context, image *model.Image) error {
	var deletionErrors []error

	// Delete original file
	if image.OriginPath != "" {
		h.logger.Info("Deleting original file",
			slog.String("image_id", image.ID),
			slog.String("path", image.OriginPath))

		exists, err := h.storage.ObjectExists(ctx, h.bucketName, image.OriginPath)
		if err != nil {
			deletionErrors = append(deletionErrors, fmt.Errorf("failed to check origin file existence: %w", err))
		} else if exists {
			if err := h.storage.DeleteObject(ctx, h.bucketName, image.OriginPath); err != nil {
				deletionErrors = append(deletionErrors, fmt.Errorf("failed to delete origin file: %w", err))
			} else {
				h.logger.Info("Deleted original file",
					slog.String("image_id", image.ID),
					slog.String("path", image.OriginPath))
			}
		} else {
			h.logger.Warn("Original file not found in storage",
				slog.String("image_id", image.ID),
				slog.String("path", image.OriginPath))
		}
	}

	// Delete processed file if exists
	if image.ProcessedPath != nil && *image.ProcessedPath != "" {
		h.logger.Info("Deleting processed file",
			slog.String("image_id", image.ID),
			slog.String("path", *image.ProcessedPath))

		exists, err := h.storage.ObjectExists(ctx, h.bucketName, *image.ProcessedPath)
		if err != nil {
			deletionErrors = append(deletionErrors, fmt.Errorf("failed to check processed file existence: %w", err))
		} else if exists {
			if err := h.storage.DeleteObject(ctx, h.bucketName, *image.ProcessedPath); err != nil {
				deletionErrors = append(deletionErrors, fmt.Errorf("failed to delete processed file: %w", err))
			} else {
				h.logger.Info("Deleted processed file",
					slog.String("image_id", image.ID),
					slog.String("path", *image.ProcessedPath))
			}
		} else {
			h.logger.Warn("Processed file not found in storage",
				slog.String("image_id", image.ID),
				slog.String("path", *image.ProcessedPath))
		}
	}

	// If there were any deletion errors, return them
	if len(deletionErrors) > 0 {
		// Log all errors
		for _, err := range deletionErrors {
			h.logger.Error("Storage deletion error",
				slog.String("image_id", image.ID),
				slog.String("error", err.Error()))
		}

		// Return the first error as retryable
		return NewRetryableError(
			deletionErrors[0],
			events.CategoryStorage,
			events.SeverityMedium,
		).WithContext("image_id", image.ID).
			WithContext("total_errors", len(deletionErrors))
	}

	return nil
}

// isNotFoundError checks if error is a not found error
func isNotFoundError(err error) bool {
	// Check for your custom not found error
	// Adjust based on your error handling implementation
	if customErr, ok := err.(*errors.Err); ok {
		return customErr.Type == errors.ErrorTypeNotFound
	}
	return false
}

// BatchImageDeletionHandler handles batch deletion requests
type BatchImageDeletionHandler struct {
	*BaseEventHandler
	imageRepo  port.ImageRepository
	storage    port.ObjectStorage
	bucketName string
	publisher  port.ImageEventPublisher
	logger     *slog.Logger
}

// NewBatchImageDeletionHandler creates a new batch handler
func NewBatchImageDeletionHandler(
	imageRepo port.ImageRepository,
	storage port.ObjectStorage,
	bucketName string,
	publisher port.ImageEventPublisher,
	serializer events.EventSerializer,
	telemetryPublisher port.TelemetryEventPublisher,
	logger *slog.Logger,
) *BatchImageDeletionHandler {
	return &BatchImageDeletionHandler{
		BaseEventHandler: NewBaseEventHandler(
			logger,
			serializer,
			telemetryPublisher,
			DefaultRetryConfig(),
		),
		imageRepo:  imageRepo,
		storage:    storage,
		bucketName: bucketName,
		publisher:  publisher,
		logger:     logger,
	}
}

// Handle processes batch deletion by publishing individual deletion events
func (h *BatchImageDeletionHandler) Handle(ctx context.Context, imageIDs []string) error {
	h.logger.Info("Processing batch image deletion",
		slog.Int("count", len(imageIDs)))

	var errors []error
	for _, imageID := range imageIDs {
		event := events.NewImageDeletionRequestedEvent(imageID)

		if err := h.publisher.PublishImageDeletionRequested(ctx, &event); err != nil {
			h.logger.Error("Failed to publish deletion event",
				slog.String("image_id", imageID),
				slog.String("error", err.Error()))
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("failed to publish %d deletion events", len(errors))
	}

	h.logger.Info("Successfully published all deletion events",
		slog.Int("count", len(imageIDs)))

	return nil
}
