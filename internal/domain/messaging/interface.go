package messaging

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, data []byte, attributes map[string]string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, subscription string, handler MessageHandler) error
	Stop() error
}

type MessageHandler func(ctx context.Context, data []byte, attributes map[string]string) error

type PubSubClient interface {
	Publisher
	Subscriber
}
