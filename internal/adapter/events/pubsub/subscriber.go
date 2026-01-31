// adapter/event/pubsub/subscriber.go
package pubsub

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/pubsub"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

type PubSubSubscriber struct {
	client         *pubsub.Client
	subscriptionID string
	serializer     *EventSerializer
	handler        portevent.EventHandler
}

func NewPubSubSubscriber(
	ctx context.Context,
	projectID string,
	subscriptionID string,
	handler portevent.EventHandler,

) (*PubSubSubscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &PubSubSubscriber{
		client:         client,
		subscriptionID: subscriptionID,
		serializer:     NewEventSerializer(),
		handler:        handler,
	}, nil
}

func (s *PubSubSubscriber) Subscribe(ctx context.Context, handler portevent.EventHandler) error {
	sub := s.client.Subscription(s.subscriptionID)

	// Configure settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// 1. Get event type from attributes
		eventTypeStr, ok := msg.Attributes["event_type"]
		if !ok {
			msg.Nack()
			return
		}

		eventType := domainevent.EventType(eventTypeStr)

		// 2. Deserialize
		event, err := s.serializer.Deserialize(msg.Data, eventType)
		if err != nil {
			fmt.Printf("Failed to deserialize event: %v\n", err)
			msg.Nack()
			return
		}

		// 4. Handle
		if err := handler.Handle(ctx, event); err != nil {
			return
		}

		msg.Ack()
	})
}

func (s *PubSubSubscriber) Stop() error {
	return s.client.Close()
}
