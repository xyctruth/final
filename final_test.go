package final

import (
	"fmt"
	"testing"
	"time"

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
			bus1 := New("bus1_svc", NewDB(), NewAmqp(), DefaultOptions().WithAckerNum(1).WithSubscriberNum(1).WithPurgeOnStartup(tt.purge))
			count := 0
			bus1.Subscribe("PurgeOnStartup").Handler("handler1", func(c *Context) error {
				count++
				msg := &DemoMessage{}
				msgpack.Unmarshal(c.Message.Payload, msg)
				require.Equal(t, "aaa", msg.Type)
				require.Equal(t, 100, msg.Count)
				fmt.Println("===================================")
				return nil
			})
			err := bus1.Start()
			require.Equal(t, nil, err)
			err = bus1.Shutdown()
			require.Equal(t, nil, err)

			// bus2 publish messages
			bus2 := New("test_svc", NewDB(), NewAmqp(), DefaultOptions().WithAckerNum(1).WithSubscriberNum(1).WithPurgeOnStartup(false))
			err = bus2.Start()
			require.Equal(t, nil, err)
			err = bus2.Publish("PurgeOnStartup", "handler1", NewDemoMessage("aaa", 100), message.WithConfirm(true))
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
	bus := New("test_svc", NewDB(), NewAmqp(), DefaultOptions().WithAckerNum(1).WithSubscriberNum(1).WithPurgeOnStartup(true))

	count := 0

	bus.Subscribe("Publish_SelfReceive").Handler("handler1", func(c *Context) error {
		count++
		msg := &DemoMessage{}
		msgpack.Unmarshal(c.Message.Payload, msg)
		require.Equal(t, "aaa", msg.Type)
		require.Equal(t, 100, msg.Count)
		return nil
	})

	bus.Start()

	msg := DemoMessage{Type: "aaa", Count: 100}
	msgBytes, err := msgpack.Marshal(msg)
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	time.Sleep(1 * time.Second)
	bus.Shutdown()
	require.Equal(t, 2, count)
}
