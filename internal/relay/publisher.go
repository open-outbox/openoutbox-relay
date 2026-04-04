package relay

import (
	"context"
)

// PublishError defines how the Engine should react to a failure.
type PublishError struct {
	Err         error
	IsRetryable bool   // Crucial: Tells the Engine whether to try again or DLQ.
	Code        string // e.g., "BROKER_NACK", "AUTH_EXPIRED", "VALIDATION_ERROR"
}

func (e *PublishError) Error() string { return e.Err.Error() }
func (e *PublishError) Unwrap() error { return e.Err }

// Publisher is the common interface for all egress transports.
type Publisher interface {
	// Publish blocks until the downstream system acknowledges receipt (ACK/NACK).
	Publish(ctx context.Context, event Event) error

	Close() error
}
