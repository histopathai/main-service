package vobj

import (
	"fmt"
)

type ImageStatus string

const (
	StatusPending    ImageStatus = "PENDING"    // Initial state, waiting for processing
	StatusProcessing ImageStatus = "PROCESSING" // Currently being processed
	StatusProcessed  ImageStatus = "PROCESSED"  // Successfully processed
	StatusFailed     ImageStatus = "FAILED"     // Processing failed
	StatusDeleting   ImageStatus = "DELETING"   // Marked for deletion
)

func (is ImageStatus) String() string {
	return string(is)
}

func (is ImageStatus) IsValid() bool {
	switch is {
	case StatusPending, StatusProcessing, StatusProcessed, StatusFailed, StatusDeleting:
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
