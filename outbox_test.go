package final

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/xyctruth/final/_example"
	"github.com/xyctruth/final/message"
)

func TestOutBox(t *testing.T) {
	bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithPurgeOnStartup(true))

	err := bus.outbox.init()
	require.Equal(t, nil, err)

	err = bus.outbox.staging(nil, message.NewMessage("0", "", nil))
	require.Equal(t, nil, err)
	err = bus.outbox.staging(nil, message.NewMessage("1", "", nil))
	require.Equal(t, nil, err)
	err = bus.outbox.staging(nil, message.NewMessage("2", "", nil))
	require.Equal(t, nil, err)
	err = bus.outbox.staging(nil, message.NewMessage("3", "", nil))
	require.Equal(t, nil, err)

	msgs, err := bus.outbox.take(nil, 100, time.Second)
	require.Equal(t, nil, err)
	require.Equal(t, 0, len(msgs))

	time.Sleep(time.Second)

	msgs, err = bus.outbox.take(nil, 100, time.Second)
	require.Equal(t, nil, err)
	require.Equal(t, 4, len(msgs))
	require.Equal(t, "0", msgs[0].UUID)
	require.Equal(t, "1", msgs[1].UUID)
	require.Equal(t, "2", msgs[2].UUID)
	require.Equal(t, "3", msgs[3].UUID)

	msgs, err = bus.outbox.take(nil, 100, time.Second)
	require.Equal(t, nil, err)
	require.Equal(t, 0, len(msgs))

	time.Sleep(time.Second)

	msgs, err = bus.outbox.take(nil, 100, time.Second)
	require.Equal(t, nil, err)
	require.Equal(t, 4, len(msgs))
}
