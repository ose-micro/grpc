package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/ose-micro/core/logger"
	"github.com/ose-micro/core/tracing"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/dns"
)
func newConn(conf Config, log logger.Logger, tracer tracing.Tracer) (*grpc.ClientConn, error) {
	var opts []grpc.DialOption

	// --- Security / Transport credentials ---
	if conf.Insecure {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		creds, err := loadTLSCredentials(conf)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	// --- Interceptors ---
	var unaryInterceptors []grpc.UnaryClientInterceptor
	var streamInterceptors []grpc.StreamClientInterceptor

	if tracer != nil {
		unaryInterceptors = append(unaryInterceptors, ClientTracingInterceptor(tracer))
		streamInterceptors = append(streamInterceptors, ClientStreamTracingInterceptor(tracer))
	}
	if log != nil {
		unaryInterceptors = append(unaryInterceptors, ClientLoggingInterceptor(log))
	}

	if len(unaryInterceptors) > 0 {
		opts = append(opts, grpc.WithChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		opts = append(opts, grpc.WithChainStreamInterceptor(streamInterceptors...))
	}

	// --- Retry Policy ---
	if conf.EnableRetry {
		opts = append(opts, grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{}],
				"retryPolicy": {
					"maxAttempts": 4,
					"initialBackoff": "0.1s",
					"maxBackoff": "1s",
					"backoffMultiplier": 2,
					"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
				}
			}]
		}`))
	}

	// --- Load Balancing (DNS + round_robin) ---
	if conf.ServiceName != "" {
		resolver.Register(dns.NewBuilder())
		opts = append(opts, grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, roundrobin.Name)))
	}

	// --- Dial timeout ---
	ctx := context.Background()
	if conf.TimeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(conf.TimeoutSec)*time.Second)
		defer cancel()
	}

	return grpc.DialContext(ctx, conf.Target, opts...)
}

// loadTLSCredentials loads client TLS config
func loadTLSCredentials(conf Config) (credentials.TransportCredentials, error) {
	if conf.TLSCertPath == "" {
		return credentials.NewClientTLSFromCert(nil, conf.Authority), nil
	}

	pemData, err := os.ReadFile(conf.TLSCertPath)
	if err != nil {
		return nil, err
	}

	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("failed to append certs")
	}

	cfg := &tls.Config{
		RootCAs:    cp,
		ServerName: conf.Authority,
	}

	return credentials.NewTLS(cfg), nil
}

func MakeCall[C any](
	parentCtx context.Context,
	conf Config,
	log logger.Logger,
	tracer tracing.Tracer,
	newClient func(*grpc.ClientConn) C,
	fn func(ctx context.Context, client C) error,
) error {
	// Set default call timeout to 5 seconds if not set
	timeout := 5 * time.Second
	if conf.CallTimeoutSec > 0 {
		timeout = time.Duration(conf.CallTimeoutSec) * time.Second
	}

	// Create context with timeout, derived from parentCtx if available
	ctx := parentCtx
	if ctx == nil {
		ctx = context.Background()
	}
	callCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Create gRPC client connection
	conn, err := newConn(conf, log, tracer)
	if err != nil {
		log.Error("failed to create gRPC client", zap.Error(err))
		return cleanGRPCError(err)
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			log.Error("failed to close gRPC client connection", zap.Error(cerr))
		}
	}()

	// Create typed client from connection
	client := newClient(conn)

	// Execute the passed function with context and client
	return fn(callCtx, client)
}
