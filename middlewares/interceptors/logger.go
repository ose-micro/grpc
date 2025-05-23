package interceptors

import (
	"context"

	"github.com/ose-micro/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func LoggingInterceptor(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		logger.Info("gRPC request", zap.String("method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			logger.Error("gRPC error", zap.String("method", info.FullMethod), zap.Error(err))
		}
		return resp, err
	}
}
