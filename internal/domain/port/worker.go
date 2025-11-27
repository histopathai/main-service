package port

import "context"

// ProcessingInput contains input for image processing
type ProcessingInput struct {
	ImageID    string
	OriginPath string
	BucketName string
}

// ProcessingResult contains the result of image processing
type ProcessingResult struct {
	ImageID       string
	ProcessedPath string
	Width         int
	Height        int
	Size          int64
	Success       bool
	Error         string
}

// ImageProcessingWorker defines the interface for image processing workers
type ImageProcessingWorker interface {
	// ProcessImage triggers async image processing
	ProcessImage(ctx context.Context, input *ProcessingInput) error
}
