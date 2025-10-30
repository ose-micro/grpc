package middlewares

import (
	"context"
	"golang.org/x/sync/semaphore"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

// Allow up to 10 requests/sec, with burst 20
var rateLimiter = rate.NewLimiter(10, 20)

// Allow max 100 concurrent operations
var sem = semaphore.NewWeighted(100)

func RateLimitInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		// 1️⃣ Rate limiter: throttle request frequency
		if !rateLimiter.Allow() {
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		// 2️⃣ Concurrency limiter: prevent goroutine overload
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, status.Error(codes.ResourceExhausted, "server busy")
		}
		defer sem.Release(1)

		// 3️⃣ Execute handler safely
		resp, err := handler(ctx, req)
		return resp, err
	}
}
