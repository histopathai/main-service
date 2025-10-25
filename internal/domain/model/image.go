package model

import "time"

type ImageStatus string

const (
	StatusUploaded   ImageStatus = "UPLOADED"
	StatusProcessing ImageStatus = "PENDING"
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
	CreatedAt     time.Time
	UpdatedAt     time.Time
}
