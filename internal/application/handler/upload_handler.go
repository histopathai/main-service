package handler

import (
	"context"
	"log"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"

	portevent "github.com/histopathai/main-service/internal/port/event"
)

type UploadHandler struct {
	subscriber portevent.EventSubscriber
	publisher  portevent.EventPublisher
	repo       port.ContentRepository
	logger     *slog.Logger
}

func NewUploadHandler(
	subscriber portevent.EventSubscriber,
	repo port.ContentRepository,
	publisher portevent.EventPublisher,
	logger *slog.Logger,
) *UploadHandler {
	return &UploadHandler{
		subscriber: subscriber,
		repo:       repo,
		publisher:  publisher,
		logger:     logger,
	}
}

func (h *UploadHandler) Start(ctx context.Context) error {
	log.Println("UploadHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *UploadHandler) Stop() error {
	log.Println("UploadHandler stopping...")
	return h.subscriber.Stop()
}

func (h *UploadHandler) Handle(ctx context.Context, event domainevent.Event) error {

	uploadedEvent, ok := event.(domainevent.UploadEvent)
	if !ok {
		h.logger.Warn("UploadHandler: received unsupported event type")
		return nil
	}

	content := &uploadedEvent.Content

	created, err := h.repo.Read(ctx, content.ID)
	if err != nil {
		return err
	}

	contentCategory := created.ContentType.GetCategory()

	if contentCategory == "image" && !created.ContentType.IsThumbnail() {
		publishEvent := domainevent.ImageProcessEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: domainevent.ImageProcessEventType,
				Timestamp: time.Now(),
			},
			Content:           *created,
			ProcessingVersion: vobj.ProcessingV2,
		}
		err := h.publisher.Publish(ctx, publishEvent)
		if err != nil {
			h.logger.Error("UploadHandler: failed to publish image process event", "error", err)
			return err
		}
	}
	return nil
}
