package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// HomeHandler is a default handler to serve up
// a home page.
func HomeHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("index.html"))
}

// LoginHandler renders login page
func LoginHandler(c buffalo.Context) error {
	return c.Render(http.StatusOK, r.HTML("login.html"))
}

// LogoutHandler clears all session information and redirects user to root.
func LogoutHandler(c buffalo.Context) error {
	// workaround for goth logout feature. originally gothic.Logout(res, req)
	for _, p := range []string{"gplus", "facebook", "github"} {
		s, err := app.SessionStore.Get(c.Request(), p+"_gothic_session")
		if err == nil {
			s.Options.MaxAge = -1
			s.Values = make(map[interface{}]interface{})
			s.Save(c.Request(), c.Response())
		}
	}
	session := c.Session()
	session.Clear()
	session.Save()
	c.Flash().Add("success", t(c, "you.have.been.successfully.logged.out"))
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
