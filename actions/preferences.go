package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/pkg/errors"

	"github.com/hyeoncheon/uart/models"
)

func membershipHandler(c buffalo.Context) error {
	member := &models.Member{}
	if c.Value("member_is_admin").(bool) && c.Param("member_id") != "" {
		member = models.GetMember(c.Param("member_id"))
	} else {
		member = currentMember(c)
	}
	if member == nil {
		c.Logger().Error("OOPS! member not found.")
		return c.Error(http.StatusNotFound, errors.New("Member Not Found"))
	}

	c.Set("member", member)
	c.Set("credentials", member.Credentials())
	return c.Render(http.StatusOK, r.HTML("membership.html"))
}

func currentMember(c buffalo.Context) *models.Member {
	return models.GetMember(c.Session().Get("member_id"))
}
