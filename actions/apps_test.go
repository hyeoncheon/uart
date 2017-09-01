package actions_test

//* Test coverage: 100%
// testing by appman(flow), other(deny, grant)
// testing with flow (processFlowAppmanRole)

import (
	"net/http"

	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/willie"
	uuid "github.com/satori/go.uuid"
)

const (
	AppName = "Testing App"
	AppCode = "testingapp"
)

func (as *ActionSuite) Test_AppsResource_A_All_As_Appman() {
	as.setupMembers()
	as.activateMember(appman)
	as.loginAs(appman)

	processFlowAppmanRole(as)
	successCreateTestingApp(as)

	// List()
	res := as.HTML("/apps").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), AppName)

	// New()
	res = as.HTML("/apps/new").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")

	app := models.GetAppByCode(AppCode)
	as.NotEqual(uuid.Nil, app.ID, "cannot found app with code %v", AppCode)
	as.NotEqual("", app.AppKey, "appKey not set! %v", app.AppKey)

	// Edit()
	res = as.HTML("/apps/%v/edit", app.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")

	// Update()
	app.Description = "Testing Description"
	res = as.HTML("/apps/%v", app.ID).Put(app)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")

	// Show()
	res = as.HTML("/apps/%v", app.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), AppName)
	as.Contains(res.Body.String(), "Testing Description")

	res = as.HTML("/apps/%v", app.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/apps", res.HeaderMap.Get("Location"))
}

func (as *ActionSuite) Test_AppsResource_J_All_As_Other() {
	as.setupMembers()
	as.loginAs(other)
	as.activateMember(other)

	// List(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps").Get()
	})

	// New(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps/new").Get()
	})

	// Create(), denied by role based blocker
	permissionDenied(as, requestCreateTestingApp)

	// Show(), denied by role based blocker
	uart := models.GetAppByCode(models.ACUART)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps/%v", uart.ID).Get()
	})

	// Edit(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps/%v/edit", uart.ID).Get()
	})

	// Update(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps/%v", uart.ID).Put(uart)
	})

	// Destroy(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/apps/%v", uart.ID).Delete()
	})
}

func (as *ActionSuite) Test_AppsResource_O_GrantFlow() {
	as.setupMembers()
	as.activateMember(appman)
	as.loginAs(appman)

	processFlowAppmanRole(as)
	successCreateTestingApp(as)

	//! requests resulting failure
	// Grant() by normal member, invalid key
	res := as.HTML("/grant/%v", "FOOBAR").Get()
	as.Equal(http.StatusTemporaryRedirect, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Revoke() by normal member, invalid id
	res = as.HTML("/revoke/%v", uuid.Nil).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	//! successful flow
	app := models.GetAppByCode(AppCode)
	as.NotEqual(uuid.Nil, app.ID, "cannot found app with code %v", AppCode)
	as.NotEqual("", app.AppKey, "appKey not set! %v", app.AppKey)

	// Grant() by normal member, success
	res = as.HTML("/grant/%v", app.AppKey).Get()
	as.Equal(http.StatusTemporaryRedirect, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Revoke() by normal member, success
	res = as.HTML("/revoke/%v", app.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
}

//** test functions -------------------------------------------------

func requestCreateTestingApp(as *ActionSuite) *willie.Response {
	app := &models.App{
		Name:        AppName,
		Code:        AppCode,
		Description: "",
		SiteURL:     "http://localhost:3000",
		CallbackURL: "http://localhost:3000/callback",
	}
	return as.HTML("/apps").Post(app)
}

func successCreateTestingApp(as *ActionSuite) {
	res := requestCreateTestingApp(as)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")
}
