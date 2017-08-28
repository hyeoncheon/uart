package models

// TODO REVIEW REQUIRED
// Test coverage: 100% (without interface methods)

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/satori/go.uuid"
)

// common constants
const (
	DefaultSortCredentials = "created_at"
)

// Credential is the model for oauth2 information from 3rd party providers
type Credential struct {
	ID           uuid.UUID `json:"id" db:"id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
	MemberID     uuid.UUID `json:"member_id" db:"member_id"`
	Provider     string    `json:"provider" db:"provider"`
	UserID       string    `json:"user_id" db:"user_id"`
	Name         string    `json:"name" db:"name"`
	Email        string    `json:"email" db:"email"`
	AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
	IsAuthorized bool      `json:"is_authorized" db:"is_authorized"`
	IsPrimary    bool      `json:"is_primary" db:"is_primary"`
}

//** rendering helpers for templates --------------------------------

// String returns pretty printable string of this model.
func (c Credential) String() string {
	return c.Provider + "/" + c.UserID
}

//** actions, relational accessor and functions below:

// Owner find and return associated member instance
func (c Credential) Owner() *Member {
	member := &Member{}
	err := DB.Find(member, c.MemberID)
	if err != nil {
		log.Error("cannot found associated member: ", err)
	}
	return member
}

// OwnerID returns id of associated member
func (c Credential) OwnerID() uuid.UUID {
	return c.Owner().ID
}

//** implementations for interfaces ---------------------------------

// QueryParams implements Belonging interface
func (c *Credential) QueryParams() QueryParams {
	return QueryParams{}
}

// QueryParams implements Belonging interface
func (c *Credentials) QueryParams() QueryParams {
	return QueryParams{
		DefaultSort: DefaultSortMessages,
	}
}

// OwnedBy implements Belonging interface
func (c *Credential) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// OwnedBy implements Belonging interface
func (c *Credentials) OwnedBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// AccessibleBy implements Belonging interface
func (c *Credential) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

// AccessibleBy implements Belonging interface
func (c *Credentials) AccessibleBy(q *pop.Query, o Owner, f ...bool) *pop.Query {
	return q
}

//** common database/crud functions ---------------------------------

//** array model for base model -------------------------------------

// Credentials is an array of Credentials.
type Credentials []Credential

// String returns json marshalled representation of Credentials
func (c Credentials) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}
