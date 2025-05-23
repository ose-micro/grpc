package main

import (
	"log"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	ose_grpc "github.com/ose-micro/grpc"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := logger.NewZap(logger.Config{
		Environment: "",
		Level: "info",
	})
	tracer, _ := tracing.NewOtel(tracing.Config{
		Endpoint: "localhost:4317",
		ServiceName: "ose-grpc-test",
		SampleRatio: 1.0,
	}, logger)

	err := ose_grpc.StartServer(ose_grpc.Config{
		Port: 50051,
	},ose_grpc.Param{
		RegisterFn: func(s *grpc.Server) {
			log.Println("Inside register")
			// pb.RegisterYourServiceServer(s, &YourService{})
		},
		Logger: logger,
		Tracer: tracer,
	})

	if err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}
}
