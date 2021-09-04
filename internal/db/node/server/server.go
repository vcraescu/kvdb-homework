package server

import (
	"context"
	"emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/node"
	"emag-homework/internal/db/store"
	"fmt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ v1.NodeServer = (*NodeServer)(nil)

type NodeServer struct {
	store node.Store
	id    string
}

func NewNodeServer(store node.Store, id string) *NodeServer {
	return &NodeServer{
		store: store,
		id:    id,
	}
}

func (s *NodeServer) Put(_ context.Context, req *v1.PutRequest) (*v1.PutResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is missing")
	}

	if req.Version == 0 {
		return nil, status.Error(codes.InvalidArgument, "version is missing")
	}

	err := s.store.Put(store.Entry{
		Key:     req.Key,
		Value:   req.Value,
		Version: req.Version,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &v1.PutResponse{}, nil
}

func (s *NodeServer) Get(_ context.Context, req *v1.GetRequest) (*v1.GetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key is missing")
	}

	entry := s.store.Get(req.Key)
	if entry == nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("%q not found", req.Key))
	}

	return &v1.GetResponse{
		Value:   entry.Value,
		Version: entry.Version,
	}, nil
}

func (s *NodeServer) Healthz(_ context.Context, _ *v1.HealthzRequest) (*v1.HealthzResponse, error) {
	return &v1.HealthzResponse{
		Code: v1.HealthzResponse_HEALTHZ_OK,
		Id:   s.id,
	}, nil
}
