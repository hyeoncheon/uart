package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

func (ms *ModelSuite) Test_Common() {
	app := models.NewApp(app.Name, app.Code, app.Description, "http://localhost", "http://localhost/auth")
	verrs, err := models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	apps := &models.Apps{}
	err = models.SelectByAttrs(apps, map[string]interface{}{
		"code": "testapp",
	})
	ms.NoError(err)
	ms.Equal(app.Name, (*apps)[0].Name)

	am := models.Marshal(app)
	ms.Contains(am, `"code":"testapp"`)
}
