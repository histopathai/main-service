package event

import (
	"context"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
)

type EventPublisher interface {
	Publish(ctx context.Context, event domainevent.Event) error
}

type TopicResolver interface {
	ResolveTopic(eventType domainevent.EventType) string
}
