package handler

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service/internal/application/usecase"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

type ImageProcessHandler struct {
	subscriber   portevent.EventSubscriber
	worker       port.ImageProcessingWorker
	imageUsecase *usecase.ImageUseCase
	logger       *slog.Logger
}

func NewImageProcessHandler(
	subscriber portevent.EventSubscriber,
	worker port.ImageProcessingWorker,
	imageUsecase *usecase.ImageUseCase,
	logger *slog.Logger,
) *ImageProcessHandler {
	return &ImageProcessHandler{
		subscriber:   subscriber,
		worker:       worker,
		imageUsecase: imageUsecase,
		logger:       logger,
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

	err := h.imageUsecase.UpdateStatus(ctx, processEvent.Content.ID, vobj.StatusProcessing)

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
	subscriber     portevent.EventSubscriber
	publisher      portevent.EventPublisher
	imageUsecase   *usecase.ImageUseCase
	contentUsecase *usecase.ContentUseCase
	logger         *slog.Logger
}

func NewImageProcessCompleteHandler(
	subscriber portevent.EventSubscriber,
	publisher portevent.EventPublisher,
	imageUsecase *usecase.ImageUseCase,
	contentUsecase *usecase.ContentUseCase,
	logger *slog.Logger,
) *ImageProcessCompleteHandler {
	return &ImageProcessCompleteHandler{
		subscriber:     subscriber,
		publisher:      publisher,
		imageUsecase:   imageUsecase,
		contentUsecase: contentUsecase,
		logger:         logger,
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

		for _, content := range processCompleteEvent.Contents {

			_, err := h.contentUsecase.Create(ctx, &content)
			if err != nil {
				return err
			}

		}

	} else {

		// TODO: Handle failure
		// Will be implemented in the future
		return nil

	}

	return nil
}
