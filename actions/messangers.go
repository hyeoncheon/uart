package actions

//! WIP

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
	"github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/models"
)

// MessangersResource is the resource for the messanger model
type MessangersResource struct {
	buffalo.Resource
}

// List gets all Messangers. GET /messangers
// ADMIN PROTECTED
func (v MessangersResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messangers := &models.Messangers{}
	q := tx.PaginateFromParams(c.Params())
	err := q.All(messangers)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("messangers", messangers)
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.HTML("messangers/index.html"))
}

// Create adds a Messanger to the DB. POST /messangers
func (v MessangersResource) Create(c buffalo.Context) error {
	messanger := &models.Messanger{}
	err := c.Bind(messanger)
	if err != nil {
		return errors.WithStack(err)
	}

	me := dummyMember(c)
	messanger.MemberID = me.ID

	if messanger.Priority == models.MessangerPriority["Alert"] {
		pm := me.PrimaryAlert()
		if pm.ID == uuid.Nil {
			messanger.IsPrimary = true
		}
	}
	if messanger.Priority == models.MessangerPriority["Notification"] {
		pm := me.PrimaryNotifier()
		if pm.ID == uuid.Nil {
			messanger.IsPrimary = true
		}
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(messanger)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("messanger", messanger)
		c.Set("m_priority", models.MessangerPriority)
		c.Set("m_method", models.MessangerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messangers/new.html"))
	}
	c.Flash().Add("success", t(c, "messanger.was.created.successfully"))
	return c.Redirect(302, "/membership/me")
}

// Update changes a messanger in the DB. PUT /messangers/{messanger_id}
func (v MessangersResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	messanger := &models.Messanger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messanger, c.Param("messanger_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messanger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	err = c.Bind(messanger)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(messanger)
	if err != nil {
		c.Flash().Add("danger", t(c, "oops.cannot.update.messanger"))
		return c.Redirect(http.StatusFound, "/apps")
	}
	if verrs.HasAny() {
		c.Set("messanger", messanger)
		c.Set("m_priority", models.MessangerPriority)
		c.Set("m_method", models.MessangerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messangers/edit.html"))
	}
	c.Flash().Add("success", t(c, "messanger.was.updated.successfully"))
	return c.Redirect(302, "/membership/me")
}

// Destroy deletes a messanger from the DB. DELETE /messangers/{messanger_id}
func (v MessangersResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	isAdmin := c.Value("member_is_admin").(bool)

	messanger := &models.Messanger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messanger, c.Param("messanger_id"))
	if isAdmin {
		err = tx.Find(messanger, c.Param("messanger_id"))
	}
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messanger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	if !isAdmin && messanger.IsPrimary {
		c.Flash().Add("warning", t(c, "deleting.a.primary.is.not.allowed"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	err = tx.Destroy(messanger)
	if err != nil {
		c.Logger().Warnf("cannot delete messanger %v", messanger)
		c.Flash().Add("danger", t(c, "oops.cannot.delete.messanger"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}
	c.Flash().Add("success", t(c, "messanger.was.deleted.successfully"))
	if isAdmin && messanger.MemberID != me.ID {
		return c.Redirect(http.StatusFound, "/messangers")
	}
	return c.Redirect(302, "/membership/me")
}

// SetPrimary sets the messanger as primary (and unset others)
func (v MessangersResource) SetPrimary(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	messanger := &models.Messanger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messanger, c.Param("messanger_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messanger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	messangers := &models.Messangers{}
	tx.BelongsTo(me).Where("priority = ?", messanger.Priority).All(messangers)
	for _, m := range *messangers {
		m.IsPrimary = false
		if err := tx.Save(&m); err != nil {
			tx.TX.Rollback()
			c.Flash().Add("danger", t(c, "oops.cannot.save.others"))
			return c.Redirect(http.StatusFound, "/apps")
		}
	}
	messanger.IsPrimary = true

	verrs, err := tx.ValidateAndUpdate(messanger)
	if err != nil {
		tx.TX.Rollback()
		c.Flash().Add("danger", t(c, "oops.cannot.update.messanger"))
		return c.Redirect(http.StatusFound, "/apps")
	}
	if verrs.HasAny() {
		c.Set("messanger", messanger)
		c.Set("m_priority", models.MessangerPriority)
		c.Set("m_method", models.MessangerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messangers/edit.html"))
	}
	c.Flash().Add("success", t(c, "messanger.was.updated.successfully"))
	return c.Redirect(302, "/membership/me")
}

//** utilities

func setDefaultMessangers(c buffalo.Context, member *models.Member) {
	for _, prio := range []string{"Alert", "Notification"} {
		c.Logger().Debugf("set default messanger for member --- %v", prio)
		messanger := models.Messanger{
			MemberID:  member.ID,
			Priority:  models.MessangerPriority[prio],
			Method:    models.MessangerMethod["Email"],
			Value:     member.Email,
			IsPrimary: true,
		}
		tx := c.Value("tx").(*pop.Connection)
		tx.Save(&messanger)
	}
}
