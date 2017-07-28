package models

// TODO REVIEW REQUIRED

import (
	"encoding/json"
	"html/template"
	"time"

	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/russross/blackfriday"
	"github.com/satori/go.uuid"
)

// AccessGrant is the linkage between Member and App.
type AccessGrant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	MemberID    uuid.UUID `json:"member_id" db:"member_id"`
	Scope       string    `json:"scope" db:"scope"`
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
	return mem.String() + " granted " + app.String()
}

// Description returns formatted description of the access grant
func (g AccessGrant) Description() template.HTML {
	app := g.App()
	mem := g.Member()
	timeString := g.CreatedAt.Local().Format("06-01-02 15:04")
	mdBytes := blackfriday.MarkdownBasic([]byte(
		mem.Name + " granted scope `" + g.Scope +
			"` to " + app.String() +
			" at " + timeString,
	))
	return template.HTML(string(mdBytes))
}

//** actions, relational accessor and functions below:

// Member returns the associcated member instance
func (g AccessGrant) Member() *Member {
	member := &Member{}
	err := DB.Find(member, g.MemberID)
	if err != nil {
		return nil
	}
	return member
}

// App returns the associated app instance
func (g AccessGrant) App() *App {
	app := &App{}
	err := DB.Find(app, g.AppID)
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
