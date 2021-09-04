package bootstrap

import (
	"context"
	v1 "emag-homework/api/v1"
	"emag-homework/internal/app"
	"emag-homework/internal/app/keyword"
	"emag-homework/internal/app/server"
	"emag-homework/pkg/dbclient"
	"emag-homework/pkg/env"
	"emag-homework/pkg/log"
	"fmt"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
)

const (
	appAddressEnv = "APP_ADDRESS"
	dbAddressEnv  = "DB_ADDRESS"
)

func Bootstrap() error {
	logger := log.NewLogger()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		logger.Info("system call: %+v", <-c)
		cancel()
	}()

	address, err := env.Require(appAddressEnv)
	if err != nil {
		return err
	}

	dbAddress, err := env.Require(dbAddressEnv)
	if err != nil {
		return err
	}

	db, err := dbclient.New(dbAddress)
	if err != nil {
		return fmt.Errorf("failed to connected to db %q: %w", dbAddress, err)
	}

	repository := keyword.NewRepository(db)
	counter := keyword.NewCounter()
	srv := server.NewAppServer(repository, counter, logger)

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	return StartGRPCServer(ctx, lis, srv, logger)
}

func StartGRPCServer(ctx context.Context, lis net.Listener, srv v1.AppServer, logger app.Logger) error {
	grpcSrv := grpc.NewServer(
		grpc.UnaryInterceptor(log.GRPCUnaryServerInterceptor(logger)),
	)

	v1.RegisterAppServer(grpcSrv, srv)

	logger.Info("Start listening at %s ...", lis.Addr().String())

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
