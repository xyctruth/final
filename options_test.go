package final

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOptions(t *testing.T) {
	opt := DefaultOptions()

	opt = opt.WithNumAcker(1)
	require.Equal(t, 1, opt.NumAcker)

	opt = opt.WithNumSubscriber(1)
	require.Equal(t, 1, opt.NumSubscriber)

	opt = opt.WithOutboxScanInterval(1 * time.Second)
	require.Equal(t, 1*time.Second, opt.OutboxScanInterval)

	opt = opt.WithOutboxScanOffset(1)
	require.Equal(t, int64(1), opt.OutboxScanOffset)

	opt = opt.WithPurgeOnStartup(true)
	require.Equal(t, true, opt.PurgeOnStartup)

}
