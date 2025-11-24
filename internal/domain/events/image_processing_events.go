package events

const (
	EventTypeImageProcessingRequested EventType = "image.processing.requested.v1"
	EventTypeImageProcessingCompleted EventType = "image.processing.completed.v1"
	EventTypeImageProcessingFailed    EventType = "image.processing.failed.v1"
)

type ImageProcessingRequestedEvent struct {
	BaseEvent
	ImageID    string `json:"image-id"`
	OriginPath string `json:"origin-path"`
}

type ImageProcessingCompletedEvent struct {
	BaseEvent
	ImageID       string `json:"image-id"`
	ProcessedPath string `json:"processed-path"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Size          int64  `json:"size"`
}

type ImageProcessingFailedEvent struct {
	BaseEvent
	ImageID       string `json:"image-id"`
	Retryable     bool   `json:"retryable"`
	FailureReason string `json:"failure-reason"`
}

type ImageProcessingDLQEvent struct {
	BaseEvent
	ImageID       string `json:"image-id"`
	FailureReason string `json:"failure-reason"`
}

func NewImageProcessingRequestedEvent(imageID, originPath string) ImageProcessingRequestedEvent {
	return ImageProcessingRequestedEvent{
		BaseEvent:  NewBaseEvent(EventTypeImageProcessingRequested),
		ImageID:    imageID,
		OriginPath: originPath,
	}
}

func NewImageProcessingCompletedEvent(imageID, processedPath string, width int, height int, size int64) ImageProcessingCompletedEvent {
	return ImageProcessingCompletedEvent{
		BaseEvent:     NewBaseEvent(EventTypeImageProcessingCompleted),
		ImageID:       imageID,
		ProcessedPath: processedPath,
		Width:         width,
		Height:        height,
		Size:          size,
	}
}

func NewImageProcessingFailedEvent(imageID, failureReason string) ImageProcessingFailedEvent {
	return ImageProcessingFailedEvent{
		BaseEvent:     NewBaseEvent(EventTypeImageProcessingFailed),
		ImageID:       imageID,
		FailureReason: failureReason,
	}
}
