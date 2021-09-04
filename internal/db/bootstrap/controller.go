package bootstrap

import (
	"context"
	"emag-homework/internal/app"
	"emag-homework/internal/db/api/v1"
	"emag-homework/internal/db/controller/healthz"
	"emag-homework/internal/db/controller/node"
	"emag-homework/internal/db/controller/server"
	"emag-homework/internal/db/controller/service"
	"emag-homework/pkg/env"
	"emag-homework/pkg/log"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
)

const (
	ctrlAddressEnv = "CTRL_ADDRESS"
)

func StartController() error {
	logger := log.NewLogger()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		logger.Info("system call: %+v", <-c)
		cancel()
	}()

	nodePool := node.NewPool()
	checker := healthz.NewChecker()
	svc := service.NewController(logger, nodePool, checker)
	srv := server.NewControllerServer(svc)
	defer svc.TearDown()

	address, err := env.Require(ctrlAddressEnv)
	if err != nil {
		return err
	}

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return StartControllerGRPCServer(ctx, lis, srv, logger)
}

func StartControllerGRPCServer(
	ctx context.Context, lis net.Listener, srv v1.ControllerServer, logger app.Logger,
) error {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(log.GRPCUnaryServerInterceptor(logger)),
	)

	v1.RegisterControllerServer(grpcSrv, srv)

	logger.Info("DB started at %s ...", lis.Addr().String())

	go func() {
		if err := grpcSrv.Serve(lis); err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()

	logger.Info("graceful stop...")

	grpcSrv.GracefulStop()

	return nil
}
