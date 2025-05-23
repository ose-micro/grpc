package grpc

import (
	"context"

	"github.com/ose-micro/core/tracing"
	"google.golang.org/grpc"
)


func WithStreamTracing(tracer tracing.Tracer) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		// Wrap the stream so the new context with span is used
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrapped)
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
