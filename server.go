package grpc

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Param struct {
	Middlewares []grpc.UnaryServerInterceptor
	RegisterFn  func(*grpc.Server)
}

func StartServer(conf Config, param Param) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", conf.Port, err)
	}

	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(param.Middlewares...),
	}
	grpcServer := grpc.NewServer(serverOptions...)

	// Register health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthSrv)

	// Register reflection
	reflection.Register(grpcServer)

	// Register actual services
	param.RegisterFn(grpcServer)

	return grpcServer.Serve(listener)
}
