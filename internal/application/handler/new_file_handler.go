package handler

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/fields"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"

	portevent "github.com/histopathai/main-service/internal/port/event"
)

type NewFileHandler struct {
	subscriber portevent.EventSubscriber
	publisher  portevent.EventPublisher
	uow        port.UnitOfWorkFactory
	logger     *slog.Logger
}

func NewNewFileHandler(
	subscriber portevent.EventSubscriber,
	uow port.UnitOfWorkFactory,
	publisher portevent.EventPublisher,
	logger *slog.Logger,
) *NewFileHandler {
	return &NewFileHandler{
		subscriber: subscriber,
		publisher:  publisher,
		uow:        uow,
		logger:     logger,
	}
}

func (h *NewFileHandler) Start(ctx context.Context) error {
	h.logger.Info("NewFileHandler started, listening for events...")
	return h.subscriber.Subscribe(ctx, h)
}

func (h *NewFileHandler) Stop() error {
	h.logger.Info("NewFileHandler stopping...")
	return h.subscriber.Stop()
}

func (h *NewFileHandler) Handle(ctx context.Context, event domainevent.Event) error {
	h.logger.Info("NewFileHandler: received event", "event", event)
	newFileEvent, ok := event.(*domainevent.NewFileExistEvent)
	if !ok {
		h.logger.Warn("NewFileHandler: received unsupported event type")
		return nil
	}

	content := &newFileEvent.Content
	shouldPublish := false
	eventID := uuid.New().String()

	// ... inside WithTx ...
	uowerr := h.uow.WithTx(ctx, func(ctx context.Context) error {

		imageRepo := h.uow.GetImageRepo()

		// 1. Read (Allowed at start)
		imageEntity, err := imageRepo.Read(ctx, content.Parent.ID)
		if err != nil {
			return err
		}
		if imageEntity == nil {
			h.logger.Error("NewFileHandler: image entity not found", "error", err)
			return errors.New("image entity not found")
		}

		// Prepare updates map
		imageUpdates := make(map[string]interface{})

		// 2. Determine updates based on content type
		if content.ContentType.IsThumbnail() {
			imageUpdates[fields.ImageThumbnailContentID.DomainName()] = content.ID
			imageEntity.ThumbnailContentID = &content.ID // Update local model for completion check

		} else if content.ContentType.IsDZI() {
			imageUpdates[fields.ImageDziContentID.DomainName()] = content.ID
			imageEntity.DziContentID = &content.ID

		} else if content.ContentType.IsArchive() {
			imageUpdates[fields.ImageZipTilesContentID.DomainName()] = content.ID
			imageEntity.ZipTilesContentID = &content.ID

		} else if content.ContentType.IsOriginImage() {
			// Idempotency Check
			if imageEntity.Processing != nil && (imageEntity.Processing.Status == vobj.StatusProcessing || imageEntity.Processing.Status == vobj.StatusProcessed) {
				h.logger.Info("NewFileHandler: image already processing or processed, skipping request",
					"image_id", content.Parent.ID,
					"status", imageEntity.Processing.Status)
				return nil
			}

			imageUpdates[fields.ImageOriginContentID.DomainName()] = content.ID
			imageUpdates[fields.ImageProcessingStatus.DomainName()] = vobj.StatusProcessing
			imageUpdates[fields.ImageProcessingVersion.DomainName()] = vobj.ProcessingV2
			imageUpdates[fields.ImageProcessingActiveEventID.DomainName()] = eventID

			// Update local model
			imageEntity.OriginContentID = &content.ID
			// Note: Processing struct might be nil
			if imageEntity.Processing == nil {
				imageEntity.Processing = &vobj.ProcessingInfo{}
			}
			imageEntity.Processing.Status = vobj.StatusProcessing
			imageEntity.Processing.Version = vobj.ProcessingV2
			imageEntity.Processing.ActiveEventID = eventID

			shouldPublish = true

		} else if content.ContentType.IsIndexMap() {
			imageUpdates[fields.ImageIndexmapContentID.DomainName()] = content.ID
			imageEntity.IndexmapContentID = &content.ID
		}

		// 3. Check completion logic (using in-memory imageEntity)
		isComplete := false
		if imageEntity.Processing != nil && !content.ContentType.IsOriginImage() {
			// Only check completion for generated files, not when origin uploads (unless it somehow completes everything immediately which is impossible)
			// Actually origin upload sets status to Processing, so we shouldn't overwrite it to Processed here anyway.

			if imageEntity.Processing.Version == vobj.ProcessingV1 {
				// v1: thumbnail, dzi, tiles
				if imageEntity.ThumbnailContentID != nil && imageEntity.DziContentID != nil && imageEntity.TilesContentID != nil {
					isComplete = true
				}
			} else if imageEntity.Processing.Version == vobj.ProcessingV2 {
				// v2: thumbnail, dzi, zip, indexmap
				if imageEntity.ThumbnailContentID != nil && imageEntity.DziContentID != nil && imageEntity.ZipTilesContentID != nil && imageEntity.IndexmapContentID != nil {
					isComplete = true
				}
			}
		}

		if isComplete {
			imageUpdates[fields.ImageProcessingStatus.DomainName()] = vobj.StatusProcessed
		}

		// 4. Perform Writes (Create Content + Update Image)
		// Writes must come after all reads.

		content.CreatorID = imageEntity.ID
		_, err = h.uow.GetContentRepo().Create(ctx, content)
		if err != nil {
			return err
		}

		if len(imageUpdates) > 0 {
			if err := imageRepo.Update(ctx, content.Parent.ID, imageUpdates); err != nil {
				return err
			}
		}

		return nil
	})

	if uowerr != nil {
		return uowerr
	}

	if shouldPublish {
		// ... publish event ...
		err := h.publisher.Publish(ctx, &domainevent.ImageProcessReqEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:   eventID,
				EventType: domainevent.ImageProcessReqEventType,
				Timestamp: time.Now(),
			},
			Content:           *content,
			ProcessingVersion: vobj.ProcessingV2,
		})
		if err != nil {
			return err
		}

		h.logger.Info("NewFileHandler: published image process request event", "event_id", eventID)
	}

	return nil
}
