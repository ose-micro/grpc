package client

import (
	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func New(conf Config, log logger.Logger, tracer tracing.Tracer) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{}
	if conf.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	if tracer != nil {
		unaryInterceptors = append(unaryInterceptors, ClientTracingInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, ClientStreamTracingInterceptor(tracer))
	}
	if log != nil {
		unaryInterceptors = append(unaryInterceptors, ClientLoggingInterceptor(log))
	}

	if len(unaryInterceptors) > 0 {
		for _, interceptor := range unaryInterceptors {
			opts = append(opts, grpc.WithUnaryInterceptor(interceptor))
		}
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptors...))
	}

	return grpc.Dial(conf.Target, opts...)
}
