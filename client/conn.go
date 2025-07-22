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
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/resolver/dns"
)
func New(conf Config, log logger.Logger, tracer tracing.Tracer) (*grpc.ClientConn, error) {
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

// WithCallContext returns a new context with a timeout for a single RPC call.
func WithCallContext(conf Config) (context.Context, context.CancelFunc) {
	timeout := 5 * time.Second // default if not set
	if conf.CallTimeoutSec > 0 {
		timeout = time.Duration(conf.CallTimeoutSec) * time.Second
	}
	return context.WithTimeout(context.Background(), timeout)
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