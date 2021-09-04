package node

import "emag-homework/internal/db/store"

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

type Store interface {
	Get(k string) *store.Entry
	Put(e store.Entry) error
}
