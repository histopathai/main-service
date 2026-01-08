package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type EventPublisher interface {
	Publish(ctx context.Context, event *vobj.Event) error
	PublishBatch(ctx context.Context, events []*vobj.Event) error
	Close() error
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, eventType vobj.EventType, handler EventHandler) error
	Unsubscribe(ctx context.Context, eventType vobj.EventType) error
	Close() error
}

type EventHandler func(ctx context.Context, event *vobj.Event) error

type PublishOptions struct {
	OrderingKey string
	Retry       bool
	MaxRetries  int
	Attributes  map[string]string
}
