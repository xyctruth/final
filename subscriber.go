package final

import (
	"context"
	"math/rand"
	"time"

	"github.com/Rican7/retry"
	"github.com/Rican7/retry/backoff"
	"github.com/Rican7/retry/jitter"
	"github.com/Rican7/retry/strategy"
	"github.com/sirupsen/logrus"
	"github.com/xyctruth/final/message"
)

// subscriber 启动 Options.NumSubscriber 个 goroutine 订阅消息队列中的消息 使用router处理消息
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
	err := subscriber.bus.mqProvider.Subscribe(ctx, subscriber.id, msgs)
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

	retryAction := func(attempt uint) error {
		return subscriber.bus.router.handle(msg)
	}

	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))

	err := retry.Retry(retryAction,
		// github.com/Rican7/retry v3版本limit包含第一次尝试的次数
		strategy.Limit(subscriber.bus.opt.RetryCount+1),
		strategy.BackoffWithJitter(
			backoff.BinaryExponential(subscriber.bus.opt.RetryInterval),
			jitter.Deviation(random, 0.5),
		))

	if err != nil {
		msg.Reject()
		subscriber.logger.WithError(err).Error("Handle failure")
		return
	}
	msg.Ack()
}
