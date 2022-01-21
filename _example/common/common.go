package common

import (
	"fmt"

	"github.com/vmihailenco/msgpack/v5"
	"github.com/xyctruth/final"
)

func Middleware1(c *final.Context) error {
	fmt.Println("Middleware1 before")
	err := c.Next()
	if err != nil {
		return err
	}
	fmt.Println("Middleware1 after")
	return nil
}

func Middleware2(c *final.Context) error {
	fmt.Println("Middleware2 before")
	err := c.Next()
	if err != nil {
		return err
	}
	fmt.Println("Middleware2 after")
	return nil
}

func EchoHandler(c *final.Context) error {
	msg := &GeneralMessage{}
	err := msgpack.Unmarshal(c.Message.Payload, msg)
	if err != nil {
		return err
	}
	fmt.Println(msg)
	return nil
}

type GeneralMessage struct {
	Type  string
	Count int64
}
