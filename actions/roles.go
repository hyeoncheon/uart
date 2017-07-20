package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
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
	c.Flash().Add("success", "Role was created successfully")
	return c.Redirect(302, "/apps/%s", role.AppID)
}

// Edit renders a edit formular for a role.
func (v RolesResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	role := &models.Role{}
	err := tx.Find(role, c.Param("role_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("role", role)
	return c.Render(200, r.HTML("roles/edit.html"))
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
	c.Flash().Add("success", "Role was updated successfully")
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
	c.Flash().Add("success", "Role was destroyed successfully")
	return c.Redirect(302, "/apps/%s", role.AppID)
}
