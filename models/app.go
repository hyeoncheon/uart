package models

import (
	"encoding/json"
	"time"

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

// Grant create an access grant for given member to the app
func (a *App) Grant(member *Member) error {
	log.Infof("access grant for app %v to member %v", a, member)
	return DB.Create(&AccessGrant{
		AppID:       a.ID,
		MemberID:    member.ID,
		AccessCount: 1,
	})
}

// AddRole create role for the app.
func (a *App) AddRole(name, desc, code string, rank int) error {
	return DB.Create(&Role{
		AppID:       a.ID,
		Name:        name,
		Description: desc,
		Code:        code,
		Rank:        rank,
	})
}

// GetRole returns a named role of the app
func (a *App) GetRole(name string) *Role {
	r := &Role{}
	err := DB.BelongsTo(a).Where("name = ?", name).First(r)
	if err != nil {
		return nil
	}
	return r
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

// GetAppByName returns a app instance of given name
func GetAppByName(name string) *App {
	app := &App{}
	err := DB.Where("name = ?", name).First(app)
	if err != nil {
		return nil
	}
	return app
}

// NewApp create an app with given values.
func NewApp(name, desc, url, callback string, icon ...string) *App {
	app := &App{
		Name:        name,
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

func createUARTApp() *App {
	uart := NewApp("UART", "UART: Identity Management System", "", "", "")
	DB.Create(uart)
	uart.AddRole("Admin", "Administrator", "admin", 64)
	uart.AddRole("User Manager", "User Manager", "userman", 8)
	uart.AddRole("App Manager", "Application Manager", "appman", 4)
	uart.AddRole("Leader", "Team Leader", "leader", 2)
	uart.AddRole("User", "Normal User", "user", 1)
	uart.AddRole("Guest", "Guest, without any privileges", "guest", 0)
	return uart
}

// Apps is array of App.
type Apps []App

// String is not required by pop and may be deleted
func (a Apps) String() string {
	ja, _ := json.Marshal(a)
	return string(ja)
}

// Validate gets run every time you call a "pop.Validate" method.
func (a *App) Validate(tx *pop.Connection) (*validate.Errors, error) {
	return validate.Validate(
		&validators.StringIsPresent{Field: a.Name, Name: "Name"},
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
