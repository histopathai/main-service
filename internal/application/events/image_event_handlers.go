package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/histopathai/main-service/internal/application/commands"
	"github.com/histopathai/main-service/internal/application/service"
	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/port"
)

const (
	MaxProcessingRetries = 3
)

type ImageEventHandlers struct {
	imageService *service.ImageService
	worker       port.ImageProcessingWorker
	storage      port.Storage
	publisher    port.EventPublisher
}

func NewImageEventHandlers(
	imageService *service.ImageService,
	worker port.ImageProcessingWorker,
	storage port.Storage,
	publisher port.EventPublisher,
) *ImageEventHandlers {
	return &ImageEventHandlers{
		imageService: imageService,
		worker:       worker,
		storage:      storage,
		publisher:    publisher,
	}
}

func (h *ImageEventHandlers) HandleImageDeletionRequested(event *vobj.Event) error {
	var payload events.ImageDeletionRequestPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	for bucketName, prefixes := range payload.Targets {
		for _, prefix := range prefixes {
			if err := h.storage.DeleteByPrefix(context.Background(), bucketName, prefix); err != nil {
				log.Printf("Storage cleanup failed: bucket=%s prefix=%s err=%v", bucketName, prefix, err)
				// Consider publishing a retry event or dead letter queue
			}
		}
	}

	return nil
}

func (h *ImageEventHandlers) HandleUploaded(event *vobj.Event) error {
	var payload events.ImageUploadedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	cmd, err := commands.NewCreateImageCommand(
		payload.Name,
		payload.CreatorID,
		payload.Parent,
		payload.Format,
		payload.OriginPath,
		payload.ProcessedPath,
		payload.Width,
		payload.Height,
		payload.Size,
	)
	if err != nil {
		return fmt.Errorf("create command error: %w", err)
	}

	_, err = h.imageService.Create(context.Background(), cmd)
	if err != nil {
		return fmt.Errorf("create image error: %w", err)
	}

	if !payload.IsProcessed {
		// will be published after processing
	}
	return nil
}

func (h *ImageEventHandlers) HandleProcessingRequested(event *vobj.Event) error {
	var payload events.ImageProcessingRequestedPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	process_input := port.ProcessingInput{
		ImageID:    payload.ImageID,
		OriginPath: payload.OriginPath,
		BucketName: payload.BucketName,
		Size:       payload.Size,
	}

	if err := h.worker.ProcessImage(context.Background(), &process_input); err != nil {
		return fmt.Errorf("process image error: %w", err)
	}

	return nil
}

func (h *ImageEventHandlers) HandleProcessingCompleted(event *vobj.Event) error {
	var payload events.ImageProcessingResultPayload
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	// Update image entity
	//cmd := &commands.UpdateImageCommand{
	//		ID:            payload.ImageID,
	//		ProcessedPath: payload.ProcessedPath,
	//		Width:         payload.Width,
	//		Height:        payload.Height,
	//		Size:          payload.Size,
	//	}

	// TODO: Uncomment when UpdateUseCase is implemented
	// _, err := h.imageService.Update(context.Background(), cmd)
	// if err != nil {
	// 	return fmt.Errorf("update image error: %w", err)
	// }

	log.Printf("Image processing completed: id=%s (update pending UpdateUseCase implementation)", payload.ImageID)

	return nil
}
