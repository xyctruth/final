package message

import (
	"time"
)

//MessagePolicy BusMessage config
type MessagePolicy struct {
	Confirm bool
	Durable bool
	TTL     time.Duration
	Delay   int64
}

func DefaultMessagePolicy() *MessagePolicy {
	return &MessagePolicy{
		Confirm: true,
		Durable: true,
		TTL:     0,
	}

}

type MessagePolicyOption func(c *MessagePolicy)

// 开启Confirm，决定消息是否一定发送成功。默认不开启
func WithConfirm(use bool) MessagePolicyOption {
	return func(c *MessagePolicy) {
		c.Confirm = use
	}

}

// 是否消息持久化，默认开启
func WithDurable(use bool) MessagePolicyOption {
	return func(c *MessagePolicy) {
		c.Durable = use
	}
}

// 关联消息过期时间 默认
func WithTTL(duration time.Duration) MessagePolicyOption {
	return func(c *MessagePolicy) {
		c.TTL = duration
	}
}

// 延时队列
func WithDelay(delay int64) MessagePolicyOption {
	return func(c *MessagePolicy) {
		c.Delay = delay
	}

}
