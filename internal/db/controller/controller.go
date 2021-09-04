package controller

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}
