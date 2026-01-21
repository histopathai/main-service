package event

import "github.com/histopathai/main-service/internal/domain/vobj"

type ImageProcessingRequestedEvent struct {
	BaseEvent
	ID                string
	Content           vobj.Content
	ProcessingVersion vobj.ProcessingVersion
}

type ProcessingResult struct {
	Width  int
	Height int
	Size   int64
}

type ImageProcessingCompletedEvent struct {
	BaseEvent
	ID      string
	Content vobj.Content

	Success bool
	Result  *ProcessingResult

	FailureReason string
	Retryable     bool
}
