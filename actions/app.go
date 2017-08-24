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

	"github.com/hyeoncheon/uart/jobs"
	"github.com/hyeoncheon/uart/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var (
	ENV         = envy.Get("GO_ENV", "development")
	brandName   = envy.Get("BRAND_NAME", "UART")
	sessionName = envy.Get("SESSION_NAME", "_uart_session")
	app         *buffalo.App
	T           *i18n.Translator
)

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

		// register all taskers
		jobs.RegisterAll(app)
		models.Logger(app.Logger)

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

		// oauth provider
		initProvider(app.Logger)
		oauth := app.Group("/oauth")
		oauth.GET("/authorize", authorizeHandler)
		oauth.POST("/token", tokenHandler)
		oauth.Middleware.Skip(csrf.Middleware, tokenHandler)
		oauth.Middleware.Skip(AuthenticateHandler, tokenHandler)
		app.GET("/userinfo", userInfoHandler)
		app.Middleware.Skip(csrf.Middleware, userInfoHandler)
		app.Middleware.Skip(AuthenticateHandler, userInfoHandler)

		var r buffalo.Resource

		// Admin Resources
		r = &MembersResource{&buffalo.BaseResource{}}
		g := app.Resource("/members", r)
		g.Use(adminHandler)
		app.GET("/membership/me", membershipHandler)
		app.GET("/membership/{member_id}", membershipHandler)

		r = &CredentialsResource{&buffalo.BaseResource{}}
		g = app.Resource("/credentials", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.Destroy)

		r = &MessagesResource{&buffalo.BaseResource{}}
		g = app.Resource("/messages", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.List, r.Show)
		app.GET("/messages/{message_id}/dismiss", r.(*MessagesResource).Dismiss)

		r = &MessengersResource{&buffalo.BaseResource{}}
		g = app.Resource("/messengers", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.Create, r.Destroy, r.Update)
		g.GET("/{messenger_id}/setprimary", r.(*MessengersResource).SetPrimary)
		g.Middleware.Skip(adminHandler, r.(*MessengersResource).SetPrimary)

		r = &MessagingLogsResource{&buffalo.BaseResource{}}
		g = app.Resource("/messaging_logs", r)
		g.Use(adminHandler)

		// App Resources
		r = &AppsResource{&buffalo.BaseResource{}}
		g = app.Resource("/apps", r)
		g.Use(roleBasedLockHandler)
		app.GET("/grant/{key}", r.(*AppsResource).Grant)
		app.GET("/revoke/{app_id}", r.(*AppsResource).Revoke)

		r = &RolesResource{&buffalo.BaseResource{}}
		g = app.Resource("/roles", r)
		g.Use(roleBasedLockHandler)
		app.POST("/request", r.(*RolesResource).Request)
		app.GET("/accept/{app_id}/{rolemap_id}", r.(*RolesResource).Accept)
		app.GET("/retire/{role_id}", r.(*RolesResource).Retire)
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
