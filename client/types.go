package client

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Config struct {
	Target       string
	Insecure     bool
	TLSCertPath  string // optional path to client cert (for mTLS)
	Authority    string // for server name override
	TimeoutSec   int    // timeout per dial in seconds
	EnableRetry  bool   // enable gRPC built-in retry
	ServiceName  string // optional: for load balancing with xDS or DNS
	CallTimeoutSec int // per-RPC timeout
}

// CleanGRPCError transforms a gRPC error to a simplified error
func cleanGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		// not a grpc status error, return as is
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return errors.New("resource not found")
	case codes.InvalidArgument:
		return errors.New("invalid input provided")
	case codes.Unauthenticated:
		return errors.New("unauthenticated request")
	case codes.PermissionDenied:
		return errors.New("permission denied")
	case codes.DeadlineExceeded:
		return errors.New("request timed out")
	// add other cases as needed
	default:
		return errors.New(st.Message())
	}
}
