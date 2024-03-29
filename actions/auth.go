package actions

// TODO REVIEW REQUIRED

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

	fmt.Println("App().Host:", App().Options.Host)
	goth.UseProviders(
		gplus.New(os.Getenv("GPLUS_KEY"), os.Getenv("GPLUS_SECRET"),
			fmt.Sprintf("%s%s", App().Options.Host, "/auth/gplus/callback")),
		facebook.New(os.Getenv("FACEBOOK_KEY"), os.Getenv("FACEBOOK_SECRET"),
			fmt.Sprintf("%s%s", App().Options.Host, "/auth/facebook/callback")),
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"),
			fmt.Sprintf("%s%s", App().Options.Host, "/auth/github/callback")),
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
		member, err := createMember(c, user)
		if err != nil {
			mLogWarn(c, MsgFacAuth, "member creation failed for %v.%v",
				user.Provider, user.UserID)
			c.Flash().Add("danger", t(c, err.Error()))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		x := models.GetAppRole(models.ACUART, models.RCAdmin).Members(true)
		noteMsg(c, x, MsgFacUser, "new_member_registered", member)
		c.Flash().Add("success", t(c, "welcome.to.uart"))
		return loggedIn(c, member)
	case 1:
		cred := (*credentials)[0]
		member := cred.Owner()
		if member.Email == "" {
			mLogWarn(c, MsgFacAuth,
				"attempted to register with orphan credential %v", cred)
			c.Flash().Add("danger", t(c, "credential.exist.without.owner"))
			return c.Redirect(http.StatusTemporaryRedirect, "/login")
		}
		updateCredential(&cred, user)
		c.Flash().Add("success", t(c, "welcome.back.i.missed.you"))
		return loggedIn(c, cred.Owner())
	default:
		mLogAlert(c, MsgFacAuth, "SYSTEM ERROR: duplicated credentials")
		return c.Error(501, errors.New("SYSTEM ERROR: duplicated credentials"))
	}
}

func updateCredential(cred *models.Credential, user goth.User) {
	if user.Email != "" && cred.Email != user.Email {
		cred.Email = user.Email
	}
	if user.Name != "" && cred.Name != user.Name {
		cred.Name = user.Name
	}
	if user.Provider == "facebook" {
		if p, ok := user.RawData["picture"].(map[string]interface{}); ok {
			if d, ok := p["data"].(map[string]interface{}); ok {
				if s, ok := d["is_silhouette"].(bool); ok && s {
					cred.AvatarURL = ""
				} else {
					cred.AvatarURL = user.AvatarURL
				}
			}
		}
	}
	cred.Save()
	if cred.IsPrimary {
		member := cred.Owner()
		member.Icon = cred.AvatarURL
		member.Save()
	}
}

func createMember(c buffalo.Context, user goth.User) (*models.Member, error) {
	if user.Email == "" {
		return nil, errors.New(t(c, "unacceptable.no.email.provided"))
	}
	if user.Name == "" {
		return nil, errors.New(t(c, "unacceptable.no.name.provided"))
	}
	if user.Provider == "gplus" {
		if vm, ok := user.RawData["verified_email"].(bool); ok && !vm {
			return nil, errors.New(t(c, "unacceptable.email.not.verified"))
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
	member, err := models.CreateMember(cred)
	if err != nil {
		err = errors.New(t(c, "oops.cannot.register.a.member"))
	}
	setDefaultMessengers(c, member)
	return member, err
}

func loggedIn(c buffalo.Context, member *models.Member) error {
	c.Logger().Debug("--- member --- ", models.Marshal(member))
	mLogInfo(c, MsgFacAuth, "member %v logged in.", member)

	session := c.Session()
	origin := session.Get("origin")
	session.Delete("origin")
	c.Logger().Debugf("origin of authentication is %v", origin)
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
	return c.Redirect(http.StatusTemporaryRedirect, "%s", origin.(string))
}
