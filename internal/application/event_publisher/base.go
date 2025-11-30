package eventpublisher

import (
	"context"
	"fmt"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type EventPublisher struct {
	message port.Publisher
	topics  map[events.EventType]string
}

func NewEventPublisher(
	message port.Publisher,
	topics map[events.EventType]string,
) *EventPublisher {
	return &EventPublisher{
		message: message,
		topics:  topics,
	}
}

func (ep *EventPublisher) publishEvent(ctx context.Context, event events.Event) error {
	topicID, ok := ep.topics[event.GetEventType()]
	if !ok {

		return fmt.Errorf("no topic configured for event type: %s", event.GetEventType())
	}

	serializer := events.NewJSONEventSerializer()
	data, err := serializer.Serialize(event)
	if err != nil {
		return errors.NewInternalError(fmt.Sprintf("failed to serialize event %s", event.GetEventType()), err)
	}

	attributes := map[string]string{
		"event_type": string(event.GetEventType()),
		"event_id":   event.GetEventID(),
	}

	return ep.message.Publish(ctx, topicID, data, attributes)
}
