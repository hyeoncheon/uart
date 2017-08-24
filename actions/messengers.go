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

// MessengersResource is the resource for the messenger model
type MessengersResource struct {
	buffalo.Resource
}

// List gets all Messengers. GET /messengers
// ADMIN PROTECTED
func (v MessengersResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messengers := &models.Messengers{}
	q := tx.PaginateFromParams(c.Params())
	err := q.All(messengers)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("messengers", messengers)
	c.Set("pagination", q.Paginator)
	return c.Render(200, r.HTML("messengers/index.html"))
}

// Create adds a Messenger to the DB. POST /messengers
func (v MessengersResource) Create(c buffalo.Context) error {
	messenger := &models.Messenger{}
	err := c.Bind(messenger)
	if err != nil {
		return errors.WithStack(err)
	}

	me := dummyMember(c)
	messenger.MemberID = me.ID

	if messenger.Priority == models.MessengerPriority["Alert"] {
		pm := me.PrimaryAlert()
		if pm.ID == uuid.Nil {
			messenger.IsPrimary = true
		}
	}
	if messenger.Priority == models.MessengerPriority["Notification"] {
		pm := me.PrimaryNotifier()
		if pm.ID == uuid.Nil {
			messenger.IsPrimary = true
		}
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(messenger)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("messenger", messenger)
		c.Set("m_priority", models.MessengerPriority)
		c.Set("m_method", models.MessengerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messengers/new.html"))
	}
	c.Flash().Add("success", t(c, "messenger.was.created.successfully"))
	return c.Redirect(302, "/membership/me")
}

// Update changes a messenger in the DB. PUT /messengers/{messenger_id}
func (v MessengersResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	messenger := &models.Messenger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messenger, c.Param("messenger_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messenger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	err = c.Bind(messenger)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(messenger)
	if err != nil {
		c.Flash().Add("danger", t(c, "oops.cannot.update.messenger"))
		return c.Redirect(http.StatusFound, "/apps")
	}
	if verrs.HasAny() {
		c.Set("messenger", messenger)
		c.Set("m_priority", models.MessengerPriority)
		c.Set("m_method", models.MessengerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messengers/edit.html"))
	}
	c.Flash().Add("success", t(c, "messenger.was.updated.successfully"))
	return c.Redirect(302, "/membership/me")
}

// Destroy deletes a messenger from the DB. DELETE /messengers/{messenger_id}
func (v MessengersResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	isAdmin := c.Value("member_is_admin").(bool)

	messenger := &models.Messenger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messenger, c.Param("messenger_id"))
	if isAdmin {
		err = tx.Find(messenger, c.Param("messenger_id"))
	}
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messenger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	if !isAdmin && messenger.IsPrimary {
		c.Flash().Add("warning", t(c, "deleting.a.primary.is.not.allowed"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	err = tx.Destroy(messenger)
	if err != nil {
		c.Logger().Warnf("cannot delete messenger %v", messenger)
		c.Flash().Add("danger", t(c, "oops.cannot.delete.messenger"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}
	c.Flash().Add("success", t(c, "messenger.was.deleted.successfully"))
	if isAdmin && messenger.MemberID != me.ID {
		return c.Redirect(http.StatusFound, "/messengers")
	}
	return c.Redirect(302, "/membership/me")
}

// SetPrimary sets the messenger as primary (and unset others)
func (v MessengersResource) SetPrimary(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)

	messenger := &models.Messenger{}
	me := dummyMember(c)
	err := models.FindMyOwn(tx.Q(), me, messenger, c.Param("messenger_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.messenger.not.found"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}

	messengers := &models.Messengers{}
	tx.BelongsTo(me).Where("priority = ?", messenger.Priority).All(messengers)
	for _, m := range *messengers {
		m.IsPrimary = false
		if err := tx.Save(&m); err != nil {
			tx.TX.Rollback()
			c.Flash().Add("danger", t(c, "oops.cannot.save.others"))
			return c.Redirect(http.StatusFound, "/apps")
		}
	}
	messenger.IsPrimary = true

	verrs, err := tx.ValidateAndUpdate(messenger)
	if err != nil {
		tx.TX.Rollback()
		c.Flash().Add("danger", t(c, "oops.cannot.update.messenger"))
		return c.Redirect(http.StatusFound, "/apps")
	}
	if verrs.HasAny() {
		c.Set("messenger", messenger)
		c.Set("m_priority", models.MessengerPriority)
		c.Set("m_method", models.MessengerMethod)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messengers/edit.html"))
	}
	c.Flash().Add("success", t(c, "messenger.was.updated.successfully"))
	return c.Redirect(302, "/membership/me")
}

//** utilities

func setDefaultMessengers(c buffalo.Context, member *models.Member) {
	for _, prio := range []string{"Alert", "Notification"} {
		c.Logger().Debugf("set default messenger for member --- %v", prio)
		messenger := models.Messenger{
			MemberID:  member.ID,
			Priority:  models.MessengerPriority[prio],
			Method:    models.MessengerMethod["Email"],
			Value:     member.Email,
			IsPrimary: true,
		}
		tx := c.Value("tx").(*pop.Connection)
		tx.Save(&messenger)
	}
}
