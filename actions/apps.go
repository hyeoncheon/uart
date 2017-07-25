package actions

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// AppsResource is the resource for the app model
type AppsResource struct {
	buffalo.Resource
}

// List gets all Apps.
func (v AppsResource) List(c buffalo.Context) error {
	apps := &models.Apps{}
	searchParams, err := models.All(c, apps)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("apps", apps)
	c.Set("searchParams", searchParams)
	return c.Render(200, r.HTML("apps/index.html"))
}

// Show gets the data for one App.
func (v AppsResource) Show(c buffalo.Context) error {
	_, app, err := safeSetAppAdmin(c)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("app", app)
	c.Set("roles", app.GetRoles())
	c.Set("role", &models.Role{AppID: app.ID})
	return c.Render(200, r.HTML("apps/show.html"))
}

// New renders the formular for creating a new App.
func (v AppsResource) New(c buffalo.Context) error {
	c.Set("app", &models.App{})
	return c.Render(200, r.HTML("apps/new.html"))
}

// Create adds a App to the DB.
func (v AppsResource) Create(c buffalo.Context) error {
	app := &models.App{}
	err := c.Bind(app)
	if err != nil {
		return errors.WithStack(err)
	}
	app.GenerateKeyPair()

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(app)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("app", app)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("apps/new.html"))
	}

	// set default roles
	app.AddRole(tx, "Admin", "admin", "Administrator", 64, true)
	app.AddRole(tx, "User", "user", "Normal User", 1, true)
	app.AddRole(tx, "Guest", "guest", "Guest, No Privileges", 0, true)
	currentMember(c).AddRole(tx, app.GetRole(tx, "admin"))

	c.Flash().Add("success", t(c, "app.was.created.successfully"))
	return c.Redirect(302, "/apps/%s", app.ID)
}

// Edit renders a edit formular for a app.
func (v AppsResource) Edit(c buffalo.Context) error {
	_, app, err := safeSetAppAdmin(c)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("app", app)
	return c.Render(200, r.HTML("apps/edit.html"))
}

// Update changes a app in the DB.
func (v AppsResource) Update(c buffalo.Context) error {
	tx, app, err := safeSetAppAdmin(c)
	if err != nil {
		c.Flash().Add("danger", t(c, "app.not.found.check.your.permission"))
		return c.Redirect(http.StatusTemporaryRedirect, "/apps")
	}
	err = c.Bind(app)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(app)
	if err != nil {
		c.Flash().Add("danger", t(c, "oops.cannot.update.app"))
		return c.Redirect(http.StatusTemporaryRedirect, "/apps")
	}
	if verrs.HasAny() {
		c.Set("app", app)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("apps/edit.html"))
	}
	c.Flash().Add("success", t(c, "app.was.updated.successfully"))
	return c.Redirect(http.StatusFound, "/apps/%s", app.ID)
}

// Destroy deletes a app from the DB.
func (v AppsResource) Destroy(c buffalo.Context) error {
	tx, app, err := safeSetAppAdmin(c)
	if err != nil {
		c.Flash().Add("danger", t(c, "app.not.found.check.your.permission"))
		return c.Redirect(http.StatusFound, "/apps")
	}

	adminMember := currentMember(c)
	err = adminMember.RemoveRole(tx, app.GetRole(tx, "admin"))
	if err != nil {
		tx.TX.Rollback()
		c.Logger().Errorf("cannot remove admin role from member")
		c.Flash().Add("danger", t(c, "cannot.remove.admin.role.from.you"))
		return c.Redirect(http.StatusFound, "/apps")
	}

	err = app.Revoke(tx, adminMember)
	if err != nil {
		tx.TX.Rollback()
		c.Logger().Errorf("cannot revoke access right for admin")
		c.Flash().Add("danger", t(c, "cannot.revoke.access.right.for.you"))
		return c.Redirect(http.StatusFound, "/apps")
	}

	if cnt, _ := tx.BelongsTo(app).Count(&models.AccessGrants{}); cnt > 0 {
		tx.TX.Rollback()
		c.Logger().Errorf("cannot delete the app. %v user(s) exists!", cnt)
		c.Flash().Add("danger", t(c, "cannot.delete.the.app.user.exists"))
		return c.Redirect(http.StatusFound, "/apps")
	}

	// cleaning role and associated role maps.
	// it can be forced because there is no granted users (checked above)
	for _, role := range *app.GetRoles() {
		rolemaps := &[]models.RoleMap{}
		tx.Where("role_id = ?", role.ID).All(rolemaps)
		for _, rm := range *rolemaps {
			c.Logger().Infof("delete rmap %v: %v", rm.ID, tx.Destroy(&rm))
		}
		c.Logger().Infof("delete role %v: %v", role.ID, tx.Destroy(&role))
	}

	err = tx.Destroy(app)
	if err != nil {
		tx.TX.Rollback()
		c.Logger().Errorf("cannot delete the app %v: %v", app, err)
		c.Flash().Add("danger", t(c, "oops.cannot.delete.app"))
		return c.Redirect(http.StatusFound, "/apps")
	}
	c.Logger().Infof("app %v deleted completely", app)
	c.Flash().Add("success", t(c, "app.was.deleted.successfully"))
	return c.Redirect(http.StatusFound, "/apps")
}

// Grant adds a grant for the app to current member and set guest role.
func (v AppsResource) Grant(c buffalo.Context) error {
	// escape route first.
	origin := "/"
	if orig, ok := c.Session().Get("origin").(string); ok {
		origin = orig
	}
	c.Session().Delete("origin")
	c.Session().Save()

	app := models.GetAppByKey(c.Param("key"))
	if app == nil {
		c.Logger().Error("OOPS! cannot found app with key: ", c.Param("key"))
		c.Flash().Add("danger", t(c, "cannot.found.app"))
		return c.Redirect(http.StatusTemporaryRedirect, "%s", origin)
	}
	member := currentMember(c)
	tx := c.Value("tx").(*pop.Connection)

	err := app.Grant(tx, member)
	if err != nil {
		tx.TX.Rollback()
		c.Logger().Errorf("cannot grant %v to %v: %v", app, member, err)
		c.Flash().Add("danger", t(c, "oops.cannot.grant"))
		return c.Redirect(http.StatusTemporaryRedirect, "%s", origin)
	}
	c.Logger().Infof("app %v granted to member %v", app, member)

	guest := app.GetRole(tx, "guest")
	err = member.AddRole(tx, guest)
	if err != nil {
		tx.TX.Rollback()
		c.Logger().Error("cannot add a role to user: ", err)
		c.Flash().Add("danger", t(c, "oops.cannot.assign.a.role"))
		return c.Redirect(http.StatusTemporaryRedirect, "%s", origin)
	}
	c.Logger().Infof("role %v added to %v", guest, member)

	return c.Redirect(http.StatusTemporaryRedirect, "%s", origin)
}

// Revoke serve /revoke/{app_id} to revoke access grant for the current member
func (v AppsResource) Revoke(c buffalo.Context) error {
	tx, app, err := safeSetApp(c)
	if err != nil {
		c.Flash().Add("warning", t(c, "cannot.revoke.your.access.right"))
		return c.Redirect(http.StatusTemporaryRedirect, "/membership/me")
	}
	member := currentMember(c)
	err = app.Revoke(tx, member)
	if err != nil {
		return errors.New("cannot revoke")
	}
	return c.Redirect(http.StatusTemporaryRedirect, "/membership/me")
}

// utilities
func safeSetAppAdmin(c buffalo.Context) (*pop.Connection, *models.App, error) {
	tx := c.Value("tx").(*pop.Connection)
	app := &models.App{}
	err := pop.Q(tx).
		LeftJoin("roles", "roles.app_id = apps.id").
		LeftJoin("role_maps", "role_maps.role_id = roles.id").
		Where("role_maps.member_id = ?", currentMember(c).ID).
		Where("roles.code = ?", "admin").
		Find(app, c.Param("app_id"))
	if err != nil {
		c.Logger().Error("cannot found app with your right: ", err)
		err = errors.New("App Not Found")
	}
	return tx, app, err
}

func safeSetApp(c buffalo.Context) (*pop.Connection, *models.App, error) {
	tx := c.Value("tx").(*pop.Connection)
	app := &models.App{}
	err := pop.Q(tx).
		LeftJoin("roles", "roles.app_id = apps.id").
		LeftJoin("role_maps", "role_maps.role_id = roles.id").
		Where("role_maps.member_id = ?", currentMember(c).ID).
		Where("roles.code != ?", "admin").
		Find(app, c.Param("app_id"))
	if err != nil {
		c.Logger().Error("cannot found app with your right: ", err)
		err = errors.New("App Not Found")
	}
	return tx, app, err
}
