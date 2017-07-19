package models

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// Role is used to set member's privilege for each apps.
type Role struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Code        string    `json:"code" db:"code"`
	Rank        int       `json:"rank" db:"rank"`
}

// String returns pretty printable string of this model.
func (r Role) String() string {
	return r.App().String() + "." + r.Name
}

// App returns an app instance of the role
func (r Role) App() *App {
	app := &App{}
	DB.Find(app, r.AppID)
	return app
}

// Roles is array of Role.
type Roles []Role

// String is not required by pop and may be deleted
func (r Roles) String() string {
	jr, _ := json.Marshal(r)
	return string(jr)
}

// Validate gets run every time you call a "pop.Validate" method.
func (r *Role) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: r.Name, Name: "Name"},
		&validators.StringIsPresent{Field: r.Code, Name: "Code"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (r *Role) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (r *Role) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

//// Role Map

// RoleMap is a mapping object for role and member.
type RoleMap struct {
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
	RoleID    uuid.UUID `json:"role_id" db:"role_id"`
	MemberID  uuid.UUID `json:"member_id" db:"member_id"`
}
