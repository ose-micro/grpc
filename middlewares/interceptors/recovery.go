package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/ose-micro/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func RecoveryInterceptor(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("panic recovered", zap.String("method", info.FullMethod), zap.Any("recover", r), zap.ByteString("stack", debug.Stack()))
				err = fmt.Errorf("internal server error")
			}
		}()
		return handler(ctx, req)
	}
}