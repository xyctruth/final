package message

import (
	"sync"

	uuidtools "github.com/satori/go.uuid"
)

type (
	Message struct {
		UUID    string
		Topic   string
		SvcName string
		Header  Header
		Payload []byte
		Policy  *Policy

		AckChan    chan struct{} `msgpack:"-"`
		RejectChan chan struct{} `msgpack:"-"`

		mutex sync.Mutex
	}

	Header map[string]interface{}
)

func NewMessage(uuid, topic string, payload []byte, opts ...PolicyOption) *Message {
	if uuid == "" {
		uuid = uuidtools.NewV4().String()
	}
	msg := &Message{
		UUID:       uuid,
		Topic:      topic,
		Payload:    payload,
		AckChan:    make(chan struct{}),
		RejectChan: make(chan struct{}),
		Header:     make(map[string]interface{}),
	}

	messagePolicy := DefaultMessagePolicy()
	for _, opt := range opts {
		opt(messagePolicy)
	}
	msg.Policy = messagePolicy
	return msg
}

func (m *Message) Reset(uuid, topic string, payload []byte, opts ...PolicyOption) {
	if uuid == "" {
		uuid = uuidtools.NewV4().String()
	}
	m.UUID = uuid
	m.Topic = topic
	m.Payload = payload

	messagePolicy := DefaultMessagePolicy()
	for _, opt := range opts {
		opt(messagePolicy)
	}
	m.Policy = messagePolicy
}

func (m *Message) Ack() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	close(m.AckChan)
}

func (m *Message) Reject() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	close(m.RejectChan)
}

func (m *Message) Acked() <-chan struct{} {
	return m.AckChan
}

func (m *Message) Rejected() <-chan struct{} {
	return m.RejectChan
}

func (m Header) Get(key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}

func (m Header) Set(key string, value interface{}) {
	m[key] = value
}
