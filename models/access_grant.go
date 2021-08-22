package models

// TODO REVIEW REQUIRED
// Test coverage: 100% (without interface methods)

import (
	"html/template"
	"time"

	"github.com/gobuffalo/pop/v5"
	"github.com/gobuffalo/validate/v3"
	"github.com/gofrs/uuid"
	"github.com/russross/blackfriday/v2"
)

const (
	grantsDefaultSort = "created_at desc"
)

// AccessGrant is the linkage between Member and App.
type AccessGrant struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	AppID       uuid.UUID `json:"app_id" db:"app_id"`
	MemberID    uuid.UUID `json:"member_id" db:"member_id"`
	Scope       string    `json:"scope" db:"scope"`
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
	mdBytes := blackfriday.Run([]byte(
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

// Validate gets run every time you call a "pop.Validate" method.
func (g *AccessGrant) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(), nil
}
