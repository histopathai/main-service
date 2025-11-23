package events

import "context"

type Publisher interface {
	Publish(ctx context.Context, topic string, data []byte, attributes map[string]string) error
}

type Subscriber interface {
	Subscribe(ctx context.Context, subscription string, handler EventHandler) error
	Stop() error
}

type EventHandler func(ctx context.Context, data []byte, attributes map[string]string) error

type PubSubClient interface {
	Publisher
	Subscriber
}

type ImageEventPublisher interface {
	PublishImageProcessingRequested(ctx context.Context, event *ImageProcessingRequestedEvent) error
	PublishImageDeletionRequested(ctx context.Context, event *ImageDeletionRequestedEvent) error
}
