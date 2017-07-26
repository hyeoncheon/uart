package actions

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
	return c.Redirect(302, "/apps/%s", role.AppID)
}

// Update changes a role in the DB.
func (v RolesResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
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
	return c.Redirect(302, "/apps/%s", role.AppID)
}

// Destroy deletes a role from the DB.
func (v RolesResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(role)
	if err != nil {
		return errors.WithStack(err)
	}

	// TODO: cleanup rolemaps for this deleted role
	c.Flash().Add("success", t(c, "role.was.destroyed.successfully"))
	return c.Redirect(302, "/apps/%s", role.AppID)
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
	} else {
		rolemap.IsActive = true
		err = tx.Save(rolemap)
		if err != nil {
			c.Logger().Errorf("OOPS! cannot save rolemap id %v: %v", rmID, err)
			c.Flash().Add("danger", t(c, "oops.cannot.proceed.acception"))
		} else {
			c.Flash().Add("success", t(c, "request.accepted.successfully"))
		}
	}
	return c.Redirect(http.StatusFound, "/apps/%s", appID)
}

// Request creates role assignments for the member's request.
func (v RolesResource) Request(c buffalo.Context) error {
	var roleIDs []string
	if err := c.Request().ParseForm(); err == nil {
		roleIDs = c.Request().Form["role_id"]
	}
	c.Logger().Debugf("%v roles are requested", len(roleIDs))

	tx := c.Value("tx").(*pop.Connection)
	member := currentMember(c)
	if !member.IsActive {
		c.Flash().Add("danger", t(c, "eep.how.can.you.reach.here"))
		return c.Redirect(http.StatusFound, "/membership/me")
	}
	for _, rID := range roleIDs {
		role := &models.Role{}
		err := tx.Find(role, rID)
		if err != nil {
			c.Logger().Error("OOPS! role not found! ", err)
			c.Flash().Add("danger", t(c, "oops.cannot.found.the.role"))
			break
		}

		err = member.AddRole(tx, role)
		if err != nil {
			tx.TX.Rollback()
			c.Logger().Errorf("cannot assign a role %v to %v. error: %v",
				role, member, err)
			c.Flash().Add("danger", t(c, "cannot.add.a.role"))
			break
		}
		c.Flash().Add("success", t(c, "role.request.finished.successfully"))
	}
	return c.Redirect(http.StatusFound, "/membership/me")
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
			c.Logger().Debug("no assigned role remind. revoke automatically.")
			app := &models.App{}
			err := tx.Find(app, role.AppID)
			if err != nil {
				c.Logger().Warnf("cannot found app with id '%v'", role.AppID)
			} else {
				c.Logger().Debugf("trying to revoke %v@%v", member, app)
				err = app.Revoke(tx, member)
				if err != nil {
					c.Logger().Warnf("cannot revoke %v@%v.", member, app)
				} else {
					c.Logger().Info("no remining roles, revoked.")
				}
			}
		}
		c.Flash().Add("success", t(c, "role.removed.from.you.successfully"))
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/membership/me")
}
