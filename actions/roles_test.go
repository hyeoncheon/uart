package actions_test

// Test coverage: 100%

import (
	"net/http"
	"net/url"

	"github.com/markbates/willie"
	uuid "github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/models"
)

var roleTemplate = models.Role{
	Name: "Tester",
	Code: "tester",
	Rank: 8,
}

func (as *ActionSuite) Test_RolesResource_A_Protected() {
	as.setupMembers()
	as.loginAs(other)

	role := roleTemplate

	// Create(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles").Post(&role)
	})

	existingRole := models.GetAppRole(models.ACUART, models.RCUser)
	as.NotEqual(existingRole.ID, uuid.Nil)

	// Update(), denied by role based blocker
	existingRole.Name = "tester"
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", existingRole.ID).Put(existingRole)
	})

	// Destroy(), denied by role based blocker
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", existingRole.ID).Delete()
	})

	roleRequest := url.Values{}
	roleRequest.Add("role_id", existingRole.ID.String())

	// Request() by normal member, not allowed (inactive user)
	res := as.HTML("/request/roles").Post(&roleRequest)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	// Retire() by normal member, not allowed (inactive user)
	res = as.HTML("/request/roles/%v/retire", existingRole.ID).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
}

func (as *ActionSuite) Test_RolesResource_B_AsAppMan() {
	as.setupMembers()
	processFlowAppmanRole(as)

	as.loginAs(appman) //! login as appman

	successCreateTestingApp(as)
	app := models.GetAppByCode(AppCode)

	rolem := roleTemplate
	role := &rolem

	// Create() by appman, without AppID
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles").Post(role)
	})

	// Create() by appman, with AppID but no right for the app
	role.AppID = models.GetAppByCode(models.ACUART).ID
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles").Post(role)
	})

	// Create() by appman, with AppID and right
	role.AppID = app.ID
	res := as.HTML("/roles").Post(role)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")
	// then
	role = app.GetRole(as.DB, "tester")
	as.NotEqual(uuid.Nil, role.ID)
	as.Equal("Tester", role.Name)

	// Update() by appman
	as.Equal("Tester", role.Name)
	role.Name = "Perfect Tester"
	res = as.HTML("/roles/%v", role.ID).Put(role)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")
	// then
	role = app.GetRole(as.DB, "tester")
	as.NotEqual(uuid.Nil, role.ID)
	as.Equal("Perfect Tester", role.Name)

	// Update() with invalid id
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", uuid.Nil).Put(role)
	})

	// Update() with invalid value
	role.Name = ""
	res = as.HTML("/roles/%v", role.ID).Put(role)
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// Create() by appman, with AppID and right but data error
	role.Name = ""
	res = as.HTML("/roles").Post(role)
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// Update() on role of others, denied
	UARTRole := models.GetAppRole(models.ACUART, models.RCUser)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", UARTRole.ID).Put(UARTRole)
	})

	// Update() on read only role, denied
	userRole := app.GetRole(as.DB, models.RCUser)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", userRole.ID).Put(userRole)
	})

	// Destroy() by appman
	res = as.HTML("/roles/%v", role.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")

	// Destroy() on read only role, denied
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", userRole.ID).Delete()
	})

	// Destroy() on role of others, denied
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", UARTRole.ID).Delete()
	})

	// Destroy() with invalid id
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/%v", uuid.Nil).Delete()
	})

	res = as.HTML("/apps/%v", app.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "User")
	as.NotContains(res.Body.String(), "Perfect Tester")
}

func (as *ActionSuite) Test_RolesResource_C_RoleRequestCycle() {
	as.setupMembers()
	processFlowAppmanRole(as)

	as.loginAs(appman)

	successCreateTestingApp(as)
	app := models.GetAppByCode(AppCode)
	as.NotEqual(uuid.Nil, app.ID)
	role := app.GetRole(as.DB, models.RCUser)
	as.NotEqual(uuid.Nil, role.ID)

	as.loginAs(other) //! login as other and request a role
	as.activateMember(other)

	// Request() by normal member
	roleRequest := url.Values{}
	roleRequest.Add("role_id", role.ID.String())
	// then
	res := as.HTML("/request/roles").Post(&roleRequest)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	roleMap := &models.RoleMap{}
	err := as.DB.Where("role_id = ? AND member_id = ?", role.ID, other.ID).First(roleMap)
	as.NoError(err)
	as.Equal(roleMap.RoleID, role.ID)
	as.Equal(false, roleMap.IsActive) // inactive request

	// Request() by normal member
	uartRole := models.GetAppRole(models.ACUART, models.RCUser)
	uartRoleRequest := url.Values{}
	uartRoleRequest.Add("role_id", uartRole.ID.String())
	// then
	res = as.HTML("/request/roles").Post(&uartRoleRequest)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	uartRoleMap := &models.RoleMap{}
	err = as.DB.Where("role_id = ? AND member_id = ?", uartRole.ID, other.ID).First(uartRoleMap)
	as.NoError(err)
	as.Equal(uartRoleMap.RoleID, uartRole.ID)
	as.Equal(false, uartRoleMap.IsActive) // inactive request

	as.loginAs(appman) //! login as appman and accept request

	// Accept() with invalid rolemap
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/accept/%v/%v", app.ID, uuid.Nil).Get()
	})

	// Accept() with invalid appID
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/accept/%v/%v", uuid.Nil, roleMap.ID).Get()
	})

	// Accept() on others app
	uart := models.GetAppByCode(models.ACUART)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/roles/accept/%v/%v", uart.ID, uartRoleMap.ID).Get()
	})

	// Accept() by appman
	res = as.HTML("/roles/accept/%v/%v", app.ID, roleMap.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")
	err = as.DB.Where("role_id = ? AND member_id = ?", role.ID, other.ID).First(roleMap)
	as.NoError(err)
	as.Equal(roleMap.RoleID, role.ID)
	as.Equal(true, roleMap.IsActive) // inactive request

	as.loginAs(other) //! login as other and retire a role

	// Retire() by normal member, allowed
	res = as.HTML("/request/roles/%v/retire", role.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	as.Equal(0, len(*other.AppRoles(app.ID, true)))

	// Retire() by normal member, invalid ID
	res = as.HTML("/request/roles/%v/retire", uuid.Nil).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))
}

/**/

//** test functions -------------------------------------------------

// used by Test_AppsResource_A_All_As_Appman
// used by Test_AppsResource_O_GrantFlow
func processFlowAppmanRole(as *ActionSuite) {
	as.activateMember(appman)
	as.loginAs(appman) //! login as appman

	role := models.GetAppRole(models.ACUART, models.RCAppMan)
	as.NotEqual(uuid.Nil, role.ID, "cannot get appman role")
	roleRequest := url.Values{}
	roleRequest.Add("role_id", role.ID.String())

	res := as.HTML("/request/roles").Post(roleRequest)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	roleMap := &models.RoleMap{}
	err := as.DB.Where("role_id = ? AND member_id = ?", role.ID, appman.ID).First(roleMap)
	as.NoError(err)
	as.Equal(false, roleMap.IsActive) // inactive request, OK

	// Accept() by normal user not allowed
	uart := models.GetAppByCode(models.ACUART)
	res = as.HTML("/roles/accept/%v/%v", uart.ID, roleMap.ID).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	as.loginAs(admin) //! login as admin and accept role request

	// Accept() by admin
	res = as.HTML("/roles/accept/%v/%v", uart.ID, roleMap.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Contains(res.HeaderMap.Get("Location"), "/apps/")

	err = as.DB.Where("role_id = ? AND member_id = ?", role.ID, appman.ID).First(roleMap)
	as.NoError(err)
	as.Equal(true, roleMap.IsActive) // activated rolemap, OK
}
