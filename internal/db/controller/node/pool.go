package node

import (
	v1 "emag-homework/internal/db/api/v1"
	"fmt"
	"google.golang.org/grpc"
	"sync"
)

type Pool struct {
	mu         sync.RWMutex
	nodes      map[string]*Item
	newNodesCh chan NewNodeEvent
	closed     bool
}

type NewNodeEvent struct {
	ID     string
	Client v1.NodeClient
}

func NewPool() *Pool {
	return &Pool{
		nodes:      make(map[string]*Item),
		newNodesCh: make(chan NewNodeEvent),
	}
}

func (p *Pool) Select() []*Item {
	p.mu.RLock()
	defer p.mu.RUnlock()

	items := make([]*Item, 0)

	for _, node := range p.nodes {
		if node.IsReady() {
			items = append(items, node)
		}
	}

	return items
}

func (p *Pool) Add(id, address string) (*Item, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("pool is closed")
	}

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("cannot connect to node %s", address)
	}

	client := v1.NewNodeClient(conn)

	p.nodes[id] = &Item{
		id:     id,
		conn:   conn,
		client: client,
	}

	go func() {
		p.newNodesCh <- NewNodeEvent{
			ID:     id,
			Client: client,
		}
	}()

	return p.nodes[id], nil
}

func (p *Pool) Remove(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.nodes, id)

	return nil
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	close(p.newNodesCh)

	for _, node := range p.nodes {
		_ = node.close()
	}

	return nil
}

func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.nodes)
}

func (p *Pool) MarkError(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	node, ok := p.nodes[id]
	if !ok {
		return
	}

	node.MarkError()
}

func (p *Pool) MarkReady(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	node, ok := p.nodes[id]
	if !ok {
		return
	}

	node.MarkReady()
}
