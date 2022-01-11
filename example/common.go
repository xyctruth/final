package example

import (
	"database/sql"
	"fmt"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/xyctruth/final"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB() *sql.DB {
	db, err := sql.Open("mysql", "root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4")
	if err != nil {
		panic(err)
	}
	return db
}

func NewGormDB() *gorm.DB {
	gormDB, err := gorm.Open(mysql.Open("root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func Middleware1(c *final.Context) error {
	err := c.Next()
	if err != nil {
		return err
	}
	return nil
}

func Middleware2(c *final.Context) error {
	err := c.Next()
	if err != nil {
		return err
	}
	return nil
}

type Test struct {
	Id   int64
	Name string
}

func Handler1(c *final.Context) error {
	msg := &DemoMessage{}
	msgpack.Unmarshal(c.Message.Payload, msg)
	fmt.Println(msg)
	return nil
}

type DemoMessage struct {
	Type  string
	Count int64
}
