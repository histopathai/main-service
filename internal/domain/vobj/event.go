package vobj

import (
	"encoding/json"
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
	ID        string    `json:"id"`
	Type      EventType `json:"type"`
	Timestamp int64     `json:"timestamp"`
	Payload   []byte    `json:"payload"`
}

func NewEvent(eventType EventType, payload interface{}) (*Event, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Timestamp: time.Now().Unix(),
		Payload:   data,
	}, nil
}

func (e *Event) GetID() string {
	return e.ID
}

func (e *Event) GetType() EventType {
	return e.Type
}

func (e *Event) GetTimestamp() time.Time {
	return time.Unix(e.Timestamp, 0)
}

func (e *Event) IsValid() bool {
	return e.Type.IsValid() && e.ID != "" && e.Timestamp > 0
}

func (e *Event) UnmarshalPayload(v interface{}) error {
	return json.Unmarshal(e.Payload, v)
}
