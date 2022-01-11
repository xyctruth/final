package final

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/xyctruth/final/message"

	"github.com/sirupsen/logrus"
	"github.com/xyctruth/final/mq"
)

type (
	Bus struct {
		svcName    string  // service name
		opt        Options // Bus Options
		db         *sql.DB
		mqProvider mq.IProvider // mq provider

		router      *router       // router 是handler的路由程序，帮助消息的到正确的handler处理
		outbox      *outbox       // outbox db发件箱，在未收到ack前消息会保存在 outbox 中
		subscribers []*subscriber // subscriber 启动 Options.SubscriberNum 个 goroutine 订阅消息队列中的消息 使用 router 处理消息
		publisher   *publisher    // publisher 发送消息到消息队列中
		ackers      []*acker      // acker 启动 Options.AckerNum 个goroutine接收消息队列ack消息后，Done掉 outbox 中的消息记录

		logger  *logrus.Entry
		msgPool sync.Pool
		cancel  context.CancelFunc
	}

	TxBus struct {
		bus   *Bus
		tx    *sql.Tx
		msgs  []*message.Message
		mutex sync.Mutex
	}
)

// 初始化Bus
func New(svcName string, db *sql.DB, mqProvider mq.IProvider, opt Options) *Bus {
	logger := &logrus.Logger{
		Out: os.Stdout,
		Formatter: &logrus.TextFormatter{
			FullTimestamp: true,
		},
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.InfoLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	logEntry := logger.WithField("final", svcName)

	var bus = &Bus{
		svcName:    svcName,
		db:         db,
		mqProvider: mqProvider,
		opt:        opt,
		logger:     logEntry,
		router:     newRouter(),
	}

	// create outbox
	bus.outbox = newOutBox(svcName, bus)

	// create subscribers
	bus.subscribers = make([]*subscriber, 0, bus.opt.SubscriberNum)
	for i := 0; i < bus.opt.SubscriberNum; i++ {
		bus.subscribers = append(bus.subscribers, newSubscriber(fmt.Sprintf("%s_subscribers_%d", svcName, i), bus))
	}

	// create publisher
	bus.publisher = newPublisher(bus)

	// create acker
	bus.ackers = make([]*acker, 0, bus.opt.AckerNum)
	for i := 0; i < bus.opt.AckerNum; i++ {
		bus.ackers = append(bus.ackers, newAcker(fmt.Sprintf("%s_acker_%d", svcName, i), bus))
	}

	bus.msgPool.New = func() interface{} {
		return bus.allocateMessage()
	}

	bus.logger.Info("Bus init success")

	return bus
}

func (bus *Bus) Start() error {
	var err error

	err = bus.initProvider()
	if err != nil {
		return err
	}

	err = bus.outbox.init()
	if err != nil {
		return err
	}

	ctx := context.Background()
	ctx, bus.cancel = context.WithCancel(ctx)

	for _, subscriber := range bus.subscribers {
		err = subscriber.Start(ctx)
		if err != nil {
			return err
		}
	}

	err = bus.publisher.Start(ctx)
	if err != nil {
		return err
	}

	for _, acker := range bus.ackers {
		err = acker.Start(ctx)
		if err != nil {
			return err
		}
	}

	bus.logger.Info("Bus start success")
	return nil
}

func (bus *Bus) Shutdown() error {
	bus.cancel()
	err := bus.mqProvider.Exit()
	if err != nil {
		bus.logger.WithError(err).Error("Bus stop error!!!")
	}
	bus.logger.Info("Bus stop !!!")
	return err
}

func (bus *Bus) Subscribe(topic string) *routerTopic {
	if topic, ok := bus.router.topics[topic]; ok {
		return topic
	}

	newTopic := &routerTopic{
		name: topic,
		bus:  bus,
	}
	bus.router.topics[topic] = newTopic
	return newTopic
}

func (bus *Bus) Publish(topic string, handler string, payload []byte, opts ...message.MessagePolicyOption) error {
	var err error

	msg := bus.msgPool.Get().(*message.Message)
	msg.Reset("", topic, handler, payload, opts...)

	if msg.Policy.Confirm {
		err = bus.outbox.staging(nil, msg)
		if err != nil {
			return err
		}
	}
	go func() {
		bus.publisher.publish(msg)
		bus.msgPool.Put(msg)
	}()

	return nil
}

func (bus *Bus) WithTx(tx *sql.Tx) *TxBus {
	txBus := &TxBus{
		bus:  bus,
		tx:   tx,
		msgs: make([]*message.Message, 0),
	}
	return txBus
}

func (bus *Bus) Transaction(tx *sql.Tx, fc func(txBus *TxBus) error) error {
	txBus := &TxBus{
		bus:  bus,
		tx:   tx,
		msgs: make([]*message.Message, 0),
	}

	err := fc(txBus)

	if err != nil {
		txBus.RollBack()
		return err
	}

	return txBus.Commit()
}

func (txBus *TxBus) Publish(topic string, handler string, payload []byte, opts ...message.MessagePolicyOption) error {
	txBus.mutex.Lock()
	defer txBus.mutex.Unlock()

	msg := txBus.bus.msgPool.Get().(*message.Message)
	msg.Reset("", topic, handler, payload, opts...)

	err := txBus.bus.outbox.staging(txBus.tx, msg)
	if err != nil {
		return err
	}
	txBus.msgs = append(txBus.msgs, msg)
	return nil
}

func (txBus *TxBus) Commit() error {
	err := txBus.tx.Commit()
	if err != nil {
		return err
	}
	txBus.publish()
	return nil
}

func (txBus *TxBus) RollBack() error {
	err := txBus.tx.Rollback()
	if err != nil {
		return err
	}
	return nil

}

func (txBus *TxBus) publish() {
	go func() {
		txBus.bus.publisher.publish(txBus.msgs...)
		for _, msg := range txBus.msgs {
			txBus.bus.msgPool.Put(msg)
		}
	}()
}

func (bus *Bus) initProvider() error {
	var err error

	if bus.mqProvider == nil {
		return errors.New("mqProvider is nil")
	}

	topics := make([]string, 0, len(bus.router.topics))
	for topic := range bus.router.topics {
		topics = append(topics, topic)
	}

	err = bus.mqProvider.Init(bus.svcName, bus.opt.PurgeOnStartup, topics)
	if err != nil {
		return err
	}

	if bus.db == nil {
		return errors.New("db is nil")
	}

	return nil
}

func (engine *Bus) allocateMessage() *message.Message {
	return &message.Message{
		AckChan:    make(chan struct{}),
		RejectChan: make(chan struct{}),
		Header:     make(map[string]interface{}),
	}
}
