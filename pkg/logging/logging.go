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

func NewLogs(module, method string) *Logger {
	once.Do(func() {
		log.Logger = log.Output(zerolog.SyncWriter(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}))
	})
	return &Logger{module: module, method: method}
}

func (l Logger) addMetadata(e *zerolog.Event) *zerolog.Event {
	return e.Str("module", l.module).Str("method", l.method)
}

func (l Logger) GetInfo() *zerolog.Event {
	if l.infoLog == nil {
		l.infoLog = l.addMetadata(log.Info())
	}
	return l.infoLog
}

func (l Logger) GetWarn() *zerolog.Event {
	if l.warnLog == nil {
		l.warnLog = l.addMetadata(log.Warn())
	}
	return l.warnLog
}

func (l Logger) GetError() *zerolog.Event {
	if l.errorLog == nil {
		l.errorLog = l.addMetadata(log.Error())
	}
	return l.errorLog
}
