package event

import (
	"context"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
)

type EventHandler interface {
	Handle(ctx context.Context, event domainevent.Event) error
}

type EventSubscriber interface {
	Subscribe(ctx context.Context, handler EventHandler) error
	Stop() error
}
