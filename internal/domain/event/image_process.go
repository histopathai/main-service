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
	ID                string
	ProcessingVersion vobj.ProcessingVersion
	Contents          []model.Content

	Success       bool
	Result        *ProcessResult
	FailureReason string
	Retryable     bool
}

type ImageProcessDlqEvent struct {
	BaseEvent
	ID            string
	Content       model.Content
	FailureReason string
	Retryable     bool
}
