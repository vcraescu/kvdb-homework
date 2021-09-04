package store

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

var _ Logger = (*noOpLogger)(nil)

type noOpLogger struct{}

func (n noOpLogger) Info(format string, v ...interface{}) {}

func (n noOpLogger) Error(format string, v ...interface{}) {}
