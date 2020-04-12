package ws

import (
	"errors"

	"github.com/gorilla/websocket"
)

// EventHandlerFunc defines the handler used by gin middleware as return value.
type EventHandlerFunc func(ctx *Context)
type SessionHandlerFunc func(session *Session)
type CloseHandlerFunc func(ctx *Session, code int, text string) error
type ErrorHandlerFunc func(ctx *Context, err error)
type EventHandlersChain []EventHandlerFunc

// Last returns the last handler in the chain. ie. the last handler is the main own.
func (c EventHandlersChain) Last() EventHandlerFunc {
	if length := len(c); length > 0 {
		return c[length-1]
	}
	return nil
}

// Close codes defined in RFC 6455, section 11.7.
// Duplicate of codes from gorilla/websocket for convenience.
const (
	CloseNormalClosure           = 1000
	CloseGoingAway               = 1001
	CloseProtocolError           = 1002
	CloseUnsupportedData         = 1003
	CloseNoStatusReceived        = 1005
	CloseAbnormalClosure         = 1006
	CloseInvalidFramePayloadData = 1007
	ClosePolicyViolation         = 1008
	CloseMessageTooBig           = 1009
	CloseMandatoryExtension      = 1010
	CloseInternalServerErr       = 1011
	CloseServiceRestart          = 1012
	CloseTryAgainLater           = 1013
	CloseTLSHandshake            = 1015
)

// Duplicate of codes from gorilla/websocket for convenience.
var validReceivedCloseCodes = map[int]bool{
	// see http://www.iana.org/assignments/websocket/websocket.xhtml#close-code-number

	CloseNormalClosure:           true,
	CloseGoingAway:               true,
	CloseProtocolError:           true,
	CloseUnsupportedData:         true,
	CloseNoStatusReceived:        false,
	CloseAbnormalClosure:         false,
	CloseInvalidFramePayloadData: true,
	ClosePolicyViolation:         true,
	CloseMessageTooBig:           true,
	CloseMandatoryExtension:      true,
	CloseInternalServerErr:       true,
	CloseServiceRestart:          true,
	CloseTryAgainLater:           true,
	CloseTLSHandshake:            false,
}

type filterFunc func(*Session) bool

// Broadcast broadcasts a text message to all sessions.
func (e *Engine) Broadcast(msg []byte) error {
	if e.hub.closed() {
		return errors.New("engine instance is closed")
	}

	message := &envelope{t: websocket.TextMessage, msg: msg}
	e.hub.broadcast <- message

	return nil
}

// BroadcastFilter broadcasts a text message to all sessions that fn returns true for.
func (e *Engine) BroadcastFilter(msg []byte, fn func(*Session) bool) error {
	if e.hub.closed() {
		return errors.New("engine instance is closed")
	}

	message := &envelope{t: websocket.TextMessage, msg: msg, filter: fn}
	e.hub.broadcast <- message

	return nil
}

// BroadcastOthers broadcasts a text message to all sessions except session s.
func (e *Engine) BroadcastOthers(msg []byte, s *Session) error {
	return e.BroadcastFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// BroadcastMultiple broadcasts a text message to multiple sessions given in the sessions slice.
func (e *Engine) BroadcastMultiple(msg []byte, sessions []*Session) error {
	for _, sess := range sessions {
		if writeErr := sess.Write(msg); writeErr != nil {
			return writeErr
		}
	}
	return nil
}

// BroadcastBinary broadcasts a binary message to all sessions.
func (e *Engine) BroadcastBinary(msg []byte) error {
	if e.hub.closed() {
		return errors.New("engine instance is closed")
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg}
	e.hub.broadcast <- message

	return nil
}

// BroadcastBinaryFilter broadcasts a binary message to all sessions that fn returns true for.
func (e *Engine) BroadcastBinaryFilter(msg []byte, fn func(*Session) bool) error {
	if e.hub.closed() {
		return errors.New("engine instance is closed")
	}

	message := &envelope{t: websocket.BinaryMessage, msg: msg, filter: fn}
	e.hub.broadcast <- message

	return nil
}

// BroadcastBinaryOthers broadcasts a binary message to all sessions except session s.
func (e *Engine) BroadcastBinaryOthers(msg []byte, s *Session) error {
	return e.BroadcastBinaryFilter(msg, func(q *Session) bool {
		return s != q
	})
}

// Close closes the engine instance and all connected sessions.
func (e *Engine) Close() error {
	if e.hub.closed() {
		return errors.New("engine instance is already closed")
	}

	e.hub.exit <- &envelope{t: websocket.CloseMessage, msg: []byte{}}

	return nil
}

// CloseWithMsg closes the engine instance with the given close payload and all connected sessions.
// Use the FormatCloseMessage function to format a proper close message payload.
func (e *Engine) CloseWithMsg(msg []byte) error {
	if e.hub.closed() {
		return errors.New("engine instance is already closed")
	}

	e.hub.exit <- &envelope{t: websocket.CloseMessage, msg: msg}

	return nil
}

// Len return the number of connected sessions.
func (e *Engine) Len() int {
	return e.hub.len()
}

// IsClosed returns the status of the engine instance.
func (e *Engine) IsClosed() bool {
	return e.hub.closed()
}

// FormatCloseMessage formats closeCode and text as a WebSocket close message.
func FormatCloseMessage(closeCode int, text string) []byte {
	return websocket.FormatCloseMessage(closeCode, text)
}
