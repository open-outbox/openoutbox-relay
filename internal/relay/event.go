package relay

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusPending    Status = "PENDING"    // Ready to be picked up
	StatusDelivering Status = "DELIVERING" // Currently being processed (Locked)
	StatusDelivered  Status = "DELIVERED"  // Success!
	StatusDead       Status = "DEAD"       // Failed too many times
)

func (s Status) IsTerminal() bool {
	return s == StatusDelivered || s == StatusDead
}

type Event struct {
	ID           uuid.UUID      `db:"event_id"      json:"id"`
	Type         string         `db:"event_type"    json:"type"`
	PartitionKey string         `db:"partition_key" json:"partition_key"`
	Payload      []byte         `db:"payload"       json:"payload"`
	Headers      map[string]any `db:"headers"       json:"headers"`

	// Delivery status
	Attempts int `db:"attempts"   json:"attempts"`

	// Time fields
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type FailedEvent struct {
	ID          uuid.UUID `json:"id"           db:"id"`
	NewStatus   Status    `json:"new_status"   db:"status"`
	AvailableAt time.Time `json:"available_at" db:"available_at"`
	Attempts    int       `json:"attempts"     db:"attempts"`
	LastError   string    `json:"last_error"   db:"last_error"`
}
