package eventhandlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/infrastructure/storage/firestore"
)

// ImageProcessingRequestHandler handles image processing requests
type ImageProcessingRequestHandler struct {
	*BaseEventHandler
	imageRepo  port.ImageRepository
	worker     port.ImageProcessingWorker
	storage    port.ObjectStorage
	bucketName string
	logger     *slog.Logger
}

// NewImageProcessingRequestHandler creates a new handler
func NewImageProcessingRequestHandler(
	imageRepo port.ImageRepository,
	worker port.ImageProcessingWorker,
	storage port.ObjectStorage,
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
		storage:    storage,
		bucketName: bucketName,
		logger:     logger,
	}
}

// Handle processes image processing request events
func (h *ImageProcessingRequestHandler) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	return h.HandleWithRetry(ctx, data, attributes, h.processEvent)
}

func (h *ImageProcessingRequestHandler) processEvent(ctx context.Context, data []byte, attributes map[string]string) error {
	var event events.ImageProcessingRequestedEvent
	if err := h.DeserializeEvent(data, &event); err != nil {
		return err
	}

	existingImage, err := h.imageRepo.Read(ctx, event.ImageID)
	if err != nil {
		if errors.Is(err, firestore.ErrNotFound) || strings.Contains(err.Error(), "NotFound") {
			h.logger.Warn("Image record not found in database, acknowledging message to stop retry loop",
				slog.String("image_id", event.ImageID),
				slog.String("error", err.Error()))
			return nil
		}
		return NewRetryableError(fmt.Errorf("failed to read image state: %w", err), events.CategoryDatabase, events.SeverityHigh)
	}

	if existingImage.Status == model.StatusDeleting {
		h.logger.Info("Skipping processing for image marked as DELETING",
			slog.String("image_id", event.ImageID))
		return nil
	}

	h.logger.Info("Processing image processing request",
		slog.String("image_id", event.ImageID),
		slog.String("origin_path", event.OriginPath))

	exists, err := h.storage.ObjectExists(ctx, h.bucketName, event.OriginPath)
	if err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to check object existence: %w", err),
			events.CategoryStorage,
			events.SeverityHigh,
		).WithContext("origin_path", event.OriginPath)
	}

	if !exists {
		h.logger.Error("File does not exist in storage, aborting processing",
			slog.String("image_id", event.ImageID),
			slog.String("origin_path", event.OriginPath))

		return NewNonRetryableError(
			fmt.Errorf("file not found in bucket %s", h.bucketName),
			events.CategoryStorage,
			events.SeverityHigh,
		).WithContext("origin_path", event.OriginPath)
	}

	objMetadata, err := h.storage.GetObjectMetadata(ctx, h.bucketName, event.OriginPath)
	if err != nil {
		return NewRetryableError(
			fmt.Errorf("failed to get object metadata: %w", err),
			events.CategoryStorage,
			events.SeverityHigh,
		).WithContext("origin_path", event.OriginPath)
	}

	imageSize := objMetadata.Size
	h.logger.Info("Retrieved image size from GCS",
		slog.String("image_id", event.ImageID),
		slog.Int64("size", imageSize))

	if err := h.updateImageStatus(ctx, event.ImageID, model.StatusProcessing); err != nil {
		h.logger.Warn("Could not update image status to PROCESSING",
			slog.String("error", err.Error()))
	}

	workerInput := &port.ProcessingInput{
		ImageID:    event.ImageID,
		OriginPath: event.OriginPath,
		BucketName: h.bucketName,
		Size:       imageSize,
	}

	if err := h.worker.ProcessImage(ctx, workerInput); err != nil {
		h.logger.Error("Worker failed to process image",
			slog.String("image_id", event.ImageID),
			slog.String("error", err.Error()))

		_ = h.updateImageStatusWithFailure(ctx, event.ImageID, err.Error())

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
