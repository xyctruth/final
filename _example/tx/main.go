package main

import (
	"database/sql"
	"time"

	"github.com/xyctruth/final/_example/common"

	"github.com/xyctruth/final/_example"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final"
)

func main() {
	go send()
	go sendGorm()
	go receive()
	select {}

}

func send() {
	bus := final.New("send_svc", _example.NewDB(), _example.NewAmqp(), final.DefaultOptions().WithPurgeOnStartup(true))
	err := bus.Start()
	if err != nil {
		panic(err)
	}
	defer bus.Shutdown()
	for true {
		tx, _ := _example.NewDB().Begin()

		/* return err rollback，return nil commit */
		err := bus.Transaction(tx, func(txBus *final.TxBus) error {
			_, err := tx.Exec("INSERT INTO local_business (remark) VALUE (?)", "sql local business")
			if err != nil {
				return err
			}

			msg := common.GeneralMessage{Type: "sql transaction message", Count: 100}
			msgBytes, _ := msgpack.Marshal(msg)
			err = txBus.Publish("topic1", msgBytes)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	}
}

func sendGorm() {
	bus := final.New("send_gorm_svc", _example.NewDB(), _example.NewAmqp(), final.DefaultOptions().WithPurgeOnStartup(true))
	err := bus.Start()
	if err != nil {
		panic(err)
	}
	defer bus.Shutdown()
	for true {

		tx := _example.NewGormDB().Begin()

		/* return err rollback，return nil commit */
		err = bus.Transaction(tx.Statement.ConnPool.(*sql.Tx), func(txBus *final.TxBus) error {
			result := tx.Create(&_example.LocalBusiness{
				Remark: "gorm local business",
			})
			if result.Error != nil {
				return result.Error
			}

			msg := common.GeneralMessage{Type: "gorm transaction message", Count: 100}
			msgBytes, _ := msgpack.Marshal(msg)
			err = txBus.Publish("topic1", msgBytes)
			if err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			panic(err)
		}

		time.Sleep(1 * time.Second)
	}
}

func receive() {
	bus := final.New("receive_svc", _example.NewDB(), _example.NewAmqp(), final.DefaultOptions().WithPurgeOnStartup(true))
	bus.Subscribe("topic1").Middleware(common.Middleware1, common.Middleware2).Handler(common.EchoHandler)
	err := bus.Start()
	if err != nil {
		panic(err)
	}
	defer bus.Shutdown()
	select {}
}
