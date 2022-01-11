package final

import (
	"database/sql"

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

type DemoMessage struct {
	Type  string
	Count int64
}
