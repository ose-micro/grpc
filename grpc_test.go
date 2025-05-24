package grpc_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/ose-micro/grpc/go_gen"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func startTestServer(t *testing.T) (addr string, cleanup func()) {
	lis, err := net.Listen("tcp", ":0") // random port
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	go_gen.RegisterTestServiceServer(grpcServer, &testPingServer{})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			t.Logf("gRPC server stopped: %v", err)
		}
	}()

	return lis.Addr().String(), func() {
		grpcServer.GracefulStop()
	}
}

type testPingServer struct {
	go_gen.UnimplementedTestServiceServer
}

func (s *testPingServer) Ping(ctx context.Context, req *go_gen.Empty) (*go_gen.Pong, error) {
	return &go_gen.Pong{Msg: "pong"}, nil
}

func TestPing(t *testing.T) {
	addr, cleanup := startTestServer(t)
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, addr, grpc.WithInsecure(), grpc.WithBlock())
	require.NoError(t, err)
	defer conn.Close()

	client := go_gen.NewTestServiceClient(conn)
	res, err := client.Ping(ctx, &go_gen.Empty{})
	require.NoError(t, err)
	require.Equal(t, "pong", res.Msg)
}
