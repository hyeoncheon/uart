package utils

// partially tested with actions/misc_test.go

import (
	"fmt"
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/pop"

	"github.com/hyeoncheon/uart/models"
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

// InvalidAccess make a log for system error and return redirect
//  url: url redirect to
//  form: format of the log message
//  data...: arguments for log message
func InvalidAccess(c buffalo.Context, url, form string, data ...interface{}) error {
	return OOPS(c, http.StatusFound, url, "access violation", form, data...)
}

// SOOPS make a log for system error and return redirect
//  form: format of the log message
//  data...: arguments for log message
func SOOPS(c buffalo.Context, form string, data ...interface{}) error {
	origin := "/" // TODO origin check
	return OOPS(c, http.StatusFound, origin, "system", form, data...)
}

// DOOPS make a log for database error and return redirect.
//  form: format of the log message
//  data...: arguments for log message
func DOOPS(c buffalo.Context, form string, data ...interface{}) error {
	origin := "/" // TODO origin check
	return OOPS(c, http.StatusFound, origin, "database", form, data...)
}

// OOPS is a common error log and redirector.
func OOPS(c buffalo.Context, ec int, url, dom, form string, data ...interface{}) error {
	errorID := fmt.Sprintf("%v-%v-%v", dom, c.Value("request_id"), ec)
	l := c.Logger().WithField("domain", dom).WithField("error_id", errorID)
	l.Errorf("OOPS! "+dom+" error: "+form, data...)
	c.Flash().Add("danger", "OOPS! Cannot serve the request!")
	c.Response().Header().Set("X-UART-Error", errorID)
	return c.Redirect(ec, url)
}
