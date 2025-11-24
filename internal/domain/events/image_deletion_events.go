package events

const (
	EventTypeImageDeletionRequested EventType = "image.deletion.requested.v1"
)

type ImageDeletionRequestedEvent struct {
	BaseEvent
	ImageID string `json:"image-id"`
}

func NewImageDeletionRequestedEvent(imageID string) ImageDeletionRequestedEvent {
	return ImageDeletionRequestedEvent{
		BaseEvent: NewBaseEvent(EventTypeImageDeletionRequested),
		ImageID:   imageID,
	}
}
