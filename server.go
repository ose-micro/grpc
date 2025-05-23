package grpc

import (
	"fmt"
	"net"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"github.com/ose-micro/grpc/middlewares/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Param struct {
	Middlewares []grpc.ServerOption
	RegisterFn  func(*grpc.Server)
	Logger      logger.Logger
	Tracer      tracing.Tracer
}

func StartServer(conf Config, param Param) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", conf.Port, err)
	}

	middlewares := param.Middlewares
	if param.Logger != nil {
		middlewares = append(middlewares, grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(param.Logger),
			interceptors.RecoveryInterceptor(param.Logger),
			WithTracing(param.Tracer),
		))
		middlewares = append(middlewares, grpc.StreamInterceptor(
			WithStreamTracing(param.Tracer),
		))
	}
	
	grpcServer := grpc.NewServer(middlewares...)

	// Register health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	// Register reflection
	reflection.Register(grpcServer)

	// Register actual services
	param.RegisterFn(grpcServer)

	param.Logger.Info(fmt.Sprintf("Server started on port: %d", conf.Port))

	return grpcServer.Serve(listener)
}
