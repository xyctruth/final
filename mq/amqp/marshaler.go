package amqp

import (
	"strconv"
	"time"

	"github.com/streadway/amqp"
	"github.com/xyctruth/final/message"
)

func NewMessageFromDelivery(delivery amqp.Delivery) *message.Message {
	msg := message.NewMessage(
		delivery.MessageId,
		castToString(delivery.Headers["x-final-msg-topic"]),
		delivery.Body,
	)

	return msg
}

func NewPublishingFromMessage(msg *message.Message) amqp.Publishing {
	headers := amqp.Table{
		"x-final-msg-topic": msg.Topic,
	}

	publishing := amqp.Publishing{
		Body:        msg.Payload,
		ReplyTo:     msg.SvcName,
		MessageId:   msg.UUID,
		ContentType: "string",
		Headers:     headers,
	}

	if msg.Policy.Durable {
		publishing.DeliveryMode = amqp.Persistent
	} else {
		publishing.DeliveryMode = amqp.Transient
	}

	if msg.Policy.TTL > 0 {
		ms := int64(msg.Policy.TTL / time.Millisecond)
		publishing.Headers["x-message-ttl"] = strconv.FormatInt(ms, 10)
	}

	return publishing
}

func castToString(i interface{}) string {
	v, ok := i.(string)
	if !ok {
		return ""
	}
	return v
}
