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
	newFileEvent, ok := event.(*domainevent.NewFileExistEvent)
	if !ok {
		h.logger.Warn("NewFileHandler: received unsupported event type")
		return nil
	}

	content := &newFileEvent.Content

	uowerr := h.uow.WithTx(ctx, func(ctx context.Context) error {

		imageRepo := h.uow.GetImageRepo()

		imageEntity, err := imageRepo.Read(ctx, content.Parent.ID)

		if err != nil {
			return err
		}

		if imageEntity == nil {
			h.logger.Error("NewFileHandler: image entity not found", "error", err)
			return errors.New("image entity not found")
		}

		// Save Content

		created, err := h.uow.GetContentRepo().Create(ctx, content)
		if err != nil {
			return err
		}
		if created.ContentType.IsThumbnail() {

			if err := imageRepo.Update(ctx, created.Parent.ID, map[string]interface{}{
				fields.ImageThumbnailContentID.DomainName(): created.ID}); err != nil {
				return err
			}

		} else if created.ContentType.IsDZI() {
			if err := imageRepo.Update(ctx, created.Parent.ID, map[string]interface{}{
				fields.ImageDziContentID.DomainName(): created.ID}); err != nil {
				return err
			}

		} else if created.ContentType.IsArchive() {
			if err := imageRepo.Update(ctx, created.Parent.ID, map[string]interface{}{
				fields.ImageZipTilesContentID.DomainName(): created.ID}); err != nil {
				return err
			}

		} else if created.ContentType.IsOriginImage() {
			if err := imageRepo.Update(ctx, created.Parent.ID, map[string]interface{}{
				fields.ImageOriginContentID.DomainName():   created.ID,
				fields.ImageProcessingStatus.DomainName():  vobj.StatusProcessing,
				fields.ImageProcessingVersion.DomainName(): vobj.ProcessingV2,
			}); err != nil {
				return err
			}

		} else if created.ContentType.IsIndexMap() {
			if err := imageRepo.Update(ctx, created.Parent.ID, map[string]interface{}{
				fields.ImageIndexmapContentID.DomainName(): created.ID}); err != nil {
				return err
			}

		}
		return nil
	})

	if uowerr != nil {
		return uowerr
	}

	if content.ContentType.IsOriginImage() {
		err := h.publisher.Publish(ctx, domainevent.ImageProcessReqEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: domainevent.ImageProcessReqEventType,
				Timestamp: time.Now(),
			},
			Content:           *content,
			ProcessingVersion: vobj.ProcessingV2,
		})
		if err != nil {
			return err
		}
	} else {
		// if image thumbnail, dzi, archive, indexmap are set,
		// then update image status to ready

		imageEntity, err := h.uow.GetImageRepo().Read(ctx, content.Parent.ID)
		if err != nil {
			return err
		}
		if imageEntity == nil {
			h.logger.Error("NewFileHandler: image entity not found", "error", err)
			return errors.New("image entity not found")
		}

		if imageEntity.ThumbnailContentID != nil && imageEntity.DziContentID != nil && imageEntity.ZipTilesContentID != nil && imageEntity.IndexmapContentID != nil {
			if err := h.uow.GetImageRepo().Update(ctx, content.Parent.ID, map[string]interface{}{
				fields.ImageProcessingStatus.DomainName(): vobj.StatusProcessed}); err != nil {
				return err
			}
		}

	}

	return nil
}
