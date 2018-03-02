package actions_test

import (
	"net/http"

	"github.com/gobuffalo/uuid"
)

func (as *ActionSuite) Test_Membership_A_Membership() {
	as.setupMembers()
	as.activateMember(appman)

	as.loginAs(admin)

	// membershipHandler() for appman, by admin, allowed
	res := as.HTML("/membership/%v", uuid.Nil).Get()
	as.Equal(http.StatusNotFound, res.Code)

	// membershipHandler() for appman, by admin, allowed
	res = as.HTML("/membership/%v", appman.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), appman.String())
	as.Contains(res.Body.String(), "not yours")

	for _, app := range *appman.GrantedApps() {
		as.Contains(res.Body.String(), app.Name)
	}
	for _, role := range *appman.Roles() {
		as.Contains(res.Body.String(), role.Name)
	}
	for _, cred := range *appman.Credentials() {
		as.Contains(res.Body.String(), cred.Name)
	}

	as.loginAs(other)

	// List() by normal member
	res = as.HTML("/membership/me").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), other.String())
	as.NotContains(res.Body.String(), "not yours")
	as.Contains(res.Body.String(), "locked or new")
}
