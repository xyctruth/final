package final

import (
	"database/sql"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/xyctruth/final/mq"
	"github.com/xyctruth/final/mq/amqp_provider"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewGormDB() *gorm.DB {
	gormDB, err := gorm.Open(mysql.Open("root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func NewDB() *sql.DB {
	db, err := sql.Open("mysql", "root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4")
	if err != nil {
		panic(err)
	}
	return db
}

func NewAmqp() mq.IProvider {
	mq := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	return mq
}

type DemoMessage struct {
	Type  string
	Count int
}

func NewDemoMessage(t string, count int) []byte {
	msg := DemoMessage{Type: t, Count: count}
	msgBytes, _ := msgpack.Marshal(msg)
	return msgBytes
}
