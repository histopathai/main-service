package vobj

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeImageUploaded            EventType = "image.uploaded.v1"
	EventTypeImageProcessingRequested EventType = "image.processing.requested.v1"
	EventTypeImageProcessingCompleted EventType = "image.processing.result.v1"
	EventTypeImageDeletionRequested   EventType = "image.deletion.requested.v1"
)

func (et EventType) String() string {
	return string(et)
}

func (et EventType) IsValid() bool {
	switch et {
	case EventTypeImageUploaded,
		EventTypeImageProcessingRequested,
		EventTypeImageProcessingCompleted,
		EventTypeImageDeletionRequested:
		return true
	default:
		return false
	}
}

type Event struct {
	ID        string
	Type      EventType
	Timestamp int64
	Payload   []byte
}

func NewEvent(eventType EventType, payload []byte) *Event {
	id := uuid.New().String()
	timestamp := time.Now().UnixMilli()
	return &Event{
		ID:        id,
		Type:      eventType,
		Timestamp: timestamp,
		Payload:   payload,
	}
}

func (e *Event) GetID() string {
	return e.ID
}

func (e *Event) GetType() EventType {
	return e.Type
}

func (e *Event) GetTimestamp() int64 {
	return e.Timestamp
}
func (e *Event) GetPayload() []byte {
	return e.Payload
}

func (e *Event) SetPayload(payload []byte) {
	e.Payload = payload
}

func (e *Event) IsValid() bool {
	return e.Type.IsValid() && e.ID != "" && e.Timestamp > 0
}

func (e *Event) String() string {
	return string(e.Type)
}

func (e *Event) Equals(other *Event) bool {
	if other == nil {
		return false
	}
	return e.ID == other.ID &&
		e.Type == other.Type &&
		e.Timestamp == other.Timestamp
}

func (e *Event) Clone() *Event {
	return &Event{
		ID:        e.ID,
		Type:      e.Type,
		Timestamp: e.Timestamp,
		Payload:   append([]byte(nil), e.Payload...),
	}
}
