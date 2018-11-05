package actions_test

//* Test coverage: 100%
// testing by appman, other

import (
	"net/http"

	"github.com/gobuffalo/httptest"
	"github.com/gobuffalo/uuid"

	"github.com/hyeoncheon/uart/models"
)

var credAdmin = &models.Credential{
	Provider:  "dummy",
	UserID:    "A1234567890",
	Name:      "Tony Stark",
	Email:     "tony@example.com",
	AvatarURL: "https://i.imgur.com/6GipSpg.jpg",
}

var credMember = &models.Credential{
	Provider:  "dummy",
	UserID:    "A0123456789",
	Name:      "Black Widow",
	Email:     "bwidow@example.com",
	AvatarURL: "https://i.imgur.com/L5eq3S4.jpg",
}

var credOther = &models.Credential{
	Provider:  "dummy",
	UserID:    "A9876543210",
	Name:      "Peter Parker",
	Email:     "spider@example.com",
	AvatarURL: "https://i.imgur.com/L5eq3S4.jpg",
}

func (as *ActionSuite) Test_CredentialsResource_A_All_As_Admin() {
	as.setupMembers()
	as.loginAs(admin)

	// List()
	res := as.HTML("/credentials").Get()
	as.Equal(200, res.Code)
	as.Contains(res.Body.String(), credAdmin.Provider+"/"+credAdmin.UserID)

	// Destroy(), denied by current rule. cannot delete credential by self.
	cred := (*admin.Credentials())[0]
	as.NotEqual(uuid.Nil, cred.ID)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/credentials/%v", cred.ID).Delete()
	})
	as.Equal(1, admin.CredentialCount())

	// Destroy(), error, not exists
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/credentials/%v", uuid.Nil).Delete()
	})

	// Destroy(), allowed
	cred = (*other.Credentials())[0]
	as.NotEqual(uuid.Nil, cred.ID)
	res = as.HTML("/credentials/%v", cred.ID).Delete()
	as.Equal(http.StatusSeeOther, res.Code)
	as.Equal("/credentials", res.HeaderMap.Get("Location"))
	as.Equal(0, other.CredentialCount())
}

func (as *ActionSuite) Test_CredentialsResource_B_All_As_Other() {
	as.setupMembers()
	as.activateMember(other)
	as.loginAs(other)

	// List(), denied by admin protector
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/credentials").Get()
	})

	// Destroy(), denied by current rule. cannot delete credential by self.
	cred := (*other.Credentials())[0]
	as.NotEqual(uuid.Nil, cred.ID)
	permissionDenied(as, func(*ActionSuite) *willie.Response {
		return as.HTML("/credentials/%v", cred.ID).Delete()
	})
	as.Equal(1, other.CredentialCount())
}
