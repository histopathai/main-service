package pubsub

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/histopathai/main-service-refactor/internal/domain/events"
)

type PublishOptions struct {
	OrderingKey string
	Retry       bool
	MaxRetries  int
}

type GooglePubSubClient struct {
	client     *pubsub.Client
	logger     *slog.Logger
	serializer events.EventSerializer
}

func NewGooglePubSubClient(ctx context.Context, projectID string, logger *slog.Logger) (*GooglePubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return &GooglePubSubClient{
		client:     client,
		logger:     logger,
		serializer: events.NewJSONEventSerializer(),
	}, nil
}

func (g *GooglePubSubClient) PublishEvent(ctx context.Context, topicID string, event events.Event) error {
	data, err := g.serializer.Serialize(event)
	if err != nil {
		g.logger.Error("Failed to serialize event", "error", err, "eventType", event.GetEventType())
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	attributes := map[string]string{
		"event_type": string(event.GetEventType()),
		"event_id":   event.GetEventID(),
		"timestamp":  event.GetTimestamp().Format(time.RFC3339),
	}

	return g.Publish(ctx, topicID, data, attributes)
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
		g.logger.Error("Failed to publish message",
			"error", err,
			"topic", topicID,
			"eventType", attributes["event_type"])
		return fmt.Errorf("failed to publish message: %w", err)
	}

	g.logger.Info("Message published successfully",
		"topic", topicID,
		"eventType", attributes["event_type"],
		"eventID", attributes["event_id"])
	return nil
}

func (g *GooglePubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler events.EventHandler) error {
	sub := g.client.Subscription(subscriptionID)

	// Configure subscription settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute

	g.logger.Info("Starting subscription", "subscription", subscriptionID)

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		startTime := time.Now()

		g.logger.Debug("Received message",
			"subscription", subscriptionID,
			"messageID", msg.ID,
			"eventType", msg.Attributes["event_type"])

		err := handler(ctx, msg.Data, msg.Attributes)

		duration := time.Since(startTime)

		if err != nil {
			g.logger.Error("Message handler error",
				"error", err,
				"subscription", subscriptionID,
				"messageID", msg.ID,
				"eventType", msg.Attributes["event_type"],
				"duration", duration)
			msg.Nack()
			return
		}

		g.logger.Info("Message processed successfully",
			"subscription", subscriptionID,
			"messageID", msg.ID,
			"eventType", msg.Attributes["event_type"],
			"duration", duration)
		msg.Ack()
	})
}

func (g *GooglePubSubClient) Stop() error {
	return g.client.Close()
}
