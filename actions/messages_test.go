package actions_test

//* Test coverage: Exported Handlers
// testing list, show, delete and dismiss.
// need to be add private function with additional file.

import (
	"net/http"

	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/willie"
)

func (as *ActionSuite) Test_MessagesResource_A_ListShow_A() {
	as.setupMembers()
	as.loginAs(admin)

	// make member status changed message
	other.IsActive = true
	res := as.HTML("/members/%v", other.ID).Put(other)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))
	as.DB.Reload(other)
	as.Equal(true, other.IsActive)

	as.loginAs(admin) //! login as admin again, simulator limitation
	// List()
	res = as.HTML("/messages").Get()
	as.Equal(http.StatusOK, res.Code)
	as.NotContains(res.Body.String(), "Member Status Changed: ")

	message := &models.Message{}
	err := as.DB.Where("priority = ?", models.MsgPriNote).First(message)
	as.NoError(err)

	// Show() denied, this message is for other not admin
	//! it also generate error log for admin
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/messages/%v", message.ID).Get()
	})

	// List()
	res = as.HTML("/messages").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "access violation: message by ")

	message = &models.Message{}
	err = as.DB.Where("priority = ?", models.MsgPriErr).First(message)
	as.NoError(err)

	// Dismiss()
	res = as.HTML("/messages/%v", message.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
}

func (as *ActionSuite) Test_MessagesResource_A_ListShow_B() {
	as.setupMembers()
	as.loginAs(admin)

	// make member status changed message
	other.IsActive = true
	res := as.HTML("/members/%v", other.ID).Put(other)
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/members", res.HeaderMap.Get("Location"))
	as.DB.Reload(other)
	as.Equal(true, other.IsActive)

	as.loginAs(other) //! login as other, the rcpt.
	// List()
	res = as.HTML("/messages").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Member Status Changed: ")

	message := &models.Message{}
	err := as.DB.Where("priority = ?", models.MsgPriNote).First(message)
	as.NoError(err)

	// Show() with id
	res = as.HTML("/messages/%v", message.ID).Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), "Member Status Changed: ")
	as.Contains(res.Body.String(), "Membership was Activated by Admin")

	// Dismiss()
	res = as.HTML("/messages/%v/dismiss", message.ID).Get()
	as.Equal(http.StatusSeeOther, res.Code)

	// List()
	res = as.HTML("/messages").Get()
	as.Equal(http.StatusOK, res.Code)
	as.NotContains(res.Body.String(), "Member Status Changed: ")
}
