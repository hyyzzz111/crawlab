package ws

import (
	"github.com/panjf2000/ants/v2"
	"sync"
)

type hub struct {
	sessions      sync.Map
	broadcast     chan *envelope
	register      chan *Session
	unregister    chan *Session
	exit          chan *envelope
	open          bool
	rwMutex       *sync.RWMutex
	broadcastPool *ants.Pool
}

func newHub() (*hub, error) {
	pool, err := ants.NewPool(10000)
	if err != nil {
		return nil, err
	}
	return &hub{
		sessions:      sync.Map{},
		broadcast:     make(chan *envelope),
		register:      make(chan *Session),
		unregister:    make(chan *Session),
		exit:          make(chan *envelope),
		open:          true,
		rwMutex:       &sync.RWMutex{},
		broadcastPool: pool,
	}, nil
}

func (h *hub) run() {
loop:
	for {
		select {
		case s := <-h.register:
			h.sessions.Store(s, true)
		case s := <-h.unregister:
			h.sessions.Delete(s)
		case m := <-h.broadcast:
			h.sessions.Range(func(key, value interface{}) bool {
				session := key.(*Session)
				if m.filter == nil || m.filter(session) {
					h.broadcastPool.Submit(func() {
						session.writeMessage(m)
					})
				}
				return true
			})
		case m := <-h.exit:
			h.rwMutex.Lock()
			h.sessions.Range(func(key, value interface{}) bool {
				session := key.(*Session)
				session.writeMessage(m)
				Close(session)
				return true
			})
			h.sessions = sync.Map{}
			h.open = false
			h.rwMutex.Unlock()
			break loop
		}
	}
}

func (h *hub) closed() bool {
	h.rwMutex.RLock()
	defer h.rwMutex.RUnlock()
	return !h.open
}

func (h *hub) len() (l int) {
	h.sessions.Range(func(key, value interface{}) bool {
		l++
		return true
	})

	return l
}
