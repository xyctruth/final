package final

import (
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

type DemoMessage struct {
	Type  string
	Count int64
}
