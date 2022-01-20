package main

import (
	"time"

	"github.com/xyctruth/final/_example/common"

	"github.com/xyctruth/final/_example"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final"
	"github.com/xyctruth/final/message"
)

func main() {
	go send()
	go receive()
	select {}
}

func send() {
	bus := final.New("send_svc", _example.NewDB(), _example.NewAmqp(), final.DefaultOptions())
	err := bus.Start()
	if err != nil {
		panic(err)
	}
	defer bus.Shutdown()
	for true {
		msg := common.DemoMessage{Type: "simple message", Count: 100}
		msgBytes, _ := msgpack.Marshal(msg)
		err := bus.Publish("topic1", msgBytes, message.WithConfirm(true))
		if err != nil {
			panic(err)
		}
		time.Sleep(1 * time.Second)
	}
}

func receive() {
	bus := final.New("receive_svc", _example.NewDB(), _example.NewAmqp(), final.DefaultOptions())
	bus.Subscribe("topic1").Middleware(common.Middleware1, common.Middleware2).Handler(common.EchoHandler)
	err := bus.Start()
	if err != nil {
		panic(err)
	}
	defer bus.Shutdown()
	select {}
}
