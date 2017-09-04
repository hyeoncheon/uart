package actions_test

import (
	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

func (as *ActionSuite) Test_UARTAdmins() {
	members := utils.UARTAdmins(as.DB)
	as.Equal(0, len(*members))

	as.setupMembers()
	members = utils.UARTAdmins(as.DB)
	as.Equal(1, len(*members))

	adminRole := models.GetAppRole(models.ACUART, models.RCAdmin)
	appman.AddRole(as.DB, adminRole, true)
	members = utils.UARTAdmins(as.DB)
	as.Equal(2, len(*members))
}
