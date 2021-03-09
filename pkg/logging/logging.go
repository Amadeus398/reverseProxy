package logging

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"sync"
	"time"
)

var (
	once sync.Once
)

type Logger struct {
	infoLog  *zerolog.Event
	warnLog  *zerolog.Event
	errorLog *zerolog.Event
	module   string
	method   string
}

// NewLogs returns pointer to Logger
// with a method and a module
func NewLogs(module, method string) *Logger {
	once.Do(func() {
		log.Logger = log.Output(zerolog.SyncWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}))
	})
	return &Logger{module: module, method: method}
}

// addMetadata adds module and method to each
// Get logger call
func (l Logger) addMetadata(e *zerolog.Event) *zerolog.Event {
	return e.Str("module", l.module).Str("method", l.method)
}

// GetInfo returns pointer to info logger
func (l Logger) GetInfo() *zerolog.Event {
	if l.infoLog == nil {
		l.infoLog = l.addMetadata(log.Info())
	}
	return l.infoLog
}

// GetWarn returns pointer to warn logger
func (l Logger) GetWarn() *zerolog.Event {
	if l.warnLog == nil {
		l.warnLog = l.addMetadata(log.Warn())
	}
	return l.warnLog
}

// GetError returns pointer to error logger
func (l Logger) GetError() *zerolog.Event {
	if l.errorLog == nil {
		l.errorLog = l.addMetadata(log.Error())
	}
	return l.errorLog
}
