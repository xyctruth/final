package final

import (
	"testing"
	"time"

	"github.com/xyctruth/final/mq/amqp_provider"
)

func TestBus(t *testing.T) {
	var err error

	mqProvider, err := amqp_provider.NewProvider("amqp://user:62qJWqxMVV@test.rabbitmq.jia-huang.com:5672/xyc_final")
	if err != nil {
		t.Error(err)
	}

	bus := New("test_svc", nil, mqProvider)

	//topic1~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
	bus.Topic("topic1").Middleware(middleware1, middleware2).Handler("handler1", handler1)

	err = bus.Start()
	if err != nil {
		t.Error(err)
	}

	for {
		err = bus.Topic("topic1").Publish("handler1", []byte("payload~~~"))
		if err != nil {
			t.Error(err)
		}

		err = bus.Topic("topic1").Publish("handler2", []byte("payload~~~"))
		if err != nil {
			t.Error(err)
		}

		time.Sleep(1 * time.Second)
	}

	select {}
}
