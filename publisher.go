package final

import (
	"context"
	"sync"

	"github.com/xyctruth/final/message"

	"github.com/sirupsen/logrus"
)

// publisher 发送消息到消息队列中
type publisher struct {
	logger   *logrus.Entry
	ackers   []*acker
	ack      chan uint64
	nack     chan uint64
	pending  sync.Map
	sequence uint64
	bus      *Bus
}

func newPublisher(bus *Bus) *publisher {
	return &publisher{
		logger: bus.logger.WithFields(logrus.Fields{
			"module": "publisher",
		}),
		ack:    make(chan uint64, 10000),
		nack:   make(chan uint64, 10000),
		ackers: make([]*acker, 0),
		bus:    bus,
	}
}

func (p *publisher) Start(ctx context.Context) error {
	p.bus.mqProvider.NotifyConfirm(p.ack, p.nack)

	p.logger.Info("Publisher start success")

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.logger.Info("Publisher stop success")
				return
			}
		}
	}()

	return nil
}

func (p *publisher) publish(msgs ...*message.Message) {
	for _, msg := range msgs {
		err := p.bus.mqProvider.Publish(msg)
		if err != nil {
			p.logger.WithError(err).Error("mqProvider publish failure")
			continue
		}

		if msg.Policy.Confirm {
			p.sequence++
			p.pending.Store(p.sequence, msg.Header.Get("record_id"))
		}
	}
}

func (p *publisher) confirm(ack uint64) error {
	recordID, _ := p.pending.Load(ack)
	p.logger.
		WithField("ack", ack).
		WithField("recordID", recordID).
		WithField("pending", p.pending).
		Info("ack received")

	if recordID == nil {
		go func() {
			p.ack <- ack
		}()
	}

	if err := p.bus.outbox.done(nil, recordID); err != nil {
		p.logger.WithError(err).
			WithField("ack", ack).
			WithField("recordID", recordID).
			Error("Failed to delete record")
	}

	p.pending.Delete(ack)
	return nil
}
