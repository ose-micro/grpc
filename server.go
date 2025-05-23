package grpc

import (
	"net"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"github.com/ose-micro/grpc/middlewares/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Params struct {
	Middlewares []grpc.ServerOption
	Logger      logger.Logger
	Tracer      tracing.Tracer
}

type Server struct {
	grpcServer *grpc.Server
	log        logger.Logger
	tracer     tracing.Tracer
}

func (s Server) Server() *grpc.Server {
	return s.grpcServer
}

func (s Server) Serve(lis net.Listener, RegisterFn func(*grpc.Server)) error {
	// // Register health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(s.grpcServer, healthSrv)

	// // Register reflection
	reflection.Register(s.grpcServer)

	// // Register actual services
	RegisterFn(s.grpcServer)
	return s.grpcServer.Serve(lis)
}

func (s Server) Stop() error {
	s.grpcServer.GracefulStop()
	s.log.Info("gRPC server stopped")
	return nil
}

func New(params Params) (*Server, error) {
	middlewares := params.Middlewares
	if params.Logger != nil {
		middlewares = append(middlewares, grpc.ChainUnaryInterceptor(
			interceptors.LoggingInterceptor(params.Logger),
			interceptors.RecoveryInterceptor(params.Logger),
			WithTracing(params.Tracer),
		))
		middlewares = append(middlewares, grpc.StreamInterceptor(
			WithStreamTracing(params.Tracer),
		))
	}

	grpcServer := grpc.NewServer(middlewares...)
	return &Server{
		grpcServer: grpcServer,
		log:        params.Logger,
		tracer:     params.Tracer,
	}, nil
}
