package events

import (
	"fmt"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type EventRouter struct {
	topicMap map[vobj.EventType]string
}

func NewEventRouter() *EventRouter {
	return &EventRouter{
		topicMap: map[vobj.EventType]string{
			vobj.EventTypeImageUploaded:            "image-uploaded",
			vobj.EventTypeImageProcessingRequested: "image-processing-requested",
			vobj.EventTypeImageProcessingCompleted: "image-processing-completed",
			vobj.EventTypeImageDeletionRequested:   "image-deletion-requested",
		},
	}
}

func (r *EventRouter) GetTopic(eventType vobj.EventType) (string, error) {
	topic, ok := r.topicMap[eventType]
	if !ok {
		return "", fmt.Errorf("no topic mapped for event type: %s", eventType)
	}
	return topic, nil
}

func (r *EventRouter) RegisterTopic(eventType vobj.EventType, topicID string) {
	r.topicMap[eventType] = topicID
}
