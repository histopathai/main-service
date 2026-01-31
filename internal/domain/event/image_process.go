package event

import (
	"github.com/histopathai/main-service/internal/domain/model"
	"github.com/histopathai/main-service/internal/domain/vobj"
)

type ImageProcessReqEvent struct {
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
}
