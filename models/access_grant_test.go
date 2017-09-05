package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
	uuid "github.com/satori/go.uuid"
)

func (ms *ModelSuite) Test_AccessGrant() {
	app = models.NewApp(app.Name, app.Code, app.Description, "http://localhost", "http://localhost/auth")
	verrs, err := models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	verrs, err = models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	member.Grant(models.DB, app, "profile")
	grant := &models.AccessGrant{}
	err = models.DB.Where("member_id = ?", member.ID).First(grant)
	ms.NoError(err)

	gs := grant.String()
	ms.Contains(gs, "Dummy Member")
	ms.Contains(gs, "granted Test App")

	gd := grant.Description()
	ms.Contains(gd, "<p>Dummy Member granted scope <code>profile</code> to Test App")
	ms.Equal(member.ID, grant.Member().ID)
	ms.Equal(app.ID, grant.App().ID)
}

func (ms *ModelSuite) Test_AccessGrant_Invalid() {
	ag := &models.AccessGrant{
		AppID:    uuid.Nil,
		MemberID: uuid.Nil,
	}
	str := ag.String()
	ms.Equal("Broken Access Grant!", str)
}
