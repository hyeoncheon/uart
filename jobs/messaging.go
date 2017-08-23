package jobs

//! WIP

import (
	"fmt"
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/models"
)

// default constants
const (
	MessagingWorker = "worker.Messaging"
)

type shooter func(string, *models.Messages) error

var shooters = map[string]shooter{
	"mail": mailer,
}

// RegisterMessaging register this background job handler.
func (h *Handler) RegisterMessaging() error {
	h.Name = MessagingWorker
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

	me := &models.Member{ID: args["member_id"].(uuid.UUID)}

	for _, priority := range []int{models.MsgPriNote, models.MsgPriAlert} {
		logger.Debugf("proceed messages with priority %v...", priority)
		messages := &models.Messages{}
		models.DB.BelongsToThrough(me, models.MessageMaps{}).
			Where("messages.is_log = ?", false).
			Where("message_maps.is_sent = ?", false).
			Where("message_maps.is_read = ?", false).
			Where("messages.priority = ?", priority).All(messages)
		if len(*messages) == 0 {
			logger.Debugf("no message found for priority %v. ignore", priority)
			continue
		}

		messangers := me.Messangers(priority)
		for _, messanger := range *messangers {
			if f, ok := shooters[messanger.Method]; ok {
				f(messanger.Value, messages)
			} else {
				logger.Error("EMERGENCY: CANNOT SHOOT A BALL!!! ", messanger)
				mesg := "EMERGENCY! CANNOT SHOOT A BALL!"
				cont := fmt.Sprintf("Emergency! cannot shoot a ball\n\n%v",
					messanger)
				app := models.GetAppByCode(models.ACUART)
				role := app.GetRole(models.DB, models.RCAdmin)
				rcpts := role.Members(true)
				models.NewMessage(models.DB, me.ID, rcpts, nil, mesg, cont,
					models.ACUART, models.MsgFacMesg, models.MsgPriAlert,
					false)
			}
		}
	}

	countSuccess(MessagingWorker)
	return nil
}

// mailer is a message shooter for mail method.
func mailer(target string, messages *models.Messages) error {
	logger.Debug("sending a mail to ", target)
	if len(*messages) == 1 {
		logger.Debugf("...single message %v", (*messages)[0])
	} else {
		for _, message := range *messages {
			logger.Debugf("...one of messages %v", message)
		}
	}
	return nil
}
