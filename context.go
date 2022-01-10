package final

import "github.com/xyctruth/final/message"

type (
	Context struct {
		Topic   string
		Key     string
		Message *message.Message
		// middleware and handler
		handlers []HandlerFunc
		index    int
	}
)

func newContext(m *message.Message, handlers []HandlerFunc) *Context {
	topic := m.Topic
	key := m.Handler

	return &Context{
		Topic:    topic,
		Key:      key,
		Message:  m,
		handlers: handlers,
		index:    -1,
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
	topic := m.Topic
	key := m.Handler

	c.Topic = topic
	c.Key = key
	c.Message = m
	c.handlers = handlers
	c.index = -1
}
