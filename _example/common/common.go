package common

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final"
)

func Middleware1(c *final.Context) error {
	err := c.Next()
	if err != nil {
		return err
	}
	return nil
}

func Middleware2(c *final.Context) error {
	err := c.Next()
	if err != nil {
		return err
	}
	return nil
}

func EchoHandler(c *final.Context) error {
	msg := &DemoMessage{}
	msgpack.Unmarshal(c.Message.Payload, msg)
	fmt.Println(msg)
	return nil
}

type DemoMessage struct {
	Type  string
	Count int64
}
