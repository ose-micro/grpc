package grpc

import (
	"fmt"
	"net"
	"net/http"
	"runtime"
	"time"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"github.com/ose-micro/grpc/middlewares/interceptors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

type Params struct {
	UnaryInterceptors  []grpc.UnaryServerInterceptor
	StreamInterceptors []grpc.StreamServerInterceptor
	Logger             logger.Logger
	Tracer             tracing.Tracer
	EnablePprof        bool // Optional: enable pprof
	MonitorGoroutines  bool // Optional: monitor goroutine count
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
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle: 5 * time.Minute,
			MaxConnectionAge:  15 * time.Minute,
			Time:              2 * time.Minute,
			Timeout:           20 * time.Second,
		}),
	}

	grpcServer := grpc.NewServer(opts...)

	s := &Server{
		grpcServer: grpcServer,
		log:        params.Logger,
	}

	if params.EnablePprof {
		s.enablePprof()
	}

	if params.MonitorGoroutines {
		s.monitorGoroutines()
	}

	return s, nil
}

func (s *Server) Serve(lis net.Listener, registerFn func(*grpc.Server)) error {
	if lis == nil {
		return fmt.Errorf("listener is nil")
	}

	// Register built-in services
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(s.grpcServer, healthSrv)
	reflection.Register(s.grpcServer)

	// Register app services
	registerFn(s.grpcServer)

	s.log.Info("gRPC server started", "addr", lis.Addr().String())
	return s.grpcServer.Serve(lis)
}

func (s *Server) Stop() error {
	if s.grpcServer != nil {
		s.log.Info("Stopping gRPC server...")
		s.grpcServer.GracefulStop()
		s.log.Info("gRPC server stopped")
	}
	return nil
}

func (s *Server) Instance() *grpc.Server {
	return s.grpcServer
}

// Optional: Enable pprof server
func (s *Server) enablePprof() {
	go func() {
		s.log.Info("pprof listening at http://localhost:6060/debug/pprof/")
		_ = http.ListenAndServe("localhost:6060", nil)
	}()
}

// Optional: Periodically logs number of active goroutines
func (s *Server) monitorGoroutines() {
	go func() {
		for range time.Tick(10 * time.Second) {
			s.log.Info("goroutine count", "count", runtime.NumGoroutine())
		}
	}()
}
