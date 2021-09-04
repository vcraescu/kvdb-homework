package bootstrap

import (
	"context"
	v1 "emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/node"
	"emag-homework/internal/db/node/server"
	"emag-homework/internal/db/store"
	"emag-homework/pkg/env"
	"emag-homework/pkg/log"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"time"
)

const (
	nodeAddressEnv = "NODE_ADDRESS"
	storePathEnv   = "STORE_PATH"
)

func StartNode() error {
	logger := log.NewLogger()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		logger.Info("system call: %+v", <-c)
		cancel()
	}()

	address, err := env.Require(nodeAddressEnv)
	if err != nil {
		return err
	}

	storePath, err := env.Require(storePathEnv)
	if err != nil {
		return err
	}

	ctrlAddress, err := env.Require(ctrlAddressEnv)
	if err != nil {
		return err
	}

	cc, err := grpc.Dial(ctrlAddress, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer cc.Close()

	ctrlClient := v1.NewControllerClient(cc)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	s, err := store.New(store.WithFilename(storePath), store.WithLogger(logger))
	if err != nil {
		return fmt.Errorf("failed creating store: %w", err)
	}
	defer s.Close()

	nodeInfo := server.NodeInfo{
		ID:      server.GenerateID(),
		Address: lis.Addr().String(),
	}
	srv := server.NewNodeServer(s, nodeInfo.ID)

	doneCh := make(chan struct{}, 1)
	errCh := make(chan error, 1)

	go func() {
		if err := StartNodeGRPCServer(ctx, lis, srv, logger); err != nil {
			errCh <- err

			close(errCh)
		}

		close(doneCh)
	}()

	var registered bool

	defer func() {
		if !registered {
			return
		}

		if err := server.Unregister(context.Background(), nodeInfo, ctrlClient, logger); err != nil {
			logger.Error(err.Error())
		}
	}()

	t := time.NewTicker(time.Second * 5)

	if err := server.Register(ctx, nodeInfo, ctrlClient, logger); err != nil {
		logger.Error(err.Error())

		registered = false
	}

	registered = true

	for {
		select {
		case <-doneCh:
			return nil
		case err := <-errCh:
			return err
		case <-t.C:
			if err := server.Register(ctx, nodeInfo, ctrlClient, logger); err != nil {
				logger.Error(err.Error())

				registered = false

			}

			registered = true
		}
	}
}

func StartNodeGRPCServer(
	ctx context.Context, lis net.Listener, srv v1.NodeServer, logger node.Logger,
) error {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(log.GRPCUnaryServerInterceptor(logger)),
	)

	v1.RegisterNodeServer(grpcSrv, srv)

	logger.Info("node started at %s ...", lis.Addr().String())

	errCh := make(chan error)

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err

			close(errCh)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("graceful stop...")

		grpcSrv.GracefulStop()
	case err := <-errCh:
		return err
	}

	return nil
}
