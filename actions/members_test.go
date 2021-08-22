package actions_test

//* Test coverage: 100%
// testing create, update, list by admin
// testing mark as deleted, delete permanently
// testing blocking by admin protection

import (
	"net/http"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"
)

func (as *ActionSuite) Test_MembersResource_A_CreateUpdateList() {
	as.setupMembers()
	as.loginAs(admin)

	// List() by admin, allowed
	res := as.HTML("/members").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), admin.Name)
	as.Contains(res.Body.String(), other.Name)

	// Edit() by admin, allowed
	res = as.HTML("/members/%v/edit", other.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), other.Name)
	as.Contains(res.Body.String(), "form action=")

	// Update() by admin, with unacceptable value, denied
	other.Name = ""
	res = as.HTML("/members/%v", other.ID).Put(other)
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// Update() by admin, allowed
	other.Name = "Other Name"
	res = as.HTML("/members/%v", other.ID).Put(other)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))
	as.DB.Reload(other)
	as.Equal("Other Name", other.Name)

	// Update() none existing member
	res = as.HTML("/members/%v/edit", uuid.Nil).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))

	// Update() none existing member
	res = as.HTML("/members/%v", uuid.Nil).Put(other)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))

	// Destroy() none existing member
	res = as.HTML("/members/%v", uuid.Nil).Delete()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))
}

func (as *ActionSuite) Test_MembersResource_B_Delete() {
	as.setupMembers()
	as.loginAs(admin)

	// Destroy() by admin, denied (admin cannot be deleted directly)
	res := as.HTML("/members/%v", admin.ID).Delete()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))

	// Destroy() by admin, allowed, but just mark as deleted
	res = as.HTML("/members/%v", other.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))

	err := as.DB.Reload(other)
	as.NoError(err)
	as.Equal(false, other.IsActive)
	as.Contains(other.Name, "-Deleted")
	cred := (*other.Credentials())[0]
	as.Contains(cred.UserID, "-DLTD")
	as.Equal(0, other.AccessGrantCount())
	as.Equal(0, len((*other.Roles())))

	as.loginAs(admin) //! login as admin again (just for simulator)

	// Destroy() by admin, allowed, delete permanently!
	res = as.HTML("/members/%v", other.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(other)
	as.Error(err)
}

func (as *ActionSuite) Test_MembersResource_J_InvalidAccess() {
	as.setupMembers()
	as.activateMember(other)
	as.loginAs(other)

	// List(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/members").Get()
	})

	// Edit(), denied by role based blocker TODO: open it later?
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/members/%v/edit", other.ID).Get()
	})

	// Update(), denied by role based blocker TODO: open it later?
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/members/%v", other.ID).Put(other)
	})

	// Destroy(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/members/%v", other.ID).Delete()
	})
}
