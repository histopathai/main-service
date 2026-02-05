package vobj

import (
	"time"
)

type ProcessingVersion string

const (
	ProcessingV1 ProcessingVersion = "v1"
	ProcessingV2 ProcessingVersion = "v2"
)

func (pv ProcessingVersion) String() string {
	return string(pv)
}

func (pv ProcessingVersion) IsValid() bool {
	switch pv {
	case ProcessingV1, ProcessingV2:
		return true
	default:
		return false
	}
}

type ProcessingInfo struct {
	Status          ImageStatus
	Version         ProcessingVersion
	FailureReason   *string
	RetryCount      int
	LastProcessedAt time.Time
	ActiveEventID   string
}

func (pi *ProcessingInfo) IsRetryable(maxRetries int) bool {
	if pi.Status == StatusDeleting {
		return false
	}
	if pi.Status == StatusFailed && pi.RetryCount < maxRetries {
		return true
	}
	return false
}

func (pi *ProcessingInfo) MarkForRetry() {
	pi.Status = StatusProcessing
	pi.RetryCount++
	now := time.Now()
	pi.LastProcessedAt = now
}

func (pi *ProcessingInfo) MarkAsProcessing() {
	pi.Status = StatusProcessing
	now := time.Now()
	pi.LastProcessedAt = now
}

func (pi *ProcessingInfo) MarkAsProcessed(version ProcessingVersion) {
	pi.Status = StatusProcessed
	pi.Version = version
	now := time.Now()
	pi.LastProcessedAt = now
	pi.FailureReason = nil
}

func (pi *ProcessingInfo) MarkAsFailed(reason string) {
	pi.Status = StatusFailed
	pi.FailureReason = &reason
	now := time.Now()
	pi.LastProcessedAt = now
}
