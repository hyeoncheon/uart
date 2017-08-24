package jobs

//! WIP

import (
	"fmt"
	"os"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/shurcooL/github_flavored_markdown"
	mailgun "gopkg.in/mailgun/mailgun-go.v1"

	"github.com/hyeoncheon/uart/models"
	"github.com/hyeoncheon/uart/utils"
)

// default constants
const (
	MessagingWorker = "worker.Messaging"
	startMessage    = `Starting UART messaging subsystem...`
	signature       = `UART, an Hyeoncheon Project`
)

type shooter func(models.Messanger, *models.MessagingLog, *models.Messages) error

var mailSender string
var shooters = map[string]shooter{
	"mail": mailer,
}

// RegisterMessaging register this background job handler.
func (h *Handler) RegisterMessaging() error {
	h.Name = MessagingWorker

	mailSender = os.Getenv("MAIL_SENDER")
	logger.Infof("messaging: set %v as mail sender", mailSender)

	adms := utils.UARTAdmins(models.DB)
	for _, adm := range *adms {
		a := adm.PrimaryAlert().Value
		if a == "" {
			logger.Warnf("OOPS! super admin %v has no alerters!", adm)
			continue
		}
		m := prepareMail("Starting UART", startMessage, "", mailSender, a)
		_, _, err := shootMailgun(m)
		if err != nil {
			logger.Fatal("cannot initialize mailer: ", err)
		}
	}

	return w.Register(MessagingWorker, messagingHandler)
}

// QueueMessaging enqueues new mailing job of messages for given member.
func QueueMessaging(id interface{}) error {
	logger.Debugf("queueing %v for %v...", MessagingWorker, id)
	return w.PerformIn(worker.Job{
		Queue:   DefaultQueue,
		Handler: MessagingWorker,
		Args: worker.Args{
			"member_id": id,
		},
	}, 30*time.Second)
}

func messagingHandler(args worker.Args) error {
	logger.Debugf("%v invoked with: %v", MessagingWorker, models.Marshal(args))
	defer printStatistics(MessagingWorker)
	countRunning(MessagingWorker)

	me := models.GetMember(args["member_id"])

	for _, priority := range []int{models.MsgPriNote, models.MsgPriAlert} {
		logger.Debugf("proceed messages with priority %v...", priority)
		messages := &models.Messages{}
		models.DB.BelongsToThrough(me, models.MessageMaps{}).
			Where("messages.is_log = ?", false).
			Where("message_maps.is_sent = ?", false).
			Where("message_maps.is_read = ?", false).
			Where("messages.priority = ?", priority).
			Order("messages.created_at desc").All(messages)
		if len(*messages) == 0 {
			logger.Debugf("no message found for priority %v. ignore", priority)
			continue
		}

		messangers := me.Messangers(priority)
		for _, m := range *messangers {
			messagingLog := &models.MessagingLog{
				Status:  "none",
				Method:  m.Method,
				SentFor: me.String(),
				SentTo:  m.Value,
			}

			if f, ok := shooters[m.Method]; ok {
				if err := f(m, messagingLog, messages); err != nil {
					logger.Errorf("cannot shoot a message with %v (%v)", m, err)
				} else {
					for _, m := range *messages {
						me.MessageMarkAsSent(m.ID)
					}
				}
			} else {
				logger.Error("FATAL: cannot found a shooter: ", m)
				mesg := "FATAL! cannot found a shooter!"
				cont := fmt.Sprintf("FATAL! cannot shoot a ball\n\n%v", m)
				rcpts := utils.UARTAdmins(models.DB)
				models.NewMessage(models.DB, me.ID, rcpts, nil, mesg, cont,
					models.ACUART, models.MsgFacMesg, models.MsgPriAlert,
					false)
			}

			logger.Debugf("messaing: %v", messagingLog)
			if err := models.DB.Save(messagingLog); err != nil {
				logger.Error("cannot save messaging log: ", err)
			}
		}
	}

	countSuccess(MessagingWorker)
	return nil
}

//** Messaging Handler for Email method (currently by mailgun)

// mailer is a message shooter for mail method.
func mailer(messanger models.Messanger, log *models.MessagingLog, messages *models.Messages) error {
	logger.Debug("sending a mail to ", messanger)
	fm := (*messages)[0]
	m := prepareMail(fm.Subject, fm.Content, "", mailSender, messanger.Value)
	if len(*messages) > 1 {
		m.text = "You have messages!\n------\n"
	} else {
		m.text = "You have message!\n------\n"
	}
	for _, message := range *messages {
		logger.Debugf("...one of messages %v", message)
		m.text += "\n" + message.Content + "\n------\n"
	}
	m.html = string(github_flavored_markdown.Markdown([]byte(m.text)))
	m.text += "\n-- \n" + signature
	m.html += string(github_flavored_markdown.Markdown([]byte(signature)))
	if len(*messages) > 1 {
		m.subject += fmt.Sprintf(" (and %v more messages)", len(*messages)-1)
	}
	log.Subject = m.subject

	response, queueID, err := shootMailgun(m)
	if err != nil {
		log.Status = "error"
		logger.Error("cannot send email: ", err)
		return err
	}
	log.Response = response
	log.QueueID = queueID
	log.Status = "sent"
	for _, message := range *messages {
		log.Notes += message.String() + ";\n"
	}
	return nil
}

// shootMailgun sends a message using mailgun API.
func shootMailgun(m *mail) (resp, id string, err error) {
	mg, err := mailgun.NewMailgunFromEnv()
	if err != nil {
		logger.Error("FATAL! cannot setup mailgun from env: ", err)
		return "", "", err
	}
	logger.Infof("about to send a mail [%v] to %v", m.subject, m.rcpt)
	message := mailgun.NewMessage(m.sender, m.subject, m.text, m.rcpt...)
	if len(m.html) > 0 {
		message.SetHtml(m.html)
	}
	for _, bcc := range m.bccs {
		logger.Debugf("add %v as bcc...", bcc)
		message.AddBCC(bcc)
	}
	logger.Infof("shoot the gun...")
	return mg.Send(message)
}

type mail struct {
	subject string
	sender  string
	rcpt    []string
	bccs    []string
	text    string
	html    string
}

func prepareMail(subj, text, html, sndr, rcpt string, bccs ...string) *mail {
	rcpts := []string{rcpt}
	return &mail{
		subject: subj,
		sender:  sndr,
		rcpt:    rcpts,
		bccs:    bccs,
		text:    text,
		html:    html,
	}
}
