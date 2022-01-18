package final

import "time"

type Options struct {
	PurgeOnStartup bool // 启动Bus时是否清除遗留的消息，包含（mq遗留的消息，和本地消息表遗留的消息）

	MaxRetryCount int           // 重试次数
	RetryDuration time.Duration // 重试间隔

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
		MaxRetryCount:      3,
		RetryDuration:      10 * time.Millisecond,
		NumAcker:           5,
		OutboxScanOffset:   500,
		OutboxScanInterval: 1 * time.Minute,
	}
}

// WithNumSubscriber sets the number of subscriber
// 1 subscriber run 1 goroutine
// The default value of NumSubscriber is 5.
func (opt Options) WithNumSubscriber(val int) Options {
	opt.NumSubscriber = val
	return opt
}

// WithNumAcker sets the number of acker
// 1 acker run 1 goroutine
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
