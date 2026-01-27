package pubsub

import (
	"time"

	domainevent "github.com/histopathai/main-service/internal/domain/event"
)

type RetryPolicy struct {
	MaxAttempts       int
	BaseBackoff       time.Duration
	MaxBackoff        time.Duration
	BackoffMultiplier float64
}

var DefaultRetryPolicies = map[domainevent.EventType]RetryPolicy{
	domainevent.ImageProcessCompleteEventType: {
		MaxAttempts:       5,
		BaseBackoff:       1 * time.Second,
		MaxBackoff:        60 * time.Second,
		BackoffMultiplier: 2.0,
	},
	domainevent.ImageProcessReqEventType: {
		MaxAttempts:       3,
		BaseBackoff:       2 * time.Second,
		MaxBackoff:        30 * time.Second,
		BackoffMultiplier: 2.0,
	},
}

func (p *RetryPolicy) CalculateBackoff(attemptCount int) time.Duration {
	backoff := time.Duration(float64(p.BaseBackoff) *
		(p.BackoffMultiplier * float64(attemptCount)))

	if backoff > p.MaxBackoff {
		return p.MaxBackoff
	}
	return backoff
}
