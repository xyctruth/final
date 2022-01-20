package final

import "time"

type Options struct {
	PurgeOnStartup bool // 启动Bus时是否清除遗留的消息，包含（mq遗留的消息，和本地消息表遗留的消息）

	// RetryCount retry count of message processing failed
	// Does not include the first attempt
	RetryCount uint
	// RetryCount retry interval of message processing failed
	RetryInterval time.Duration

	// outbox opt
	OutboxScanInterval time.Duration // 扫描outbox没有收到ack的消息间隔
	OutboxScanOffset   int64         // 扫描outbox没有收到ack的消息

	NumSubscriber int // subscriber number
	NumAcker      int // acker number
}

// DefaultOptions bus 默认配置
func DefaultOptions() Options {
	return Options{
		PurgeOnStartup:     false,
		NumSubscriber:      5,
		RetryCount:         3,
		RetryInterval:      10 * time.Millisecond,
		NumAcker:           5,
		OutboxScanOffset:   500,
		OutboxScanInterval: 1 * time.Minute,
	}
}

// WithRetryCount sets the retry count of message processing failed
// Does not include the first attempt
// The default value of RetryCount is 3.
func (opt Options) WithRetryCount(val uint) Options {
	opt.RetryCount = val
	return opt
}

// WithRetryInterval sets the retry interval of message processing failed
// The default value of RetryInterval is 10 millisecond.
func (opt Options) WithRetryInterval(val time.Duration) Options {
	opt.RetryInterval = val
	return opt
}

// WithNumSubscriber sets the number of subscriber
// Each subscriber runs in an independent goroutine
// The default value of NumSubscriber is 5.
func (opt Options) WithNumSubscriber(val int) Options {
	opt.NumSubscriber = val
	return opt
}

// WithNumAcker sets the number of acker
// Each acker runs in an independent goroutine
// The default value of NumAcker is 5.
func (opt Options) WithNumAcker(val int) Options {
	opt.NumAcker = val
	return opt
}

// WithOutboxScanInterval 设置扫描outbox没有收到ack的消息间隔
// The default value of OutboxScanInterval is 1 minute.
func (opt Options) WithOutboxScanInterval(val time.Duration) Options {
	opt.OutboxScanInterval = val
	return opt
}

// WithOutboxScanOffset  设置扫描outbox没有收到ack的消息
// The default value of OutboxScanOffset is 500.
func (opt Options) WithOutboxScanOffset(val int64) Options {
	opt.OutboxScanOffset = val
	return opt
}

// WithPurgeOnStartup  设置启动Bus时是否清除遗留的消息
// 包含（mq遗留的消息，和本地消息表遗留的消息）
// The default value of PurgeOnStartup is false.
func (opt Options) WithPurgeOnStartup(val bool) Options {
	opt.PurgeOnStartup = val
	return opt
}
