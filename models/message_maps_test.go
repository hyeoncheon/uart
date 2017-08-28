package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/hyeoncheon/uart/models"
)

var mm = &models.MessageMap{
	IsSent: true,
	IsRead: true,
	IsBCC:  false,
}

func (ms *ModelSuite) Test_MessageMap() {
	mstr := mm.String()
	ms.Contains(mstr, `"is_sent":true`)
	// Nothing to test :-(
}
