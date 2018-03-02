package models

// Test coverage: 100% (without interface methods)

import (
	"time"

	"github.com/gobuffalo/pop"
	"github.com/gobuffalo/uuid"
	"github.com/gobuffalo/validate"
	"github.com/gobuffalo/validate/validators"
)

// Doc is a structure for documentations
type Doc struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	MemberID    uuid.UUID `json:"member_id" db:"member_id"`
	Type        string    `json:"type" db:"type"`
	Category    string    `json:"category" db:"category"`
	Subject     string    `json:"subject" db:"subject"`
	Slug        string    `json:"slug" db:"slug"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	IsPublished bool      `json:"is_published" db:"is_published"`
	NewCategory string    `json:"new_category" db:"-"`
	NewSubject  string    `json:"new_subject" db:"-"`
}

//** rendering helpers for templates --------------------------------

// String returns title of the document
func (d Doc) String() string {
	return d.Title
}

// AuthorName returns name of the author associated to the doc.
func (d Doc) AuthorName() string {
	return GetMember(d.MemberID).Name
}

//** relational accessor and functions ------------------------------

//** implementations for interfaces ---------------------------------

// QueryParams implements Belonging interface
func (d *Doc) QueryParams() QueryParams {
	return QueryParams{}
}

// OwnedBy implements Belonging interface
func (d *Doc) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

// AccessibleBy implements Belonging interface
func (d *Doc) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q.BelongsTo(o)
}

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// Docs is an array of Docs
type Docs []Doc

// Validate gets run every time you call a "pop.Validate" method.
func (d *Doc) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: d.Type, Name: "Type"},
		&validators.StringIsPresent{Field: d.Category, Name: "Category"},
		&validators.StringIsPresent{Field: d.Subject, Name: "Subject"},
		&validators.StringIsPresent{Field: d.Slug, Name: "Slug"},
		&validators.StringIsPresent{Field: d.Title, Name: "Title"},
		&validators.StringIsPresent{Field: d.Content, Name: "Content"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (d *Doc) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (d *Doc) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
