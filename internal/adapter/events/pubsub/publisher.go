// adapter/event/pubsub/publisher.go
package pubsub

import (
	"context"
	"time"

	"cloud.google.com/go/pubsub"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

type PubSubPublisher struct {
	client        *pubsub.Client
	serializer    *EventSerializer
	topicResolver portevent.TopicResolver
}

func NewPubSubPublisher(ctx context.Context, projectID string) (*PubSubPublisher, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &PubSubPublisher{
		client:        client,
		serializer:    NewEventSerializer(),
		topicResolver: NewTopicResolver(projectID),
	}, nil
}

func (p *PubSubPublisher) Publish(ctx context.Context, event domainevent.Event) error {
	// 1. Serialize event
	data, err := p.serializer.Serialize(event)
	if err != nil {
		return err
	}

	// 2. Resolve topic
	topicID := p.topicResolver.ResolveTopic(event.GetEventType())
	topic := p.client.Topic(topicID)
	defer topic.Stop()

	// 3. Prepare message
	msg := &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"event_type": string(event.GetEventType()),
			"event_id":   event.GetEventID(),
			"timestamp":  event.GetTimestamp().Format(time.RFC3339),
		},
	}

	// 4. Publish
	result := topic.Publish(ctx, msg)
	_, err = result.Get(ctx)
	return err
}

func (p *PubSubPublisher) Stop() error {
	return p.client.Close()
}
