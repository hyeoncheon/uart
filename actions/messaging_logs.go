package actions

//* Test coverage: 100%

import (
	"net/http"

	"github.com/gobuffalo/buffalo"
	"github.com/hyeoncheon/uart/models"
	"github.com/markbates/pop"
	"github.com/pkg/errors"
)

// MessagingLogsResource is the resource for the messaging_log model
type MessagingLogsResource struct {
	buffalo.Resource
}

// List gets all MessagingLogs. GET /messaging_logs
//! ADMIN PROTECTED
func (v MessagingLogsResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messagingLogs := &models.MessagingLogs{}
	q := tx.PaginateFromParams(c.Params())
	err := q.Order("updated_at desc").All(messagingLogs)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Set("messagingLogs", messagingLogs)
	c.Set("pagination", q.Paginator)
	return c.Render(http.StatusOK, r.HTML("messaging_logs/index.html"))
}

// Destroy deletes a messaging_log from the DB. DELETE /messaging_logs/{messaging_log_id}
//! ADMIN PROTECTED
func (v MessagingLogsResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messagingLog := &models.MessagingLog{}
	err := tx.Find(messagingLog, c.Param("messaging_log_id"))
	if err != nil {
		return errors.WithStack(err)
	}
	err = tx.Destroy(messagingLog)
	if err != nil {
		return errors.WithStack(err)
	}
	c.Flash().Add("success", "MessagingLog was destroyed successfully")
	return c.Redirect(http.StatusSeeOther, "/messaging_logs")
}
