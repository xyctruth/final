package final

import (
	"context"

	"github.com/sirupsen/logrus"
)

// acker 启动  AckerConfig.Num 个goroutine接收 publisher 的ack消息后 Done掉 outbox 中的消息记录
type acker struct {
	outbox *outbox
	logger *logrus.Entry

	// 退出信号
	exit chan bool

	id        string
	publisher *publisher
}

func newAcker(id string, outbox *outbox, publisher *publisher, logger *logrus.Entry) *acker {
	s := &acker{
		logger: logger.WithFields(logrus.Fields{
			"module":   "acker",
			"acker_id": id,
		}),
		outbox:    outbox,
		publisher: publisher,
		exit:      make(chan bool),
		id:        id,
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
			case ack, ok := <-acker.publisher.ack:
				if !ok {
					acker.logger.Error("acker.ack close")
					return
				}
				err := acker.publisher.confirm(ack)
				if err != nil {
					acker.logger.WithError(err).Error("acker  confirm error")
				}
			case nack, ok := <-acker.publisher.nack:
				if !ok {
					acker.logger.Error("acker.nack close")
					return
				}
				acker.logger.WithField("nack", nack).Error("nack received")
				acker.logger.WithField("channel_len", len(acker.publisher.nack)).Debug("length of nack channel")

			}
		}

	}()
	return nil
}
