package event

import "time"

type RetryMetadata struct {
	AttemptCount    int
	MaxAttempts     int
	LastAttemptAt   time.Time
	FirstAttemptAt  time.Time
	BackoffDuration time.Duration
}

func (r *RetryMetadata) ShouldRetry() bool {
	return r.AttemptCount < r.MaxAttempts
}

func (r *RetryMetadata) IncrementAttempt() {
	r.AttemptCount++
	r.LastAttemptAt = time.Now()

	if r.FirstAttemptAt.IsZero() {
		r.FirstAttemptAt = time.Now()
	}

	// Exponential backoff: 2^attempt * base (e.g., 1s, 2s, 4s, 8s...)
	r.BackoffDuration = time.Duration(1<<uint(r.AttemptCount)) * time.Second
}
