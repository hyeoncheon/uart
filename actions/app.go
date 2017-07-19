package actions

import (
	"log"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"

	"github.com/hyeoncheon/uart/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var ENV = envy.Get("GO_ENV", "development")
var brandName = envy.Get("BRAND_NAME", "UART")
var sessionName = envy.Get("SESSION_NAME", "_uart_session")
var app *buffalo.App
var T *i18n.Translator

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.Automatic(buffalo.Options{
			Env:          ENV,
			SessionName:  sessionName,
			SessionStore: newSessionStore(ENV),
		})
		// Automatically save the session if the underlying
		// Handler does not return an error.
		app.Use(middleware.SessionSaver)

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		if ENV != "test" {
			// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
			// Remove to disable this.
			app.Use(csrf.Middleware)
		}

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.NewBox("../locales"), "en-US"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())

		app.GET("/", HomeHandler)
		app.GET("/login", LoginHandler)
		app.GET("/logout", LogoutHandler)

		app.ServeFiles("/assets", packr.NewBox("../public/assets"))

		// authentication
		auth := app.Group("/auth")
		auth.GET("/{provider}", buffalo.WrapHandlerFunc(gothic.BeginAuthHandler))
		auth.GET("/{provider}/callback", AuthCallback)

		app.Use(AuthenticateHandler)
		app.Middleware.Skip(AuthenticateHandler, HomeHandler)
		app.Middleware.Skip(AuthenticateHandler, LoginHandler)
		app.Middleware.Skip(AuthenticateHandler, LogoutHandler)

		app.Use(contextHandler) // just after authentication

		var r buffalo.Resource

		// Admin Resources
		r = &MembersResource{&buffalo.BaseResource{}}
		g := app.Resource("/members", r)
		g.Use(adminHandler)
		app.GET("/preferences", preferencesHandler)
		app.GET("/preferences/{member_id}", preferencesHandler)
	}

	return app
}

func newSessionStore(env string) sessions.Store {
	secret := envy.Get("SESSION_SECRET", "")
	if env == "production" && secret == "" {
		log.Fatal("set SESSION_SECRET environtmental variable for security!")
	}
	cookieStore := sessions.NewCookieStore([]byte(secret))
	cookieStore.MaxAge(60 * 60 * 1)
	return cookieStore
}
