package models

// TODO REVIEW REQUIRED

import (
	"encoding/json"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// AccessGrant is the linkage between Member and App.
type AccessGrant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	MemberID    uuid.UUID `json:"member_id" db:"member_id"`
	IsRevoked   bool      `json:"is_revoked" db:"is_revoked"`
	ExpiresIn   time.Time `json:"expires_in" db:"expires_in"`
	RevokedAt   time.Time `json:"revoked_at" db:"revoked_at"`
	AccessCount int       `json:"access_count" db:"access_count"`
}

// String returns pretty printable string of this model.
func (g AccessGrant) String() string {
	app := g.App()
	mem := g.Member()
	if app == nil || mem == nil {
		return "Broken Access Grant!"
	}
	return app.String() + " to " + mem.String()
}

//** actions, relational accessor and functions below:

// Member returns the associcated member instance
func (g AccessGrant) Member() *Member {
	member := &Member{}
	err := DB.BelongsTo(&g).First(member)
	if err != nil {
		return nil
	}
	return member
}

// App returns the associated app instance
func (g AccessGrant) App() *App {
	app := &App{}
	err := DB.BelongsTo(&g).First(app)
	if err != nil {
		return nil
	}
	return app
}

//** array model for base model --------------------------------------------

// AccessGrants is array of AccessGrants
type AccessGrants []AccessGrant

// String is not required by pop and may be deleted
func (g AccessGrants) String() string {
	jg, _ := json.Marshal(g)
	return string(jg)
}

// Validate gets run every time you call a "pop.Validate" method.
func (g *AccessGrant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.TimeIsPresent{Field: g.ExpiresIn, Name: "ExpiresIn"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (g *AccessGrant) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (g *AccessGrant) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
