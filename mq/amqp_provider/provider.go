package amqp_provider

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/xyctruth/final/message"

	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"github.com/xyctruth/final/mq"
)

type Provider struct {
	exit chan bool

	log *logrus.Entry

	// conn
	conn *amqp.Connection
	// 初始化channel
	initChannel *amqp.Channel
	// 发布channel
	publishChannel *amqp.Channel
	// 发布channel nowait
	publishNoWaitChannel *amqp.Channel

	//notify ch
	connErrors    chan *amqp.Error
	connBlocks    chan amqp.Blocking
	channelErrors chan *amqp.Error

	svcName string
	topics  []string

	//启动时是否清除
	purge bool

	queue           amqp.Queue
	dlxQueue        amqp.Queue
	queueName       string
	dlxQueueName    string
	dlxExchangeName string
}

func NewProvider(connStr string) (mq.IProvider, error) {
	conn, err := amqp.DialConfig(connStr, amqp.Config{
		Heartbeat: 10 * time.Minute,
	})
	if err != nil {
		return nil, err
	}
	log := &logrus.Logger{}
	return &Provider{
		log: log.WithFields(logrus.Fields{
			"module": "amqp_provider",
		}),
		exit: make(chan bool),
		conn: conn,
	}, nil
}

func (provider *Provider) Init(svcName string, purge bool, topics []string) error {
	var err error

	provider.purge = purge
	provider.topics = topics
	provider.svcName = svcName
	provider.queueName = svcName
	provider.dlxQueueName = fmt.Sprintf("%s_dlx", svcName)
	provider.dlxExchangeName = fmt.Sprintf("%s_exchange_dlx", svcName)

	provider.connErrors = make(chan *amqp.Error)
	provider.connBlocks = make(chan amqp.Blocking)
	provider.conn.NotifyClose(provider.connErrors)
	provider.conn.NotifyBlocked(provider.connBlocks)

	provider.channelErrors = make(chan *amqp.Error)
	if provider.initChannel, err = provider.conn.Channel(); err != nil {
		return err
	}
	provider.initChannel.NotifyClose(provider.channelErrors)

	if provider.publishChannel, err = provider.conn.Channel(); err != nil {
		return err
	}

	err = provider.publishChannel.Confirm(false)
	if err != nil {
		return err
	}

	if provider.publishNoWaitChannel, err = provider.conn.Channel(); err != nil {
		return err
	}

	err = provider.initQueue()
	if err != nil {
		return err
	}

	err = provider.initExchange()
	if err != nil {
		return err
	}

	err = provider.initDlxQueue()
	if err != nil {
		return err
	}

	err = provider.initDlxExchange()
	if err != nil {
		return err
	}

	go provider.monitorAMQPErrors()

	return nil
}

func (provider *Provider) Publish(message *message.Message) error {
	publishing := NewPublishingFromMessage(message)

	var channel *amqp.Channel

	if message.Policy.Confirm {
		channel = provider.publishChannel
	} else {
		channel = provider.publishNoWaitChannel
	}

	return channel.Publish(
		message.Topic, //exchange
		message.Topic, //key
		false,         //开启强制消息投递（mandatory为设置为true），但消息未被路由至任何一个queue，则回退一条消息到channel.NotifyReturn
		false,         //当immediate标志位设置为true时，如果exchange在将消息路由到queue(s)时发现对于的queue上么有消费者，那么这条消息不会放入队列中。当与消息routeKey关联的所有queue（一个或者多个）都没有消费者时，该消息会通过basic.return方法返还给生产者。
		publishing,    //msg
	)
}

func (provider *Provider) NotifyConfirm(ack, nack chan uint64) {
	provider.publishChannel.NotifyConfirm(ack, nack)
}

func (provider *Provider) Subscribe(ctx context.Context, consumerTag string, msgs chan *message.Message) error {
	channel, err := provider.conn.Channel()
	if err != nil {
		provider.log.WithError(err).Error("Failed to get initChannel")
		return err
	}

	channelErrors := make(chan *amqp.Error)
	provider.initChannel.NotifyClose(channelErrors)

	err = channel.Qos(1, 0, false)
	if err != nil {
		provider.log.WithError(err).Error("Failed to set initChannel qos ")
	}

	deliveries, err := provider.initConsumer(consumerTag, channel)
	if err != nil {
		provider.log.WithError(err).Error("Failed to init consumer")
		return err
	}

	go func() {
		for {
			select {
			case <-provider.exit:
				return
			case delivery := <-deliveries:
				msg := NewMessageFromDelivery(delivery)
				select {
				case <-provider.exit:
					return
				case msgs <- msg:
					provider.log.WithField("uuid", msg.UUID).Trace("Handler sent to consumer")
				}

				select {
				case <-provider.exit:
					return
				case <-msg.Acked():
					provider.log.WithField("uuid", msg.UUID).Trace("Handler Ack")
					err := delivery.Ack(false)
					if err != nil {
						provider.log.WithError(err).Error("Failed ack message")
					}
				case <-msg.Rejected():
					provider.log.WithField("uuid", msg.UUID).Trace("Handler rejectch")
					err := delivery.Reject(false)
					if err != nil {
						provider.log.WithError(err).Error("Failed reject message")
					}
				}
			case amqpErr, ok := <-channelErrors:
				provider.log.WithField("amqp_error", amqpErr).Error("Handler Ack")
				if !ok {
					return
				}

			}
		}
	}()

	return nil
}

func (provider *Provider) initConsumer(consumerTag string, channel *amqp.Channel) (<-chan amqp.Delivery, error) {
	deliveries, e := channel.Consume(provider.queueName, /*queue*/
		consumerTag, /*consumer*/
		false,       /*autoAck*/
		false,       /*exclusive*/
		false,       /*noLocal*/
		false,       /*noWait*/
		nil /*args* amqp.Table*/)
	return deliveries, e
}

func (provider *Provider) initQueue() error {
	var err error

	if provider.purge {
		_, err := provider.initChannel.QueueDelete(
			provider.queueName,
			false, /*ifUnused*/
			false, /*ifEmpty*/
			false /*noWait*/)
		if err != nil {
			return err
		}
	}

	args := amqp.Table{"x-dead-letter-exchange": provider.dlxExchangeName}
	provider.queue, err = provider.initChannel.QueueDeclare(
		provider.queueName,
		true,  /*durable*/
		false, /*autoDelete*/
		false, /*exclusive*/
		false, /*noWait*/
		args /*args*/)

	return err
}

func (provider *Provider) initExchange() error {
	for _, topic := range provider.topics {
		err := provider.initChannel.ExchangeDeclare(topic, /*name*/
			"topic", /*kind*/
			true,    //设置是否持久
			false,   //设置是否自动删除
			false,   /*internal*/
			false,   // 当noWait为true时，声明时无需等待服务器的确认
			nil /*args amqp.Table*/)
		if err != nil {
			return err
		}
		err = provider.bindQueue(provider.queueName, topic, topic)
		if err != nil {
			return err
		}

	}
	return nil
}

func (provider *Provider) initDlxQueue() error {
	var err error

	if provider.purge {
		_, err := provider.initChannel.QueueDelete(provider.dlxQueueName, false /*ifUnused*/, false /*ifEmpty*/, false /*noWait*/)
		if err != nil {
			return err
		}
	}

	provider.dlxQueue, err = provider.initChannel.QueueDeclare(
		provider.dlxQueueName,
		true,  /*durable*/
		false, /*autoDelete*/
		false, /*exclusive*/
		false, /*noWait*/
		nil)

	return err
}

func (provider *Provider) initDlxExchange() error {
	err := provider.initChannel.ExchangeDeclare(provider.dlxExchangeName, /*name*/
		"fanout", /*kind*/
		true,     //设置是否持久
		false,    //设置是否自动删除
		false,    /*internal*/
		false,    // 当noWait为true时，声明时无需等待服务器的确认
		nil /*args amqp.Table*/)
	if err != nil {
		return err
	}
	err = provider.bindQueue(provider.dlxQueueName, "", provider.dlxExchangeName)
	if err != nil {
		return err
	}
	return nil
}

func (provider *Provider) bindQueue(queue, topic, exchange string) error {
	return provider.initChannel.QueueBind(queue, topic, exchange, false /*noWait*/, nil /*args*/)
}

func (provider *Provider) monitorAMQPErrors() {
	defer func() {
		if p := recover(); p != nil {
			err := errors.New(fmt.Sprintf("%v\n%s", p, debug.Stack()))
			provider.log.WithError(err).Error("panic monitor amqp errors")
		}
	}()

	for {
		select {
		case <-provider.exit:
			return
		case blocked := <-provider.connBlocks:
			provider.log.WithField("reason", blocked.Reason).WithField("active", blocked.Active).Warn("connBlocks warn")
		case amqpErr, ok := <-provider.connErrors:
			provider.log.WithField("amqp_error", amqpErr).Error("connErrors error")
			if !ok {
				return
			}
		case amqpErr, ok := <-provider.channelErrors:
			provider.log.WithField("amqp_error", amqpErr).Error("channelErrors error")
			if !ok {
				return
			}
		}
	}
}