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
	_, app, err := safeSetApp(c)
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
	app.AddRole(tx, "Admin", "admin", "Administrator", 64)
	app.AddRole(tx, "User", "user", "Normal User", 1)
	app.AddRole(tx, "Guest", "guest", "Guest, without any privileges", 0)
	currentMember(c).AddRole(tx, app.GetRole(tx, "admin"))

	c.Flash().Add("success", t(c, "app.was.created.successfully"))
	return c.Redirect(302, "/apps/%s", app.ID)
}

// Edit renders a edit formular for a app.
func (v AppsResource) Edit(c buffalo.Context) error {
	_, app, err := safeSetApp(c)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("app", app)
	return c.Render(200, r.HTML("apps/edit.html"))
}

// Update changes a app in the DB.
func (v AppsResource) Update(c buffalo.Context) error {
	tx, app, err := safeSetApp(c)
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
	tx, app, err := safeSetApp(c)
	if err != nil {
		c.Flash().Add("danger", t(c, "app.not.found.check.your.permission"))
		return c.Redirect(http.StatusTemporaryRedirect, "/apps")
	}
	err = tx.Destroy(app)
	if err != nil {
		c.Flash().Add("danger", t(c, "oops.cannot.delete.app"))
		return c.Redirect(http.StatusTemporaryRedirect, "/apps")
	}
	c.Flash().Add("success", t(c, "app.was.deleted.successfully"))
	return c.Redirect(http.StatusFound, "/apps")
}

// utilities
func safeSetApp(c buffalo.Context) (*pop.Connection, *models.App, error) {
	tx := c.Value("tx").(*pop.Connection)
	app := &models.App{}
	err := pop.Q(tx).
		LeftJoin("roles", "roles.app_id = apps.id").
		LeftJoin("role_maps", "role_maps.role_id = roles.id").
		Where("role_maps.member_id = ?", currentMember(c).ID).
		Where("roles.code = ?", "admin").
		Find(app, c.Param("app_id"))

	/*
		err := tx.BelongsToThrough(currentMember(c), &models.RoleMap{}).
			Find(app, c.Param("app_id"))
	*/
	if err != nil {
		c.Logger().Error("cannot found app with your right: ", err)
	}
	return tx, app, err
}
