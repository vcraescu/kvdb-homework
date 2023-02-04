package log

import (
	"io"
	"log"
	"os"
)

type logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

var _ logger = (*Logger)(nil)

type Logger struct {
	info  *log.Logger
	error *log.Logger
}

func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stderr, "[INFO] ", log.LstdFlags),
		error: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
	}
}

func NewNopLogger() *Logger {
	l := NewLogger()

	l.info.SetOutput(io.Discard)
	l.error.SetOutput(io.Discard)

	return l
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.info.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.error.Printf(format, v...)
}
