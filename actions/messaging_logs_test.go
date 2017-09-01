package actions_test

//* Test coverage: 100%

import (
	"net/http"

	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/willie"
)

var mlogTemplate = models.MessagingLog{
	Status:  "sent",
	Subject: "Testing one two three",
}

func (as *ActionSuite) Test_MessagingLogsResource() {
	as.setupMembers()
	as.loginAs(other)

	mlog := mlogTemplate
	as.DB.Save(&mlog)

	// List(), denied by admin protect
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/messaging_logs").Get()
	})

	// Destroy(), denied by admin protect
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/messaging_logs/%v", mlog.ID).Delete()
	})

	as.loginAs(admin) //! login as admin

	// List() as admin
	res := as.HTML("/messaging_logs").Get()
	as.Equal(http.StatusOK, res.Code)
	as.Contains(res.Body.String(), mlog.Subject)

	// Destroy() as admin
	res = as.HTML("/messaging_logs/%v", mlog.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/messaging_logs", res.HeaderMap.Get("Location"))
	err := as.DB.Reload(&mlog)
	as.Error(err)
}
