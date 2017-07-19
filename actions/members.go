package actions

import (
	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// MembersResource is the resource for the member model
type MembersResource struct {
	buffalo.Resource
}

// List gets all Members.
// ADMIN PROTECTED
func (v MembersResource) List(c buffalo.Context) error {
	members := &models.Members{}
	searchParams, err := models.All(c, members)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("members", members)
	c.Set("searchParams", searchParams)
	return c.Render(200, r.HTML("members/index.html"))
}

// New renders the formular for creating a new Member.
// TODO implement mail based local authentication based on this.
func (v MembersResource) New(c buffalo.Context) error {
	c.Set("member", &models.Member{})
	return c.Render(200, r.HTML("members/new.html"))
}

// Create adds a Member to the DB.
// TODO implement mail based local authentication based on this.
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

// Edit renders a edit formular for a member.
// ADMIN PROTECTED
func (v MembersResource) Edit(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	err := tx.Find(member, c.Param("member_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("member", member)
	c.Set("memberStatus", models.MemberStatus)
	return c.Render(200, r.HTML("members/edit.html"))
}

// Update changes a member in the DB.
// ADMIN PROTECTED
func (v MembersResource) Update(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	err := tx.Find(member, c.Param("member_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = c.Bind(member)
	if err != nil {
		return errors.WithStack(err)
	}
	verrs, err := tx.ValidateAndUpdate(member)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("member", member)
		c.Set("memberStatus", models.MemberStatus)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("members/edit.html"))
	}
	c.Flash().Add("success", "Member was updated successfully")
	return c.Redirect(302, "/members")
}

// Destroy deletes a member from the DB.
// ADMIN PROTECTED
func (v MembersResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	member := &models.Member{}
	err := tx.Find(member, c.Param("member_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(member)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "Member was destroyed successfully")
	return c.Redirect(302, "/members")
}
