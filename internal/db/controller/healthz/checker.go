package healthz

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"sync"
	"time"
)

const (
	defaultCheckInterval = time.Second * 10
	defaultCheckTimeout  = time.Second * 2
)

type CheckFunc func(ctx context.Context, req *v1.HealthzRequest) (*v1.HealthzResponse, error)

type CheckEvent struct {
	ID  string
	Err error
	Res *v1.HealthzResponse
}

type Config struct {
	CheckInterval time.Duration
	CheckTimeout  time.Duration
}

type Option func(cfg *Config)

type checkItem struct {
	tick *time.Ticker
	id   string
	fn   CheckFunc
}

type Checker struct {
	checkers      map[string]*checkItem
	mu            sync.Mutex
	newCheckCh    chan *checkItem
	eventCh       chan CheckEvent
	doneCh        chan struct{}
	checkInterval time.Duration
	checkTimeout  time.Duration
}

func NewChecker(opts ...Option) *Checker {
	cfg := &Config{
		CheckInterval: defaultCheckInterval,
		CheckTimeout:  defaultCheckTimeout,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &Checker{
		checkers:      make(map[string]*checkItem),
		newCheckCh:    make(chan *checkItem, 100),
		eventCh:       make(chan CheckEvent, 100),
		doneCh:        make(chan struct{}),
		checkInterval: cfg.CheckInterval,
		checkTimeout:  cfg.CheckTimeout,
	}
}

func WithCheckInterval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.CheckInterval = interval
	}
}

func WithCheckTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		cfg.CheckTimeout = timeout
	}
}

func (c *Checker) Add(id string, fn CheckFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, ok := c.checkers[id]; ok {
		return
	}

	item := &checkItem{
		tick: time.NewTicker(c.checkInterval),
		id:   id,
		fn:   fn,
	}

	c.checkers[id] = item

	go func() {
		c.newCheckCh <- item
	}()
}

func (c *Checker) Remove(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.checkers[id]
	if ok {
		item.tick.Stop()
	}

	delete(c.checkers, id)
}

func (c *Checker) Start() {
	for item := range c.newCheckCh {
		go c.checkItem(item)
	}
}

func (c *Checker) Events() <-chan CheckEvent {
	return c.eventCh
}

func (c *Checker) Stop() {
	close(c.doneCh)
	close(c.newCheckCh)
}

func (c *Checker) sendEvent(id string, res *v1.HealthzResponse, err error) {
	go func() {
		c.eventCh <- CheckEvent{
			ID:  id,
			Err: err,
			Res: res,
		}
	}()
}

func (c *Checker) checkItem(item *checkItem) {
	for {
		select {
		case <-c.doneCh:
			return
		case <-item.tick.C:
			ctx, cancel := context.WithTimeout(context.Background(), c.checkTimeout)
			defer cancel()

			res, err := item.fn(ctx, &v1.HealthzRequest{})
			c.sendEvent(item.id, res, err)
		}
	}
}
