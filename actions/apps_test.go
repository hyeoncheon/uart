package actions_test

//* Test coverage: 100%
// testing by appman(flow), other(deny, grant)
// testing with flow (processFlowAppmanRole)

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/httptest"
	"github.com/gofrs/uuid"

	"github.com/hyeoncheon/uart/models"
)

const (
	AppName = "Testing App"
	AppCode = "testingapp"
)

func (as *ActionSuite) Test_AppsResource_A_All_As_Appman() {
	as.setupMembers()
	as.activateMember(appman)
	processFlowAppmanRole(as)

	as.loginAs(admin) //! login as admin ----------------------------
	uart := models.GetAppByCode(models.ACUART)

	// List() by admin
	res := as.HTML("/apps").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), fmt.Sprintf(`href="/apps/%v/edit`, uart.ID))

	// Edit()
	res = as.HTML("/apps/%v/edit", uart.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")

	// caution: uart is a soft readonly app. this code is just for test.
	uart.SiteURL = "siteurl"
	uart.CallbackURL = "callback"
	res = as.HTML("/apps/%v", uart.ID).Put(uart)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")

	// Show() by admin
	res = as.HTML("/apps/%v", uart.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), uart.Name)

	as.loginAs(appman) //! login as appman --------------------------
	successCreateTestingApp(as)
	app := models.GetAppByCode(AppCode)

	res = as.HTML("/apps").Post(&models.App{})
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// List()
	res = as.HTML("/apps").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), fmt.Sprintf(`href="/apps/%v/edit`, app.ID))
	as.NotContains(res.Body.String(), fmt.Sprintf(`href="/apps/%v/edit`, uart.ID))

	// New()
	res = as.HTML("/apps/new").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "form action=")

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

	// Update(), with invalid value
	app.Name = ""
	res = as.HTML("/apps/%v", app.ID).Put(app)
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// Show()
	res = as.HTML("/apps/%v", app.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), AppName)
	as.Contains(res.Body.String(), "Testing Description")

	// Edit(), denied by ownership function
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v/edit", uart.ID).Get()
	})

	// Update(), denied by ownership function
	uart.Description = "I will rule you"
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Put(uart)
	})

	// Show(), denied by ownership function
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Get()
	})

	// Delete(), denied by ownership function
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Delete()
	})

	res = as.HTML("/apps/%v", app.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/apps", res.HeaderMap.Get("Location"))
}

func (as *ActionSuite) Test_AppsResource_J_All_As_Other() {
	as.setupMembers()
	as.loginAs(other) //! login as other
	as.activateMember(other)

	// List(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps").Get()
	})

	// New(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/new").Get()
	})

	// Create(), denied by role based blocker
	permissionDenied(as, requestCreateTestingApp)

	// Show(), denied by role based blocker
	uart := models.GetAppByCode(models.ACUART)
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Get()
	})

	// Edit(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v/edit", uart.ID).Get()
	})

	// Update(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Put(uart)
	})

	// Destroy(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *httptest.Response {
		return as.HTML("/apps/%v", uart.ID).Delete()
	})
}

func (as *ActionSuite) Test_AppsResource_O_GrantFlow() {
	as.setupMembers()
	as.activateMember(appman)
	as.loginAs(appman) //! login as appman

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

/**/

//** test functions -------------------------------------------------

func requestCreateTestingApp(as *ActionSuite) *httptest.Response {
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
