package runtime

import (
	"crawlab/pkg/core/broker"

	"sync"
)

//go:generate enumer -type EnvMode -linecomment
const (
	Development EnvMode = iota
	Production
	Test
)

type EnvMode int

var modeListenerManager = &modeChangeListenerManager{mode: Development}

type modeChangeListenerManager struct {
	mode EnvMode
	mu   sync.RWMutex
}

func (m *modeChangeListenerManager) Mode() EnvMode {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.mode
}

func (m *modeChangeListenerManager) SetMode(newMode EnvMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.mode != newMode {
		_ = Bus.Publish(ModeChange, &broker.Message{
			Header: nil,
			Body:   newMode,
		})
	}
}
func (m *modeChangeListenerManager) Register(name string, handler broker.Handler) (broker.Subscriber, error) {
	return Bus.Subscribe(ModeChange, handler)
}

func SetMode(newMode EnvMode) {
	modeListenerManager.SetMode(newMode)
}
func GetMode() EnvMode {
	return modeListenerManager.Mode()
}

func IsRelease() bool {
	return GetMode() != Development && GetMode() != Test
}
