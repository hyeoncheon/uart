package models_test

// Test coverage: 100% (Nothing to test)

import (
	"github.com/hyeoncheon/uart/models"
)

var messagingLog = &models.MessagingLog{
	Subject: "New Member...",
	SentFor: "Tony...",
	Status:  "sent",
}

func (ms *ModelSuite) Test_MessagingLog() {
	m := messagingLog
	ms.Equal(m.Subject+" sent for "+m.SentFor, m.String())
}
