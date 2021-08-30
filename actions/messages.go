package actions

//! WIP
//* Use Belonging Interface
//* Test coverage: Exported Handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"text/template"

	"github.com/gobuffalo/buffalo"
	"github.com/gobuffalo/flect"
	"github.com/gobuffalo/pop/v5"
	"github.com/pkg/errors"

	"github.com/hyeoncheon/uart/jobs"
	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

// constants for messaging/logging subsystem
const (
	MsgFacCore  = "core"
	MsgFacAuth  = "auth"
	MsgFacApp   = "app"
	MsgFacUser  = "user"
	MsgFacMesg  = "messaging"
	MsgFacCron  = "scheduler"
	MsgFacSecu  = "security"
	MsgPriEmerg = 0 // RESERVED
	MsgPriAlert = 1 // for alert
	MsgPriCrit  = 2 // FATAL
	MsgPriErr   = 3
	MsgPriWarn  = 4
	MsgPriNote  = 5 // for notification
	MsgPriInfo  = 6
	MsgPriDebug = 7
)

// MessagesResource is the resource for the message model
type MessagesResource struct {
	buffalo.Resource
}

// List gets all Messages. GET /messages
func (v MessagesResource) List(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messages := &models.Messages{}
	q := tx.PaginateFromParams(c.Params())
	err := models.AllMy(q, dummyMember(c), messages, false)
	if err != nil {
		return utils.DOOPS(c, "while listing messages (params: %v, error: %v)", c.Params(), err)
	}
	c.Set("messages", messages)
	c.Set("pagination", q.Paginator)
	return c.Render(http.StatusOK, r.HTML("messages/index.html"))
}

// Show gets the data for one Message. GET /messages/{message_id}
func (v MessagesResource) Show(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	message := &models.Message{}
	err := models.FindMy(tx.Q(), dummyMember(c), message, c.Param("message_id"))
	if err != nil {
		c.Flash().Add("danger", t(c, "eep.message.not.found"))
		mLogErr(c, MsgFacSecu, "access violation: message by %v", currentMember(c))
		return c.Redirect(http.StatusFound, "/")
	}
	c.Set("message", message)
	return c.Render(http.StatusOK, r.HTML("messages/show.html"))
}

// New renders the formular for creating a new Message. GET /messages/new
//! NOT USED, do we need to communicate between members or admin to member?
/*
func (v MessagesResource) New(c buffalo.Context) error {
	c.Set("message", &models.Message{})
	return c.Render(200, r.HTML("messages/new.html"))
}
*/

// Create adds a Message to the DB. POST /messages
//! NOT USED, do we need to communicate between members or admin to member?
/*
func (v MessagesResource) Create(c buffalo.Context) error {
	message := &models.Message{}
	err := c.Bind(message)
	if err != nil {
		return errors.WithStack(err)
	}

	tx := c.Value("tx").(*pop.Connection)
	verrs, err := tx.ValidateAndCreate(message)
	if err != nil {
		return errors.WithStack(err)
	}
	if verrs.HasAny() {
		c.Set("message", message)
		c.Set("errors", verrs)
		return c.Render(422, r.HTML("messages/new.html"))
	}
	c.Flash().Add("success", "Message was created successfully")
	return c.Redirect(302, "/messages/%s", message.ID)
}
*/

// Dismiss changes status of message map. GET /messages/{message_id}/dismiss
func (v MessagesResource) Dismiss(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	messageMap := &models.MessageMap{}
	err := tx.BelongsTo(dummyMember(c)).
		Where("message_id = ?", c.Param("message_id")).First(messageMap)
	if err != nil {
		return utils.InvalidAccess(c, "/messages", "while find message: %v, error: %v", c.Param("message_id"), err)
	}

	messageMap.IsRead = true
	err = tx.Save(messageMap)
	if err != nil {
		return utils.DOOPS(c, "while saving message map: %v, error: %v", messageMap, err)
	}
	c.Flash().Add("success", t(c, "message.dismissed"))
	return c.Redirect(http.StatusSeeOther, "/messages")
}

// Destroy deletes a message from the DB. DELETE /messages/{message_id}
// ADMIN PROTECTED
func (v MessagesResource) Destroy(c buffalo.Context) error {
	tx := c.Value("tx").(*pop.Connection)
	message := &models.Message{}
	err := tx.Find(message, c.Param("message_id"))
	if err != nil {
		return utils.InvalidAccess(c, "/messages", "while find message: %v, error: %v", c.Param("message_id"), err)
	}
	err = tx.Destroy(message)
	if err != nil {
		return utils.DOOPS(c, "while deleting message: %v, error: %v", message, err)
	}
	c.Flash().Add("success", "Message was destroyed successfully")
	return c.Redirect(http.StatusSeeOther, "/messages")
}

//** utilities

// mLog* replaces default, simple Logger.* functions with message.
func mLogInfo(c buffalo.Context, fac, form string, args ...interface{}) error {
	c.Logger().WithField("category", fac).Infof(form, args...)
	return mLog(c, MsgPriInfo, fac, form, args...)
}

func mLogWarn(c buffalo.Context, fac, form string, args ...interface{}) error {
	c.Logger().WithField("category", fac).Warnf(form, args...)
	return mLog(c, MsgPriWarn, fac, form, args...)
}

func mLogErr(c buffalo.Context, fac, form string, args ...interface{}) error {
	c.Logger().WithField("category", fac).Errorf(form, args...)
	return mLog(c, MsgPriErr, fac, form, args...)
}

func mLogNote(c buffalo.Context, fac, form string, args ...interface{}) error {
	c.Logger().WithField("category", fac).Infof(form, args...)
	return mLog(c, MsgPriNote, fac, form, args...)
}

func mLogAlert(c buffalo.Context, fac, form string, args ...interface{}) error {
	c.Logger().WithField("category", fac).Errorf(form, args...)
	return mLog(c, MsgPriAlert, fac, form, args...)
}

// mLog create simple, common log message for UART system administrators
func mLog(c buffalo.Context, p int, fac, form string, args ...interface{}) error {
	tx := c.Value("tx").(*pop.Connection)
	mesg := fmt.Sprintf(form, args...)
	rcpts := utils.UARTAdmins(tx)
	m := models.NewMessage(tx, dummyMember(c).ID, rcpts, nil, mesg, "",
		models.ACUART, fac, p, true)
	if m == nil {
		tx.TX.Rollback()
		c.Logger().Error("cannot create new message")
		return errors.New("cannot create new message")
	}
	return nil
}

// appMsg log and create a message for application admins (not system level)
func appMsg(c buffalo.Context, r *models.Members, content, form string, args ...interface{}) error {
	c.Logger().WithField("category", MsgFacApp).Infof(form, args...)
	tx := c.Value("tx").(*pop.Connection)
	mesg := fmt.Sprintf(form, args...)
	m := models.NewMessage(tx, dummyMember(c).ID, r, nil, mesg, content,
		models.ACUART, MsgFacApp, MsgPriNote, false)
	if m == nil {
		tx.TX.Rollback()
		c.Logger().Error("cannot create new message")
		return errors.New("cannot create new message")
	}

	for _, member := range *r {
		jobs.QueueMessaging(member.ID)
	}

	return nil
}

// for template based messages
//
// xMsg create template based message.
func xMsg(c buffalo.Context, r *models.Members, app, fac string, prio int, mesg, tpl string, data interface{}) error {
	tmpl, err := template.ParseFiles(uartHome + "/messages/" + tpl + ".tpl")
	if err != nil {
		c.Logger().WithField("color", "RED").Error("TEMPLATE: PARSE ERROR")
		return err
	}
	buf := &bytes.Buffer{}
	if err := tmpl.Execute(buf, data); err != nil {
		c.Logger().WithField("color", "YELLOW").Error("TEMPLATE: EXEC ERROR")
		return err
	}
	content := buf.String()
	c.Logger().Debugf("------ message content:\n%v\n------", content)

	tx := c.Value("tx").(*pop.Connection)
	m := models.NewMessage(tx, dummyMember(c).ID, r, nil, mesg, content,
		app, fac, prio, false)
	if m == nil {
		tx.TX.Rollback()
		c.Logger().Error("cannot create new formatted message")
		return errors.New("cannot create new formatted message")
	}

	for _, member := range *r {
		jobs.QueueMessaging(member.ID)
	}

	return nil
}

// Inventory used as template inventory for messaging subsystem
type Inventory interface {
	String() string
}

// noteMsg logs and creates info level message with template
func noteMsg(c buffalo.Context, r *models.Members, fac, tpl string, data Inventory) error {
	mesg := flect.Titleize(tpl) + ": " + data.String()
	c.Logger().WithField("category", fac).Info(mesg)
	return xMsg(c, r, models.ACUART, fac, MsgPriNote, mesg, tpl, data)
}

// alertMsg logs and creates warning level message with template
func alertMsg(c buffalo.Context, r *models.Members, fac, tpl string, data Inventory) error {
	mesg := flect.Titleize(tpl) + ": " + data.String()
	c.Logger().WithField("category", fac).Warn(mesg)
	return xMsg(c, r, models.ACUART, fac, MsgPriAlert, mesg, tpl, data)
}
