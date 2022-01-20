package final

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	opt := DefaultOptions()

	require.Equal(t, uint(3), opt.RetryCount)
	opt = opt.WithRetryCount(5)
	require.Equal(t, uint(5), opt.RetryCount)

	require.Equal(t, 10*time.Millisecond, opt.RetryInterval)
	opt = opt.WithRetryInterval(100 * time.Millisecond)
	require.Equal(t, 100*time.Millisecond, opt.RetryInterval)

	require.Equal(t, 5, opt.NumAcker)
	opt = opt.WithNumAcker(1)
	require.Equal(t, 1, opt.NumAcker)

	require.Equal(t, 5, opt.NumSubscriber)
	opt = opt.WithNumSubscriber(1)
	require.Equal(t, 1, opt.NumSubscriber)

	require.Equal(t, 1*time.Minute, opt.OutboxScanInterval)
	opt = opt.WithOutboxScanInterval(1 * time.Second)
	require.Equal(t, 1*time.Second, opt.OutboxScanInterval)

	require.Equal(t, int64(500), opt.OutboxScanOffset)
	opt = opt.WithOutboxScanOffset(1)
	require.Equal(t, int64(1), opt.OutboxScanOffset)

	require.Equal(t, false, opt.PurgeOnStartup)
	opt = opt.WithPurgeOnStartup(true)
	require.Equal(t, true, opt.PurgeOnStartup)
}
