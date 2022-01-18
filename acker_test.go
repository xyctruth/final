package final

import (
	"testing"

	"github.com/xyctruth/final/_example"

	"github.com/stretchr/testify/require"
)

func TestAcker(t *testing.T) {

	bus := New("test_svc", _example.NewDB(), _example.NewAmqp(), DefaultOptions().WithNumAcker(10))
	require.Equal(t, 10, len(bus.ackers))
	err := bus.Start()
	require.Equal(t, nil, err)
	defer bus.Shutdown()
}
