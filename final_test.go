package final

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"github.com/xyctruth/final/_example"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/stretchr/testify/require"
	"github.com/xyctruth/final/message"
)

func TestPurgeOnStartup(t *testing.T) {
	tests := []struct {
		name  string
		purge bool
		want  int
	}{
		{name: "no_purge", purge: false, want: 1},
		{name: "purge", purge: true, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus1 := New("bus1_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(tt.purge))
			count := 0
			bus1.Subscribe("PurgeOnStartup").Handler("handler1", func(c *Context) error {
				count++
				msg := &DemoMessage{}
				err := msgpack.Unmarshal(c.Message.Payload, msg)
				require.Equal(t, nil, err)
				require.Equal(t, "message", msg.Type)
				require.Equal(t, 100, msg.Count)
				return nil
			})
			err := bus1.Start()
			require.Equal(t, nil, err)
			err = bus1.Shutdown()
			require.Equal(t, nil, err)

			// bus2 publish messages
			bus2 := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(false))
			err = bus2.Start()
			require.Equal(t, nil, err)
			err = bus2.Publish("PurgeOnStartup", "handler1", NewDemoMessage("message", 100), message.WithConfirm(true))
			require.Equal(t, nil, err)
			time.Sleep(1 * time.Second)
			err = bus2.Shutdown()
			require.Equal(t, nil, err)

			// bus1 start
			count = 0
			err = bus1.Start()
			require.Equal(t, nil, err)
			time.Sleep(1 * time.Second)
			err = bus1.Shutdown()
			require.Equal(t, nil, err)
			require.Equal(t, tt.want, count)
		})
	}
}

func TestPublish_SelfReceive(t *testing.T) {
	bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(true))

	count := 0

	bus.Subscribe("Publish_SelfReceive").Handler("handler1", func(c *Context) error {
		count++
		msg := &DemoMessage{}
		err := msgpack.Unmarshal(c.Message.Payload, msg)
		require.Equal(t, nil, err)
		require.Equal(t, "message", msg.Type)
		require.Equal(t, 100, msg.Count)
		return nil
	})

	err := bus.Start()
	require.Equal(t, nil, err)

	msg := DemoMessage{Type: "message", Count: 100}
	msgBytes, err := msgpack.Marshal(msg)
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	time.Sleep(1 * time.Second)
	err = bus.Shutdown()
	require.Equal(t, nil, err)
	require.Equal(t, 2, count)
}

func TestSqlTxMessage(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantErr   error
		wantCount int
	}{
		{name: "tx_error", err: errors.New("unknown"), wantErr: sql.ErrNoRows, wantCount: 0},
		{name: "tx_no_error", err: nil, wantErr: nil, wantCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(true))

			count := 0
			bus.Subscribe("GormTxMessage").Handler("handler1", func(c *Context) error {
				count++
				msg := &DemoMessage{}
				err := msgpack.Unmarshal(c.Message.Payload, msg)
				require.Equal(t, nil, err)
				require.Equal(t, "gorm transaction message", msg.Type)
				require.Equal(t, 100, msg.Count)
				return nil
			})

			err := bus.Start()
			require.Equal(t, nil, err)

			msg := DemoMessage{Type: "gorm transaction message", Count: 100}
			msgBytes, _ := msgpack.Marshal(msg)

			_example.InitLocalBusiness()

			db := _example.NewDB()
			tx, err := db.Begin()
			require.Equal(t, nil, err)
			localBusiness := _example.LocalBusiness{
				Remark: "gorm tx message",
			}

			err = bus.Transaction(tx, func(txBus *TxBus) error {
				result, err := tx.Exec("INSERT INTO local_business (remark) VALUE (?)", "sql message")
				if err != nil {
					return err
				}
				localBusiness.Id, _ = result.LastInsertId()

				err = txBus.Publish("GormTxMessage", "handler1", msgBytes)
				if err != nil {
					return err
				}
				return tt.err
			})

			var id int64
			var remark string
			err = db.QueryRow("SELECT * FROM local_business WHERE ID = ?", localBusiness.Id).Scan(&id, &remark)
			require.Equal(t, tt.wantErr, err)
			time.Sleep(1 * time.Second)
			err = bus.Shutdown()
			require.Equal(t, nil, err)
			require.Equal(t, tt.wantCount, count)

		})
	}

}

func TestGormTxMessage(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantErr   error
		wantCount int
	}{
		{name: "tx_error", err: errors.New("unknown"), wantErr: gorm.ErrRecordNotFound, wantCount: 0},
		{name: "tx_no_error", err: nil, wantErr: nil, wantCount: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(true))

			count := 0
			bus.Subscribe("GormTxMessage").Handler("handler1", func(c *Context) error {
				count++
				msg := &DemoMessage{}
				err := msgpack.Unmarshal(c.Message.Payload, msg)
				require.Equal(t, nil, err)
				require.Equal(t, "gorm transaction message", msg.Type)
				require.Equal(t, 100, msg.Count)
				return nil
			})

			err := bus.Start()
			require.Equal(t, nil, err)

			msg := DemoMessage{Type: "gorm transaction message", Count: 100}
			msgBytes, _ := msgpack.Marshal(msg)

			_example.InitLocalBusiness()

			gormDB := _example.NewGormDB()
			tx := gormDB.Begin()

			localBusiness := _example.LocalBusiness{
				Remark: "gorm tx message",
			}

			tx = gormDB.Begin()

			err = bus.Transaction(tx.Statement.ConnPool.(*sql.Tx), func(txBus *TxBus) error {
				result := tx.Create(&localBusiness)
				if result.Error != nil {
					return result.Error
				}

				err = txBus.Publish("GormTxMessage", "handler1", msgBytes)
				if err != nil {
					return err
				}
				return tt.err
			})

			queryLocalBusiness := _example.LocalBusiness{}
			err = gormDB.First(&queryLocalBusiness, localBusiness.Id).Error
			require.Equal(t, tt.wantErr, err)
			time.Sleep(1 * time.Second)
			err = bus.Shutdown()
			require.Equal(t, nil, err)
			require.Equal(t, tt.wantCount, count)

		})
	}

}
