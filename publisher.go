package final

import (
	"context"
	"sync"

	"github.com/xyctruth/final/message"

	"github.com/xyctruth/final/mq"

	"github.com/sirupsen/logrus"
)

// publisher 启动后循环从 outbox 中获取未接收到ack的消息，然后发送到 publisher 中
type publisher struct {
	mqProvider mq.IProvider
	logger     *logrus.Entry

	outbox *outbox

	publishMutex sync.Mutex
	confirmMutex sync.Mutex

	ackers   []*acker
	ack      chan uint64
	nack     chan uint64
	pending  map[uint64]interface{}
	sequence uint64
}

func newPublisher(mqProvider mq.IProvider, outbox *outbox, logger *logrus.Entry) *publisher {
	publisher := &publisher{
		logger: logger.WithFields(logrus.Fields{
			"module": "publisher",
		}),
		pending:    make(map[uint64]interface{}),
		ack:        make(chan uint64, 10000),
		nack:       make(chan uint64, 10000),
		ackers:     make([]*acker, 0),
		mqProvider: mqProvider,
		outbox:     outbox,
	}
	return publisher
}

func (publisher *publisher) Start(ctx context.Context) error {
	publisher.mqProvider.NotifyConfirm(publisher.ack, publisher.nack)

	publisher.logger.Info("Publisher start success")

	go func() {
		for {
			select {
			case <-ctx.Done():
				publisher.logger.Info("Publisher stop success")
				return
			}
		}
	}()

	return nil
}

func (publisher *publisher) publish(msgs ...*message.Message) {
	publisher.publishMutex.Lock()
	defer publisher.publishMutex.Unlock()

	for _, msg := range msgs {
		err := publisher.mqProvider.Publish(msg)
		if err != nil {
			publisher.logger.WithError(err).Error("mqProvider publish failure")
			continue
		}

		if msg.Policy.Confirm {
			publisher.sequence++
			publisher.pending[publisher.sequence] = msg.Header.Get("record_id")

		}
	}
}

func (publisher *publisher) confirm(ack uint64) error {
	publisher.confirmMutex.Lock()
	defer publisher.confirmMutex.Unlock()

	recordID := publisher.pending[ack]
	publisher.logger.
		WithField("ack", ack).
		WithField("recordID", recordID).
		WithField("pending", publisher.pending).
		Info("ack received")

	if recordID == nil {
		go func() {
			publisher.ack <- ack
		}()
	}

	if err := publisher.outbox.done(nil, recordID); err != nil {
		publisher.logger.WithError(err).
			WithField("ack", ack).
			WithField("recordID", recordID).
			Error("Failed to delete record")
	}
	delete(publisher.pending, ack)
	return nil
}
