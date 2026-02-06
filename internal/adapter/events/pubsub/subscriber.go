// adapter/event/pubsub/subscriber.go
package pubsub

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/google/uuid"
	domainevent "github.com/histopathai/main-service/internal/domain/event"
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
	portcache "github.com/histopathai/main-service/internal/port/cache"
	portevent "github.com/histopathai/main-service/internal/port/event"
)

type PubSubSubscriber struct {
	client         *pubsub.Client
	subscriptionID string
	serializer     *EventSerializer
	handler        portevent.EventHandler
	logger         *slog.Logger
	cache          portcache.Cache
}

func NewPubSubSubscriber(
	ctx context.Context,
	projectID string,
	subscriptionID string,
	handler portevent.EventHandler,
	cache portcache.Cache,
	logger *slog.Logger,
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
		logger:         logger,
		cache:          cache,
	}, nil
}

func (s *PubSubSubscriber) Subscribe(ctx context.Context, handler portevent.EventHandler) error {
	sub := s.client.Subscription(s.subscriptionID)

	// Configure settings
	sub.ReceiveSettings.MaxOutstandingMessages = 100
	sub.ReceiveSettings.NumGoroutines = 10
	sub.ReceiveSettings.MaxExtension = 10 * time.Minute

	return sub.Receive(ctx, func(ctx context.Context, msg *pubsub.Message) {
		// Check for GCS Notification
		if gcsEventType, ok := msg.Attributes["eventType"]; ok && gcsEventType == "OBJECT_FINALIZE" {
			if err := s.handleGCSEvent(ctx, msg, handler); err != nil {
				s.logger.Error("Failed to handle GCS event", "error", err)
				msg.Ack()
				return
			}
			msg.Ack()
			return
		}

		// 1. Get event type from attributes
		eventTypeStr, ok := msg.Attributes["event_type"]
		if !ok {
			// if event type is not found, ignore and remove the message
			s.logger.Warn("Event type not found", "message_id", msg.ID)
			msg.Ack()
			return
		}
		if eventTypeStr == "" {
			// if event type is empty, ignore and remove the message
			s.logger.Warn("Event type is empty", "message_id", msg.ID)
			msg.Ack()
			return
		}
		eventType := domainevent.EventType(eventTypeStr)

		// 2. Deserialize
		if eventType == domainevent.ImageProcessCompleteEventType {
			s.logger.Info("Received raw message", "data", string(msg.Data), "event_type", eventType)
		}
		event, err := s.serializer.Deserialize(msg.Data, eventType)
		if err != nil {
			s.logger.Error("Failed to deserialize event", "error", err, "message_id", msg.ID)
			msg.Ack()
			return
		}

		// Use subscription ID to namespace the cache key
		cacheKey := fmt.Sprintf("%s:%s", s.subscriptionID, event.GetEventID())
		exists, err := s.cache.Has(ctx, cacheKey)
		if err != nil {
			s.logger.Error("Failed to check cache", "error", err, "message_id", msg.ID)
			msg.Ack()
			return
		}
		if exists {
			s.logger.Debug("Event already processed", "message_id", msg.ID, "event_id", event.GetEventID())
			msg.Ack()
			return
		}

		if err := s.cache.Set(ctx, cacheKey, true, 2*time.Hour); err != nil {
			s.logger.Error("Failed to set cache", "error", err, "message_id", msg.ID)
			msg.Ack()
			return
		}

		// 4. Handle
		if err := handler.Handle(ctx, event); err != nil {
			s.logger.Error("Handler failed", "error", err, "message_id", msg.ID)
			msg.Ack()
			return
		}

		msg.Ack()
	})
}

func (s *PubSubSubscriber) handleGCSEvent(ctx context.Context, msg *pubsub.Message, handler portevent.EventHandler) error {
	var gcsObj struct {
		Name        string            `json:"name"`
		Bucket      string            `json:"bucket"`
		ContentType string            `json:"contentType"`
		Size        string            `json:"size"`
		Updated     time.Time         `json:"updated"`
		Metadata    map[string]string `json:"metadata"`
	}

	if err := json.Unmarshal(msg.Data, &gcsObj); err != nil {
		return fmt.Errorf("failed to unmarshal GCS object: %w", err)
	}

	// Extract IDs and Metadata
	// The object name is expected to start with "{imageID}" (UUID, 36 chars) followed by separator
	if len(gcsObj.Name) < 37 {
		return fmt.Errorf("object name too short: %s", gcsObj.Name)
	}

	imageID := gcsObj.Name[:36]

	// Use Message ID for deduplication to avoid blocking other files for the same image
	cacheKey := fmt.Sprintf("gcs:%s:%s", s.subscriptionID, msg.ID)
	exists, err := s.cache.Has(ctx, cacheKey)
	if err != nil {
		s.logger.Error("Failed to check cache", "error", err, "message_id", msg.ID)
		return fmt.Errorf("failed to check cache for message %s: %w", msg.ID, err)
	}
	if exists {
		return nil
	}

	if err := s.cache.Set(ctx, cacheKey, true, 2*time.Hour); err != nil {
		s.logger.Error("Failed to set cache", "error", err, "message_id", msg.ID)
		return fmt.Errorf("failed to set cache for message %s: %w", msg.ID, err)
	}

	// Check separator (can be - or /)
	separator := gcsObj.Name[36]
	if separator != '-' && separator != '/' {
		return fmt.Errorf("unexpected separator in object name: %s", gcsObj.Name)
	}

	fileName := gcsObj.Name[37:]

	// Validate UUID
	if _, err := uuid.Parse(imageID); err != nil {
		return fmt.Errorf("invalid UUID: %s", imageID)
	}

	size, _ := strconv.ParseInt(gcsObj.Size, 10, 64)

	// Map metadata
	var entityType, provider, targetContentType string
	if gcsObj.Metadata != nil {
		entityType = gcsObj.Metadata["entity-type"]
		provider = gcsObj.Metadata["provider"]
		targetContentType = gcsObj.Metadata["content-type"]
	}

	if targetContentType == "" {
		targetContentType = gcsObj.ContentType
	}

	contentID := uuid.New().String()
	if gcsObj.Metadata != nil && gcsObj.Metadata["id"] != "" {
		contentID = gcsObj.Metadata["id"]
	}

	content := model.Content{
		Entity: vobj.Entity{
			ID:        contentID,
			Name:      fileName,
			CreatedAt: gcsObj.Updated,
			UpdatedAt: gcsObj.Updated,
			Parent: vobj.ParentRef{
				ID:   imageID,
				Type: vobj.ParentTypeImage,
			},
			EntityType: vobj.EntityType(entityType),
			CreatorID:  gcsObj.Metadata["creator-id"],
		},
		Provider:      vobj.ContentProvider(provider),
		Path:          gcsObj.Name,
		ContentType:   vobj.ContentType(targetContentType),
		Size:          size,
		UploadPending: false,
	}

	if content.EntityType == "" {
		content.EntityType = vobj.EntityTypeContent
	}

	// Default provider if missing
	if content.Provider == "" {
		content.Provider = vobj.ContentProviderGCS
	}

	event := &domainevent.NewFileExistEvent{
		BaseEvent: domainevent.BaseEvent{
			EventID:   msg.ID,
			EventType: domainevent.NewFileExistEventType,
			Timestamp: time.Now(),
		},
		Content: content,
	}

	return handler.Handle(ctx, event)
}

func (s *PubSubSubscriber) Stop() error {
	return s.client.Close()
}
