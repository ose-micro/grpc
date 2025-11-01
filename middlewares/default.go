package middlewares

import (
	"github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/ose-micro/core/tracing"
	ose_grpc "github.com/ose-micro/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func DefaultMiddlewares(logger *zap.Logger, tracer tracing.Tracer) []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		ose_grpc.WithTracing(tracer),
		grpc_recovery.UnaryServerInterceptor(),
		RateLimitInterceptor(),
	}
}
