package events

const (
	EventTypeImageUploaded EventType = "image.uploaded.v1"
)

type ImageUploadedEvent struct {
	BaseEvent
	Name     string                `json:"name"`
	Bucket   string                `json:"bucket"`
	Metadata ImageUploadedMetadata `json:"metadata"`
}

type ImageUploadedMetadata struct {
	ImageID    string `json:"image-id"`
	EntityType string `json:"entity-type"`
	Name       string `json:"name"`
	ParentID   string `json:"parent-id"`
	CreatorID  string `json:"creator-id"`

	// Image specific metadata
	Format string `json:"format"`
	Width  *int   `json:"width,omitempty,string"`
	Height *int   `json:"height,omitempty,string"`
	Size   *int64 `json:"size,omitempty,string"`

	OriginPath string `json:"origin-path"`
	Status     string `json:"status"`
}

func NewImageUploadedEvent(
	imageID, patientID, creatorID, name, format, originPath, status, bucket string,
	width *int,
	height *int,
	size *int64,
) ImageUploadedEvent {

	return ImageUploadedEvent{

		BaseEvent: NewBaseEvent(EventTypeImageUploaded),
		Name:      originPath,
		Bucket:    bucket,
		Metadata: ImageUploadedMetadata{
			ImageID:    imageID,
			ParentID:   patientID,
			CreatorID:  creatorID,
			Name:       name,
			Format:     format,
			Width:      width,
			Height:     height,
			Size:       size,
			OriginPath: originPath,
			Status:     status,
		},
	}
}

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

const (
	EventTypeImageProcessingRequested EventType = "image.processing.requested.v1"
	EventTypeImageProcessingCompleted EventType = "image.processing.result.v1"
)

type ImageProcessingRequestedEvent struct {
	BaseEvent
	ImageID    string `json:"image-id"`
	OriginPath string `json:"origin-path"`
}

type ImageProcessingResultEvent struct {
	BaseEvent
	ImageID       string  `json:"image-id"`
	Success       bool    `json:"success"`
	ProcessedPath *string `json:"processed-path,omitempty"`
	Width         *int    `json:"width,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Size          *int64  `json:"size,omitempty"`
	FailureReason *string `json:"failure-reason,omitempty"`
	Retryable     *bool   `json:"retryable,omitempty"`
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

func NewImageProcessingSuccessEvent(imageID, processedPath string, width int, height int, size int64) ImageProcessingResultEvent {
	return ImageProcessingResultEvent{
		BaseEvent:     NewBaseEvent(EventTypeImageProcessingCompleted),
		ImageID:       imageID,
		Success:       true,
		ProcessedPath: &processedPath,
		Width:         &width,
		Height:        &height,
		Size:          &size,
	}
}

func NewImageProcessingFailureEvent(imageID, failureReason string, retryable bool) ImageProcessingResultEvent {
	return ImageProcessingResultEvent{
		BaseEvent:     NewBaseEvent(EventTypeImageProcessingCompleted),
		ImageID:       imageID,
		Success:       false,
		FailureReason: &failureReason,
		Retryable:     &retryable,
	}
}
