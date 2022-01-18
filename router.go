package final

import (
	"errors"
	"fmt"
	"sync"

	"github.com/xyctruth/final/message"
)

type (
	HandlerFunc       func(*Context) error
	HandlerMiddleware func(h HandlerFunc) HandlerFunc

	routerTopic struct {
		name        string
		middlewares []HandlerFunc
		bus         *Bus
	}

	// router 是handler的路由程序，帮助消息的到正确的handler处理
	router struct {
		handlers map[string]HandlerFunc
		topics   map[string]*routerTopic
		ctxPool  sync.Pool
	}
)

func (topic *routerTopic) Middleware(middlewares ...HandlerFunc) *routerTopic {
	topic.middlewares = append(topic.middlewares, middlewares...)
	return topic
}

func (topic *routerTopic) Handler(name string, handler HandlerFunc) {
	topic.bus.router.addRoute(topic.name, name, handler)
}

func newRouter() *router {
	s := &router{
		handlers: make(map[string]HandlerFunc),
		topics:   make(map[string]*routerTopic),
	}

	s.ctxPool.New = func() interface{} {
		return s.allocateContext()
	}
	return s
}

func (r *router) addRoute(topic, handlerName string, handler HandlerFunc) {
	r.handlers[r.buildKey(topic, handlerName)] = handler
}

func (r *router) getRoute(topic, handlerName string) HandlerFunc {
	if handler, ok := r.handlers[r.buildKey(topic, handlerName)]; ok {
		return handler
	}
	return nil
}

func (r *router) buildKey(topic, handlerName string) string {
	return fmt.Sprintf("%s/%s", topic, handlerName)
}

func (r *router) handle(msg *message.Message) error {
	var middlewares []HandlerFunc
	if topic, ok := r.topics[msg.Topic]; ok {
		middlewares = append(middlewares, topic.middlewares...)
	}

	c := r.ctxPool.Get().(*Context)
	defer r.ctxPool.Put(c)

	c.Reset(msg, middlewares)

	handle := r.getRoute(c.Topic, c.HandlerName)
	if handle != nil {
		c.handlers = append(c.handlers, handle)
	} else {
		c.handlers = append(c.handlers, func(c *Context) error {
			return errors.New("no match handler")
		})
	}
	err := c.Next()
	return err
}

func (r *router) allocateContext() *Context {
	return &Context{}
}
