package publishers

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/open-outbox/relay/internal/relay"
)

// Nats is a publisher implementation that sends events to a NATS server.
type Nats struct {
	conn *nats.Conn
}

// NewNats establishes a connection to a NATS server.
func NewNats(url string) (*Nats, error) {
	nc, err := nats.Connect(url, nats.Name("OpenOutbox-Relay"))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}
	return &Nats{conn: nc}, nil
}

// Publish sends the event payload to a NATS subject.
func (n *Nats) Publish(ctx context.Context, event relay.Event) error {

	if n.conn.IsClosed() {
		return &relay.PublishError{
			Err:         errors.New("nats: connection closed permanently"),
			IsRetryable: true,
			Code:        "NATS_FATAL_CLOSED",
		}
	}

	err := n.conn.Publish(event.Type, event.Payload)

	if err != nil {
		return &relay.PublishError{
			Err:         fmt.Errorf("nats flush failed: %w", err),
			IsRetryable: isNatsErrorRetryable(err),
			Code:        "NATS_FLUSH_ERROR",
		}
	}

	flushCtx := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		// If not, we MUST create one, otherwise NATS returns the error you're seeing.
		var cancel context.CancelFunc
		flushCtx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	if err := n.conn.FlushWithContext(flushCtx); err != nil {
		return &relay.PublishError{
			Err:         fmt.Errorf("nats flush failed: %w", err),
			IsRetryable: isNatsErrorRetryable(err),
			Code:        "NATS_FLUSH_ERROR",
		}
	}
	return nil
}

// Close gracefully shuts down the NATS connection.
// It drains any remaining buffered messages and closes the underlying socket.
func (n *Nats) Close() error {
	if n.conn != nil {
		// Option: You could call n.conn.Drain() here if you wanted
		// to ensure all pending publishes are flushed before closing.
		n.conn.Close()
	}
	return nil
}

func isNatsErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	if isContextError(err) {
		return true
	}

	switch {
	// Group 1: Connectivity & Cluster Health
	case errors.Is(err, nats.ErrConnectionClosed),
		errors.Is(err, nats.ErrConnectionReconnecting),
		errors.Is(err, nats.ErrDisconnected),
		errors.Is(err, nats.ErrNoServers),
		errors.Is(err, nats.ErrStaleConnection):
		return true

	// Group 2: Capacity & Performance
	case errors.Is(err, nats.ErrTimeout),
		errors.Is(err, nats.ErrMaxConnectionsExceeded),
		errors.Is(err, nats.ErrMaxAccountConnectionsExceeded),
		errors.Is(err, nats.ErrNoResponders):
		return true
	}

	// Underlying network/socket issues are always worth a retry
	var netErr net.Error
	if errors.As(err, &netErr) {
		return true
	}

	return false
}
