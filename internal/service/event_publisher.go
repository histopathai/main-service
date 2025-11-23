package service

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type EventPublisher struct {
	pubsub events.Publisher
	topic  map[events.EventType]string
}

func NewEventPublisher(
	pubsub events.Publisher,
	topic map[events.EventType]string,
) events.ImageEventPublisher {
	return &EventPublisher{
		pubsub: pubsub,
		topic:  topic,
	}
}
func (ep *EventPublisher) PublishImageProcessingRequested(ctx context.Context, event *events.ImageProcessingRequestedEvent) error {
	err := ep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("Failed to publish ImageProcessingRequested event: %v", err)
	}
	return nil
}
func (ep *EventPublisher) PublishImageDeletionRequested(ctx context.Context, event *events.ImageDeletionRequestedEvent) error {
	err := ep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("Failed to publish ImageDeletionRequested event: %v", err)
	}
	return nil
}

func (ep *EventPublisher) publishEvent(ctx context.Context, event events.Event) error {

	topicID, ok := ep.topic[event.GetEventType()]
	if !ok {
		return nil
	}

	serializer := events.NewJSONEventSerializer()
	data, err := serializer.Serialize(event)
	if err != nil {
		message := fmt.Sprintf("No topic configured for event type %s", event.GetEventType())
		return errors.NewNotFoundError(message)
	}

	attributes := map[string]string{
		"event_type": string(event.GetEventType()),
		"event_id":   event.GetEventID(),
	}

	return ep.pubsub.Publish(ctx, topicID, data, attributes)
}
