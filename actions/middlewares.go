package actions

// TODO REVIEW REQUIRED

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	uuid "github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/models"
)

// LoginAsTester is helper middleware for testing (simulate authcallback)
func LoginAsTester(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		if ENV == "test" {
			member := &models.Member{}
			models.DB.Order("updated_at desc").First(member)
			if ENV == "test" && member.ID != uuid.Nil {
				c.Logger().Info("### ------ LoginAsTester: ", member)
				c.Session().Set("member_id", member.ID)
				c.Session().Set("member_name", member.Name)
				c.Session().Set("member_mail", member.Email)
				c.Session().Set("member_icon", member.Icon)
				c.Session().Set("member_roles", member.GetAppRoleCodes(models.ACUART))
				c.Flash().Add("danger", "TEST AUTHENTICATED")
			}
		}
		return next(c)
	}
}

// AuthenticateHandler protect all application pages from unauthorized access.
func AuthenticateHandler(next buffalo.Handler) buffalo.Handler {
	return func(c buffalo.Context) error {
		memberID := c.Session().Get("member_id")
		if memberID == nil {
			c.Session().Set("origin", c.Request().RequestURI)
			c.Logger().Warn("unauthorized access to ", c.Request().RequestURI)
			c.Flash().Add("danger", t(c, "login.required"))
			return c.Redirect(http.StatusFound, "/login")
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
			c.Logger().Debug("storing roles on context: ", roles)
			for _, role := range roles {
				c.Set("role_"+role, true)
				if role == models.RCAdmin {
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
			return c.Redirect(http.StatusFound, "/")
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
				"apps":  models.RCAppMan,
				"roles": models.RCAppMan,
			}
			if p := perms[pos]; p != "" {
				if c.Value("role_"+p) == nil {
					c.Logger().Warnf("%v has no permission for %v",
						currentMember(c), pos)
					c.Flash().Add("danger", t(c, "you.dont.have.permission"))
					return c.Redirect(http.StatusFound, "/")
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
