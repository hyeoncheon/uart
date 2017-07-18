package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
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

// GetMember find and return associated member instance
func (c Credential) GetMember() (*Member, error) {
	member := &Member{}
	err := DB.Find(member, c.MemberID)
	return member, err
}

// Credentials is an array of Credentials.
type Credentials []Credential

// String is not required by pop and may be deleted
func (c Credentials) String() string {
	jc, _ := json.Marshal(c)
	return string(jc)
}

// Validate gets run every time you call a "pop.Validate" method.
func (c *Credential) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: c.Provider, Name: "Provider"},
		&validators.StringIsPresent{Field: c.UserID, Name: "UserID"},
		&validators.StringIsPresent{Field: c.Name, Name: "Name"},
		&validators.StringIsPresent{Field: c.Email, Name: "Email"},
		&validators.StringIsPresent{Field: c.AvatarURL, Name: "AvatarURL"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (c *Credential) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (c *Credential) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
