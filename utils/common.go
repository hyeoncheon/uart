package utils

import (
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
)

// UARTAdmins returns registered administrators of UART app.
func UARTAdmins(tx *pop.Connection) *models.Members {
	if app := models.GetAppByCode(models.ACUART); app != nil {
		if role := app.GetRole(tx, models.RCAdmin); role != nil {
			return role.Members(true)
		}
	}
	return &models.Members{}
}
