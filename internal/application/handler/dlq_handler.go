package handler

import (
	"context"
	"log/slog"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

// ImageProcessDlqHandler handles events that failed max retries
type ImageProcessDlqHandler struct {
	subscriber portevent.EventSubscriber
	imageRepo  port.ImageRepository
	logger     *slog.Logger
}

func NewImageProcessDlqHandler(
	subscriber portevent.EventSubscriber,
	imageRepo port.ImageRepository,
	logger *slog.Logger,
) *ImageProcessDlqHandler {
	return &ImageProcessDlqHandler{
		subscriber: subscriber,
		imageRepo:  imageRepo,
		logger:     logger,
	}
}

func (h *ImageProcessDlqHandler) Start(ctx context.Context) error {
	h.logger.Info("ImageProcessDlqHandler started, listening for DLQ events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *ImageProcessDlqHandler) Stop() error {
	h.logger.Info("ImageProcessDlqHandler stopping...")
	return h.subscriber.Stop()
}

func (h *ImageProcessDlqHandler) Handle(ctx context.Context, event domainevent.Event) error {
	dlqEvent, ok := event.(domainevent.ImageProcessDlqEvent)
	if !ok {
		h.logger.Warn("ImageProcessDlqHandler: received unsupported event type")
		return nil
	}

	h.logger.Error("Image processing permanently failed - sent to DLQ",
		slog.String("image_id", dlqEvent.ImageID),
		slog.String("original_event_id", dlqEvent.OriginalEventID),
		slog.String("failure_reason", dlqEvent.FailureReason),
		slog.Int("retry_attempts", dlqEvent.RetryMetadata.AttemptCount),
		slog.String("processing_version", string(dlqEvent.ProcessingVersion)),
	)

	// Update image status to permanently failed
	updates := map[string]any{
		fields.ImageProcessingStatus.DomainName(): vobj.StatusFailedPermanent,
	}

	if err := h.imageRepo.Update(ctx, dlqEvent.ImageID, updates); err != nil {
		h.logger.Error("Failed to update image status in DLQ handler",
			slog.String("image_id", dlqEvent.ImageID),
			slog.String("error", err.Error()))
		return err
	}

	// TODO: Send alert/notification for manual intervention
	// TODO: Store in DLQ table for analysis and potential replay

	return nil
}
