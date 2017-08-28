package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

func (ms *ModelSuite) Test_Belonging() {
	verrs, err := models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	// FindMyOwn, AllByOwn
	message1.MemberID = member.ID
	verrs, err = models.DB.ValidateAndCreate(message1)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	message := &models.Message{}
	err = models.FindMyOwn(models.DB.Q(), member, message, message1.ID)
	ms.NoError(err)
	ms.Equal(message1.ID, message.ID)

	messages := &models.Messages{}
	err = models.AllMyOwn(models.DB.Q(), member, messages)
	ms.NoError(err)
	ms.Equal(1, len(*messages))

	// FindMy, AllMy
	verrs, err = models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	app.AddRole(models.DB, "User", "user", "user", 1, false)
	role := app.GetRole(models.DB, "user")
	member.AddRole(models.DB, role, true)

	ap := &models.App{}
	err = models.FindMy(models.DB.Q(), member, ap, app.ID)
	ms.NoError(err)
	ms.Equal(app.ID, ap.ID)

	apps := &models.Apps{}
	err = models.AllMy(models.DB.Q(), member, apps)
	ms.NoError(err)
	ms.Equal(1, len(*apps))
}
