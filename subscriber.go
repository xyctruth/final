package final

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/xyctruth/final/message"
)

// subscriber 启动 Options.SubscriberNum 个 goroutine 订阅消息队列中的消息 使用router处理消息
type subscriber struct {
	logger *logrus.Entry
	id     string
	bus    *Bus
}

func newSubscriber(id string, bus *Bus) *subscriber {
	s := &subscriber{
		logger: bus.logger.WithFields(logrus.Fields{
			"module":        "subscriber",
			"subscriber_id": id,
		}),
		id:  id,
		bus: bus,
	}
	return s
}

func (subscriber *subscriber) Start(ctx context.Context) error {
	msgs := make(chan *message.Message)
	err := subscriber.bus.mqProvider.Subscribe(context.Background(), subscriber.id, msgs)
	if err != nil {
		subscriber.logger.WithError(err).Error("Subscriber start failure")
		return err
	}
	subscriber.logger.Info("Subscriber start success")
	go func() {
		for {
			select {
			case <-ctx.Done():
				subscriber.logger.Info("Subscriber stop success")
				return
			case msg := <-msgs:
				subscriber.processMessage(msg)
			}
		}
	}()

	return nil
}

func (subscriber *subscriber) processMessage(msg *message.Message) {
	subscriber.logger.Info("processMessage")
	err := subscriber.bus.router.handle(msg)
	if err != nil {
		msg.Reject()
		subscriber.logger.WithError(err).Error("Handle failure")
		return
	}
	msg.Ack()
}
