package vobj

import (
	"fmt"
	"time"
)

type ImageStatus string

func (s ImageStatus) String() string {
	return string(s)
}

func NewImageStatusFromString(statusStr string) (ImageStatus, error) {
	switch statusStr {
	case string(StatusUploaded):
		return StatusUploaded, nil
	case string(StatusProcessing):
		return StatusProcessing, nil
	case string(StatusProcessed):
		return StatusProcessed, nil
	case string(StatusFailed):
		return StatusFailed, nil
	case string(StatusDeleting):
		return StatusDeleting, nil
	default:
		return "", fmt.Errorf("invalid image status: %s", statusStr)
	}
}

const (
	StatusUploaded   ImageStatus = "UPLOADED"
	StatusProcessing ImageStatus = "PROCESSING"
	StatusProcessed  ImageStatus = "PROCESSED"
	StatusFailed     ImageStatus = "FAILED"
	StatusDeleting   ImageStatus = "DELETING"
)

type ImageProcessReport struct {
	Status          ImageStatus
	FailureReason   *string
	RetryCount      int
	LastProcessedAt *time.Time
}

func (r *ImageProcessReport) IsRetryable(maxRetries int) bool {
	// Prevent retry if status is DELETING
	if r.Status == StatusDeleting {
		return false
	}
	if r.Status == StatusFailed && r.RetryCount < maxRetries {
		return true
	}
	return false
}

func (r *ImageProcessReport) MarkForRetry() {
	r.Status = StatusProcessing
	r.RetryCount++
	now := time.Now()
	r.LastProcessedAt = &now
}

func (r *ImageProcessReport) MarkAsProcessed(processedPath string) {
	r.Status = StatusProcessed
	now := time.Now()
	r.LastProcessedAt = &now
}

func (r *ImageProcessReport) MarkAsFailed(reason string) {
	r.Status = StatusFailed
	r.FailureReason = &reason
	now := time.Now()
	r.LastProcessedAt = &now
}

func (r *ImageProcessReport) MarkAsDeleting() {
	r.Status = StatusDeleting
	now := time.Now()
	r.LastProcessedAt = &now
}
