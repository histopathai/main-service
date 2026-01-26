// adapter/event/pubsub/subscriber.go
package pubsub

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/model"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

type PubSubSubscriber struct {
	client         *pubsub.Client
	subscriptionID string
	serializer     *EventSerializer
	handler        portevent.EventHandler
	retryPolicies  map[domainevent.EventType]RetryPolicy
	dlqPublisher   portevent.EventPublisher
}

func NewPubSubSubscriber(
	ctx context.Context,
	projectID string,
	subscriptionID string,
	handler portevent.EventHandler,
	retryPolicies map[domainevent.EventType]RetryPolicy,
	dlqPublisher portevent.EventPublisher,
) (*PubSubSubscriber, error) {
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &PubSubSubscriber{
		client:         client,
		subscriptionID: subscriptionID,
		serializer:     NewEventSerializer(),
		handler:        handler,
		retryPolicies:  retryPolicies,
		dlqPublisher:   dlqPublisher,
	}, nil
}

func (s *PubSubSubscriber) Subscribe(ctx context.Context, handler portevent.EventHandler) error {
	sub := s.client.Subscription(s.subscriptionID)

	// Configure settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// 1. Get event type from attributes
		eventTypeStr, ok := msg.Attributes["event_type"]
		if !ok {
			msg.Nack()
			return
		}

		eventType := domainevent.EventType(eventTypeStr)

		// 2. Get retry count
		retryCount := s.getRetryCount(msg)

		// 3. Deserialize
		event, err := s.serializer.Deserialize(msg.Data, eventType)
		if err != nil {
			fmt.Printf("Failed to deserialize event: %v\n", err)
			msg.Nack()
			return
		}

		// 4. Handle
		if err := handler.Handle(ctx, event); err != nil {
			s.handleFailure(ctx, msg, event, eventType, retryCount, err)
			return
		}

		msg.Ack()
	})
}

func (s *PubSubSubscriber) handleFailure(
	ctx context.Context,
	msg *pubsub.Message,
	event domainevent.Event,
	eventType domainevent.EventType,
	retryCount int,
	handlerErr error,
) {
	policy, exists := s.retryPolicies[eventType]
	if !exists {
		policy = RetryPolicy{MaxAttempts: 3, BaseBackoff: 1 * time.Second}
	}

	// Check if we should retry
	if retryCount < policy.MaxAttempts {
		// Nack to trigger retry (Pub/Sub will redeliver)
		fmt.Printf("Retrying event (attempt %d/%d): %v\n", retryCount+1, policy.MaxAttempts, handlerErr)
		msg.Nack()
		return
	}

	// Max retries exhausted - send to DLQ
	fmt.Printf("Max retries exhausted for event, sending to DLQ: %v\n", handlerErr)
	s.publishToDLQ(ctx, event, handlerErr, retryCount)
	msg.Ack() // Ack to remove from subscription
}

func (s *PubSubSubscriber) getRetryCount(msg *pubsub.Message) int {
	if countStr, ok := msg.Attributes["retry_count"]; ok {
		count, _ := strconv.Atoi(countStr)
		return count
	}
	// Use delivery attempt from Pub/Sub
	if msg.DeliveryAttempt != nil {
		return int(*msg.DeliveryAttempt) - 1
	}
	return 0
}

func (s *PubSubSubscriber) publishToDLQ(ctx context.Context, event domainevent.Event, err error, retryCount int) {
	if s.dlqPublisher == nil {
		fmt.Printf("DLQ publisher not configured, cannot send event to DLQ\n")
		return
	}

	// Create DLQ event based on original event type
	var dlqEvent domainevent.Event

	switch e := event.(type) {
	case *domainevent.ImageProcessCompleteEvent:
		retryMeta := &domainevent.RetryMetadata{
			AttemptCount:   retryCount,
			MaxAttempts:    s.retryPolicies[e.EventType].MaxAttempts,
			LastAttemptAt:  time.Now(),
			FirstAttemptAt: time.Now().Add(-time.Duration(retryCount) * time.Minute), // Approximate
		}

		var content model.Content
		if len(e.Contents) > 0 {
			content = e.Contents[0]
		}

		dlqEvent = &domainevent.ImageProcessDlqEvent{
			BaseEvent: domainevent.BaseEvent{
				EventID:   uuid.New().String(),
				EventType: domainevent.ImageProcessDlqEventType,
				Timestamp: time.Now(),
			},
			ImageID:           e.ImageID,
			Content:           content,
			ProcessingVersion: e.ProcessingVersion,
			FailureReason:     err.Error(),
			Retryable:         e.Retryable,
			RetryMetadata:     retryMeta,
			OriginalEventID:   e.EventID,
		}
	}

	if dlqEvent != nil {
		if err := s.dlqPublisher.Publish(ctx, dlqEvent); err != nil {
			fmt.Printf("Failed to publish to DLQ: %v\n", err)
		}
	}
}

func (s *PubSubSubscriber) Stop() error {
	return s.client.Close()
}
