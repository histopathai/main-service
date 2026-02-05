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

type ImageProcessHandler struct {
	subscriber portevent.EventSubscriber
	worker     port.ImageProcessingWorker
	imageRepo  port.ImageRepository
	logger     *slog.Logger
}

func NewImageProcessHandler(
	subscriber portevent.EventSubscriber,
	worker port.ImageProcessingWorker,
	imageRepo port.ImageRepository,
	logger *slog.Logger,
) *ImageProcessHandler {
	return &ImageProcessHandler{
		subscriber: subscriber,
		worker:     worker,
		imageRepo:  imageRepo,
		logger:     logger,
	}
}

func (h *ImageProcessHandler) Start(ctx context.Context) error {
	h.logger.Info("ImageProcessHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *ImageProcessHandler) Stop() error {
	h.logger.Info("ImageProcessHandler stopping...")
	return h.subscriber.Stop()
}

func (h *ImageProcessHandler) Handle(ctx context.Context, event domainevent.Event) error {
	h.logger.Info("ImageProcessHandler: received event", "event", event)
	processEvent, ok := event.(*domainevent.ImageProcessReqEvent)
	if !ok {
		h.logger.Warn("ImageProcessHandler: received unsupported event type")
		return nil
	}

	// 1. Read Image Entity to check ActiveEventID
	imageEntity, err := h.imageRepo.Read(ctx, processEvent.Content.Parent.ID)
	if err != nil {
		return err
	}
	if imageEntity == nil {
		h.logger.Error("ImageProcessHandler: image entity not found", "image_id", processEvent.Content.Parent.ID)
		return nil // Non-retryable if entity missing
	}

	// 2. Check for Stale/Duplicate Event
	// If the active event ID in DB does not match this event's ID, it means another request (newer or older winning race) took precedence.
	if imageEntity.Processing != nil && imageEntity.Processing.ActiveEventID != "" {
		if imageEntity.Processing.ActiveEventID != processEvent.EventID {
			h.logger.Warn("ImageProcessHandler: skipping stale/duplicate event",
				"current_active_id", imageEntity.Processing.ActiveEventID,
				"event_id", processEvent.EventID)
			return nil
		}
	}

	updates := map[string]any{
		fields.ImageProcessingStatus.DomainName():  vobj.StatusProcessing,
		fields.ImageProcessingVersion.DomainName(): processEvent.ProcessingVersion,
	}

	// We technically don't need to update again if NewFileHandler did, but we might want to update timestamps or verify connection.
	// Actually, let's keep the update to ensure "Processing" state and timestamp refreshed if we want.
	// But minimal is fine.

	err = h.imageRepo.Update(ctx, processEvent.Content.Parent.ID, updates)
	if err != nil {
		return err
	}

	err = h.worker.ProcessImage(ctx, processEvent.Content, processEvent.ProcessingVersion)
	if err != nil {
		return err
	}

	return nil
}
