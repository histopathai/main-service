package model

import "time"

type ImageStatus string

const (
	StatusUploaded   ImageStatus = "UPLOADED"
	StatusProcessing ImageStatus = "PROCESSING"
	StatusProcessed  ImageStatus = "PROCESSED"
	StatusFailed     ImageStatus = "FAILED"
)

type Image struct {
	ID            string
	PatientID     string
	CreatorID     string
	FileName      string
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

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (i *Image) IsRetryable(maxRetries int) bool {
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
