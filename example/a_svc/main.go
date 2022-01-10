package main

import (
	"github.com/xyctruth/final/message"

	"github.com/xyctruth/final/example"

	"github.com/xyctruth/final"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func main() {
	gormDB := example.NewDB()
	db, _ := gormDB.DB()

	mqProvider, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	if err != nil {
		panic(err)
	}

	bus := final.New("a_svc", db, mqProvider)

	bus.Topic("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", example.Handler1)

	err = bus.Start()
	if err != nil {
		panic(err)
	}

	msgs := make([]byte, 2000000, 2000000)

	for i := 0; i < 10000000000000000; i++ {
		err = bus.Publish("topic1", "handler1", msgs, message.WithConfirm(true))
		if err != nil {
			panic(err)
		}
	}
	bus.Shutdown()

	select {}

}
