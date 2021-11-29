package actions_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/suite/v4"
	"github.com/gofrs/uuid"

	"github.com/hyeoncheon/uart/actions"
	"github.com/hyeoncheon/uart/models"
)

type ActionSuite struct {
	*suite.Action
}

func Test_ActionSuite(t *testing.T) {
	as := &ActionSuite{suite.NewAction(actions.App())}
	suite.Run(t, as)
}

var (
	admin  = &models.Member{}
	appman = &models.Member{}
	other  = &models.Member{}
)

func (as *ActionSuite) setupMembers() {
	var err error
	admin, err = models.CreateMember(credAdmin)
	as.NoError(err, "cannot create member admin: %v", admin)
	time.Sleep(1 * time.Second)

	appman, err = models.CreateMember(credMember)
	as.NoError(err, "cannot create member appman: %v", appman)
	other, err = models.CreateMember(credOther)
	as.NoError(err, "cannot create member other: %v", other)
}

func (as *ActionSuite) loginAs(member *models.Member) {
	as.NotEqual(uuid.Nil, member.ID, "member not setted %v", member.Name)
	time.Sleep(1000 * time.Millisecond) // limitation of simulation
	member.Mobile = time.Now().Format(time.RFC3339)[0:7]
	err := as.DB.Save(member)
	as.NoError(err, "simulated login failed for %v/%v: %v", member.Name, member.Mobile, err)
}

func (as *ActionSuite) activateMember(member *models.Member) {
	as.NotEqual(uuid.Nil, member.ID, "member not setted %v", member.Name)
	member.IsActive = true
	err := as.DB.Save(member)
	as.NoError(err, "cannot activate member %v: %v", member, err)
}

type requestFunc func(*ActionSuite) *httptest.Response

func permissionDenied(as *ActionSuite, fn func(*ActionSuite) *httptest.Response) {
	res := fn(as)
	as.Equal(http.StatusFound, res.Code)
	as.Equal("/", res.HeaderMap.Get("Location"))
}
