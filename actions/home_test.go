package actions_test

//* Test coverage: 100%

import (
	"net/http"
)

func (as *ActionSuite) Test_HomeHandler_A_BeforeLoggedIn() {
	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Welcome to")
	as.Contains(res.Body.String(), "Login")
}

func (as *ActionSuite) Test_HomeHandler_B_AfterLoggedIn() {
	as.setupMembers()
	as.loginAs(other)

	res := as.HTML("/").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), other.Name)
}

func (as *ActionSuite) Test_LoginHandler() {
	res := as.HTML("/login").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Use your social ID")
}

func (as *ActionSuite) Test_LogoutHandler() {
	as.setupMembers()
	as.loginAs(other)

	res := as.HTML("/logout").Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))
}
