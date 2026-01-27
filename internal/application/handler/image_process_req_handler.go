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
	processEvent, ok := event.(*domainevent.ImageProcessReqEvent)
	if !ok {
		h.logger.Warn("ImageProcessHandler: received unsupported event type")
		return nil
	}

	updates := map[string]any{
		fields.ImageProcessingStatus.DomainName():  vobj.StatusProcessing,
		fields.ImageProcessingVersion.DomainName(): processEvent.ProcessingVersion,
	}

	err := h.imageRepo.Update(ctx, processEvent.Content.ID, updates)

	if err != nil {
		return err
	}

	err = h.worker.ProcessImage(ctx, processEvent.Content, processEvent.ProcessingVersion)
	if err != nil {
		return err
	}

	return nil
}
