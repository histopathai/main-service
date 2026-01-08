package events

import "github.com/histopathai/main-service/internal/domain/vobj"

type ImageUploadedEvent struct {
	vobj.Event
	Name     string
	Bucket   string
	Metadata ImageUploadedMetadata
}

type ImageUploadedMetadata struct {
	imageID   string
	CreatorID string
	Parent    vobj.ParentRef
	Format    string
	Name      string

	Width         *int
	Height        *int
	Size          *int64
	OriginPath    string
	ProcessedPath *string
	Status        string
}

func NewImageUploadedEvent(
	imageID, creatorID, name, format, originPath, status, bucket string,
	parent vobj.ParentRef,
	width *int,
	height *int,
	size *int64,
	processedPath *string,
) *ImageUploadedEvent {
	metadata := ImageUploadedMetadata{
		imageID:       imageID,
		CreatorID:     creatorID,
		Parent:        parent,
		Format:        format,
		Name:          name,
		Width:         width,
		Height:        height,
		Size:          size,
		OriginPath:    originPath,
		ProcessedPath: processedPath,
		Status:        status,
	}

	return &ImageUploadedEvent{
		Event: vobj.Event{
			Type: vobj.EventTypeImageUploaded,
		},
		Name:     name,
		Bucket:   bucket,
		Metadata: metadata,
	}
}

type ImageDeletionRequestEvent struct {
	vobj.Event
	ImageID string
}

func NewImageDeletionRequestEvent(imageID string) *ImageDeletionRequestEvent {
	return &ImageDeletionRequestEvent{
		Event: vobj.Event{
			Type: vobj.EventTypeImageDeletionRequested,
		},
		ImageID: imageID,
	}
}

type ImageProcessingRequestedEvent struct {
	vobj.Event
	ImageID    string
	OriginPath string
}

func NewImageProcessingRequestedEvent(imageID, originPath string) *ImageProcessingRequestedEvent {
	return &ImageProcessingRequestedEvent{
		Event: vobj.Event{
			Type: vobj.EventTypeImageProcessingRequested,
		},
		ImageID:    imageID,
		OriginPath: originPath,
	}
}

type ImageProcessingResultEvent struct {
	vobj.Event
	ImageID       string
	Success       bool
	ProcessedPath *string
	Width         *int
	Height        *int
	Size          *int64
	FailureReason *string
	Retryable     *bool
}

type ImageProcessingDLQEvent struct {
	vobj.Event
	ImageID       string
	FailureReason string
}

func NewImageProcessingSuccessEvent(
	imageID string,
	processedPath string,
	width, height *int,
	size *int64,
) *ImageProcessingResultEvent {
	return &ImageProcessingResultEvent{
		Event: vobj.Event{
			Type: vobj.EventTypeImageProcessingCompleted,
		},
		ImageID:       imageID,
		Success:       true,
		ProcessedPath: &processedPath,
		Width:         width,
		Height:        height,
		Size:          size,
	}
}

func NewImageProcessingFailureEvent(
	imageID, failureReason string,
	retryable bool,
) *ImageProcessingResultEvent {
	return &ImageProcessingResultEvent{
		Event: vobj.Event{
			Type: vobj.EventTypeImageProcessingCompleted,
		},
		ImageID:       imageID,
		Success:       false,
		FailureReason: &failureReason,
		Retryable:     &retryable,
	}
}
