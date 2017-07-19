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
	session := c.Session()
	session.Clear()
	session.Save()
	c.Flash().Add("success", t(c, "you.have.been.successfully.logged.out"))
	return c.Redirect(http.StatusTemporaryRedirect, "/")
}
