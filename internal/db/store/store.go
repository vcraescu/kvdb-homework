package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

const (
	defaultFlushInterval = time.Second * 3
)

type Entry struct {
	Key     string
	Value   []byte
	Version int64
}

type Config struct {
	Logger        Logger
	FlushInterval time.Duration
	NoPersist     bool
	Filename      string
}

type Option func(cfg *Config)

type Store struct {
	data          map[string]Entry
	mu            sync.Mutex
	fd            *os.File
	filename      string
	flushCh       chan struct{}
	logger        Logger
	flushInterval time.Duration
}

func New(opts ...Option) (*Store, error) {
	cfg := &Config{
		Logger:        noOpLogger{},
		FlushInterval: defaultFlushInterval,
	}

	for _, opt := range opts {
		opt(cfg)
	}

	s := &Store{
		data:          make(map[string]Entry),
		filename:      cfg.Filename,
		logger:        cfg.Logger,
		flushInterval: cfg.FlushInterval,
		flushCh:       make(chan struct{}),
	}

	if !s.IsPersisted() {
		s.flushCh = make(chan struct{}, 100)
	}

	if err := s.setup(); err != nil {
		return nil, err
	}

	return s, nil
}

func WithFilename(filename string) Option {
	return func(cfg *Config) {
		cfg.Filename = filename
	}
}

func WithLogger(logger Logger) Option {
	return func(cfg *Config) {
		cfg.Logger = logger
	}
}

func WithFlushInternval(interval time.Duration) Option {
	return func(cfg *Config) {
		cfg.FlushInterval = interval
	}
}

func (s *Store) Get(k string) *Entry {
	s.mu.Lock()
	defer s.mu.Unlock()

	e, ok := s.data[k]
	if !ok {
		return nil
	}

	return &e
}

func (s *Store) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.data)
}

func (s *Store) Del(k string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, k)

	s.notifyFlush()

	return nil
}

func (s *Store) Put(entry Entry) error {
	if entry.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if entry.Version == 0 {
		return fmt.Errorf("version cannot be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	defer s.notifyFlush()

	if found, ok := s.data[entry.Key]; ok && found.Version > entry.Version {
		return nil
	}

	s.data[entry.Key] = entry

	return nil
}

func (s *Store) notifyFlush() {
	go func() {
		if !s.IsPersisted() {
			return
		}

		s.flushCh <- struct{}{}
	}()
}

func (s *Store) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.flush()
}

func (s *Store) flush() error {
	if !s.IsPersisted() {
		return nil
	}

	s.logger.Info("flushing to disk...")

	if err := s.fd.Truncate(0); err != nil {
		return fmt.Errorf("failed to truncate: %w", err)
	}

	if _, err := s.fd.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek: %w", err)
	}

	if err := json.NewEncoder(s.fd).Encode(s.data); err != nil {
		return fmt.Errorf("failed json encoding store data: %w", err)
	}

	return s.fd.Sync()
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fd == nil {
		return nil
	}

	if err := s.flush(); err != nil {
		return fmt.Errorf("failed flushing: %w", err)
	}

	if err := s.fd.Close(); err != nil {
		return fmt.Errorf("failed closing file: %w", err)
	}

	s.fd = nil

	return nil
}

func (s *Store) Clean() error {
	_ = s.Close()

	return os.Remove(s.filename)
}

func (s *Store) setup() error {
	if !s.IsPersisted() {
		return nil
	}

	if err := s.load(); err != nil {
		return fmt.Errorf("failed loading store: %w", err)
	}

	go s.startFlushing()

	return nil
}

func (s *Store) load() error {
	var err error

	s.fd, err = os.OpenFile(s.filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return fmt.Errorf("failed opening file %q: %w", s.filename, err)
	}

	if err := json.NewDecoder(s.fd).Decode(&s.data); err != nil {
		if err != io.EOF {
			_ = s.fd.Close()

			return fmt.Errorf("failed json decoding store data: %w", err)
		}
	}

	return nil
}

func (s *Store) startFlushing() {
	var count int
	t := time.NewTimer(s.flushInterval)

	for range s.flushCh {
		count++

		select {
		case <-t.C:
			if count <= 0 {
				break
			}

			if err := s.Flush(); err != nil {
				s.logger.Error("failed flush: %v", err)

				break
			}

			count = 0
		default:
			if count <= 100 {
				break
			}

			if err := s.Flush(); err != nil {
				s.logger.Error("failed flush: %v", err)

				break
			}

			count = 0
			t.Reset(s.flushInterval)
		}
	}
}

func (s *Store) IsPersisted() bool {
	return s.filename != ""
}
