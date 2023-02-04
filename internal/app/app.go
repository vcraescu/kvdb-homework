package app

import "context"

type KeywordRepository interface {
	Increment(ctx context.Context, keyword string, increment int) error
	Find(ctx context.Context, keyword string) (int, error)
}

type KeywordCounter interface {
	Count(ctx context.Context, s string) (map[string]int, error)
}

type Logger interface {
	Info(format string, v ...interface{})
	Error(format string, v ...interface{})
}

