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
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	Logger             logger.Logger
	Tracer             tracing.Tracer
}

type Server struct {
	grpcServer *grpc.Server
	log        logger.Logger
}

func New(params Params) (*Server, error) {
	var unaryInts []grpc.UnaryServerInterceptor
	var streamInts []grpc.StreamServerInterceptor

	// Default interceptors
	if params.Logger != nil {
		unaryInts = append(unaryInts,
			interceptors.LoggingInterceptor(params.Logger),
			interceptors.RecoveryInterceptor(params.Logger),
		)
	}
	if params.Tracer != nil {
		unaryInts = append(unaryInts, WithTracing(params.Tracer))
		streamInts = append(streamInts, WithStreamTracing(params.Tracer))
	}

	// Append user-defined interceptors
	unaryInts = append(unaryInts, params.UnaryInterceptors...)
	streamInts = append(streamInts, params.StreamInterceptors...)

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInts...),
		grpc.ChainStreamInterceptor(streamInts...),
	}

	grpcServer := grpc.NewServer(opts...)

	return &Server{
		grpcServer: grpcServer,
		log:        params.Logger,
	}, nil
}

func (s *Server) Serve(lis net.Listener, registerFn func(*grpc.Server)) error {
	// Register built-in services
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(s.grpcServer, healthSrv)
	reflection.Register(s.grpcServer)

	// Register application services
	registerFn(s.grpcServer)

	s.log.Info("gRPC server started", "addr", lis.Addr().String())
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() error {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
		s.log.Info("gRPC server stopped")
	}
	return nil
}

func (s *Server) Instance() *grpc.Server {
	return s.grpcServer
}
