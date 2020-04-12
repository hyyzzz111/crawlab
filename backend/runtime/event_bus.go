package runtime

import (
	"crawlab/pkg/core/broker"
	"crawlab/pkg/core/broker/memory"
	"crawlab/pkg/core/codec/gob"
	"github.com/davecgh/go-spew/spew"
)

type BusEvent int

//go:generate enumer -type BusEvent -linecomment
const (
	ModeChange BusEvent = iota
)

type eventBus struct {
	inner broker.Broker
}

var Bus = &eventBus{
	inner: memory.NewBrokerAndConnect(func(options *broker.Options) {
		options.Logger = Logger
		options.ErrorHandler = func(event broker.Event) error {
			spew.Dump(event)
			return nil
		}
		options.BodyCodec = &gob.Gob{}
	}),
}

func (e *eventBus) Publish(topic BusEvent, message *broker.Message) error {
	return e.inner.Publish(topic.String(), message)
}
func (e *eventBus) Subscribe(topic BusEvent, handler broker.Handler) (broker.Subscriber, error) {
	return e.inner.Subscribe(topic.String(), handler)
}
