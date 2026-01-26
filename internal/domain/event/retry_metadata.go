package event

import "time"

// RetryMetadata tracks retry attempts for events
type RetryMetadata struct {
	AttemptCount    int           // Current retry attempt (0 = first attempt)
	MaxAttempts     int           // Maximum allowed retries
	LastAttemptAt   time.Time     // Timestamp of last retry
	FirstAttemptAt  time.Time     // Timestamp of first attempt
	BackoffDuration time.Duration // Current backoff duration
}

// ShouldRetry determines if event should be retried
func (r *RetryMetadata) ShouldRetry() bool {
	return r.AttemptCount < r.MaxAttempts
}

// IncrementAttempt increases retry counter and updates backoff
func (r *RetryMetadata) IncrementAttempt() {
	r.AttemptCount++
	r.LastAttemptAt = time.Now()

	if r.FirstAttemptAt.IsZero() {
		r.FirstAttemptAt = time.Now()
	}

	// Exponential backoff: 2^attempt * base (e.g., 1s, 2s, 4s, 8s...)
	r.BackoffDuration = time.Duration(1<<uint(r.AttemptCount)) * time.Second
}
