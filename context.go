package final

import (
	"github.com/xyctruth/final/message"
)

type (
	Context struct {
		Topic       string
		HandlerName string
		Key         string
		Message     *message.Message
		// middleware and handler
		handlers []HandlerFunc
		index    int
	}
)

func newContext(m *message.Message, handlers []HandlerFunc) *Context {
	return &Context{
		Topic:       m.Topic,
		HandlerName: m.Handler,
		Message:     m,
		handlers:    handlers,
		index:       -1,
	}
}

func (c *Context) Next() error {
	c.index++
	s := len(c.handlers)
	for ; c.index < s; c.index++ {
		err := c.handlers[c.index](c)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Context) Reset(m *message.Message, handlers []HandlerFunc) {
	c.Topic = m.Topic
	c.HandlerName = m.Handler
	c.Message = m
	c.handlers = handlers
	c.index = -1
}
