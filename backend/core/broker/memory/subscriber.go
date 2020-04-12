package memory

import (
	"crawlab/core/broker"
	"sync"
)

type memorySubscriber struct {
	id      string
	topic   string
	exit    chan bool
	closed  bool
	handler broker.Handler
	opts    broker.SubscribeOptions
	*sync.Mutex
}


func (m *memorySubscriber) Options() broker.SubscribeOptions {
	return m.opts
}

func (m *memorySubscriber) Topic() string {
	return m.topic
}

func (m *memorySubscriber) Unsubscribe() error {
	m.exit <- true
	m.closed = true
	return nil
}
