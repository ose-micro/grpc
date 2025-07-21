package client

type Config struct {
	Target       string
	Insecure     bool
	TLSCertPath  string // optional path to client cert (for mTLS)
	Authority    string // for server name override
	TimeoutSec   int    // timeout per dial in seconds
	EnableRetry  bool   // enable gRPC built-in retry
	ServiceName  string // optional: for load balancing with xDS or DNS
}
