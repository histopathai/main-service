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

type DeleteHandler struct {
	subscriber   portevent.EventSubscriber
	imageUsecase *usecase.ImageUseCase
	repo         port.Repository[*model.Content]
	logger       *slog.Logger
}

func NewDeleteHandler(
	subscriber portevent.EventSubscriber,
	imageUsecase *usecase.ImageUseCase,
	repo port.Repository[*model.Content],
	logger *slog.Logger,
) *DeleteHandler {
	return &DeleteHandler{
		subscriber:   subscriber,
		imageUsecase: imageUsecase,
		repo:         repo,
		logger:       logger,
	}
}

func (h *DeleteHandler) Start(ctx context.Context) error {
	h.logger.Info("DeleteHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *DeleteHandler) Stop() error {
	h.logger.Info("DeleteHandler stopping...")
	return h.subscriber.Stop()
}

func (h *DeleteHandler) Handle(ctx context.Context, event domainevent.Event) error {
	deleteEvent, ok := event.(domainevent.DeleteEvent)
	if !ok {
		h.logger.Warn("DeleteHandler: received unsupported event type")
		return nil
	}
	contentID := deleteEvent.Content.ID

	//TODO: Hard delete will be implemented in the future
	if err := h.repo.SoftDelete(ctx, contentID); err != nil {
		return err
	}

	return nil
}
