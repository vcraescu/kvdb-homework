package server

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/controller/service"
)

var _ v1.ControllerServer = (*ControllerServer)(nil)

type ControllerServer struct {
	service *service.Controller
}

func NewControllerServer(service *service.Controller) *ControllerServer {
	return &ControllerServer{
		service: service,
	}
}

func (s *ControllerServer) Put(ctx context.Context, req *v1.PutRequest) (*v1.PutResponse, error) {
	return s.service.Put(ctx, req)
}

func (s *ControllerServer) Get(ctx context.Context, req *v1.GetRequest) (*v1.GetResponse, error) {
	return s.service.Get(ctx, req)
}

func (s *ControllerServer) RegisterNode(
	ctx context.Context, req *v1.RegisterNodeRequest,
) (*v1.RegisterNodeResponse, error) {
	return s.service.RegisterNode(ctx, req)
}

func (s *ControllerServer) UnregisterNode(
	ctx context.Context, req *v1.UnregisterNodeRequest,
) (*v1.UnregisterNodeResponse, error) {
	return s.service.UnregisterNode(ctx, req)
}
