package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

var message1 = &models.Message{
	Subject:  "subject of message",
	Content:  "content of message",
	AppCode:  "uart",
	Facility: models.MsgFacMesg,
	Priority: models.MsgPriWarn,
	IsLog:    true,
}

func (ms *ModelSuite) Test_Message() {
	ms.Equal("Warn:subject of message", message1.String())
	ms.Equal("Warn", message1.PriorityString())

	app = models.NewApp(app.Name, app.Code, app.Description, "http://localhost", "http://localhost/auth")
	verrs, err := models.DB.ValidateAndCreate(app)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	message1.AppCode = app.Code
	ms.Equal(app.Name, message1.AppName())

	verrs, err = models.DB.ValidateAndCreate(member)
	ms.NoError(err)
	ms.False(verrs.HasAny())

	message1.MemberID = member.ID
	ms.Equal(member.Name, message1.Owner().Name)

	members := &models.Members{}
	models.DB.All(members)
	m := models.NewMessage(models.DB, member.ID, members, nil, message1.Subject, message1.Content, app.Code, models.MsgFacSecu, models.MsgPriAlert, true)
	ms.NotNil(m)

	mm := m.MemberMap(member.ID)
	ms.NotNil(mm)
	ms.Equal(member.ID, mm.MemberID)
}
