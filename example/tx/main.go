package main

import (
	"database/sql"
	"time"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final"
	"github.com/xyctruth/final/example"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func main() {
	go send()
	go receive()
	select {}

}

func send() {
	db, _ := example.NewDB().DB()
	mq, _ := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("send_svc", db, mq, final.DefaultOptions())
	bus.Start()
	defer bus.Shutdown()
	for true {
		msg := example.DemoMessage{Type: "aaa", Count: 100}
		msgBytes, _ := msgpack.Marshal(msg)
		gormTX := example.NewDB().Begin()
		/* return err rollback，return nil commit */
		bus.Transaction(gormTX.Statement.ConnPool.(*sql.Tx), func(txBus *final.TxBus) error {
			err := gormTX.Table("test").Create(&example.Test{Name: "aaa"}).Error
			if err != nil {
				return err
			}

			err = txBus.Publish("topic1", "handler1", msgBytes)
			if err != nil {
				return err
			}
			return nil
		})

		time.Sleep(2 * time.Second)
	}
}

func receive() {
	db, _ := example.NewDB().DB()
	mq, _ := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("receive_svc", db, mq, final.DefaultOptions())
	bus.Topic("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", example.Handler1)
	bus.Start()
	defer bus.Shutdown()
	select {}
}
