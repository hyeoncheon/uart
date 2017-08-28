package models_test

// Test coverage: 100% (Nothing to test)

import (
	"github.com/hyeoncheon/uart/models"
)

var messagingLog = &models.MessagingLog{
	Status: "sent",
}

func (ms *ModelSuite) Test_MessagingLog() {
	// Nothing to test
	ms.Equal("sent", messagingLog.Status)
}
