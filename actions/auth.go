package actions

import (
	"fmt"
	"os"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/facebook"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/gplus"
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
	// Do something with the user, maybe register them/sign them in
	return c.Render(200, r.JSON(user))
}
