package repository

import "context"

type MessageBrokerAdapter interface {
	Publish(ctx context.Context, topicID string, message []byte) (string, error)
	Subscribe(ctx context.Context, subscriptionID string, handler func(ctx context.Context, message []byte) error) error
}

type MessageBroker struct {
	adapter MessageBrokerAdapter
}

func NewMessageBroker(adapter MessageBrokerAdapter) *MessageBroker {
	return &MessageBroker{
		adapter: adapter,
	}
}

func (mb *MessageBroker) PublishMessage(ctx context.Context, topicID string, message []byte) (string, error) {
	return mb.adapter.Publish(ctx, topicID, message)
}

func (mb *MessageBroker) SubscribeToMessages(ctx context.Context, subscriptionID string, handler func(ctx context.Context, message []byte) error) error {
	return mb.adapter.Subscribe(ctx, subscriptionID, handler)
}
