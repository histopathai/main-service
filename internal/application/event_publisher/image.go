package eventpublisher

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type ImageEventPublisher struct {
	*EventPublisher
}

func NewImageEventPublisher(
	message port.Publisher,
	topics map[events.EventType]string,
) *ImageEventPublisher {
	basePublisher := NewEventPublisher(message, topics)
	return &ImageEventPublisher{
		EventPublisher: basePublisher,
	}
}

func (iep *ImageEventPublisher) PublishImageProcessingRequested(
	ctx context.Context,
	event *events.ImageProcessingRequestedEvent,
) error {

	err := iep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("failed to publish image processing requested event", err)
	}
	return nil
}

func (iep *ImageEventPublisher) PublishImageDeletionRequested(
	ctx context.Context,
	event *events.ImageDeletionRequestedEvent,
) error {

	err := iep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("failed to publish image deletion requested event", err)
	}
	return nil
}
