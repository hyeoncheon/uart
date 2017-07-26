package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/satori/go.uuid"
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

// String returns pretty printable string of this model.
func (c Credential) String() string {
	return c.Provider + "/" + c.UserID
}

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

// Credentials is an array of Credentials.
type Credentials []Credential

// String is not required by pop and may be deleted
func (cs Credentials) String() string {
	jc, _ := json.Marshal(cs)
	return string(jc)
}

const credentialsDefaultSort = "created_at"

// SearchParams implementation (Searchable)
func (cs Credentials) SearchParams(c buffalo.Context) SearchParams {
	sp := newSearchParams(c)
	sp.DefaultSort = credentialsDefaultSort
	return sp
}

// QueryAndParams implementation (Searchable)
func (cs Credentials) QueryAndParams(c buffalo.Context) (*pop.Query, SearchParams) {
	sp := newSearchParams(c)
	sp.DefaultSort = credentialsDefaultSort
	q := DB.Q()
	return q, sp
}
