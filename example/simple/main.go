package main

import (
	"fmt"
	"time"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/xyctruth/final"
	"github.com/xyctruth/final/example"
	"github.com/xyctruth/final/message"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func main() {
	go send()
	go receive()
	select {}

}

type DemoMessage struct {
	Type  string
	Count int64
}

func send() {
	db := example.NewSqlDB()
	mq, _ := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("send_svc", db, mq)
	bus.Start()
	defer bus.Shutdown()
	for true {
		msg := DemoMessage{Type: "aaa", Count: 100}
		msgBytes, _ := msgpack.Marshal(msg)
		err := bus.Publish("topic1", "handler1", msgBytes, message.WithConfirm(true))
		if err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
	}
}

func receive() {
	db, _ := example.NewDB().DB()
	mq, _ := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("receive_svc", db, mq)
	bus.Topic("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", Handler1)
	bus.Start()
	defer bus.Shutdown()
	select {}
}

func Handler1(c *final.Context) error {
	msg := &DemoMessage{}
	msgpack.Unmarshal(c.Message.Payload, msg)
	fmt.Println(msg)
	return nil
}
