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
	processEvent, ok := event.(domainevent.ImageProcessEvent)
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

type ImageProcessCompleteHandler struct {
	subscriber  portevent.EventSubscriber
	publisher   portevent.EventPublisher
	imageRepo   port.ImageRepository
	contentRepo port.ContentRepository
	logger      *slog.Logger
}

func NewImageProcessCompleteHandler(
	subscriber portevent.EventSubscriber,
	publisher portevent.EventPublisher,
	imageRepo port.ImageRepository,
	contentRepo port.ContentRepository,
	logger *slog.Logger,
) *ImageProcessCompleteHandler {
	return &ImageProcessCompleteHandler{
		subscriber: subscriber,
		publisher:  publisher,
		imageRepo:  imageRepo,
		logger:     logger,
	}
}

func (h *ImageProcessCompleteHandler) Start(ctx context.Context) error {
	h.logger.Info("ImageProcessCompleteHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *ImageProcessCompleteHandler) Stop() error {
	h.logger.Info("ImageProcessCompleteHandler stopping...")
	return h.subscriber.Stop()
}

func (h *ImageProcessCompleteHandler) Handle(ctx context.Context, event domainevent.Event) error {
	processCompleteEvent, ok := event.(domainevent.ImageProcessCompleteEvent)
	if !ok {
		h.logger.Warn("ImageProcessCompleteHandler: received unsupported event type")
		return nil
	}

	if processCompleteEvent.Success {
		imageUpdates := map[string]any{}
		for _, content := range processCompleteEvent.Contents {

			_, err := h.contentRepo.Create(ctx, &content)
			if err != nil {
				return err
			}

			if content.ContentType.IsThumbnail() {
				imageUpdates[fields.ImageThumbnailContentID.DomainName()] = content.ID
			}

			if content.ContentType.IsDZI() {
				imageUpdates[fields.ImageDziContentID.DomainName()] = content.ID
			}

			if content.ContentType.IsIndexMap() {
				imageUpdates[fields.ImageIndexmapContentID.DomainName()] = content.ID
			}

			if content.ContentType.IsTiles() {
				imageUpdates[fields.ImageTilesContentID.DomainName()] = content.ID
			}

			if content.ContentType.IsArchive() {
				imageUpdates[fields.ImageZipTilesContentID.DomainName()] = content.ID
			}

		}

		imageUpdates[fields.ImageProcessingStatus.DomainName()] = vobj.StatusProcessed
		err := h.imageRepo.Update(ctx, processCompleteEvent.ImageID, imageUpdates)
		if err != nil {
			return err
		}

	} else {

		// TODO: Handle failure
		// Will be implemented in the future
		return nil

	}

	return nil
}
