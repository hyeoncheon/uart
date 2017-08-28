package models

// Test coverage: 100% (without interface methods)

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// MessengerPriority is a map for priority string to code.
var MessengerPriority = map[string]int{
	"Alert":        1,
	"Notification": 5,
	"Disabled":     8,
}

// MessengerMethod is a map for method name to code string.
var MessengerMethod = map[string]string{
	"Email": "mail",
}

// MessengerMethodReverse is a reverse map for method code to name.
var MessengerMethodReverse = map[string]string{
	"mail": "Email",
}

const (
	messengersDefaultSort = "priority, is_primary desc, created_at desc"
)

// Messenger is a structure for messaging methods
type Messenger struct {
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

// String returns json marshalled representation of messenger
func (m Messenger) String() string {
	attr := fmt.Sprintf(" (%v/%v)", MsgPriReverse[m.Priority], m.IsPrimary)
	return MessengerMethodReverse[m.Method] + " to " + m.Value + attr
}

//** implementations for interfaces ---------------------------------

// QueryParams implements Belonging interface
func (m *Messenger) QueryParams() QueryParams {
	return QueryParams{}
}

// OwnedBy implements Belonging interface
func (m *Messenger) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

// AccessibleBy implements Belonging interface
func (m *Messenger) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// Messengers is an array of Messengers
type Messengers []Messenger

// String is not required by pop and may be deleted
func (m Messengers) String() string {
	jm, _ := json.Marshal(m)
	return string(jm)
}

// Validate gets run every time you call a "pop.Validate" method.
func (m *Messenger) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.IntIsPresent{Field: m.Priority, Name: "Priority"},
		&validators.StringIsPresent{Field: m.Method, Name: "Method"},
		&validators.StringIsPresent{Field: m.Value, Name: "Value"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (m *Messenger) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (m *Messenger) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
