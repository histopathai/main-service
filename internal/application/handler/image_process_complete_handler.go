package handler

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/port"

	portevent "github.com/histopathai/main-service/internal/port/event"
)

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
		subscriber:  subscriber,
		publisher:   publisher,
		imageRepo:   imageRepo,
		contentRepo: contentRepo,
		logger:      logger,
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
	h.logger.Info("ImageProcessCompleteHandler: received event", "event", event)
	processCompleteEvent, ok := event.(*domainevent.ImageProcessCompleteEvent)
	if !ok {
		h.logger.Warn("ImageProcessCompleteHandler: received unsupported event type")
		return nil
	}

	if processCompleteEvent.Success {
		for _, content := range processCompleteEvent.Contents {

			event := domainevent.NewFileExistEvent{
				BaseEvent: domainevent.BaseEvent{
					EventID:   uuid.New().String(),
					EventType: domainevent.NewFileExistEventType,
					Timestamp: time.Now(),
				},
				Content: content,
			}

			if err := h.publisher.Publish(ctx, &event); err != nil {
				return err
			}
		}

	} else {

		h.logger.Error("Image processing failed",
			slog.String("image_id", processCompleteEvent.ImageID),
			slog.String("reason", processCompleteEvent.FailureReason),
		)

	}

	return nil
}
