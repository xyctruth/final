package main

import (
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

func send() {
	db := example.NewDB()
	mq := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("send_svc", db, mq, final.DefaultOptions())
	bus.Start()
	defer bus.Shutdown()
	for true {
		msg := example.DemoMessage{Type: "aaa", Count: 100}
		msgBytes, _ := msgpack.Marshal(msg)
		err := bus.Publish("topic1", "handler1", msgBytes, message.WithConfirm(true))
		if err != nil {
			panic(err)
		}
		time.Sleep(2 * time.Second)
	}
}

func receive() {
	db := example.NewDB()
	mq := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	bus := final.New("receive_svc", db, mq, final.DefaultOptions())
	bus.Subscribe("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", example.Handler1)
	bus.Start()
	defer bus.Shutdown()
	select {}
}
