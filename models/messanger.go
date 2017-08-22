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

// MessangerPriority is a map for priority string to code.
var MessangerPriority = map[string]int{
	"Alert":        1,
	"Notification": 5,
	"Disabled":     8,
}

// MessangerMethod is a map for method name to code string.
var MessangerMethod = map[string]string{
	"Email": "mail",
}

const (
	messangersDefaultSort = "priority, is_primary desc, created_at desc"
)

// Messanger is a structure for messaging methods
type Messanger struct {
	ID        uuid.UUID `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
	Priority  int       `json:"priority" db:"priority"`
	Method    string    `json:"method" db:"method"`
	Value     string    `json:"value" db:"value"`
	IsPrimary bool      `json:"is_primary" db:"is_primary"`
}

//** rendering helpers for templates --------------------------------

// String returns json marshalled representation of messanger
func (m Messanger) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

//** implementations for interfaces ---------------------------------

// QueryParams implements Belonging interface
func (m *Messanger) QueryParams() QueryParams {
	return QueryParams{}
}

// OwnedBy implements Belonging interface
func (m *Messanger) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

// AccessibleBy implements Belonging interface
func (m *Messanger) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// Messangers is an array of Messangers
type Messangers []Messanger

// String is not required by pop and may be deleted
func (m Messangers) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate" method.
func (m *Messanger) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.Priority, Name: "Priority"},
		&validators.StringIsPresent{Field: m.Method, Name: "Method"},
		&validators.StringIsPresent{Field: m.Value, Name: "Value"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (m *Messanger) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (m *Messanger) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
