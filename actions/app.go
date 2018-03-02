package actions

import (
	"log"
	"os"
	"path"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/buffalo/middleware"
	"github.com/gobuffalo/buffalo/middleware/csrf"
	"github.com/gobuffalo/buffalo/middleware/i18n"
	"github.com/gobuffalo/envy"
	"github.com/gobuffalo/packr"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth/gothic"
	"github.com/pkg/errors"

	"github.com/hyeoncheon/uart/jobs"
	"github.com/hyeoncheon/uart/models"
)

// ENV is used to help switch settings based on where the
// application is being run. Default is "development".
var (
	ENV         = envy.Get("GO_ENV", "development")
	brandName   = envy.Get("BRAND_NAME", "UART")
	sessionName = envy.Get("SESSION_NAME", "_uart_session")
	pname       string
	uartHome    = envy.Get("UART_HOME", "")
	app         *buffalo.App
	T           *i18n.Translator
)

// App is where all routes and middleware for buffalo
// should be defined. This is the nerve center of your
// application.
func App() *buffalo.App {
	if app == nil {
		app = buffalo.New(buffalo.Options{
			Env:          ENV,
			SessionName:  sessionName,
			SessionStore: newSessionStore(ENV),
		})

		pname = path.Base(os.Args[0])
		app.Logger.Infof("UART executed as %v (in %v mode)...", pname, ENV)
		if uartHome == "" {
			uartHome, _ = os.Getwd()
		}
		app.Logger.Info("UART Home is ", uartHome)

		if _, err := os.Stat(uartHome + "/messages/"); err != nil {
			app.Stop(errors.New("abort! message template not found"))
		}

		// register all taskers
		jobs.RegisterAll(app)
		models.Logger(app.Logger)

		if ENV == "development" {
			app.Use(middleware.ParameterLogger)
		}

		// Protect against CSRF attacks. https://www.owasp.org/index.php/Cross-Site_Request_Forgery_(CSRF)
		// Remove to disable this.
		app.Use(csrf.New)

		// Wraps each request in a transaction.
		//  c.Value("tx").(*pop.PopTransaction)
		// Remove to disable this.
		app.Use(middleware.PopTransaction(models.DB))

		// Setup and use translations:
		var err error
		if T, err = i18n.New(packr.NewBox(uartHome+"/locales"), "en-US"); err != nil {
			app.Stop(err)
		}
		app.Use(T.Middleware())

		app.GET("/", HomeHandler)
		app.GET("/login", LoginHandler)
		app.GET("/logout", LogoutHandler)

		// authentication
		auth := app.Group("/auth")
		auth.GET("/{provider}", buffalo.WrapHandlerFunc(gothic.BeginAuthHandler))
		auth.GET("/{provider}/callback", AuthCallback)

		if ENV == "test" {
			app.Use(LoginAsTester)
		}

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
		//oauth.Middleware.Skip(csrf.Middleware, tokenHandler)
		oauth.Middleware.Skip(AuthenticateHandler, tokenHandler)
		app.GET("/userinfo", userInfoHandler)
		//app.Middleware.Skip(csrf.Middleware, userInfoHandler)
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

		r = &DocsResource{&buffalo.BaseResource{}}
		g = app.Resource("/docs", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.List, r.Show)
		g.GET("/publish", r.(*DocsResource).Publish)

		r = &MessagesResource{&buffalo.BaseResource{}}
		g = app.Resource("/messages", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.List, r.Show)
		g.GET("/dismiss", r.(*MessagesResource).Dismiss)
		g.Middleware.Skip(adminHandler, r.(*MessagesResource).Dismiss)

		r = &MessengersResource{&buffalo.BaseResource{}}
		g = app.Resource("/messengers", r)
		g.Use(adminHandler)
		g.Middleware.Skip(adminHandler, r.Create, r.Destroy, r.Update)
		g.GET("/setprimary", r.(*MessengersResource).SetPrimary)
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
		//! FIXME: URLs are not pretty :-(
		app.GET("/requests/{rolemap_id}/accept", r.(*RolesResource).Accept)
		app.POST("/requests/roles", r.(*RolesResource).Request)
		app.GET("/requests/roles/{role_id}/retire", r.(*RolesResource).Retire)

		// move to end of the routing :-(
		app.ServeFiles("/", assetsBox)
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
