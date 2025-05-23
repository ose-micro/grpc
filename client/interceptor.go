package client

import (
	"context"
	"time"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ClientTracingInterceptor injects trace context
func ClientTracingInterceptor(tracer tracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		ctx, span := tracer.Start(ctx, method)
		defer span.End()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		// simulate trace propagation
		md.Set("trace-id", span.SpanContext().TraceID().String())
		ctx = metadata.NewOutgoingContext(ctx, md)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ClientStreamTracingInterceptor adds tracing to stream calls
func ClientStreamTracingInterceptor(tracer tracing.Tracer) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer,
		opts ...grpc.CallOption) (grpc.ClientStream, error) {
		ctx, span := tracer.Start(ctx, method)
		defer span.End()
		return streamer(ctx, desc, cc, method, opts...)
	}
}

// ClientLoggingInterceptor logs client requests
func ClientLoggingInterceptor(logger logger.Logger) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		duration := time.Since(start)
		logger.Info("gRPC client call", zap.String("method", method), zap.Duration("duration", duration), zap.Error(err))
		return err
	}
}
