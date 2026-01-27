package handler

import (
	"context"
	"log/slog"

	"github.com/histopathai/main-service/internal/application/usecase"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/port"

	portevent "github.com/histopathai/main-service/internal/port/event"
)

type DeleteFileHandler struct {
	subscriber   portevent.EventSubscriber
	imageUsecase *usecase.ImageUseCase
	repo         port.Repository[*model.Content]
	logger       *slog.Logger
}

func NewDeleteFileHandler(
	subscriber portevent.EventSubscriber,
	imageUsecase *usecase.ImageUseCase,
	repo port.Repository[*model.Content],
	logger *slog.Logger,
) *DeleteFileHandler {
	return &DeleteFileHandler{
		subscriber:   subscriber,
		imageUsecase: imageUsecase,
		repo:         repo,
		logger:       logger,
	}
}

func (h *DeleteFileHandler) Start(ctx context.Context) error {
	h.logger.Info("DeleteFileHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *DeleteFileHandler) Stop() error {
	h.logger.Info("DeleteFileHandler stopping...")
	return h.subscriber.Stop()
}

func (h *DeleteFileHandler) Handle(ctx context.Context, event domainevent.Event) error {
	deleteEvent, ok := event.(*domainevent.DeleteFileEvent)
	if !ok {
		h.logger.Warn("DeleteFileHandler: received unsupported event type")
		return nil
	}
	contentID := deleteEvent.Content.ID

	//TODO: Hard delete will be implemented in the future
	if err := h.repo.SoftDelete(ctx, contentID); err != nil {
		return err
	}

	return nil
}
