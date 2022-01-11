package final

import "time"

type Options struct {
	PurgeOnStartup bool // 启动Bus时是否清除遗留的消息，包含（mq遗留的消息，和本地消息表遗留的消息）

	MaxRetryCount int           // 重试次数
	RetryDuration time.Duration // 重试间隔

	// outbox opt
	OutboxScanInterval time.Duration // 扫描outbox没有收到ack的消息间隔
	OutboxScanOffset   int64         // 扫描outbox没有收到ack的消息

	// subscriber opt
	SubscriberNum int // subscriber number
	// acker opt
	AckerNum int // acker number

}

// bus 默认配置
func DefaultOptions() Options {
	return Options{
		PurgeOnStartup:     false,
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

// WithOutboxScanInterval 设置扫描outbox没有收到ack的消息间隔
func (opt Options) WithOutboxScanInterval(val time.Duration) Options {
	opt.OutboxScanInterval = val
	return opt
}

// WithOutboxScanOffset  设置扫描outbox没有收到ack的消息
func (opt Options) WithOutboxScanOffset(val int64) Options {
	opt.OutboxScanOffset = val
	return opt
}

// WithPurgeOnStartup  设置启动Bus时是否清除遗留的消息
// 包含（mq遗留的消息，和本地消息表遗留的消息）
func (opt Options) WithPurgeOnStartup(val bool) Options {
	opt.PurgeOnStartup = val
	return opt
}
