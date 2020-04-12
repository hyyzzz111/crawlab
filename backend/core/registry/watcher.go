package registry

import "time"

//go:generate enumer -type EventType -linecomment
const (
	Create EventType = iota
	Delete
	Update
	Unknown
)

type EventType int

type Watcher interface {
	// Next is a blocking call
	Next() (*Result, error)
	Stop()
}
type Event struct {
	// Id is registry id
	Id string
	// Type defines type of event
	Type EventType
	// Timestamp is event timestamp
	Timestamp time.Time
	// Service is registry service
	Service *Service
}

type Result struct {
	Action  string
	Service *Service
}
