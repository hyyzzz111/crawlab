package memory

import (
	"crawlab/core/broker"
	"crawlab/core/logger"
)

type memoryEvent struct {
	*logger.LoggerWrapper
	opts    broker.Options
	topic   string
	err     error
	message interface{}
}

func (m *memoryEvent) Topic() string {
	return m.topic
}

func (m *memoryEvent) Message() *broker.Message {
	switch v := m.message.(type) {
	case *broker.Message:
		return v
	case []byte:
		msg := &broker.Message{}

		if err := m.opts.Codec.Unmarshal(v, msg); err != nil {
			m.Logger().Errorf("[memory]: failed to unmarshal: %v\n", err)
			return nil
		}

		return msg
	}

	return nil
}

func (m *memoryEvent) Ack() error {
	return nil
}

func (m *memoryEvent) Error() error {
	return m.err
}