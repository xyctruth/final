package mq

import (
	"context"

	"github.com/xyctruth/final/message"
)

type IProvider interface {
	Init(svcName string, purge bool, topics []string) error
	Publish(messages *message.Message) error
	Subscribe(ctx context.Context, consumerTag string, msgs chan *message.Message) error
	NotifyConfirm(ack, nack chan uint64)
}
