package _example

import (
	"database/sql"
	"os"

	"github.com/xyctruth/final/mq"
	"github.com/xyctruth/final/mq/amqp"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const DefaultMysqlConnStr = "root:xLrFGAzed3@tcp(localhost:3306)/final_test?parseTime=true&loc=Asia%2FShanghai&charset=utf8mb4"
const DefaultAmqpConnStr = "amqp://user:62qJWqxMVV@localhost:5672/final_test"

func NewDB() *sql.DB {
	mysqlConnStr := os.Getenv("MYSQL_CONN_STR")
	if mysqlConnStr == "" {
		mysqlConnStr = DefaultMysqlConnStr
	}

	db, err := sql.Open("mysql", mysqlConnStr)
	if err != nil {
		panic(err)
	}
	return db
}

func NewGormDB() *gorm.DB {
	mysqlConnStr := os.Getenv("MYSQL_CONN_STR")
	if mysqlConnStr == "" {
		mysqlConnStr = DefaultMysqlConnStr
	}

	gormDB, err := gorm.Open(mysql.Open(mysqlConnStr))
	if err != nil {
		panic(err)
	}
	return gormDB
}

func NewAmqp() mq.IProvider {
	amqpConnStr := os.Getenv("AMQP_CONN_STR")
	if amqpConnStr == "" {
		amqpConnStr = DefaultAmqpConnStr
	}
	return amqp.NewProvider(amqpConnStr)
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
