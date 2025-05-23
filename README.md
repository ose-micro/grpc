# ose-grpc

A standardized gRPC server module for Go microservices. Designed to work beautifully with your own tracing setup (e.g. `ose-micro`), metrics, recovery, and structured logging.

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
