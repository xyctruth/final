package message

import (
	"time"
)

// Policy Message Policy
type Policy struct {
	Confirm bool
	Durable bool
	TTL     time.Duration
	Delay   int64
}

func DefaultMessagePolicy() *Policy {
	return &Policy{
		Confirm: true,
		Durable: true,
		TTL:     0,
	}

}

type PolicyOption func(c *Policy)

// WithConfirm 开启Confirm，会使用本地消息表保存消息，联合mq的Confirm机制
func WithConfirm(use bool) PolicyOption {
	return func(c *Policy) {
		c.Confirm = use
	}

}

// WithDurable 是否消息持久化，默认开启
func WithDurable(use bool) PolicyOption {
	return func(c *Policy) {
		c.Durable = use
	}
}

// WithTTL 关联消息过期时间 默认
func WithTTL(duration time.Duration) PolicyOption {
	return func(c *Policy) {
		c.TTL = duration
	}
}

// WithDelay 延时队列
func WithDelay(delay int64) PolicyOption {
	return func(c *Policy) {
		c.Delay = delay
	}

}
