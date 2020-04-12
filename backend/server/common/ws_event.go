package common

//go:generate enumer -type WebsocketEvent -linecomment -trimprefix WsEvent
const (
	WsEventWelcome WebsocketEvent = iota
)

type WebsocketEvent int

