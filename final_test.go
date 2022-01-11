package final

import (
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/stretchr/testify/require"
	"github.com/xyctruth/final/message"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func TestPublish_SelfReceive(t *testing.T) {
	db := NewDB()
	mq, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	require.Equal(t, nil, err)

	bus := New("test_svc", db, mq, DefaultOptions().WithAckerNum(1).WithSubscriberNum(1))

	count := 0

	bus.Topic("Publish_SelfReceive").Handler("handler1", func(c *Context) error {
		count++
		msg := &DemoMessage{}
		msgpack.Unmarshal(c.Message.Payload, msg)
		require.Equal(t, "aaa", msg.Type)
		require.Equal(t, int64(100), msg.Count)
		return nil
	})

	bus.Start()
	defer bus.Shutdown()

	msg := DemoMessage{Type: "aaa", Count: 100}
	msgBytes, err := msgpack.Marshal(msg)
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	err = bus.Publish("Publish_SelfReceive", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)

	time.Sleep(1 * time.Second)

	require.Equal(t, 2, count)
}
