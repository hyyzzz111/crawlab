package logger

import "github.com/apex/log"

type LoggerWrapper struct {
	logger log.Interface
}

func NewLogger(logger log.Interface) *LoggerWrapper {
	return &LoggerWrapper{logger: logger}
}

func (l *LoggerWrapper) Logger() log.Interface {
	if l.logger == nil {
		return log.Log
	}
	return l.logger
}
