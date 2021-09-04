package dbclient

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sync"
	"time"
)

var ErrNotFound = errors.New("not found")

type Client struct {
	mu     sync.Mutex
	conn   *grpc.ClientConn
	client v1.ControllerClient
}

func New(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: v1.NewControllerClient(conn),
	}, nil
}

func (c *Client) Get(ctx context.Context, key string) ([]byte, error) {
	if c.client == nil {
		return nil, errors.New("closed connection")
	}

	res, err := c.client.Get(ctx, &v1.GetRequest{Key: key})
	if err != nil {
		if s := status.Convert(err); s != nil && s.Code() == codes.NotFound {
			return nil, ErrNotFound
		}

		return nil, fmt.Errorf("get failed: %w", err)
	}

	return res.Value, err
}

func (c *Client) Put(ctx context.Context, key string, value []byte) error {
	if c.client == nil {
		return errors.New("closed connection")
	}

	_, err := c.client.Put(ctx, &v1.PutRequest{
		Key:     key,
		Value:   value,
		Version: time.Now().UnixNano(),
	})
	if err != nil {
		return fmt.Errorf("put failed: %w", err)
	}

	return nil
}

func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		return nil
	}

	defer func() {
		c.conn = nil
	}()

	return c.conn.Close()
}
