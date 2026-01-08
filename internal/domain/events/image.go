package events

import "github.com/histopathai/main-service/internal/domain/vobj"

type ImageUploadedPayload struct {
	ImageID       string         `json:"image_id"`
	Name          string         `json:"name"`
	Bucket        string         `json:"bucket"`
	CreatorID     string         `json:"creator_id"`
	Parent        vobj.ParentRef `json:"parent"`
	Format        string         `json:"format"`
	Width         *int           `json:"width,omitempty"`
	Height        *int           `json:"height,omitempty"`
	Size          *int64         `json:"size,omitempty"`
	OriginPath    string         `json:"origin_path"`
	ProcessedPath *string        `json:"processed_path,omitempty"`
	IsProcessed   bool           `json:"is_processed"`
}

func NewImageUploadedEvent(
	imageID, creatorID, name, format, originPath, status, bucket string,
	parent vobj.ParentRef,
	width *int,
	height *int,
	size *int64,
	processedPath *string,
	isProcessed bool,
) (*vobj.Event, error) {
	payload := ImageUploadedPayload{
		ImageID:       imageID,
		CreatorID:     creatorID,
		Name:          name,
		Bucket:        bucket,
		Parent:        parent,
		Format:        format,
		Width:         width,
		Height:        height,
		Size:          size,
		OriginPath:    originPath,
		ProcessedPath: processedPath,
		IsProcessed:   isProcessed,
	}

	return vobj.NewEvent(vobj.EventTypeImageUploaded, payload)
}

type ImageDeletionRequestPayload struct {
	Targets map[string][]string `json:"targets"`
}

func NewImageDeletionRequestEvent(
	originBucket string, originPath string,
	processedBucket string, processedPath string,
) (*vobj.Event, error) {

	targets := make(map[string][]string)

	if originBucket != "" && originPath != "" {
		targets[originBucket] = append(targets[originBucket], originPath)
	}

	if processedBucket != "" && processedPath != "" {
		targets[processedBucket] = append(targets[processedBucket], processedPath)
	}

	if len(targets) == 0 {
		return nil, nil
	}

	payload := ImageDeletionRequestPayload{
		Targets: targets,
	}

	return vobj.NewEvent(vobj.EventTypeImageDeletionRequested, payload)
}

type ImageProcessingRequestedPayload struct {
	ImageID    string `json:"image_id"`
	OriginPath string `json:"origin_path"`
	BucketName string `json:"bucket_name"`
	Size       int64  `json:"size"`
}

func NewImageProcessingRequestedEvent(imageID, originPath string) (*vobj.Event, error) {
	payload := ImageProcessingRequestedPayload{
		ImageID:    imageID,
		OriginPath: originPath,
	}
	return vobj.NewEvent(vobj.EventTypeImageProcessingRequested, payload)
}

type ImageProcessingResultPayload struct {
	ImageID       string  `json:"image_id"`
	Success       bool    `json:"success"`
	ProcessedPath *string `json:"processed_path,omitempty"`
	Width         *int    `json:"width,omitempty"`
	Height        *int    `json:"height,omitempty"`
	Size          *int64  `json:"size,omitempty"`
	FailureReason *string `json:"failure_reason,omitempty"`
	Retryable     *bool   `json:"retryable,omitempty"`
}

func NewImageProcessingSuccessEvent(
	imageID string,
	processedPath string,
	width, height *int,
	size *int64,
) (*vobj.Event, error) {
	payload := ImageProcessingResultPayload{
		ImageID:       imageID,
		Success:       true,
		ProcessedPath: &processedPath,
		Width:         width,
		Height:        height,
		Size:          size,
	}
	return vobj.NewEvent(vobj.EventTypeImageProcessingCompleted, payload)
}

func NewImageProcessingFailureEvent(
	imageID, failureReason string,
	retryable bool,
) (*vobj.Event, error) {
	payload := ImageProcessingResultPayload{
		ImageID:       imageID,
		Success:       false,
		FailureReason: &failureReason,
		Retryable:     &retryable,
	}
	return vobj.NewEvent(vobj.EventTypeImageProcessingCompleted, payload)
}
