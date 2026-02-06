package event

import "time"

type Event interface {
	GetEventID() string
	GetEventType() EventType
	GetTimestamp() time.Time
}

type BaseEvent struct {
	EventID   string
	EventType EventType
	Timestamp time.Time
}

func (e BaseEvent) GetEventID() string {
	return e.EventID
}

func (e BaseEvent) GetEventType() EventType {
	return e.EventType
}

func (e BaseEvent) GetTimestamp() time.Time {
	return e.Timestamp
}
