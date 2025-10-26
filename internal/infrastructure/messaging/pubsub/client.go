package pubsub

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/pubsub"
	"github.com/histopathai/main-service-refactor/internal/domain/events"
)

type GooglePubSubClient struct {
	client *pubsub.Client
	logger *slog.Logger
}

func NewGooglePubSubClient(ctx context.Context, projectID string, logger *slog.Logger) (*GooglePubSubClient, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pubsub client: %w", err)
	}

	return &GooglePubSubClient{
		client: client,
		logger: logger,
	}, nil
}

func (g *GooglePubSubClient) Publish(ctx context.Context, topicID string, data []byte, attributes map[string]string) error {
	topic := g.client.Topic(topicID)

	result := topic.Publish(ctx, &pubsub.Message{
		Data:       data,
		Attributes: attributes,
	})

	_, err := result.Get(ctx)
	if err != nil {
		g.logger.Error("Failed to publish message", "error", err, "topic", topicID)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	g.logger.Info("Message published successfully", "topic", topicID)
	return nil
}

func (g *GooglePubSubClient) Subscribe(ctx context.Context, subscriptionID string, handler events.EventHandler) error {
	sub := g.client.Subscription(subscriptionID)

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		err := handler(ctx, msg.Data, msg.Attributes)
		if err != nil {
			g.logger.Error("Message handler error", "error", err, "subscription", subscriptionID)
			msg.Nack()
			return
		}
		msg.Ack()
	})
}

func (g *GooglePubSubClient) Stop() error {
	return g.client.Close()
}
