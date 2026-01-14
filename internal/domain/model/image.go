package model

import (
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type ImageStatus string

const (
	StatusUploaded   ImageStatus = "UPLOADED"
	StatusProcessing ImageStatus = "PROCESSING"
	StatusProcessed  ImageStatus = "PROCESSED"
	StatusFailed     ImageStatus = "FAILED"
	StatusDeleting   ImageStatus = "DELETING" // Added
)

type Image struct {
	vobj.Entity
	contentType   string
	Format        string
	Width         *int
	Height        *int
	Size          *int64
	OriginPath    string
	ProcessedPath *string
	Status        ImageStatus

	FailureReason   *string
	RetryCount      int
	LastProcessedAt *time.Time
}

func (i *Image) IsRetryable(maxRetries int) bool {
	// Prevent retry if status is DELETING
	if i.Status == StatusDeleting {
		return false
	}
	if i.Status == StatusFailed && i.RetryCount < maxRetries {
		return true
	}
	return false
}

func (i *Image) MarkForRetry() {
	i.Status = StatusProcessing
	i.RetryCount++
	now := time.Now()
	i.LastProcessedAt = &now
}
