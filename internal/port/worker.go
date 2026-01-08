package port

import "context"

type ProcessingInput struct {
	ImageID    string
	OriginPath string
	BucketName string
	Size       int64
}

type ProcessingResult struct {
	ImageID       string
	ProcessedPath string
	Width         int
	Height        int
	Size          int64
	Success       bool
	Error         string
}

type ImageProcessingWorker interface {
	ProcessImage(ctx context.Context, input *ProcessingInput) error
}
