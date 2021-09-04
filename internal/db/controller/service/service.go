package service

import (
	"context"
	"emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/controller"
	"emag-homework/internal/db/controller/healthz"
	"emag-homework/internal/db/controller/node"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
)

type HealthzChecker interface {
	Start()
	Stop()
	Add(id string, fn healthz.CheckFunc)
	Remove(id string)
	Events() <-chan healthz.CheckEvent
}

type Controller struct {
	logger         controller.Logger
	pool           NodePool
	healthzChecker HealthzChecker
}

type NodePool interface {
	io.Closer

	Add(id, address string) (*node.Item, error)
	Remove(id string) error
	Size() int
	Select() []*node.Item
	MarkError(id string)
	MarkReady(id string)
}

func NewController(logger controller.Logger, nodePool NodePool, healthzChecker HealthzChecker) *Controller {
	ctrl := &Controller{
		logger:         logger,
		pool:           nodePool,
		healthzChecker: healthzChecker,
	}

	go healthzChecker.Start()

	return ctrl
}

func (c *Controller) Put(ctx context.Context, req *v1.PutRequest) (*v1.PutResponse, error) {
	nodes := c.pool.Select()
	if len(nodes) == 0 {
		return nil, fmt.Errorf("nodes pool is empty")
	}

	var res *v1.PutResponse
	var err error

	for _, item := range nodes {
		if res, err = item.Client().Put(ctx, req); err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (c *Controller) Get(ctx context.Context, req *v1.GetRequest) (*v1.GetResponse, error) {
	items := c.pool.Select()
	if len(items) == 0 {
		return nil, fmt.Errorf("nodes pool is empty")
	}

	var res *v1.GetResponse
	var err error

	for _, item := range items {
		res, err = item.Client().Get(ctx, req)
		if err != nil {
			return nil, err
		}
	}

	return res, err
}

func (c *Controller) RegisterNode(_ context.Context, req *v1.RegisterNodeRequest) (*v1.RegisterNodeResponse, error) {
	item, err := c.pool.Add(req.Id, req.Address)
	if err != nil {
		return nil, fmt.Errorf("failed adding node to pool: %w", err)
	}

	c.healthzChecker.Add(req.Id, func(ctx context.Context, req *v1.HealthzRequest) (*v1.HealthzResponse, error) {
		return item.Client().Healthz(ctx, req)
	})

	return &v1.RegisterNodeResponse{}, nil
}

func (c *Controller) UnregisterNode(
	_ context.Context, req *v1.UnregisterNodeRequest,
) (*v1.UnregisterNodeResponse, error) {
	if err := c.pool.Remove(req.Id); err != nil {
		return nil, fmt.Errorf("failed removing node from pool: %w", err)
	}

	c.healthzChecker.Remove(req.Id)

	return &v1.UnregisterNodeResponse{}, nil
}

func (c *Controller) writeConsensus() int {
	return (c.pool.Size() + 1) / 2
}

func (c *Controller) readConsensus() int {
	return (c.pool.Size() + 1) / 2
}

func (c *Controller) TearDown() {
	c.healthzChecker.Stop()
	_ = c.pool.Close()
}

func (c *Controller) startHealthzChecker() {
	go func() {
		c.healthzChecker.Start()
	}()

	go func() {
		for evt := range c.healthzChecker.Events() {
			if evt.Err != nil {
				c.healthzChecker.Remove(evt.ID)
				c.pool.MarkError(evt.ID)

				continue
			}

			switch evt.Res.Code {
			case v1.HealthzResponse_HEALTHZ_OK:
				c.pool.MarkReady(evt.ID)
			case v1.HealthzResponse_HEALTHZ_ERROR:
				c.healthzChecker.Remove(evt.ID)
				c.pool.MarkError(evt.ID)
			}
		}
	}()
}

func isNotFound(err error) bool {
	s := status.Convert(err)
	if s == nil {
		return false
	}

	return s.Code() == codes.NotFound
}
