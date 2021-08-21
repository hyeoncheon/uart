package models_test

import (
	"github.com/gobuffalo/uuid"
	"github.com/hyeoncheon/uart/models"
)

var member = &models.Member{
	Name:  "Dummy Member",
	Email: "dummy@example.com",
	Icon:  "null.icon",
}

func (ms *ModelSuite) Test_Member() {
	// Nil
	mem := &models.Member{}
	ms.EqualValues(true, mem.IsNil())

	mem, err := models.CreateMember(cred)
	ms.NoError(err)
	ms.Equal(cred.Email, mem.Email)

	verrs, err := models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	mstr := member.String()
	ms.Contains(mstr, "Dummy Member .")

	verrs, err = models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	// Grant(), Granted()
	err = member.Grant(models.DB, app, "profile")
	ms.NoError(err)

	isGranted := member.Granted(app.ID, "profile")
	ms.True(isGranted)

	// Granted() with invalid scope
	isGranted = member.Granted(app.ID, "profile auth")
	ms.False(isGranted)

	grants := member.Grants()
	ms.NotNil(grants)
	ms.Equal(member.ID, (*grants)[0].MemberID)
	ms.Equal(app.ID, (*grants)[0].AppID)

	apps := member.GrantedApps()
	ms.NotNil(apps)
	ms.Equal(app.ID, (*apps)[0].ID)

	grantCount := member.AccessGrantCount()
	ms.Equal(1, grantCount)

	err = member.Revoke(models.DB, app)
	ms.NoError(err)

	isGranted = member.Granted(app.ID, "profile")
	ms.False(isGranted)

	// Role
	role.AppID = app.ID
	verrs, err = models.DB.ValidateAndCreate(role)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	hasRole := member.HasRole(role.ID)
	ms.False(hasRole)

	err = member.AddRole(models.DB, role, true)
	ms.NoError(err)

	hasRole = member.HasRole(role.ID)
	ms.True(hasRole)

	roles := member.AppRoles(app.ID)
	ms.NotNil(roles)
	ms.Equal(1, len(*roles))

	roles = member.Roles()
	ms.NotNil(roles)
	ms.Equal(1, len(*roles))

	roleCodes := member.GetAppRoleCodes(app.Code)
	ms.Equal(1, len(roleCodes))

	// more role
	role16.AppID = app.ID
	verrs, err = models.DB.ValidateAndCreate(role16)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	err = member.AddRole(models.DB, role16, true)
	ms.NoError(err)

	role4.AppID = app.ID
	verrs, err = models.DB.ValidateAndCreate(role4)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	err = member.AddRole(models.DB, role4, true)
	ms.NoError(err)

	roles = member.AppRoles(app.ID)
	ms.NotNil(roles)
	ms.Equal(3, len(*roles))
	ms.True((*roles)[0].Rank > (*roles)[1].Rank)
	ms.True((*roles)[1].Rank > (*roles)[2].Rank)

	// GetAppRoleCodes() with non existing app
	roleCodes = member.GetAppRoleCodes("NoWhereApp")
	ms.Equal(0, len(roleCodes))

	err = member.RemoveRole(models.DB, role)
	ms.NoError(err)

	// RemoveRole() with invalid role
	err = member.RemoveRole(models.DB, role)
	ms.Error(err)

	hasRole = member.HasRole(role.ID)
	ms.False(hasRole)

	// Credentials
	cred.MemberID = member.ID
	verrs, err = models.DB.ValidateAndCreate(cred)
	ms.Error(err) // duplicated

	creds := mem.Credentials()
	ms.Equal(1, len(*creds))

	credCount := mem.CredentialCount()
	ms.Equal(1, credCount)

	// Messaging
	verrs, err = models.DB.ValidateAndCreate(message1)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	mm := &models.MessageMap{
		MemberID:  member.ID,
		MessageID: message1.ID,
	}
	err = models.DB.Save(mm)
	ms.NoError(err)

	err = member.MessageMarkAsSent(message1.ID)
	ms.NoError(err)

	// MessageMarkAsSent() with invalid ID
	err = member.MessageMarkAsSent(uuid.Nil)
	ms.Error(err)

	alert1.MemberID = member.ID
	alert1.IsPrimary = true
	verrs, err = models.DB.ValidateAndCreate(alert1)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	alert2.MemberID = member.ID
	verrs, err = models.DB.ValidateAndCreate(alert2)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	notifier.MemberID = member.ID
	verrs, err = models.DB.ValidateAndCreate(notifier)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	messengers := member.Messengers()
	ms.Equal(3, len(*messengers))

	alerters := member.Messengers(models.MessengerPriority["Alert"])
	ms.Equal(2, len(*alerters))

	a1 := member.PrimaryAlert()
	ms.Equal(alert1.ID, a1.ID)

	n1 := member.PrimaryNotifier()
	ms.NotEqual(notifier.ID, n1.ID)
	ms.Equal(uuid.Nil, n1.ID)

	mid := member.GetID()
	ms.Equal(member.ID, mid)

	// GetMember() with invalid ID
	mem0 := models.GetMember(uuid.Nil)
	ms.Equal(uuid.Nil, mem0.ID)

	mem1 := models.GetMember(member.ID)
	ms.Equal(member.ID, mem1.ID)
}

func (ms *ModelSuite) Test_Member_InvalidAccess() {
	// Revoke() with invalid value
	mem := &models.Member{}
	app := &models.App{}
	err := mem.Revoke(models.DB, app)
	ms.Error(err)
}
