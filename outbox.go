package final

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lopezator/migrator"
	"github.com/sirupsen/logrus"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final/message"
)

const (
	OutBoxRecordStatusPending uint8 = iota // 客户端发送消息，消息表中的默认状态， 等待 mq 的 confirm ack
)

// db发件箱，在未收到ack前消息会保存在 outbox 中
type outbox struct {
	db      *sql.DB
	logger  *logrus.Entry
	svcName string
	name    string
	bus     *Bus
}

// 初始化db发件箱
func newOutBox(svcName string, bus *Bus) *outbox {
	outbox := &outbox{
		db:  bus.db,
		bus: bus,
		logger: bus.logger.WithFields(logrus.Fields{
			"module": "outbox",
		}),
		svcName: svcName,
		name:    "final_" + svcName + "_outbox",
	}
	return outbox
}

func (outbox *outbox) Start(ctx context.Context) error {
	outbox.scanning()

	outbox.logger.Info("outbox start success")
	go func() {
		loop := time.NewTicker(outbox.bus.opt.OutboxScanInterval)
		for {
			select {
			case <-ctx.Done():
				outbox.logger.Info("outbox stop success")
				return
			case <-loop.C:
				outbox.scanning()
			}
		}
	}()
	return nil
}

func (outbox *outbox) init() error {
	// Configure migrations
	m, err := migrator.New(
		migrator.TableName(fmt.Sprintf("%s_migrations", outbox.name)),
		migrator.Migrations(
			&migrator.Migration{
				Name: "init outbox table",
				Func: func(tx *sql.Tx) error {
					initSQL := `CREATE TABLE IF NOT EXISTS ` + outbox.name + `
								(
									id        bigint auto_increment primary key,
									message   longblob    null,
									status    bigint      null,
									create_at datetime(3) null,
									last_send_at datetime(3) null
								);`

					_, err := tx.Exec(initSQL)
					return err
				},
			},
		),
	)

	if err != nil {
		outbox.logger.WithError(err).Error("migrator error")
		return err
	}
	if err = m.Migrate(outbox.db); err != nil {
		outbox.logger.WithError(err).Error("migrator up error")
		return err
	}

	if outbox.bus.opt.PurgeOnStartup {
		purgeSQL := "delete from " + outbox.name
		n, err := outbox.db.Exec(purgeSQL)
		if err != nil {
			outbox.logger.WithError(err).Error("Purge error")
			return err
		}
		count, _ := n.RowsAffected()
		outbox.logger.Infof("Applied %d purge!", count)
	}

	return nil
}

// scanning scan omission message
func (outbox *outbox) scanning() {
	outbox.logger.
		WithFields(logrus.Fields{
			"offset":   outbox.bus.opt.OutboxScanOffset,
			"interval": outbox.bus.opt.OutboxScanInterval}).
		Info("scanning")

	msgs, err := outbox.take(nil, outbox.bus.opt.OutboxScanOffset, outbox.bus.opt.OutboxScanAgoTime)
	if err != nil {
		outbox.logger.WithError(err).Error("outbox take record failure")
	}
	if len(msgs) > 0 {
		outbox.bus.publisher.publish(msgs...)
	}
}

// 暂存消息到db发件箱中
func (outbox *outbox) staging(tx *sql.Tx, message *message.Message) error {
	err := outbox.transaction(tx, func(tx *sql.Tx) error {
		record, err := newOutBoxRecord(message)
		if err != nil {
			return err
		}
		result, err := tx.Exec("INSERT INTO "+outbox.name+" (message,status,create_at,last_send_at) VALUES (?,?,?,?)",
			record.Message, record.Status, record.CreateAt, record.CreateAt)
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
func (outbox *outbox) take(tx *sql.Tx, offset int64, ago time.Duration) ([]*message.Message, error) {
	msgs := make([]*message.Message, 0)

	var datetime = time.Now().Add(-ago)

	err := outbox.transaction(tx, func(tx *sql.Tx) error {
		querySQL := fmt.Sprintf("SELECT id,message,status,create_at FROM %s WHERE  status = ? AND last_send_at < ? ORDER BY id ASC LIMIT ? FOR UPDATE", outbox.name)
		rows, err := tx.Query(querySQL, OutBoxRecordStatusPending, datetime, offset)

		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil
			}
			return err
		}

		ids := make([]string, 0)
		for rows.Next() {
			var (
				id       int64
				msgBytes []byte
				status   int
				createAt time.Time
			)
			if err := rows.Scan(&id, &msgBytes, &status, &createAt); err != nil {
				outbox.logger.WithError(err).Error("row  scan error")
			}

			msg := &message.Message{}
			err := msgpack.Unmarshal(msgBytes, msg)
			if err != nil {
				panic(err)
			}
			msg.Header.Set("record_id", id)
			msgs = append(msgs, msg)
			ids = append(ids, strconv.FormatInt(id, 10))
		}

		if len(ids) > 0 {
			whereStr := strings.Join(ids, ",")
			updateSQL := fmt.Sprintf("UPDATE %s SET last_send_at = ? WHERE ID IN (%s)", outbox.name, whereStr)
			_, err = tx.Exec(updateSQL, time.Now())
			if err != nil {
				return err
			}
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
			rollbackErr := tx.Rollback()
			if rollbackErr != nil {
				outbox.logger.WithError(rollbackErr).Error("tx rollback error")
			}
			return err
		}
		return tx.Commit()
	}
	return err
}

type outBoxRecord struct {
	ID       int64     `gorm:"id"`
	Message  []byte    `gorm:"message"`
	Status   uint8     `gorm:"status"`
	CreateAt time.Time `gorm:"create_at"`
}

func newOutBoxRecord(message *message.Message) (*outBoxRecord, error) {
	messageByte, err := msgpack.Marshal(message)
	if err != nil {
		return nil, err
	}
	return &outBoxRecord{
		Message:  messageByte,
		Status:   OutBoxRecordStatusPending,
		CreateAt: time.Now(),
	}, nil
}
