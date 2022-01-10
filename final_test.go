package final

import (
	"fmt"
	"testing"
	"time"

	"github.com/vmihailenco/msgpack/v5"

	"github.com/stretchr/testify/require"
	"github.com/xyctruth/final/message"
	"github.com/xyctruth/final/mq/amqp_provider"
)

func TestPublish(t *testing.T) {
	db, err := NewDB().DB()
	require.Equal(t, nil, err)
	mq, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@localhost:5672/xyc_final")
	require.Equal(t, nil, err)
	bus := New("test_svc", db, mq)
	bus.Topic("topic1").Handler("handler1", Handler1)
	bus.Start()
	defer bus.Shutdown()
	msg := DemoMessage{Type: "aaa", Count: 100}
	msgBytes, err := msgpack.Marshal(msg)
	require.Equal(t, nil, err)
	err = bus.Publish("topic1", "handler1", msgBytes, message.WithConfirm(true))
	require.Equal(t, nil, err)
	time.Sleep(1 * time.Second)
}

func Handler1(c *Context) error {
	msg := &DemoMessage{}
	msgpack.Unmarshal(c.Message.Payload, msg)
	fmt.Println(msg)
	return nil
}
