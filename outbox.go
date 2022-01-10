package final

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	migrate "github.com/rubenv/sql-migrate"
	"github.com/xyctruth/final/message"

	"github.com/sirupsen/logrus"
)

// db发件箱，在未收到ack前消息会保存在 outbox 中
type outbox struct {
	db      *sql.DB
	logger  *logrus.Entry
	svcName string
	name    string
}

// 初始化db发件箱
func newOutBox(svcName string, db *sql.DB, logger *logrus.Entry) *outbox {
	outbox := &outbox{
		db: db,
		logger: logger.WithFields(logrus.Fields{
			"module": "outbox",
		}),
		svcName: svcName,
		name:    "final_" + svcName + "_outbox",
	}
	return outbox
}

func (outbox *outbox) migration() error {
	init := `CREATE TABLE IF NOT EXISTS ` + outbox.name + `
			(
				id        bigint auto_increment
					primary key,
				message   longblob    null,
				status    bigint      null,
				create_at datetime(3) null
			);`

	migrations := &migrate.MemoryMigrationSource{
		Migrations: []*migrate.Migration{
			{
				Id:   "init",
				Up:   []string{init},
				Down: []string{"DROP TABLE " + outbox.name},
			},
		},
	}

	n, err := migrate.Exec(outbox.db, "mysql", migrations, migrate.Up)

	if err != nil {
		outbox.logger.WithError(err).Error("migrations error")
		return err
	}
	outbox.logger.Infof("Applied %d migrations!\n", n)
	return nil
}

// 暂存消息到db发件箱中
func (outbox *outbox) staging(tx *sql.Tx, message *message.Message) error {
	err := outbox.transaction(tx, func(tx *sql.Tx) error {
		record := newRecord(message)
		result, err := tx.Exec("INSERT INTO "+outbox.name+" (message,status,create_at) VALUES (?,?,?)",
			record.Message, record.Status, record.CreateAt)
		if err != nil {
			return err
		}
		id, err := result.LastInsertId()
		if err != nil {
			return err
		}

		message.Header.Set("record_id", id)
		return nil
	})

	return err
}

// 接受到ack后 Delete掉消息记录
func (outbox *outbox) done(tx *sql.Tx, id interface{}) error {
	err := outbox.transaction(tx, func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM "+outbox.name+" WHERE ID = ?", id)
		if err != nil {
			return err
		}
		return nil
	})

	return err
}

// 获取没有收到ack的消息，准备重新发送到mq中
func (outbox *outbox) take(tx *sql.Tx, offset int64) ([]*message.Message, error) {
	msgs := make([]*message.Message, 0, 0)

	var datetime = time.Now().Add(-1 * time.Minute)

	err := outbox.transaction(tx, func(tx *sql.Tx) error {
		rows, err := tx.Query(
			" SELECT * FROM "+outbox.name+" WHERE  status = ? AND create_at < ? ORDER BY id ASC LIMIT ? FOR UPDATE",
			0, datetime, offset)

		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}

		for rows.Next() {
			var (
				id       int64
				msgBtyes []byte
				status   int
				createAt time.Time
			)
			if err := rows.Scan(&id, &msgBtyes, &status, &createAt); err != nil {
				outbox.logger.WithError(err).Error("row  scan error")
			}

			msg := &message.Message{}

			err := json.Unmarshal(msgBtyes, msg)
			if err != nil {
				outbox.logger.WithError(err).Error("json.Unmarshal(msgBtyes, msg) error")
			}
			msg.Header.Set("record_id", id)
			msgs = append(msgs, msg)
		}

		return nil
	})

	return msgs, err
}

func (outbox *outbox) transaction(tx *sql.Tx, fc func(tx *sql.Tx) error) error {
	needCommit := false
	if tx == nil {
		needCommit = true
		var err error
		tx, err = outbox.db.Begin()
		if err != nil {
			return err
		}
	}

	err := fc(tx)

	if needCommit {
		if err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit()
	}
	return err
}

func newRecord(message *message.Message) *Record {
	messageByte, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
	}
	return &Record{
		Message:  messageByte,
		CreateAt: time.Now(),
	}
}

type Record struct {
	ID       int64     `gorm:"id"`
	Message  []byte    `gorm:"message"`
	Status   int       `gorm:"status"`
	CreateAt time.Time `gorm:"create_at"`
}
