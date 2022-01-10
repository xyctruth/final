package example

import (
	"github.com/xyctruth/final"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func NewDB() *gorm.DB {
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
