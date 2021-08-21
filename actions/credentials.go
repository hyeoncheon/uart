package actions

// TODO REVIEW REQUIRED
//* Test coverage: 100%

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop/v5"

	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

// CredentialsResource is the resource for the credential model
type CredentialsResource struct {
	buffalo.Resource
}

// List gets all Credentials.
//! ADMIN PROTECTED
func (v CredentialsResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	credentials := &models.Credentials{}
	q := tx.PaginateFromParams(c.Params())
	err := models.AllMy(q, dummyMember(c), credentials, false)
	if err != nil {
		return utils.DOOPS(c, "while list creds (params: %v, error: %v)", c.Params(), err)
	}
	c.Set("credentials", credentials)
	c.Set("pagination", q.Paginator)
	return c.Render(http.StatusOK, r.HTML("credentials/index.html"))
}

// Destroy deletes a credential from the DB.
func (v CredentialsResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	credential := &models.Credential{}
	err := tx.Find(credential, c.Param("credential_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "credential.not.found"))
		return c.Redirect(http.StatusFound, "/")
	}

	if credential.MemberID == c.Value("member_id") {
		c.Logger().Warn("credential deletion blocked for ", dummyMember(c))
		c.Flash().Add("warning", "self deletion is not supported now")
		return c.Redirect(http.StatusFound, "/")
	}

	err = tx.Destroy(credential)
	if err != nil {
		return utils.DOOPS(c, "while destroy cred %v, error: %v)", credential, err)
	}
	c.Flash().Add("success", "Credential was deleted successfully")
	mLogWarn(c, MsgFacUser, "credential %v was deleted", credential)
	return c.Redirect(http.StatusSeeOther, "/credentials")
}
