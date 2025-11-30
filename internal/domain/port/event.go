package port

import (
	"context"

	"github.com/histopathai/main-service/internal/domain/events"
)

type Publisher interface {
	Publish(ctx context.Context, topic string, data []byte, attributes map[string]string) error
}

type EventHandler func(ctx context.Context, data []byte, attributes map[string]string) error

type Subscriber interface {
	Subscribe(ctx context.Context, subscription string, handler EventHandler) error
	Stop() error
}

// ImageEventPublisher handles image-related event publishing
type ImageEventPublisher interface {
	PublishImageProcessingRequested(ctx context.Context, event *events.ImageProcessingRequestedEvent) error
	PublishImageDeletionRequested(ctx context.Context, event *events.ImageDeletionRequestedEvent) error
}

// TelemetryEventPublisher handles telemetry event publishing
type TelemetryEventPublisher interface {
	PublishTelemetryDLQMessage(ctx context.Context, event *events.DLQMessageEvent) error
	PublishTelemetryError(ctx context.Context, event *events.TelemetryErrorEvent) error
}
