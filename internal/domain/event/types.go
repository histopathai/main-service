package event

type EventType string

const (
	UploadedEventType                 EventType = "uploaded.v1"
	DeleteEventType                   EventType = "delete.v1"
	ImageProcessingRequestedEventType EventType = "image.processing.requested.v1"
	ImageProcessingCompletedEventType EventType = "image.processing.completed.v1"
)
