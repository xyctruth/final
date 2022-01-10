package main

import (
	"fmt"

	"github.com/xyctruth/final/example"

	"github.com/xyctruth/final"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func middleware1() final.HandlerFunc {
	return func(c *final.Context) error {
		fmt.Println("Middleware1 start")
		err := c.Next()
		if err != nil {
			return err
		}
		fmt.Println("Middleware1 end")
		return nil
	}
}

func main() {
	gormDB := example.NewDB()
	db, _ := gormDB.DB()

	mqProvider, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	if err != nil {
		panic(err)
	}

	bus := final.New("b_svc", db, mqProvider)

	topic1 := bus.Topic("topic1").Middleware(middleware1())
	{
		topic1.Handler("handler1", func(context *final.Context) error {
			fmt.Println(context.Message.Payload)
			return nil
		})
	}

	err = bus.Start()
	if err != nil {
		panic(err)
	}

	select {}
}
