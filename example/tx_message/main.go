package main

import (
	"database/sql"
	"fmt"

	"github.com/xyctruth/final"
	"github.com/xyctruth/final/example"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func main() {
	gormDB := example.NewDB()
	db, _ := gormDB.DB()

	mqProvider, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	if err != nil {
		panic(err)
	}

	bus := final.New("tx_svc", db, mqProvider)
	bus.Topic("topic1").Middleware(example.Middleware1, example.Middleware2).Handler("handler1", example.Handler1)

	err = bus.Start()
	if err != nil {
		panic(err)
	}

	tx := gormDB.Begin()

	/* return err rollback，return nil commit */
	err = bus.Transaction(tx.Statement.ConnPool.(*sql.Tx), func(txBus *final.TxBus) error {
		err := tx.Table("test").Create(&example.Test{Name: "aaa"}).Error
		if err != nil {
			return err
		}

		err = txBus.Publish("topic1", "handler1", []byte("payload~~~"))
		if err != nil {
			return err
		}
		return nil
	})

	fmt.Println(err)

	select {}

}
