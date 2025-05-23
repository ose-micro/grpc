# Makefile for ose-grpc test proto generation
PROTO_DIR=.
PROTO_FILE=test.proto
GO_OUT=.

all: gen

gen:
	protoc \
		--go_out=$(GO_OUT) \
		--go-grpc_out=$(GO_OUT) \
		$(PROTO_DIR)/$(PROTO_FILE)

clean:
	rm -f ose_grpc/*.pb.go ose_grpc/*.grpc.pb.go
