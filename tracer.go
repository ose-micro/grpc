package grpc

import (
	"context"

	"github.com/ose-micro/core/tracing"
	"google.golang.org/grpc"
)

func WithTracing(tracer tracing.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
		}

		return resp, err
	}
}
