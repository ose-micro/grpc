package main

import (
	"net"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	ose_grpc "github.com/ose-micro/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := logger.NewZap(logger.Config{
		Environment: "",
		Level:       "info",
	})
	tracer, _ := tracing.NewOtel(tracing.Config{
		Endpoint:    "localhost:4317",
		ServiceName: "ose-grpc-test",
		SampleRatio: 1.0,
	}, logger)

	svc, err := ose_grpc.New(ose_grpc.Params{
		Logger: logger,
		Tracer: tracer,
	})
	if err != nil {
		logger.Fatal(err.Error())
	}

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		logger.Fatal("failed to listen", zap.Error(err))
	}

	if err := svc.Serve(lis, func(s *grpc.Server) {
		logger.Info("gRPC server listening on :50051")
	}); err != nil {
		logger.Fatal("gRPC server failed", zap.Error(err))
	}
}
