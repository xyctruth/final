package message

import (
	"sync"

	uuidtools "github.com/satori/go.uuid"
)

type (
	Message struct {
		UUID    string
		Topic   string
		Handler string
		SvcName string
		Header  MessageHeader
		Payload []byte
		Policy  *MessagePolicy

		AckChan    chan struct{} `json:"-"`
		RejectChan chan struct{} `json:"-"`

		mutex sync.Mutex
	}

	MessageHeader map[string]interface{}
)

func NewMessage(uuid, topic, handler string, payload []byte, opts ...MessagePolicyOption) *Message {
	if uuid == "" {
		uuid = uuidtools.NewV4().String()
	}
	msg := &Message{
		UUID:       uuid,
		Topic:      topic,
		Handler:    handler,
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

func (m *Message) Reset(uuid, topic, handler string, payload []byte, opts ...MessagePolicyOption) {
	if uuid == "" {
		uuid = uuidtools.NewV4().String()
	}
	m.UUID = uuid
	m.Topic = topic
	m.Handler = handler
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

func (m MessageHeader) Get(key string) interface{} {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}

func (m MessageHeader) Set(key string, value interface{}) {
	m[key] = value
}
