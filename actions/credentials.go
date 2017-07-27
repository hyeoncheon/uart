package actions

// TODO REVIEW REQUIRED

import (
	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// CredentialsResource is the resource for the credential model
type CredentialsResource struct {
	buffalo.Resource
}

// List gets all Credentials.
// ADMIN PROTECTED
func (v CredentialsResource) List(c buffalo.Context) error {
	credentials := &models.Credentials{}
	searchParams, err := models.All(c, credentials)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("credentials", credentials)
	c.Set("searchParams", searchParams)
	return c.Render(200, r.HTML("credentials/index.html"))
}

// Destroy deletes a credential from the DB.
func (v CredentialsResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	credential := &models.Credential{}
	err := tx.Find(credential, c.Param("credential_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(credential)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "Credential was destroyed successfully")
	return c.Redirect(302, "/credentials")
}
