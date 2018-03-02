package actions

// TODO REVIEW REQUIRED

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/uuid"
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
	if member.IsNil() {
		c.Logger().Error("OOPS! member not found.")
		return c.Error(http.StatusNotFound, errors.New("Member Not Found"))
	}

	c.Set("member", member)
	c.Set("credentials", member.Credentials())
	c.Set("roles", member.Roles())
	c.Set("grants", member.Grants())
	c.Set("messengers", member.Messengers())
	c.Set("messenger", &models.Messenger{})
	c.Set("m_priority", models.MessengerPriority)
	c.Set("m_method", models.MessengerMethod)
	return c.Render(http.StatusOK, r.HTML("membership.html"))
}

func currentMember(c buffalo.Context) *models.Member {
	return models.GetMember(c.Session().Get("member_id"))
}

func dummyMember(c buffalo.Context) *models.Member {
	dummy := &models.Member{}
	if id, ok := c.Value("member_id").(uuid.UUID); ok {
		dummy.ID = id
	}
	return dummy
}
