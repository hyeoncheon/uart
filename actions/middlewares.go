package actions

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
)

// AuthenticateHandler protect all application pages from unauthorized access.
func AuthenticateHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		memberID := c.Session().Get("member_id")
		if memberID == nil {
			c.Session().Set("origin", c.Request().RequestURI)
			c.Logger().Warn("unauthorized access to ", c.Request().RequestURI)
			c.Flash().Add("danger", t(c, "login.required"))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		return next(c)
	}
}

// contextHandler set context variables for all pages, including public pages.
// It uses session information for traditional web pages so it must be called
// after authentication handler.
func contextHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		memberID := c.Session().Get("member_id")
		if memberID != nil {
			c.Set("member_id", memberID)
			c.Set("member_name", c.Session().Get("member_name"))
			c.Set("member_mail", c.Session().Get("member_mail"))
			c.Set("member_icon", c.Session().Get("member_icon"))
			c.Set("member_roles", c.Session().Get("member_roles"))
		}
		c.Set("member_is_admin", false) // prevent nil
		if roles, ok := c.Session().Get("member_roles").([]string); ok {
			for _, role := range roles {
				c.Logger().Debug("role-----------", role)
				c.Set("role_"+role, true)
				if role == "admin" {
					c.Set("member_is_admin", true)
				}
			}
		}
		c.Set("brand_name", brandName)
		c.Set("lang", languageSelector(c))
		return next(c)
	}
}

func adminHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if val, ok := c.Value("member_is_admin").(bool); !ok || !val {
			c.Flash().Add("danger", t(c, "staff.only"))
			return c.Redirect(http.StatusTemporaryRedirect, "/")
		}
		c.Set("theme", "admin")
		return next(c)
	}
}

func roleBasedLockHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if val, ok := c.Value("member_is_admin").(bool); !ok || !val {
			pos := strings.Split(c.Value("current_path").(string), "/")[1]
			perms := map[string]string{
				"apps": "appman",
			}
			if p := perms[pos]; p != "" {
				if c.Value("role_"+p) == nil {
					c.Logger().Warnf("%v has no permission for %v",
						currentMember(c), pos)
					c.Flash().Add("danger", t(c, "you.dont.have.permission"))
					return c.Redirect(http.StatusTemporaryRedirect, "/")
				}
				c.Logger().Infof("user aquires permission %v for %v.", p, pos)
			}
		}
		return next(c)
	}
}

//helpers

func languageSelector(c buffalo.Context) string {
	// quick and dirty static ordered list of supported languages
	supportedLangs := []string{"ko-KR", "en-US"}
	acceptLangs := c.Request().Header.Get("Accept-Language")
	for _, al := range strings.Split(acceptLangs, ",") {
		al = strings.Split(al, ";")[0]
		for _, sl := range supportedLangs {
			if sl == al {
				return sl
			}
		}
	}
	return supportedLangs[0]
}
