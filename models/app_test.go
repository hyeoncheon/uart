package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/gobuffalo/uuid"

	"github.com/hyeoncheon/uart/models"
)

var app = &models.App{
	Name:        "Test App",
	Code:        "testapp",
	Description: "TestApp for testing",
}

func (ms *ModelSuite) Test_App() {
	ak := app.AppKey
	as := app.AppSecret
	app.GenerateKeyPair()
	ms.NotEqual(ak, app.AppKey)
	ms.NotEqual(as, app.AppSecret)

	app = models.NewApp(app.Name, app.Code, app.Description, "http://localhost", "http://localhost/auth", "image")
	verrs, err := models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	err = app.AddRole(models.DB, "Admin", "admin", "test admin", 1, true)
	ms.NoError(err)
	rl := app.GetRole(models.DB, "admin")
	ms.NotNil(rl)
	ms.Equal("Admin", rl.Name)

	brl := app.GetRole(models.DB, "superman")
	ms.Equal(uuid.Nil, brl.ID)

	ms.Equal(1, len(*app.GetRoles()))
	err = app.AddRole(models.DB, "Manager", "manager", "test manager", 2, true)
	ms.NoError(err)
	ms.Equal(2, len(*app.GetRoles()))

	err = member.Grant(models.DB, app, "profile")
	ms.NoError(err)
	ms.Equal(1, app.GrantsCount())

	member.AddRole(models.DB, rl)
	rm := app.Requests()
	ms.Equal(1, len(*rm))
	ms.Equal(1, app.RequestsCount())

	ap := models.GetAppByCode("testapp")
	ms.NotNil(ap)
	ms.Equal(app.Name, ap.Name)

	ap = nil
	ap = models.GetAppByKey(app.AppKey)
	ms.NotNil(ap)
	ms.Equal(app.Name, ap.Name)
}
