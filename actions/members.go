package actions

// TODO REVIEW REQUIRED
// Test coverage: 100%

import (
	"net/http"
	"strings"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v6"

	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

// MembersResource is the resource for the member model
type MembersResource struct {
	buffalo.Resource
}

// List gets all Members.
//! ADMIN PROTECTED
func (v MembersResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	members := &models.Members{}
	q := tx.PaginateFromParams(c.Params())
	err := models.AllMy(q, dummyMember(c), members, false)
	if err != nil {
		return utils.DOOPS(c, "while listing members (params: %v, error: %v)", c.Params(), err)
	}
	c.Set("members", members)
	c.Set("pagination", q.Paginator)
	return c.Render(http.StatusOK, r.HTML("members/index.html"))
}

/*
// New renders the formular for creating a new Member.
// TODO: implement mail based local authentication based on this.
func (v MembersResource) New(c buffalo.Context) error {
	c.Set("member", &models.Member{})
	return c.Render(200, r.HTML("members/new.html"))
}

// Create adds a Member to the DB.
// TODO: implement mail based local authentication based on this.
func (v MembersResource) Create(c buffalo.Context) error {
	member := &models.Member{}
	err := c.Bind(member)
	if err != nil {
		return errors.WithStack(err)
	}
	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(member)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("member", member)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("members/new.html"))
	}
	c.Flash().Add("success", "Member was created successfully")
	return c.Redirect(302, "/members/%s", member.ID)
}
*/

// Edit renders a edit formular for a member.
//! ADMIN PROTECTED
func (v MembersResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	err := tx.Find(member, c.Param("member_id"))
	if err != nil {
		c.Logger().Errorf("cannot found member with id %v", c.Param("member_id"))
		c.Flash().Add("danger", t(c, "cannot.found.member"))
		return c.Redirect(http.StatusFound, "/members")
	}
	c.Set("member", member)
	return c.Render(http.StatusOK, r.HTML("members/edit.html"))
}

// Update changes a member in the DB.
//! ADMIN PROTECTED
// TODO: update member's permissions immediately
func (v MembersResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	err := tx.Find(member, c.Param("member_id"))
	if err != nil {
		c.Logger().Errorf("cannot found member with id %v", c.Param("member_id"))
		c.Flash().Add("danger", t(c, "cannot.found.member"))
		return c.Redirect(http.StatusFound, "/members")
	}
	statusOld := member.IsActive

	err = c.Bind(member)
	if err != nil {
		return utils.SOOPS(c, "while binding member %v, error: %v", member, err)
	}
	statusNew := member.IsActive

	verrs, err := tx.ValidateAndUpdate(member)
	if err != nil {
		return utils.DOOPS(c, "while updating member %v, error: %v", member, err)
	}
	if verrs.HasAny() {
		c.Set("member", member)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("members/edit.html"))
	}

	rolemap := &models.RoleMap{}
	uRole := models.GetAppByCode("uart").GetRole(tx, models.RCUser)
	err = tx.Where("member_id = ? AND role_id = ?", member.ID, uRole.ID).
		First(rolemap)
	if err != nil {
		c.Logger().Error("cannot found rolemap for UART.User: ", err)
		c.Flash().Add("warning", t(c, "cannot.update.role.automatically"))
	} else {
		rolemap.IsActive = member.IsActive
		err = tx.Save(rolemap)
		if err != nil {
			c.Logger().Errorf("cannot save rolemap for %v: %v", uRole, err)
			c.Flash().Add("warning", t(c, "cannot.save.role.automatically"))
		}
		c.Flash().Add("info", t(c, "uart.role.also.update"))
	}
	if statusOld != statusNew {
		err := noteMsg(c, &models.Members{*member}, MsgFacUser,
			"member_status_changed", member)
		c.Logger().Debug("imessaging error: ", err)
	}

	c.Flash().Add("success", "Member was updated successfully")
	mLogInfo(c, MsgFacUser, "member %v was updated", member)
	return c.Redirect(http.StatusSeeOther, "/members")
}

// Destroy deletes a member from the DB.
//! ADMIN PROTECTED
func (v MembersResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	if err := tx.Find(member, c.Param("member_id")); err != nil {
		c.Logger().Errorf("cannot found member with id %v", c.Param("member_id"))
		c.Flash().Add("danger", t(c, "cannot.found.member"))
		return c.Redirect(http.StatusFound, "/members")
	}
	adminRole := models.GetAppRole(models.ACUART, models.RCAdmin)
	if member.HasRole(adminRole.ID) {
		c.Flash().Add("danger", t(c, "disabling.an.admin.is.not.allowed"))
		return c.Redirect(http.StatusFound, "/members")
	}

	permanent := false
	if strings.HasSuffix(member.Name, "-Deleted") {
		c.Logger().Warn("caution! permanent deletion mode selected!")
		permanent = true
	}
	var err error

	//! REMOVE RELATED THING OR JUST DISABLE THEM...
	for _, d := range *member.Credentials() {
		if !strings.HasSuffix(d.UserID, "-DLTD") {
			d.UserID = d.UserID + "-DLTD"
		}
		if permanent {
			err = tx.Destroy(&d)
		} else {
			err = tx.Save(&d)
		}
		if err != nil {
			tx.TX.Rollback()
			c.Flash().Add("danger", t(c, "cannot.inactivate.credential"))
			return c.Redirect(http.StatusFound, "/members")
		}
	}
	for _, g := range *member.Grants() {
		if err := tx.Destroy(&g); err != nil {
			tx.TX.Rollback()
			c.Flash().Add("danger", t(c, "cannot.delete.access.grant"))
			return c.Redirect(http.StatusFound, "/members")
		}
	}
	for _, r := range *member.Roles() {
		if err := member.RemoveRole(tx, &r); err != nil {
			tx.TX.Rollback()
			c.Flash().Add("danger", t(c, "cannot.remove.users.role"))
			return c.Redirect(http.StatusFound, "/members")
		}
	}
	member.IsActive = false
	member.APIKey = ""
	if !strings.HasSuffix(member.Name, "-Deleted") {
		member.Name = member.Name + "-Deleted"
	}
	if permanent {
		err = tx.Destroy(member)
	} else {
		err = tx.Save(member)
	}
	if err != nil {
		tx.TX.Rollback()
		c.Flash().Add("danger", t(c, "cannot.inactivate.member"))
		return c.Redirect(http.StatusFound, "/members")
	}
	c.Flash().Add("success", t(c, "member.was.inactivated.successfully"))
	mLogNote(c, MsgFacUser, "member %v was inactivated", member)
	return c.Redirect(http.StatusSeeOther, "/members")
}
