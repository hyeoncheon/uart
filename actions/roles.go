package actions

// TODO REVIEW REQUIRED
// Test coverage: 100%

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/markbates/pop"
	"github.com/pkg/errors"

	"github.com/hyeoncheon/uart/models"
)

// RolesResource is the resource for the role model
type RolesResource struct {
	buffalo.Resource
}

// Create adds a Role to the DB.
func (v RolesResource) Create(c buffalo.Context) error {
	role := &models.Role{}
	err := c.Bind(role)
	if err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	dmem := dummyMember(c)
	app := &models.App{}
	if err = models.FindMyOwn(tx.Q(), dmem, app, role.AppID); err != nil {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		c.Logger().Errorf("access violation: %v tried to create a role for app %v", currentMember(c), role.AppID)
		return c.Redirect(http.StatusFound, "/")
	}

	verrs, err := tx.ValidateAndCreate(role)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("role", role)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("roles/new.html"))
	}
	c.Flash().Add("success", t(c, "role.was.created.successfully"))
	c.Logger().Infof("role %v created by %v", role, c.Value("member_id"))
	return c.Redirect(http.StatusSeeOther, "/apps/%s", role.AppID)
}

// Update changes a role in the DB.
func (v RolesResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
	}

	dmem := dummyMember(c)
	app := &models.App{}
	if err = models.FindMyOwn(tx.Q(), dmem, app, role.AppID); err != nil {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		c.Logger().Errorf("access violation: %v tried to delete a role %v", currentMember(c), role)
		return c.Redirect(http.StatusFound, "/")
	}

	if role.IsReadonly == true {
		c.Flash().Add("danger", t(c, "cannot.delete.readonly.role"))
		c.Logger().Errorf("access violation: %v tried to delete a readonly role %v", currentMember(c), role)
		return c.Redirect(http.StatusFound, "/")
	}

	err = c.Bind(role)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(role)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("role", role)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("roles/edit.html"))
	}
	c.Flash().Add("success", t(c, "role.was.updated.successfully"))
	c.Logger().Infof("role %v updated by %v", role, c.Value("member_id"))
	return c.Redirect(http.StatusSeeOther, "/apps/%s", role.AppID)
}

// Destroy deletes a role from the DB.
func (v RolesResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
	}

	dmem := dummyMember(c)
	app := &models.App{}
	if err = models.FindMyOwn(tx.Q(), dmem, app, role.AppID); err != nil {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		c.Logger().Errorf("access violation: %v tried to delete a role %v", currentMember(c), role)
		return c.Redirect(http.StatusFound, "/")
	}

	if role.IsReadonly == true {
		c.Flash().Add("danger", t(c, "cannot.delete.readonly.role"))
		c.Logger().Errorf("access violation: %v tried to delete a readonly role %v", currentMember(c), role)
		return c.Redirect(http.StatusFound, "/")
	}

	err = tx.Destroy(role)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: cleanup rolemaps for this deleted role
	c.Flash().Add("success", t(c, "role.was.destroyed.successfully"))
	c.Logger().Infof("role %v deleted by %v", role, c.Value("member_id"))
	return c.Redirect(http.StatusSeeOther, "/apps/%s", role.AppID)
}

// Accept changes the status of assigned role as active
func (v RolesResource) Accept(c buffalo.Context) error {
	appID := c.Param("app_id")
	rmID := c.Param("rolemap_id")

	rolemap := &models.RoleMap{}
	tx := c.Value("tx").(*pop.Connection)
	err := tx.Find(rolemap, rmID)
	if err != nil {
		c.Logger().Errorf("OOPS! cannot found rolemap id %v: %v", rmID, err)
		c.Flash().Add("danger", t(c, "oops.cannot.found.request"))
		return c.Redirect(http.StatusFound, "/")
	}

	role := rolemap.Role()
	if role.AppID.String() != appID {
		c.Flash().Add("danger", t(c, "oops.manipulated.request"))
		c.Logger().Errorf("app_id mismatch! probably manipulated request!")
		return c.Redirect(http.StatusFound, "/")
	}
	err = models.FindMyOwn(tx.Q(), dummyMember(c), &models.App{}, appID)
	if err != nil {
		c.Flash().Add("danger", t(c, "you.have.no.right.for.this.app"))
		return c.Redirect(http.StatusFound, "/")
	}

	rolemap.IsActive = true
	err = tx.Save(rolemap)
	if err != nil {
		c.Logger().Errorf("OOPS! cannot save rolemap id %v: %v", rmID, err)
		c.Flash().Add("danger", t(c, "oops.cannot.proceed.acception"))
	} else {
		c.Flash().Add("success", t(c, "request.accepted.successfully"))
		member := rolemap.Member()
		appMsg(c, &models.Members{*member}, "",
			"role request for %v accepted!", rolemap.Role())
	}
	return c.Redirect(http.StatusSeeOther, "/apps/%s", appID)
}

// Request creates role assignments for the member's request.
func (v RolesResource) Request(c buffalo.Context) error {
	var roleIDs []string
	if err := c.Request().ParseForm(); err == nil {
		roleIDs = c.Request().Form["role_id"]
	}
	c.Logger().Infof("%v roles are requested", len(roleIDs))

	tx := c.Value("tx").(*pop.Connection)
	member := currentMember(c)
	if !member.IsActive {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		mLogErr(c, MsgFacSecu, "access violation: inactive member %v", member)
		return c.Redirect(http.StatusFound, "/membership/me")
	}
	for _, rID := range roleIDs {
		role := &models.Role{}
		err := tx.Find(role, rID)
		if err != nil {
			tx.TX.Rollback()
			c.Logger().Warn("OOPS! role not found! ", err)
			c.Flash().Add("danger", t(c, "oops.cannot.found.the.role"))
			return c.Redirect(http.StatusFound, "/membership/me")
		}

		err = member.AddRole(tx, role)
		if err != nil {
			tx.TX.Rollback()
			c.Logger().Warnf("cannot assign a role %v to %v. error: %v",
				role, member, err)
			c.Flash().Add("danger", t(c, "cannot.add.a.role"))
			return c.Redirect(http.StatusFound, "/membership/me")
		}
		c.Flash().Add("success", t(c, "role.request.finished.successfully"))

		admins := role.App().GetRole(tx, models.RCAdmin).Members(true)
		rrd := RoleRequestData{Member: member, Role: role}
		err = noteMsg(c, admins, MsgFacUser, "new_role_requested", rrd)
		if err != nil {
			c.Logger().Error("messaging error (noteMsg): ", err)
		}
	}
	return c.Redirect(http.StatusSeeOther, "/membership/me")
}

// RoleRequestData is inventory set for role request message
type RoleRequestData struct {
	Member *models.Member
	Role   *models.Role
}

func (d RoleRequestData) String() string {
	return d.Role.String() + " for " + d.Member.Name
}

// Retire remove the role of current user
func (v RolesResource) Retire(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
	}

	member := currentMember(c)
	if !member.IsActive {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		mLogErr(c, MsgFacSecu, "access violation: inactive member %v", member)
		return c.Redirect(http.StatusFound, "/membership/me")
	}
	err = member.RemoveRole(tx, role)
	if err != nil {
		c.Flash().Add("danger", t(c, "cannot.remove.this.role.from.you"))
		tx.TX.Rollback()
	} else {
		cnt, err := tx.BelongsTo(member).Where("app_id = ?", role.AppID).
			Count(&models.RoleMap{})
		if err == nil && cnt == 0 {
			c.Logger().Info("no assigned role remind. revoke automatically.")
			app := &models.App{}
			err := tx.Find(app, role.AppID)
			if err != nil {
				c.Logger().Warnf("cannot found app with id '%v'", role.AppID)
			} else {
				c.Logger().Debugf("trying to revoke %v@%v", member, app)
				err = member.Revoke(tx, app)
				if err != nil {
					c.Logger().Warnf("cannot revoke %v@%v.", member, app)
				} else {
					c.Logger().Info("no remining roles, revoked.")
				}
			}
		}
		c.Flash().Add("success", t(c, "role.removed.from.you.successfully"))
		c.Logger().Infof("member %v removed role %v", member, role)
	}
	return c.Redirect(http.StatusSeeOther, "/membership/me")
}
