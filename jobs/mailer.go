package jobs

//! WIP

import (
	"time"

	"github.com/gobuffalo/buffalo/worker"
	"github.com/satori/go.uuid"

	"github.com/hyeoncheon/uart/models"
)

// default constants
const (
	MailerWorker = "worker.Mailer"
)

func mailerHandler(args worker.Args) error {
	logger.Debugf("%v invoked with: %v", MailerWorker, models.Marshal(args))
	defer printStatistics(MailerWorker)
	countRunning(MailerWorker)

	m := &models.Member{ID: args["member_id"].(uuid.UUID)}
	messages := &models.Messages{}
	models.DB.BelongsToThrough(m, models.MessageMaps{}).
		Where("messages.is_log = ?", false).
		Where("message_maps.is_sent = ?", false).
		Where("message_maps.is_read = ?", false).
		All(messages)
	for _, m := range *messages {
		logger.Debug("message ", m)
	}
	countSuccess(MailerWorker)
	return nil
}

// QueueMailer enqueues new mailing job of messages for given member.
func QueueMailer(id interface{}) error {
	logger.Debugf("queueing %v for %v...", MailerWorker, id)
	return w.PerformIn(worker.Job{
		Queue:   DefaultQueue,
		Handler: MailerWorker,
		Args: worker.Args{
			"member_id": id,
		},
	}, 30*time.Second)
}

// RegisterMailer register this background job handler.
func (h *Handler) RegisterMailer() error {
	h.Name = MailerWorker
	return w.Register(MailerWorker, mailerHandler)
}
