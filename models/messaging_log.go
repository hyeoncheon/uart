package models

//! WIP

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
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
	jm, _ := json.Marshal(m)
	return string(jm)
}

//** implementations for interfaces ---------------------------------

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// MessagingLogs is an array of Messages
type MessagingLogs []MessagingLog

// String returns json marshalled representation of the logs
func (m MessagingLogs) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate" method.
func (m *MessagingLog) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: m.Status, Name: "Status"},
		&validators.StringIsPresent{Field: m.QueueID, Name: "QueueID"},
		&validators.StringIsPresent{Field: m.Response, Name: "Response"},
		&validators.StringIsPresent{Field: m.SentTo, Name: "SentTo"},
		&validators.StringIsPresent{Field: m.Subject, Name: "Subject"},
		&validators.StringIsPresent{Field: m.Notes, Name: "Notes"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (m *MessagingLog) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (m *MessagingLog) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
