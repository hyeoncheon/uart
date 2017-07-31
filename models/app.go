package models

// TODO REVIEW REQUIRED

import (
	"encoding/json"
	"time"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/markbates/validate"
	"github.com/markbates/validate/validators"
	"github.com/satori/go.uuid"
)

// default values
const (
	AppDefaultAdminScope = "all:all"
	AppDefaultScope      = "profile, auth:all"
	appDefaultIcon       = "/assets/images/dummy-app.png"
	appsDefaultSort      = "name"
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

// RelationalOwnerQuery returns query contains path from member to the model.
// implement belonging
func (a *App) RelationalOwnerQuery(memberID uuid.UUID) *pop.Query {
	return DB.Q().
		LeftJoin("roles", "roles.app_id = apps.id").
		LeftJoin("role_maps", "role_maps.role_id = roles.id").
		Where("role_maps.member_id = ?", memberID).
		Where("roles.code = ?", RCAdmin)
}

//** actions, relational accessor and functions below:

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

// Requests returns array of inactive rolemaps
func (a App) Requests() *[]RoleMap {
	rolemaps := &[]RoleMap{}
	err := DB.Q().
		LeftJoin("roles", "roles.id = role_maps.role_id").
		Where("roles.app_id = ?", a.ID).
		Where("role_maps.is_active = ?", false).All(rolemaps)
	if err != nil {
		log.Warn("cannot found requests. ignore: ", err)
	}
	return rolemaps
}

// GetRoles returns array of roles of the app
func (a App) GetRoles() *Roles {
	roles := &Roles{}
	DB.BelongsTo(&a).Order("rank desc").All(roles)
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

// GrantsCount returns count of access grant for the app
func (a App) GrantsCount() int {
	count, _ := DB.BelongsTo(&a).Count(&AccessGrants{})
	return count
}

// RequestsCount returns count of role requests for the app
func (a App) RequestsCount() int {
	count, _ := DB.Q().
		LeftJoin("roles", "roles.id = role_maps.role_id").
		LeftJoin("apps", "apps.id = roles.app_id").
		Where("apps.id = ?", a.ID).
		Where("role_maps.is_active = ?", false).
		Count(&RoleMap{})
	return count
}

//** Generic model operation functions below:

// GetAppByCode search and returns an app instance by given code, or nil
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

// createUARTApp create UART app and return it or nil.
func createUARTApp(tx *pop.Connection) *App {
	uart := NewApp("UART", "uart", "UART: Identity Management System", "", "")
	uart.AppIcon = hyeoncheonIcon
	err := DB.Create(uart)
	if err != nil {
		return nil
	}
	uart.AddRole(tx, "Admin", RCAdmin, "UART Administrator", 64, true)
	uart.AddRole(tx, "User Manager", RCUserMan, "UART User Manager", 32, true)
	uart.AddRole(tx, "App Manager", RCAppMan, "UART App Manager", 16, true)
	uart.AddRole(tx, "Leader", RCLeader, "Team Leader", 8, true)
	uart.AddRole(tx, "User", RCUser, "Normal User", 0, true)
	return uart
}

//** array model for base model --------------------------------------------

// Apps is array of App.
type Apps []App

// String is not required by pop and may be deleted
func (a Apps) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// SearchParams implementation (Searchable)
func (a Apps) SearchParams(c buffalo.Context) SearchParams {
	sp := newSearchParams(c)
	sp.DefaultSort = appsDefaultSort
	return sp
}

// QueryAndParams implementation (Searchable)
func (a Apps) QueryAndParams(c buffalo.Context) (*pop.Query, SearchParams) {
	sp := newSearchParams(c)
	sp.DefaultSort = appsDefaultSort
	q := DB.Q()
	if !c.Value("member_is_admin").(bool) {
		q = q.
			LeftJoin("roles", "roles.app_id = apps.id").
			LeftJoin("role_maps", "role_maps.role_id = roles.id").
			Where("role_maps.member_id = ?", c.Value("member_id")).
			Where("roles.code = ?", RCAdmin)
	}
	return q, sp
}

// Validate gets run every time you call a "pop.Validate" method.
func (a *App) Validate(tx *pop.Connection) (*validate.Errors, error) {
	if a.AppIcon == "" {
		a.AppIcon = appDefaultIcon
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
