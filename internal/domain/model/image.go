package model

import (
	"fmt"
	"time"

	"github.com/histopathai/main-service/internal/domain/vobj"
)

type ImageStatus string

func (is ImageStatus) String() string {
	return string(is)
}

func (is ImageStatus) IsValid() bool {
	switch is {
	case StatusUploaded, StatusProcessing, StatusProcessed, StatusFailed, StatusDeleting:
		return true
	default:
		return false
	}
}

func NewImageStatusFromString(s string) (ImageStatus, error) {
	is := ImageStatus(s)
	if !is.IsValid() {
		return "", fmt.Errorf("invalid ImageStatus: %s", s)
	}
	return is, nil
}

const (
	StatusUploaded   ImageStatus = "UPLOADED"
	StatusProcessing ImageStatus = "PROCESSING"
	StatusProcessed  ImageStatus = "PROCESSED"
	StatusFailed     ImageStatus = "FAILED"
	StatusDeleting   ImageStatus = "DELETING" // Added
)

type Image struct {
	vobj.Entity
	WsID          string
	ContentType   string
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
