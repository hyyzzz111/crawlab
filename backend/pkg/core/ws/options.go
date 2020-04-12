package ws

import (
	"crawlab/pkg/core/codec"
	"crawlab/pkg/core/codec/json"
	"github.com/gorilla/websocket"
	"time"
)

// WsOptions engine configuration struct.
type Options struct {
	WriteWait         time.Duration // Milliseconds until write times out.
	PongWait          time.Duration // Timeout for waiting on pong.
	PingPeriod        time.Duration // Milliseconds between pings.
	MaxMessageSize    int64         // Maximum size in bytes of a message.
	MessageBufferSize int           // The max amount of messages that can be in a sessions buffer before it starts dropping them.
	Marshaller        codec.Marshaller
}
type Option func(options *Options)

type EngineOptions struct {
	WsOptions  Options
	Upgrader   *websocket.Upgrader
	Marshaller *json.Marshaler
}
type EngineOption func(options *EngineOptions)
