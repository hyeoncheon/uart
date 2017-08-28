package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

var alert1 = &models.Messenger{
	Priority: models.MessengerPriority["Alert"],
	Method:   models.MessengerMethod["Email"],
	Value:    "alert@example.com",
}

var alert2 = &models.Messenger{
	Priority: models.MessengerPriority["Alert"],
	Method:   models.MessengerMethod["Email"],
	Value:    "alert2@example.com",
}

var notifier = &models.Messenger{
	Priority: models.MessengerPriority["Notification"],
	Method:   models.MessengerMethod["Email"],
	Value:    "note@example.com",
}

func (ms *ModelSuite) Test_Messenger() {
	ms.EqualValues("Email to alert2@example.com (Alert/false)", alert2.String())
}
