package final

import (
	"context"

	"github.com/xyctruth/final/message"

	"github.com/xyctruth/final/mq"

	"github.com/sirupsen/logrus"
)

//  subscriber 启动 SubscriberConfig.Num 个 goroutine 接收 sender的消息后 使用router处理消息
type subscriber struct {
	mqProvider mq.IProvider
	router     *router
	logger     *logrus.Entry
	id         string
}

func newSubscriber(id string, mqProvider mq.IProvider, router *router, logger *logrus.Entry) *subscriber {
	s := &subscriber{
		logger: logger.WithFields(logrus.Fields{
			"module":        "subscriber",
			"subscriber_id": id,
		}),
		mqProvider: mqProvider,
		router:     router,
		id:         id,
	}
	return s
}

func (subscriber *subscriber) Start(ctx context.Context) error {
	msgs := make(chan *message.Message)
	err := subscriber.mqProvider.Subscribe(context.Background(), subscriber.id, msgs)
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
	err := subscriber.router.handle(msg)
	if err != nil {
		msg.Reject()
		subscriber.logger.WithError(err).Error("Handle failure")
		return
	}
	msg.Ack()
}
