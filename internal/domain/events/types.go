package events

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

const (
	EventTypeImageUploaded            EventType = "image.uploaded.v1"
	EventTypeImageProcessingRequested EventType = "image.processing.requested.v1"
	EventTypeImageProcessingCompleted EventType = "image.processing.completed.v1"
	EventTypeImageProcessingFailed    EventType = "image.processing.failed.v1"
)

type BaseEvent struct {
	EventID   string
	EventType EventType
	Timestamp time.Time
}

func NewBaseEvent(eventType EventType) BaseEvent {
	return BaseEvent{
		EventID:   uuid.New().String(),
		EventType: eventType,
		Timestamp: time.Now(),
	}
}

type ImageUploadedEvent struct {
	// Unique identifier for the event
	BaseEvent
	// Details about the uploaded image
	ImageID    string `json:"image-id"`
	PatientID  string `json:"patient-id"`
	CreatorID  string `json:"creator-id"`
	FileName   string `json:"file-name"`
	Format     string `json:"format"`
	Width      *int   `json:"width,omitempty"`
	Height     *int   `json:"height,omitempty"`
	Size       *int64 `json:"size,omitempty"`
	OriginPath string `json:"origin-path"`
	Status     string `json:"status"`
}

func NewImageUploadedEvent(
	imageID, patientID, creatorID, fileName, format string,
	width *int, height *int, size *int64,
	originPath, status string,
) ImageUploadedEvent {
	return ImageUploadedEvent{
		BaseEvent:  NewBaseEvent(EventTypeImageUploaded),
		ImageID:    imageID,
		PatientID:  patientID,
		CreatorID:  creatorID,
		FileName:   fileName,
		Format:     format,
		Width:      width,
		Height:     height,
		Size:       size,
		OriginPath: originPath,
		Status:     status,
	}
}

type ImageProcessingRequestedEvent struct {
	BaseEvent
	// Details about the image to be processed
	ImageID    string `json:"image-id"`
	OriginPath string `json:"origin-path"`
}

func NewImageProcessingRequestedEvent(imageID, originPath string) ImageProcessingRequestedEvent {
	return ImageProcessingRequestedEvent{
		BaseEvent:  NewBaseEvent(EventTypeImageProcessingRequested),
		ImageID:    imageID,
		OriginPath: originPath,
	}
}

type ImageProcessingFailedEvent struct {
	// Unique identifier for the event
	BaseEvent
	// Details about the failed image processing
	ImageID       string `json:"image-id"`
	FailureReason string `json:"failure-reason"`
}

func NewImageProcessingFailedEvent(imageID, failureReason string) ImageProcessingFailedEvent {
	return ImageProcessingFailedEvent{
		BaseEvent:     NewBaseEvent(EventTypeImageProcessingFailed),
		ImageID:       imageID,
		FailureReason: failureReason,
	}
}

type Event interface {
	GetEventID() string
	GetEventType() EventType
	GetTimestamp() time.Time
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
