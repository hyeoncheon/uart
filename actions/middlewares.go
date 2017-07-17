package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
)

// AuthenticateHandler protect all application pages from unauthorized access.
func AuthenticateHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		memberID := c.Session().Get("member_id")
		if memberID == nil {
			c.Logger().Warn("unauthorized access to ", c.Request().RequestURI)
			c.Flash().Add("danger", t(c, "login.required"))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		return next(c)
	}
}
