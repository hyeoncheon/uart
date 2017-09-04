package models

// Test coverage: 100% (Nothing to test)

import (
	"time"

	"github.com/satori/go.uuid"
)

// MessagingLog is a structure for log of messaging.
type MessagingLog struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	Status    string    `json:"status" db:"status"`
	QueueID   string    `json:"queue_id" db:"queue_id"`
	Response  string    `json:"response" db:"response"`
	Method    string    `json:"method" db:"method"`
	SentFor   string    `json:"sent_for" db:"sent_for"`
	SentTo    string    `json:"sent_to" db:"sent_to"`
	Subject   string    `json:"subject" db:"subject"`
	Notes     string    `json:"notes" db:"notes"`
}

//** rendering helpers for templates --------------------------------

// String returns representation of the log
func (m MessagingLog) String() string {
	return m.Subject + " sent for " + m.SentFor
}

//** implementations for interfaces ---------------------------------

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// MessagingLogs is an array of Messages
type MessagingLogs []MessagingLog
