package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/histopathai/main-service-refactor/internal/domain/events"
	"github.com/histopathai/main-service-refactor/internal/shared/errors"
)

type EventPublisher struct {
	pubsub events.Publisher
	logger *slog.Logger
	topic  map[events.EventType]string
}

func NewEventPublisher(
	pubsub events.Publisher,
	logger *slog.Logger,
	topic map[events.EventType]string,
) *EventPublisher {
	return &EventPublisher{
		pubsub: pubsub,
		logger: logger,
		topic:  topic,
	}
}
func (ep *EventPublisher) PublishImageProcessingRequested(ctx context.Context, event *events.ImageProcessingRequestedEvent) error {
	return ep.publishEvent(ctx, event)
}

func (ep *EventPublisher) publishEvent(ctx context.Context, event events.Event) error {

	topicID, ok := ep.topic[event.GetEventType()]
	if !ok {
		ep.logger.Error("no topic configured for event type", slog.String("event_type", string(event.GetEventType())))
		return nil
	}

	serializer := events.NewJSONEventSerializer()
	data, err := serializer.Serialize(event)
	if err != nil {
		message := fmt.Sprintf("No topic configured for event type %s", event.GetEventType())
		ep.logger.Error(message, slog.String("event_type", string(event.GetEventType())))
		return errors.NewNotFoundError(message)
	}

	attributes := map[string]string{
		"event_type": string(event.GetEventType()),
		"event_id":   event.GetEventID(),
	}

	return ep.pubsub.Publish(ctx, topicID, data, attributes)
}
