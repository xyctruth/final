package final

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

//  looper
type looper struct {
	outbox    *outbox
	publisher *publisher
	logger    *logrus.Entry
	config    *LooperConfig
}

func newLooper(outbox *outbox, publisher *publisher, config *LooperConfig, logger *logrus.Entry) *looper {
	l := &looper{
		logger: logger.WithFields(logrus.Fields{
			"module": "looper",
		}),
		outbox:    outbox,
		publisher: publisher,
		config:    config,
	}

	return l
}

func (looper *looper) Start(ctx context.Context) error {

	looper.scanning()

	looper.logger.Info("Looper start success")

	go func() {
		loop := time.NewTicker(looper.config.Interval)
		for {
			select {
			case <-ctx.Done():
				looper.logger.Info("Publisher stop success")

				return
			case <-loop.C:
				looper.scanning()
			}
		}
	}()
	return nil
}

func (looper *looper) scanning() {
	msgs, err := looper.outbox.take(nil, int64(looper.config.Offset))
	if err != nil {
		looper.logger.WithError(err).Error("outbox take record failure")
	}
	if msgs != nil && len(msgs) > 0 {
		looper.publisher.publish(msgs...)
	}
}
