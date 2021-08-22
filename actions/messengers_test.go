package actions_test

//* Test coverage: 100% of Exported Handlers
// testing owner's flow, admin flow, and denied when requested by the others.
// needed to add private functions

import (
	"net/http"

	"github.com/gofrs/uuid"

	"github.com/hyeoncheon/uart/models"
)

var msgrTemplate = models.Messenger{
	Priority: models.MessengerPriority["Alert"],
	Method:   models.MessengerMethod["Email"],
	Value:    "alert@example.com",
}

func (as *ActionSuite) Test_MessengersResource_A_Owner() {
	as.setupMembers()
	setupTwoMessengerForMember(as, appman)
	prim := appman.PrimaryAlert()
	scnd := &(*appman.Messengers())[1]

	as.loginAs(appman)

	// Create() without Value, denied
	res := as.HTML("/messengers").Post(&models.Messenger{
		Priority: models.MessengerPriority["Alert"],
		Method:   models.MessengerMethod["Email"],
	})
	as.Equal(http.StatusUnprocessableEntity, res.Code)

	// List() by normal member, denied (ADMIN PROTECTED)
	res = as.HTML("/messengers").Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))

	// Update() by owner, allowed
	as.Equal("alert@example.com", prim.Value)
	prim.Value = "alert2@example.com"
	res = as.HTML("/messengers/%v", prim.ID).Put(prim)
	as.Equal(http.StatusSeeOther, res.Code)
	err := as.DB.Reload(prim)
	as.NoError(err)
	as.Equal("alert2@example.com", prim.Value)

	// Update() by owner, but invalid value
	prim.Value = ""
	res = as.HTML("/messengers/%v", prim.ID).Put(prim)
	as.Equal(http.StatusUnprocessableEntity, res.Code)
	err = as.DB.Reload(prim)
	as.NoError(err)
	as.Equal("alert2@example.com", prim.Value)

	// SetPrimary() non primary messenger by owner, allowed
	res = as.HTML("/messengers/%v/setprimary", scnd.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(scnd)
	as.NoError(err)
	as.Equal(true, scnd.IsPrimary)
	err = as.DB.Reload(prim)
	as.NoError(err)
	as.Equal(false, prim.IsPrimary)

	// Destroy() primary messenger, denied
	res = as.HTML("/messengers/%v", scnd.ID).Delete()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(scnd)
	as.NoError(err)
	as.Equal("alert@example.com", scnd.Value)

	// Destroy() non primary messenger by owner, allowed
	res = as.HTML("/messengers/%v", prim.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(prim)
	as.Error(err)
}

func (as *ActionSuite) Test_MessengersResource_B_Other() {
	as.setupMembers()
	setupTwoMessengerForMember(as, appman)
	prim := appman.PrimaryAlert()
	scnd := &(*appman.Messengers())[1]

	as.loginAs(other) //! Action by the others

	// Update() by the others, denied
	prim.Value = "alert3@example.com"
	res := as.HTML("/messengers/%v", prim.ID).Put(prim)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	err := as.DB.Reload(prim)
	as.NoError(err)
	as.Equal("alert@example.com", prim.Value)

	// SetPrimary() non primary messenger by the others, denied
	res = as.HTML("/messengers/%v/setprimary", prim.ID).Get()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))

	err = as.DB.Reload(scnd)
	as.NoError(err)
	as.Equal(false, scnd.IsPrimary)
	err = as.DB.Reload(prim)
	as.NoError(err)
	as.Equal(true, prim.IsPrimary)

	// Destroy() non primary messenger by the others, denied
	res = as.HTML("/messengers/%v", scnd.ID).Delete()
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(scnd)
	as.NoError(err)
	as.Equal("alert@example.com", scnd.Value)
}

func (as *ActionSuite) Test_MessengersResource_C_Admin() {
	as.setupMembers()
	setupTwoMessengerForMember(as, appman)
	prim := appman.PrimaryAlert()
	scnd := &(*appman.Messengers())[1]

	as.loginAs(admin) //! login as admin

	// List() by as admin, allowed
	res := as.HTML("/messengers").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), prim.Value)
	as.Contains(res.Body.String(), scnd.Value)

	// Destroy() primary messenger, allowed
	err := as.DB.Reload(prim)
	as.NoError(err)
	res = as.HTML("/messengers/%v", prim.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/messengers", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(prim)
	as.Error(err)

	// Destroy() second messenger, allowed
	err = as.DB.Reload(scnd)
	as.NoError(err)
	res = as.HTML("/messengers/%v", scnd.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/messengers", res.HeaderMap.Get("Location"))
	err = as.DB.Reload(scnd)
	as.Error(err)
}

func setupTwoMessengerForMember(as *ActionSuite, m *models.Member) {
	as.loginAs(m)

	// Create() by normal member, allowed
	res := as.HTML("/messengers").Post(&msgrTemplate)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	prim := m.PrimaryAlert()
	as.NotEqual(uuid.Nil, prim.ID)
	as.Equal(true, prim.IsPrimary)

	// Create() by normal member, second messenger
	res = as.HTML("/messengers").Post(&msgrTemplate)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	scnd := &(*m.Messengers())[1]
	as.NotEqual(uuid.Nil, scnd.ID)

	// Create() by normal member, notification messenger
	notification := msgrTemplate
	notification.Priority = models.MessengerPriority["Notification"]
	res = as.HTML("/messengers").Post(&notification)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/membership/me", res.HeaderMap.Get("Location"))
	noti := m.PrimaryNotifier()
	as.NotEqual(uuid.Nil, noti.ID)
}
