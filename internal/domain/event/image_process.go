package event

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type ImageProcessEvent struct {
	BaseEvent
	model.Content

	ProcessingVersion vobj.ProcessingVersion
}

type ProcessResult struct {
	Width  int
	Height int
	Size   int64
}

type ImageProcessCompleteEvent struct {
	BaseEvent
	ImageID           string
	ProcessingVersion vobj.ProcessingVersion
	Contents          []model.Content

	Success       bool
	Result        *ProcessResult
	FailureReason string
	Retryable     bool

	// Retry tracking
	RetryMetadata *RetryMetadata
}

type ImageProcessDlqEvent struct {
	BaseEvent
	ImageID           string
	Content           model.Content
	ProcessingVersion vobj.ProcessingVersion
	FailureReason     string
	Retryable         bool
	RetryMetadata     *RetryMetadata // Track all retry attempts
	OriginalEventID   string         // Reference to original event
}
