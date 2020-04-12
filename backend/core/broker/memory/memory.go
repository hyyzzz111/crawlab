package memory

import (
	"context"
	"crawlab/core/broker"
	"crawlab/core/logger"
	"errors"
	"github.com/apex/log"
	"github.com/panjf2000/ants/v2"

	"github.com/google/uuid"
	"math/rand"
	"sync"
	"time"
)

type memoryBroker struct {
	opts broker.Options
	*logger.LoggerWrapper
	addr string
	sync.RWMutex
	connected   bool
	Subscribers map[string][]*memorySubscriber
	pool        *ants.Pool
	channelSize int
}

//func(m *memorySubscriber) sensed(){
//	m.opts.Sensed = true
//}
func (m *memoryBroker) Options() broker.Options {
	return m.opts
}

func (m *memoryBroker) Address() string {
	return m.addr
}

func (m *memoryBroker) Connect() error {
	m.connected = true
	return nil
}

func (m *memoryBroker) Disconnect() error {
	m.Lock()
	defer m.Unlock()

	if !m.connected {
		return nil
	}

	m.connected = false

	return nil
}

func (m *memoryBroker) Init(opts ...broker.Option) error {
	for _, o := range opts {
		o(&m.opts)
	}
	return nil
}
func (m *memoryBroker) subscribe(ctx context.Context, topic string, msg *broker.Message) bool {

	subs, ok := m.Subscribers[topic]

	if !ok {
		return true
	}
	var v interface{}
	if m.opts.BodyCodec !=nil{
		buf,err:= m.opts.BodyCodec.Marshal(msg.Body)
		if err != nil {
			return false
		}
		msg.Body = buf
	}
	if m.opts.Codec != nil {
		buf, err := m.opts.Codec.Marshal(msg)
		if err != nil {
			m.Logger().Errorf("%s", err.Error())
			return false
		}
		v = buf
	} else {
		v = msg
	}

	p := &memoryEvent{
		topic:   topic,
		message: v,
		opts:    m.opts,
	}
	for _, sub := range subs {
		select {
		case <-ctx.Done():
			break
		default:
			if sub.closed {
				continue
			}
			if err := sub.handler(p); err != nil {
				p.err = err
				if eh := m.opts.ErrorHandler; eh != nil {
					_ = eh(p)
					continue
				}
				m.Logger().Errorf("%s", err.Error())
			}else{
				if sub.Options().Once {
					_ = sub.Unsubscribe()
				}
			}

		}
	}

	return true
}
func (m *memoryBroker) Publish(topic string, msg *broker.Message, opts ...broker.PublishOption) error {
	m.RLock()
	if !m.connected {
		m.RUnlock()
		return errors.New("not connected")
	}
	m.RUnlock()

	return m.pool.Submit(func() {
		m.subscribe(m.opts.Context, topic, msg)
	})
}

func (m *memoryBroker) Subscribe(topic string, handler broker.Handler, opts ...broker.SubscribeOption) (broker.Subscriber, error) {
	m.RLock()
	if !m.connected {
		m.RUnlock()
		return nil, errors.New("not connected")
	}
	m.RUnlock()

	var options broker.SubscribeOptions
	for _, o := range opts {
		o(&options)
	}

	sub := &memorySubscriber{
		exit:    make(chan bool, 1),
		id:      uuid.New().String(),
		topic:   topic,
		handler: handler,
		opts:    options,
	}

	m.Lock()
	m.Subscribers[topic] = append(m.Subscribers[topic], sub)
	m.Unlock()

	go func() {
		<-sub.exit
		m.Lock()
		var newSubscribers []*memorySubscriber
		for _, sb := range m.Subscribers[topic] {
			if sb.id == sub.id {
				continue
			}
			newSubscribers = append(newSubscribers, sb)
		}
		m.Subscribers[topic] = newSubscribers
		m.Unlock()
	}()

	return sub, nil
}

func (m *memoryBroker) String() string {
	return "memory"
}
func NewBrokerAndConnect(opts ...broker.Option) broker.Broker {
	b :=NewBroker(opts...)
	_ = b.Connect()
	return b
}
func NewBroker(opts ...broker.Option) broker.Broker {
	options := broker.Options{
		Context: context.Background(),
		Logger:  log.Log,
	}

	rand.Seed(time.Now().UnixNano())
	for _, o := range opts {
		o(&options)
	}
	pool, _ := ants.NewPool(10000)
	return &memoryBroker{
		opts:        options,
		pool:        pool,
		Subscribers: make(map[string][]*memorySubscriber),
	}
}
