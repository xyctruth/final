package final

import "time"

type BusConfig struct {
	// worker数量
	SubscriberConfig *SubscriberConfig
	// outbox 配置
	OutboxConfig *OutboxConfig
	// publisher 配置
	LooperConfig *LooperConfig
	// acker 配置
	AckerConfig *AckerConfig
	// acker 配置
	ReceiverConfig *SubscriberConfig
}

type SubscriberConfig struct {
	Num uint32
	// 重试次数
	MaxRetryCount uint
	// 重试间隔
	RetryDuration time.Duration
}

type OutboxConfig struct {
}

type LooperConfig struct {
	// 是否开启Looper
	Enable bool
	// 每次获取db的message的数量
	Offset uint64
	// 获取db的message的间隔
	Interval time.Duration
}

type AckerConfig struct {
	// Acker goroutine的数量
	Num uint32
}

// bus 默认配置
func DefaultBusConfig() *BusConfig {
	return &BusConfig{
		SubscriberConfig: &SubscriberConfig{
			Num:           5,
			MaxRetryCount: 3,
			RetryDuration: 10 * time.Millisecond,
		},
		AckerConfig: &AckerConfig{
			Num: 5,
		},
		OutboxConfig: &OutboxConfig{},
		LooperConfig: &LooperConfig{
			Enable:   true,
			Offset:   500,
			Interval: 1 * time.Minute,
		},
	}
}

// bus 可选配置
type BusConfigOption func(c *BusConfig)

// 最大重试次数
func WithMaxRetryCount(maxRetryCount uint) BusConfigOption {
	return func(c *BusConfig) {
		c.SubscriberConfig.MaxRetryCount = maxRetryCount
	}
}

// 重试间隔
func WithRetryDuration(retryDuration time.Duration) BusConfigOption {
	return func(c *BusConfig) {
		c.SubscriberConfig.RetryDuration = retryDuration
	}
}
