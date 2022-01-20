package final

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final/_example"
	"github.com/xyctruth/final/message"
	"gorm.io/gorm"
)

func TestGeneral(t *testing.T) {
	bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(1).WithNumSubscriber(1).WithPurgeOnStartup(true))

	count := 0

	bus.Subscribe("General").Handler(func(c *Context) error {
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

	err = bus.Publish("General", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	err = bus.Publish("General", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	time.Sleep(1 * time.Second)
	err = bus.Shutdown()
	require.Equal(t, nil, err)
	require.Equal(t, 2, count)
}

func TestRetry(t *testing.T) {

	tests := []struct {
		name  string
		retry uint
		want  int
	}{
		{name: "retry count 0", retry: 0, want: 1},
		{name: "retry count 1", retry: 1, want: 2},
		{name: "retry count 2", retry: 2, want: 3},
		{name: "retry count 3", retry: 3, want: 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithRetryCount(tt.retry).WithPurgeOnStartup(true))
			count := 0
			bus.Subscribe("Retry").Handler(func(c *Context) error {
				count++
				return errors.New("error")
			})

			err := bus.Start()
			require.Equal(t, nil, err)

			msg := DemoMessage{Type: "message", Count: 100}
			msgBytes, err := msgpack.Marshal(msg)
			require.Equal(t, nil, err)

			err = bus.Publish("Retry", msgBytes, message.WithConfirm(true))
			require.Equal(t, nil, err)

			time.Sleep(1 * time.Second)
			err = bus.Shutdown()
			require.Equal(t, nil, err)
			require.Equal(t, tt.want, count)
		})
	}

}

func TestWithPurgeOnStartup(t *testing.T) {
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
			bus1.Subscribe("PurgeOnStartup").Handler(func(c *Context) error {
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
			err = bus2.Publish("PurgeOnStartup", NewDemoMessage("message", 100), message.WithConfirm(true))
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
			bus.Subscribe("SqlTxMessage").Handler(func(c *Context) error {
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

				err = txBus.Publish("SqlTxMessage", msgBytes)
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
			bus.Subscribe("GormTxMessage").Handler(func(c *Context) error {
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

				err = txBus.Publish("GormTxMessage", msgBytes)
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
