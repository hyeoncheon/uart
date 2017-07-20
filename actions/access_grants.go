package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// AccessGrantsResource is the resource for the access_grant model
type AccessGrantsResource struct {
	buffalo.Resource
}

// List gets all AccessGrants.
func (v AccessGrantsResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	accessGrants := &models.AccessGrants{}
	err := tx.Order("create_at desc").All(accessGrants)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("accessGrants", accessGrants)
	return c.Render(200, r.HTML("access_grants/index.html"))
}

// Show gets the data for one AccessGrant.
func (v AccessGrantsResource) Show(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	accessGrant := &models.AccessGrant{}
	err := tx.Find(accessGrant, c.Param("access_grant_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("accessGrant", accessGrant)
	return c.Render(200, r.HTML("access_grants/show.html"))
}

// New renders the formular for creating a new AccessGrant.
func (v AccessGrantsResource) New(c buffalo.Context) error {
	c.Set("accessGrant", &models.AccessGrant{})
	return c.Render(200, r.HTML("access_grants/new.html"))
}

// Create adds a AccessGrant to the DB.
func (v AccessGrantsResource) Create(c buffalo.Context) error {
	accessGrant := &models.AccessGrant{}
	err := c.Bind(accessGrant)
	if err != nil {
		return errors.WithStack(err)
	}
	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(accessGrant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("accessGrant", accessGrant)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("access_grants/new.html"))
	}
	c.Flash().Add("success", "AccessGrant was created successfully")
	return c.Redirect(302, "/access_grants/%s", accessGrant.ID)
}

// Edit renders a edit formular for a access_grant.
func (v AccessGrantsResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	accessGrant := &models.AccessGrant{}
	err := tx.Find(accessGrant, c.Param("access_grant_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("accessGrant", accessGrant)
	return c.Render(200, r.HTML("access_grants/edit.html"))
}

// Update changes a access_grant in the DB.
func (v AccessGrantsResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	accessGrant := &models.AccessGrant{}
	err := tx.Find(accessGrant, c.Param("access_grant_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.Bind(accessGrant)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(accessGrant)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("accessGrant", accessGrant)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("access_grants/edit.html"))
	}
	c.Flash().Add("success", "AccessGrant was updated successfully")
	return c.Redirect(302, "/access_grants/%s", accessGrant.ID)
}

// Destroy deletes a access_grant from the DB.
func (v AccessGrantsResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	accessGrant := &models.AccessGrant{}
	err := tx.Find(accessGrant, c.Param("access_grant_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(accessGrant)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "AccessGrant was destroyed successfully")
	return c.Redirect(302, "/access_grants")
}
