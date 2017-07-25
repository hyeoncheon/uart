package models

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// App is model for application which can be authenticated with uart.
type App struct {
	ID          uuid.UUID `json:"id" db:"id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	Name        string    `json:"name" db:"name"`
	Code        string    `json:"code" db:"code"`
	Description string    `json:"description" db:"description"`
	AppKey      string    `json:"app_key" db:"app_key"`
	AppSecret   string    `json:"app_secret" db:"app_secret"`
	SiteURL     string    `json:"site_url" db:"site_url"`
	CallbackURL string    `json:"callback_url" db:"callback_url"`
	AppIcon     string    `json:"app_icon" db:"app_icon"`
}

// String returns pretty printable string of this model.
func (a App) String() string {
	return a.Name
}

//// actions and relational functions below:

// Grant create an access grant for given member to the app
func (a *App) Grant(tx *pop.Connection, member *Member) error {
	log.Infof("access grant for app %v to member %v", a, member)
	return tx.Create(&AccessGrant{
		AppID:       a.ID,
		MemberID:    member.ID,
		AccessCount: 1,
	})
}

// Revoke decouples the app and given member, returns database status
func (a *App) Revoke(tx *pop.Connection, member *Member) error {
	log.Infof("revoke access to %v by %v", a.Name, member.Name)
	grant := &AccessGrant{}
	err := tx.BelongsTo(member).Where("app_id = ?", a.ID).First(grant)
	if err != nil {
		log.Errorf("cannot found grant for app %v to %v: %v", a, member, err)
	}
	return tx.Destroy(grant)
}

// AddRole create role for the app.
func (a *App) AddRole(tx *pop.Connection, n, c, d string, r int, o bool) error {
	return tx.Create(&Role{
		AppID:       a.ID,
		Name:        n,
		Code:        c,
		Description: d,
		Rank:        r,
		IsReadonly:  o,
	})
}

// GetRole returns a role with given code of the app or nil.
func (a *App) GetRole(tx *pop.Connection, code string) *Role {
	r := &Role{}
	err := tx.BelongsTo(a).Where("code = ?", code).First(r)
	if err != nil {
		return nil
	}
	return r
}

// GetRoles returns array of roles of the app
func (a *App) GetRoles() *Roles {
	roles := &Roles{}
	DB.BelongsTo(a).Order("rank desc").All(roles)
	return roles
}

// MemberRoles returns roles of the app which is associated with given member
func (a App) MemberRoles(memberID interface{}) *Roles {
	member := &Member{}
	roles := &Roles{}
	err := DB.Find(member, memberID)
	if err != nil {
		log.Errorf("cannot found member with given id %v: %v", memberID, err)
		return roles
	}
	DB.BelongsToThrough(member, &RoleMap{}).
		Where("roles.app_id = ?", a.ID).All(roles)
	var rs = []string{}
	for _, r := range *roles {
		rs = append(rs, r.Name)
	}
	return roles
}

// GenerateKeyPair generates key and secret for the app.
func (a *App) GenerateKeyPair() {
	key := randString(48)
	for DB.Where("app_key = ?", key).First(&App{}) == nil {
		key = randString(48)
		log.Debug("duplicated key. try again")
	}
	a.AppKey = key

	secret := randString(64)
	for DB.Where("app_secret = ?", secret).First(&App{}) == nil {
		secret = randString(64)
		log.Debug("duplicated secretcc. try again")
	}
	a.AppSecret = secret
}

// GrantCount returns count of access grant for the app
func (a App) GrantCount() int {
	count, _ := DB.BelongsTo(&a).Count(&AccessGrants{})
	return count
}

//// Generic model operation functions below:

// GetAppByCode returns an app instance has given code
func GetAppByCode(code string) *App {
	app := &App{}
	err := DB.Where("code = ?", code).First(app)
	if err != nil {
		return nil
	}
	return app
}

// GetAppByKey returns an app instance has given app_key or nil.
func GetAppByKey(key string) *App {
	app := &App{}
	err := DB.Where("app_key = ?", key).First(app)
	if err != nil {
		return nil
	}
	return app
}

// NewApp create an app with given values.
func NewApp(name, code, desc, url, callback string, icon ...string) *App {
	app := &App{
		Name:        name,
		Code:        code,
		Description: desc,
		SiteURL:     url,
		CallbackURL: callback,
	}
	if len(icon) == 1 {
		app.AppIcon = icon[0]
	}
	app.GenerateKeyPair()
	return app
}

const hyeoncheonIcon = "/assets/images/hyeoncheon-icon.png"

func createUARTApp(tx *pop.Connection) *App {
	uart := NewApp("UART", "uart", "UART: Identity Management System", "", "")
	uart.AppIcon = hyeoncheonIcon
	DB.Create(uart)
	uart.AddRole(tx, "Admin", "admin", "UART Administrator", 64, true)
	uart.AddRole(tx, "User Manager", "userman", "UART User Manager", 32, true)
	uart.AddRole(tx, "App Manager", "appman", "UART App Manager", 16, true)
	uart.AddRole(tx, "Leader", "leader", "Team Leader", 2, true)
	uart.AddRole(tx, "User", "user", "Normal User", 1, true)
	uart.AddRole(tx, "Guest", "guest", "Guest, No Privileges", 0, true)
	return uart
}

// Apps is array of App.
type Apps []App

// String is not required by pop and may be deleted
func (a Apps) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

const appsDefaultSort = "name"

// SearchParams implementation (Searchable)
func (a Apps) SearchParams(c buffalo.Context) SearchParams {
	sp := newSearchParams(c)
	sp.DefaultSort = appsDefaultSort
	return sp
}

const defaultAppIcon = "/assets/images/dummy-app.png"

// Validate gets run every time you call a "pop.Validate" method.
func (a *App) Validate(tx *pop.Connection) (*validate.Errors, error) {
	if a.AppIcon == "" {
		a.AppIcon = defaultAppIcon
	}
	return validate.Validate(
		&validators.StringIsPresent{Field: a.Name, Name: "Name"},
		&validators.StringIsPresent{Field: a.Code, Name: "Code"},
		&validators.StringIsPresent{Field: a.AppKey, Name: "AppKey"},
		&validators.StringIsPresent{Field: a.AppSecret, Name: "AppSecret"},
		&validators.StringIsPresent{Field: a.SiteURL, Name: "SiteUrl"},
		&validators.StringIsPresent{Field: a.CallbackURL, Name: "CallbackUrl"},
		&validators.StringIsPresent{Field: a.AppIcon, Name: "AppIcon"},
	), nil
}

// ValidateSave gets run every time you call "pop.ValidateSave" method.
func (a *App) ValidateSave(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}

// ValidateUpdate gets run every time you call "pop.ValidateUpdate" method.
func (a *App) ValidateUpdate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.NewErrors(), nil
}
