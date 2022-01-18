package _example

import (
	"database/sql"

	"github.com/xyctruth/final/mq"
	"github.com/xyctruth/final/mq/amqp_provider"
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

func NewAmqp() mq.IProvider {
	return amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
}

func NewGormDB() *gorm.DB {
	gormDB, err := gorm.Open(mysql.Open("root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func InitLocalBusiness() {
	err := NewGormDB().AutoMigrate(&LocalBusiness{})
	if err != nil {
		return
	}
}

type LocalBusiness struct {
	Id     int64
	Remark string
}

func (LocalBusiness) TableName() string {
	return "local_business"
}
