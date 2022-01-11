package final

import "time"

type Options struct {
	MaxRetryCount int           // 重试次数
	RetryDuration time.Duration // 重试间隔

	// outbox opt
	OutboxScanInterval time.Duration // scan omission message interval
	OutboxScanOffset   int64         // scan omission message number

	// subscriber opt
	SubscriberNum int // subscriber number
	// acker opt
	AckerNum int // acker number

}

// bus 默认配置
func DefaultOptions() Options {
	return Options{
		SubscriberNum:      5,
		MaxRetryCount:      3,
		RetryDuration:      10 * time.Millisecond,
		AckerNum:           5,
		OutboxScanOffset:   500,
		OutboxScanInterval: 1 * time.Minute,
	}
}

func (opt Options) WithSubscriberNum(val int) Options {
	opt.SubscriberNum = val
	return opt
}

func (opt Options) WithAckerNum(val int) Options {
	opt.AckerNum = val
	return opt
}

func (opt Options) WithOutboxScanInterval(val time.Duration) Options {
	opt.OutboxScanInterval = val
	return opt
}

func (opt Options) WithOutboxScanOffset(val int64) Options {
	opt.OutboxScanOffset = val
	return opt
}
