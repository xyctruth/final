package final

import (
	"github.com/vmihailenco/msgpack/v5"
)

type DemoMessage struct {
	Type  string
	Count int
}

func NewDemoMessage(t string, count int) []byte {
	msg := DemoMessage{Type: t, Count: count}
	msgBytes, _ := msgpack.Marshal(msg)
	return msgBytes
}
