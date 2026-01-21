package port

import (
	"context"
)

type Serializer interface {
	Serialize(event interface{}) ([]byte, error)
	Deserialize(data []byte, v interface{}) error
}

type Publisher interface {
	Publish(ctx context.Context, topic string, data []byte, attributes map[string]string) error
}

type EventHandler func(ctx context.Context, data []byte, attributes map[string]string) error

type Subscriber interface {
	Subscribe(ctx context.Context, subscription string, handler EventHandler) error
	Stop() error
}
