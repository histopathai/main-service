package adapter

import (
	"context"
	"fmt"
	"log/slog"

	"cloud.google.com/go/pubsub"
)

type GooglePubSubAdapter struct {
	client *pubsub.Client
}

func NewGooglePubSubAdapter(client *pubsub.Client) *GooglePubSubAdapter {
	return &GooglePubSubAdapter{
		client: client,
	}
}

func (g *GooglePubSubAdapter) Publish(ctx context.Context, topicID string, data []byte) (string, error) {

	topic := g.client.Topic(topicID)
	result := topic.Publish(ctx, &pubsub.Message{
		Data: data,
	})

	messageId, err := result.Get(ctx)
	if err != nil {
		slog.Error("Pub/Sub publish error", "topic", topicID, "error", err)
		return "", fmt.Errorf("failed to publish message to topic %s: %w", topicID, err)
	}
	slog.Debug("Pub/Sub message published", "topic", topicID, "messageId", messageId)
	return messageId, nil
}

func (g *GooglePubSubAdapter) Subscribe(ctx context.Context, subscriptionID string,
	handler func(ctx context.Context, data []byte) error) error {

	sub := g.client.Subscription(subscriptionID)

	err := sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		slog.Debug("Pub/Sub message received", "subscription", subscriptionID, "messageId", msg.ID)

		err := handler(ctx, msg.Data)
		if err != nil {
			slog.Error("Pub/Sub message process error, nacking message", "messageId", msg.ID, "error", err)
			msg.Nack()
		} else {
			slog.Debug("Pub/Sub message processed successfully, acking message", "messageId", msg.ID)
			msg.Ack()
		}
	})

	if err != nil {
		slog.Error("Pub/Sub subscription receive error", "subscription", subscriptionID, "error", err)
		return err
	}

	slog.Info("Pub/Sub subscription receive ended", "subscription", subscriptionID)
	return nil
}
