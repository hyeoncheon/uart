package jobs

import (
	"time"

	"github.com/gobuffalo/buffalo/worker"

	"github.com/hyeoncheon/uart/models"
)

// default constants
const (
	TimerWorker = "worker.Timer"
)

func timerHandler(args worker.Args) error {
	logger.Debugf("%v invoked with: %v", TimerWorker, models.Marshal(args))
	defer printStatistics(TimerWorker)
	countRunning(TimerWorker)
	countSuccess(TimerWorker)
	return QueueTimer(args["sec"].(time.Duration))
}

// QueueTimer enqueues new sample timer job to queue.
func QueueTimer(sec time.Duration) error {
	return w.PerformIn(worker.Job{
		Queue:   DefaultQueue,
		Handler: TimerWorker,
		Args: worker.Args{
			"sec": sec,
		},
	}, sec*time.Second)
}

// RegisterTimer register this background job handler.
func (h *Handler) RegisterTimer() error {
	h.Name = TimerWorker
	if err := w.Register(TimerWorker, timerHandler); err != nil {
		logger.Errorf("cannot register background handler %v", TimerWorker)
	}
	return QueueTimer(300)
}
