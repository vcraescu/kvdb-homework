package keyword

import (
	"context"
	"emag-homework/pkg/dbclient"
	"errors"
	"fmt"
	"strconv"
	"sync"
)

var ErrNotFound = errors.New("not found")

type DB interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Put(ctx context.Context, key string, value []byte) error
}

type Repository struct {
	db DB
}

func NewRepository(db DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Increment(ctx context.Context, keyword string, increment int) error {
	keyword, err := clean(keyword)
	if err != nil {
		return fmt.Errorf("failed cleaning up the text: %w", err)
	}

	v, err := r.Find(ctx, keyword)
	if err != nil {
		fmt.Println(err)
		if !errors.Is(err, dbclient.ErrNotFound) {
			return err
		}
	}

	return r.db.Put(ctx, keyword, []byte(fmt.Sprint(v+increment)))
}

func (r *Repository) Find(ctx context.Context, keyword string) (int, error) {
	keyword, err := clean(keyword)
	if err != nil {
		return 0, fmt.Errorf("failed cleaning up the text: %w", err)
	}

	b, err := r.db.Get(ctx, keyword)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(string(b))
}

type InMemRepository struct {
	data map[string]int
	mu   sync.RWMutex
}

func NewInMemRepository(data map[string]int) *InMemRepository {
	if data == nil {
		data = make(map[string]int)
	}

	return &InMemRepository{
		data: data,
	}
}

func (r *InMemRepository) Increment(_ context.Context, keyword string, increment int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.data[keyword] += increment

	return nil
}

func (r *InMemRepository) Find(_ context.Context, keyword string) (int, error) {
	r.mu.RLock()
	defer r.mu.RLock()

	v, ok := r.data[keyword]
	if !ok {
		return 0, ErrNotFound
	}

	return v, nil
}
