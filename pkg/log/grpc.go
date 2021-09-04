package log

import (
	"context"
	"google.golang.org/grpc"
	"time"
)

func GRPCUnaryServerInterceptor(logger logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		logger.Info("GRPC_IN - Method: %s", info.FullMethod)

		// Calls the handler
		h, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			logger.Error("GRPC_OUT - Method: %s\tDuration: %s\tError: %v", info.FullMethod, duration, err)
		} else {
			logger.Info("GRPC_OUT - Method: %s\tDuration: %s", info.FullMethod, duration)
		}

		return h, err
	}
}
