package models_test

// Test coverage: 100% (without interface methods)

import (
	"github.com/gobuffalo/uuid"

	"github.com/hyeoncheon/uart/models"
)

var cred = &models.Credential{
	Provider: "dummy",
	UserID:   "dummy-001",
	Name:     "Dummy",
	Email:    "dummy@example.com",
}

func (ms *ModelSuite) Test_Credential() {
	ms.Equal("dummy/dummy-001", cred.String())

	mem, err := models.CreateMember(cred)
	ms.NoError(err)
	ms.Equal(cred.Email, mem.Email)

	ms.Equal(mem.ID, cred.Owner().ID)
	ms.Equal(mem.ID, cred.OwnerID())
}

func (ms *ModelSuite) Test_Credential_Invalid() {
	cred := &models.Credential{
		MemberID: uuid.Nil,
	}
	member := cred.Owner()
	ms.Equal(uuid.Nil, member.ID)
}
