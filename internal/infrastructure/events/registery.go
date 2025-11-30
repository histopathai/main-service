package events

import (
	"context"
	"sync"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type EventRegistry struct {
	handlers map[events.EventType][]port.EventHandler
	mu       sync.RWMutex
}

func NewEventRegistry() *EventRegistry {
	return &EventRegistry{
		handlers: make(map[events.EventType][]port.EventHandler),
	}
}

func (er *EventRegistry) Register(eventType events.EventType, handler port.EventHandler) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.handlers[eventType] = append(er.handlers[eventType], handler)
}

func (er *EventRegistry) Handle(ctx context.Context, data []byte, attributes map[string]string) error {
	eventTypeStr, ok := attributes["event_type"]
	if !ok {
		return errors.NewInternalError("event_type attribute missing in event attributes", nil)
	}

	eventType := events.EventType(eventTypeStr)
	er.mu.RLock()
	handlers, exists := er.handlers[eventType]
	er.mu.RUnlock()

	if !exists || len(handlers) == 0 {
		return errors.NewInternalError("no handlers registered for event type", nil)
	}

	var lastErr error
	for _, handler := range handlers {
		if err := handler(ctx, data, attributes); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
