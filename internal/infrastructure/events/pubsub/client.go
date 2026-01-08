package pubsub

import (
	"context"
	"fmt"
	"sync"

	"cloud.google.com/go/pubsub"
	"github.com/histopathai/main-service/internal/domain/vobj"
	"github.com/histopathai/main-service/internal/events"
	"github.com/histopathai/main-service/internal/port"
)

type GooglePubSubPublisher struct {
	client   *pubsub.Client
	topicMap map[vobj.EventType]*pubsub.Topic
	router   *events.EventRouter
}

func NewGooglePubSubPublisher(ctx context.Context, projectID string, router *events.EventRouter) (*GooglePubSubPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, mapPubSubError(err, "failed to create pubsub client")
	}

	return &GooglePubSubPublisher{
		client:   client,
		topicMap: make(map[vobj.EventType]*pubsub.Topic),
		router:   router,
	}, nil
}

func (p *GooglePubSubPublisher) Publish(ctx context.Context, event *vobj.Event) error {
	topicID, err := p.router.GetTopic(event.Type)
	if err != nil {
		return err
	}

	topic, err := p.getTopic(topicID)
	if err != nil {
		return err
	}

	attributes := map[string]string{
		"event_type": string(event.Type),
		"event_id":   event.ID,
		"timestamp":  fmt.Sprintf("%d", event.Timestamp),
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data:       event.Payload,
		Attributes: attributes,
	})

	if _, err := result.Get(ctx); err != nil {
		return mapPubSubError(err, "failed to publish event")
	}

	return nil
}

func (p *GooglePubSubPublisher) PublishBatch(ctx context.Context, events []*vobj.Event) error {
	for _, event := range events {
		if err := p.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

func (p *GooglePubSubPublisher) getTopic(topicID string) (*pubsub.Topic, error) {
	topic := p.client.Topic(topicID)
	exists, err := topic.Exists(context.Background())
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, fmt.Errorf("topic %s does not exist", topicID)
	}
	return topic, nil
}

func (p *GooglePubSubPublisher) Close() error {
	for _, topic := range p.topicMap {
		topic.Stop()
	}
	return p.client.Close()
}

// Subscriber implementation
type GooglePubSubSubscriber struct {
	client        *pubsub.Client
	handlers      map[vobj.EventType]port.EventHandler
	subscriptions map[vobj.EventType]*pubsub.Subscription
	mu            sync.RWMutex
}

func NewGooglePubSubSubscriber(ctx context.Context, projectID string) (*GooglePubSubSubscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, mapPubSubError(err, "failed to create pubsub client")
	}

	return &GooglePubSubSubscriber{
		client:        client,
		handlers:      make(map[vobj.EventType]port.EventHandler),
		subscriptions: make(map[vobj.EventType]*pubsub.Subscription),
	}, nil
}

func (s *GooglePubSubSubscriber) Subscribe(ctx context.Context, eventType vobj.EventType, handler port.EventHandler) error {
	s.mu.Lock()
	s.handlers[eventType] = handler
	s.mu.Unlock()

	subscriptionID := fmt.Sprintf("%s-subscription", eventType)
	sub := s.client.Subscription(subscriptionID)

	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10

	go func() {
		err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
			event := &vobj.Event{
				ID:        msg.Attributes["event_id"],
				Type:      eventType,
				Timestamp: parseTimestamp(msg.Attributes["timestamp"]),
				Payload:   msg.Data,
			}

			if err := handler(ctx, event); err != nil {
				msg.Nack()
				return
			}
			msg.Ack()
		})
		if err != nil {
			// Log error
		}
	}()

	return nil
}

func (s *GooglePubSubSubscriber) Unsubscribe(ctx context.Context, eventType vobj.EventType) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.handlers, eventType)
	if sub, ok := s.subscriptions[eventType]; ok {
		delete(s.subscriptions, eventType)
		// PubSub subscription'ı otomatik olarak kapanacak
		_ = sub
	}
	return nil
}

func (s *GooglePubSubSubscriber) Close() error {
	return s.client.Close()
}

func parseTimestamp(ts string) int64 {
	var timestamp int64
	fmt.Sscanf(ts, "%d", &timestamp)
	return timestamp
}
