# ose-grpc

> A standardized gRPC server module for Go microservices. Designed to work beautifully with your own tracing setup (e.g. `ose-micro`), metrics, recovery, and structured logging.

[![Go Reference](https://pkg.go.dev/badge/github.com/ose-micro/grpc.svg)](https://pkg.go.dev/github.com/ose-micro/grpc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ose-micro/grpc)](https://goreportcard.com/report/github.com/ose-micro/grpc)
[![License](https://img.shields.io/github/license/ose-micro/grpc)](LICENSE)

## ✨ Features

- Custom tracing interceptor (OpenTelemetry-compatible)
- Logging via Zap (with context)
- Panic recovery
- Prometheus metrics-ready
- Health + reflection auto-registration
- Unary and stream interceptors supported

## 🚀 Usage

```go
logger, _ := logger.NewZap(logger.Config{
    Environment: "",
    Level: "info",
})
tracer, _ := tracing.NewOtel(tracing.Config{
    Endpoint: "localhost:4317",
    ServiceName: "ose-grpc-test",
    SampleRatio: 1.0,
}, logger)

if err := ose_grpc.StartServer(ose_grpc.Config{
		Port: 50051,
	},ose_grpc.Param{
		RegisterFn: func(s *grpc.Server) {
			log.Println("Inside register")
			// pb.RegisterYourServiceServer(s, &YourService{})
		},
		Logger: logger,
		Tracer: tracer,
	}); err != nil {
    logger.Fatal(err)
}
```
