package node

import (
	v1 "emag-homework/internal/db/api/v1"
	"google.golang.org/grpc"
	"sync"
)

const (
	readyNodeStatus   = 0
	errorNodeStatus   = 1
	pendingNodeStatus = 2
)

type Item struct {
	id     string
	conn   *grpc.ClientConn
	client v1.NodeClient
	status int
	mu     sync.RWMutex
}

func (n *Item) ID() string {
	return n.id
}

func (n *Item) close() error {
	if n.conn == nil {
		return nil
	}

	return n.conn.Close()
}

func (n *Item) IsReady() bool {
	return n.status == readyNodeStatus
}

func (n *Item) MarkError() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.status = errorNodeStatus
}

func (n *Item) MarkReady() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.status = readyNodeStatus
}

func (n *Item) Client() v1.NodeClient {
	return n.client
}
