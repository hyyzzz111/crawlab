package ws

import (
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/panjf2000/ants/v2"
	"math"
	"net/http"
	"sync"
	"time"
)

const abortIndex int8 = math.MaxInt8 / 2

// Session wrapper around websocket connections.
type Session struct {
	Request       *http.Request
	Keys          map[string]interface{}
	conn          *websocket.Conn
	output        chan *envelope
	engine        *Engine
	open          bool
	rwmutex       *sync.RWMutex
	coroutinePool *ants.Pool
}

func (s *Session) writeMessage(message *envelope) {
	if s.closed() {
		fmt.Println(errors.New("tried to write to closed a session"))
		return
	}

	select {
	case s.output <- message:
	default:
		fmt.Println(errors.New("session message buffer is full"))
	}
}

func (s *Session) writeRaw(message *envelope) (err error) {
	if s.closed() {
		return errors.New("tried to write to a closed session")
	}

	err = s.conn.SetWriteDeadline(time.Now().Add(s.engine.options.WriteWait))
	if err != nil {
		return err
	}
	err = s.conn.WriteMessage(message.t, message.msg)

	if err != nil {
		return err
	}

	return nil
}

func (s *Session) closed() bool {
	s.rwmutex.RLock()
	defer s.rwmutex.RUnlock()

	return !s.open
}

func (s *Session) close() {
	if !s.closed() {
		s.rwmutex.Lock()
		s.open = false
		Close(s.conn)
		close(s.output)
		s.rwmutex.Unlock()
	}
}

func (s *Session) ping() error {
	return s.writeRaw(&envelope{t: websocket.PingMessage, msg: []byte{}})
}

func (s *Session) writePump() {
	ticker := time.NewTicker(s.engine.options.PingPeriod)
	defer ticker.Stop()

loop:
	for {
		select {
		case msg, ok := <-s.output:
			if !ok {
				break loop
			}

			err := s.writeRaw(msg)

			if err != nil {
				fmt.Println(err)
				//s.engine.errorHandler(s, err)
				break loop
			}

			if msg.t == websocket.CloseMessage {
				break loop
			}
			//
			//if msg.t == websocket.TextMessage {
			//	s.engine.messageSentHandler(s, msg.msg)
			//}
			//
			//if msg.t == websocket.BinaryMessage {
			//	s.engine.messageSentHandlerBinary(s, msg.msg)
			//}
		case <-ticker.C:
			_ = s.ping()
		}
	}
}

func (s *Session) readPump() {
	s.conn.SetReadLimit(s.engine.options.MaxMessageSize)
	_ = s.conn.SetReadDeadline(time.Now().Add(s.engine.options.PongWait))

	s.conn.SetPongHandler(func(string) error {
		_ = s.conn.SetReadDeadline(time.Now().Add(s.engine.options.PongWait))
		if s.engine.pongHandler != nil {
			s.engine.pongHandler(s)
		}
		return nil
	})

	if s.engine.closeHandler != nil {
		s.conn.SetCloseHandler(func(code int, text string) error {
			return s.engine.closeHandler(s, code, text)
		})
	}

	for {
		t, message, err := s.conn.ReadMessage()

		if err != nil {
			fmt.Println(err)
			break
		}
		ctx := s.engine.pool.Get().(*Context)
		ctx.reset()
		ctx.data = message
		ctx.session = s
		ctx.Keys = s.Keys
		switch t {
		case websocket.TextMessage:
			ctx.msgType = Text
		case websocket.BinaryMessage:
			ctx.msgType = Binary
		}
		s.engine.messageHandler(ctx)
		s.engine.pool.Put(ctx)
	}
}

// Write writes message to session.
func (s *Session) Write(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.TextMessage, msg: msg})

	return nil
}

// WriteBinary writes a binary message to session.
func (s *Session) WriteBinary(msg []byte) error {
	if s.closed() {
		return errors.New("session is closed")
	}

	s.writeMessage(&envelope{t: websocket.BinaryMessage, msg: msg})

	return nil
}

// Close closes session.
func (s *Session) Close() error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: []byte{}})

	return nil
}

// CloseWithMsg closes the session with the provided payload.
// Use the FormatCloseMessage function to format a proper close message payload.
func (s *Session) CloseWithMsg(msg []byte) error {
	if s.closed() {
		return errors.New("session is already closed")
	}

	s.writeMessage(&envelope{t: websocket.CloseMessage, msg: msg})

	return nil
}

// Set is used to store a new key/value pair exclusivelly for this session.
// It also lazy initializes s.Keys if it was not used previously.
func (s *Session) Set(key string, value interface{}) {
	if s.Keys == nil {
		s.Keys = make(map[string]interface{})
	}

	s.Keys[key] = value
}

// Get returns the value for the given key, ie: (value, true).
// If the value does not exists it returns (nil, false)
func (s *Session) Get(key string) (value interface{}, exists bool) {
	if s.Keys != nil {
		value, exists = s.Keys[key]
	}

	return
}

// MustGet returns the value for the given key if it exists, otherwise it panics.
func (s *Session) MustGet(key string) interface{} {
	if value, exists := s.Get(key); exists {
		return value
	}

	panic("Key \"" + key + "\" does not exist")
}

// IsClosed returns the status of the connection.
func (s *Session) IsClosed() bool {
	return s.closed()
}
