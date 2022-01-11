package final

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAcker(t *testing.T) {
	bus := New("test_svc", NewDB(), NewAmqp(), DefaultOptions().WithAckerNum(10))
	require.Equal(t, 10, len(bus.ackers))
	err := bus.Start()
	require.Equal(t, nil, err)
	defer bus.Shutdown()
}
