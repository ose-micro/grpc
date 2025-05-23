package main

import (
	"log"

	ose_grpc "github.com/ose-micro/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, _ := zap.NewProduction()

	err := ose_grpc.StartServer(ose_grpc.Config{
		Port: 50051,
	},ose_grpc.Param{
		Middlewares: ose_grpc.DefaultMiddlewares(logger),
		RegisterFn: func(s *grpc.Server) {
			log.Println("Inside register")
			// pb.RegisterYourServiceServer(s, &YourService{})
		},
	})

	if err != nil {
		log.Fatalf("failed to start gRPC server: %v", err)
	}
}
