package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

var role = &models.Role{
	Name:        "User",
	Code:        models.RCUser,
	Description: "User of the App",
	Rank:        8,
	IsReadonly:  false,
}

func (ms *ModelSuite) Test_Role() {
	app = models.NewApp(app.Name, app.Code, app.Description, "http://localhost", "http://localhost/auth")
	verrs, err := models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	role.AppID = app.ID
	verrs, err = models.DB.ValidateAndCreate(role)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	ms.Equal(app.ID, role.App().ID)
	ms.Equal(app.Name+"."+role.Name, role.String())
	ms.Equal(0, len(*role.Members()))

	verrs, err = models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	err = member.AddRole(models.DB, role, true)
	ms.NoError(err)

	ms.Equal(1, len(*role.Members()))
	ms.Equal(1, len(*role.Members(true)))
	ms.Equal(0, len(*role.Members(false)))
	ms.Equal(1, role.MemberCount(true))
	ms.Equal(0, role.MemberCount(false))

	rl := models.GetAppRole(app.Code, role.Code)
	ms.NotNil(rl)
	ms.Equal(role.Name, rl.Name)

	rm := &models.RoleMap{}
	err = models.DB.Where("role_id = ? AND member_id = ?", role.ID, member.ID).First(rm)
	ms.NoError(err)
	ms.Equal(role.ID, rm.Role().ID)
	ms.Equal(member.ID, rm.Member().ID)
}
