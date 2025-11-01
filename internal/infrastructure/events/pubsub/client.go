package pubsub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/histopathai/main-service/internal/domain/events"
)

type PublishOptions struct {
	OrderingKey string
	Retry       bool
	MaxRetries  int
}

type GooglePubSubClient struct {
	client     *pubsub.Client
	serializer events.EventSerializer
}

func NewGooglePubSubClient(ctx context.Context, projectID string) (*GooglePubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, mapPubSubError(err, "failed to create pubsub client")
	}

	return &GooglePubSubClient{
		client:     client,
		serializer: events.NewJSONEventSerializer(),
	}, nil
}

func (g *GooglePubSubClient) PublishEvent(ctx context.Context, topicID string, event events.Event) error {
	data, err := g.serializer.Serialize(event)
	if err != nil {
		return mapPubSubError(err, "failed to serialize event")
	}

	attributes := map[string]string{
		"event_type": string(event.GetEventType()),
		"event_id":   event.GetEventID(),
		"timestamp":  event.GetTimestamp().Format(time.RFC3339),
	}

	err = g.Publish(ctx, topicID, data, attributes)
	if err != nil {
		return mapPubSubError(err, "failed to publish event")
	}
	return nil
}

func (g *GooglePubSubClient) Publish(ctx context.Context, topicID string, data []byte, attributes map[string]string) error {
	topic := g.client.Topic(topicID)
	defer topic.Stop()

	// Enable message ordering if ordering key is present
	if orderingKey, ok := attributes["ordering_key"]; ok && orderingKey != "" {
		topic.EnableMessageOrdering = true
	}

	result := topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})

	_, err := result.Get(ctx)
	if err != nil {
		return mapPubSubError(err, "failed to publish message")
	}

	return nil
}

func (g *GooglePubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler events.EventHandler) error {
	sub := g.client.Subscription(subscriptionID)

	// Configure subscription settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {

		err := handler(ctx, msg.Data, msg.Attributes)

		if err != nil {
			msg.Nack()
			return
		}

		msg.Ack()
	})
}

func (g *GooglePubSubClient) Stop() error {
	return g.client.Close()
}
