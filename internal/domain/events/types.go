package events

import "time"

const (
	EventTypeImageUploaded            = "image.uploaded.v1"
	EventTypeImageProcessingRequested = "image.processing.requested.v1"
	EventTypeImageProcessingCompleted = "image.processing.completed.v1"
	EventTypeImageProcessingFailed    = "image.processing.failed.v1"
)

type ImageUploadedEvent struct {
	// Unique identifier for the event
	EventID   string
	EventType string
	Timestamp time.Time
	// Details about the uploaded image
	ImageID    string
	PatientID  string
	CreatorID  string
	FileName   string
	Format     string
	Width      *int
	Height     *int
	Size       *int64
	OriginPath string
	Status     string
}

type ImageProcessingRequestedEvent struct {
	// Unique identifier for the event
	EventID   string
	EventType string
	Timestamp time.Time
	// Details about the image to be processed
	ImageID string
}

type ImageProcessingCompletedEvent struct {
	// Unique identifier for the event
	EventID   string
	EventType string
	Timestamp time.Time
	// Details about the processed image
	ImageID       string
	ProcessedPath string
}
type ImageProcessingFailedEvent struct {
	// Unique identifier for the event
	EventID   string
	EventType string
	Timestamp time.Time
	// Details about the failed image processing
	ImageID       string
	FailureReason string
}
