package actions

import (
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/jinzhu/copier"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gplus"

	"github.com/hyeoncheon/uart/models"
)

func init() {
	gothic.Store = App().SessionStore

	goth.UseProviders(
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"),
			fmt.Sprintf("%s%s", App().Host, "/auth/gplus/callback")),
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"),
			fmt.Sprintf("%s%s", App().Host, "/auth/facebook/callback")),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"),
			fmt.Sprintf("%s%s", App().Host, "/auth/github/callback")),
	)
}

// AuthCallback is a callback handler for oauth2 authentication
func AuthCallback(c buffalo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		return c.Error(401, err)
	}
	c.Logger().Debugf("raw user data: %v", r.JSON(user))

	credentials := &models.Credentials{}
	if err := models.SelectByAttrs(credentials, map[string]interface{}{
		"provider": user.Provider,
		"user_id":  user.UserID,
	}); err != nil {
		return c.Error(501, err)
	}

	switch len(*credentials) {
	case 0:
		member, err := createMember(user)
		if err != nil {
			c.Flash().Add("danger", t(c, err.Error()))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		c.Flash().Add("success", t(c, "welcome.to.uart"))
		return loggedIn(c, member)
	case 1:
		member := (*credentials)[0].Owner()
		if member.Email == "" {
			c.Flash().Add("danger", t(c, "credential.exist.without.owner"))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		c.Flash().Add("success", t(c, "welcome.back.i.missed.you"))
		return loggedIn(c, member)
	default:
		return c.Error(501, errors.New("SYSTEM ERROR: duplicated credentials"))
	}
}

func createMember(user goth.User) (*models.Member, error) {
	if user.Email == "" {
		return nil, errors.New("unacceptable.no.email.provided")
	}
	if user.Name == "" {
		return nil, errors.New("unacceptable.no.name.provided")
	}
	if user.Provider == "gplus" {
		if vm, ok := user.RawData["verified_email"].(bool); ok && !vm {
			return nil, errors.New("unacceptable.email.not.verified")
		}
	}

	cred := &models.Credential{}
	copier.Copy(cred, user)
	cred.IsAuthorized = true
	cred.IsPrimary = true

	if user.Provider == "facebook" {
		if p, ok := user.RawData["picture"].(map[string]interface{}); ok {
			if d, ok := p["data"].(map[string]interface{}); ok {
				if s, ok := d["is_silhouette"].(bool); ok && s {
					cred.AvatarURL = ""
				}
			}
		}
	}
	return models.CreateMember(cred), nil
}

func loggedIn(c buffalo.Context, member *models.Member) error {
	c.Logger().Debug("--- member --- ", models.Marshal(member))
	c.Logger().Infof("member %v logged in.", member)

	session := c.Session()
	origin := session.Get("origin")
	session.Delete("origin")
	c.Logger().Infof("origin of authentication is %v", origin)
	if origin == nil {
		origin = "/"
	}
	session.Set("member_id", member.ID)
	session.Set("member_name", member.Name)
	session.Set("member_mail", member.Email)
	session.Set("member_icon", member.Icon)
	session.Set("member_roles", member.GetAppRoleCodes("uart"))
	err := session.Save()
	if err != nil {
		c.Logger().Error("SYSTEM ERROR: cannot save session! ", err)
	}
	return c.Redirect(http.StatusTemporaryRedirect, origin.(string))
}
