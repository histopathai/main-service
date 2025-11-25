package eventpublisher

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
	"github.com/histopathai/main-service/internal/domain/port"
	"github.com/histopathai/main-service/internal/shared/errors"
)

type TelemetryEventPublisher struct {
	*EventPublisher
}

func NewTelemetryEventPublisher(
	message port.Publisher,
	topics map[events.EventType]string,
) *TelemetryEventPublisher {
	basePublisher := NewEventPublisher(message, topics)
	return &TelemetryEventPublisher{
		EventPublisher: basePublisher,
	}
}

func (tep *TelemetryEventPublisher) PublishTelemetryDLQMessage(
	ctx context.Context,
	event *events.DLQMessageEvent,
) error {

	err := tep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("failed to publish telemetry DLQ message event", err)
	}
	return nil
}

func (tep *TelemetryEventPublisher) PublishTelemetryError(
	ctx context.Context,
	event *events.TelemetryErrorEvent,
) error {

	err := tep.publishEvent(ctx, event)
	if err != nil {
		return errors.NewInternalError("failed to publish telemetry error event", err)
	}
	return nil
}
