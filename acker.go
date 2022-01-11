package final

import (
	"context"

	"github.com/sirupsen/logrus"
)

// acker 启动 Options.AckerNum 个goroutine接收消息队列ack消息后，Done掉 outbox 中的消息记录
type acker struct {
	logger *logrus.Entry

	// 退出信号
	exit chan bool

	id  string
	bus *Bus
}

func newAcker(id string, bus *Bus) *acker {
	s := &acker{
		logger: bus.logger.WithFields(logrus.Fields{
			"module":   "acker",
			"acker_id": id,
		}),
		exit: make(chan bool),
		id:   id,
		bus:  bus,
	}
	return s
}

func (acker *acker) Start(ctx context.Context) error {
	acker.logger.Info("Acker start success")

	go func() {
		for {
			select {
			case <-ctx.Done():
				acker.logger.Info("Acker stop success")
				return
			case ack, ok := <-acker.bus.publisher.ack:
				if !ok {
					acker.logger.Error("acker.ack close")
					return
				}
				err := acker.bus.publisher.confirm(ack)
				if err != nil {
					acker.logger.WithError(err).Error("acker  confirm error")
				}
			case nack, ok := <-acker.bus.publisher.nack:
				if !ok {
					acker.logger.Error("acker.nack close")
					return
				}
				acker.logger.WithField("nack", nack).Error("nack received")
				acker.logger.WithField("channel_len", len(acker.bus.publisher.nack)).Debug("length of nack channel")

			}
		}

	}()
	return nil
}
