package relay

import (
	"math"
	"math/rand"
	"time"
)

// RetryPolicy defines the contract for calculating backoff durations between
// event delivery attempts. It allows the relay engine to be flexible with
// different retry strategies (e.g., constant, linear, or exponential).
type RetryPolicy interface {
	// NextBackoff returns the duration to wait before the next attempt and
	// a boolean indicating whether the retry limit has been reached.
	NextBackoff(attempts int) (time.Duration, bool)
}

// ExponentialBackoff implements a binary exponential backoff strategy with
// randomized jitter. This is the recommended policy for high-throughput
// production systems to avoid overwhelming downstream brokers after a failure.
type ExponentialBackoff struct {
	// MaxAttempts is the hard limit for delivery attempts.
	MaxAttempts int
	// BaseDelay is the initial backoff duration (e.g., 1s).
	BaseDelay time.Duration
	// MaxDelay is the maximum duration any single backoff can reach.
	MaxDelay time.Duration
	// Jitter is a factor (0.0 to 1.0) used to randomize the backoff interval.
	Jitter float64
}

// NextBackoff calculates the next delay using the formula: BaseDelay * 2^(attempts-1).
// It ensures the delay does not exceed MaxDelay and applies a random jitter
// to stagger retries across multiple relay instances.
func (p ExponentialBackoff) NextBackoff(attempts int) (time.Duration, bool) {
	if attempts >= p.MaxAttempts {
		return 0, false
	}

	// Calculate Exponential Base
	// 2^(attempts-1) * BaseDelay
	exp := math.Pow(2, float64(attempts-1))
	delay := time.Duration(float64(p.BaseDelay) * exp)

	if delay > p.MaxDelay {
		delay = p.MaxDelay
	}

	jitterMax := int64(float64(delay) * p.Jitter)
	var jitter time.Duration
	if jitterMax > 0 {
		jitter = time.Duration(rand.Int63n(jitterMax))
	}

	return delay + jitter, true
}
