package interceptors

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TimingUnaryClient logs every outgoing call and its latency.
func TimingUnaryClient(
	ctx context.Context,
	method string,
	req, reply any,
	cc *grpc.ClientConn,
	invoker grpc.UnaryInvoker,
	opts ...grpc.CallOption,
) error {
	start := time.Now()
	err := invoker(ctx, method, req, reply, cc, opts...)
	log.Printf("client %s took=%s code=%s", method, time.Since(start), status.Code(err))
	return err
}

// AuthUnaryClient attaches a bearer token to every outgoing call.
func AuthUnaryClient(token string) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// DialWithChain wires the chain into a client connection.
func DialWithChain(addr, token string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr,
		grpc.WithChainUnaryInterceptor(
			TimingUnaryClient,
			AuthUnaryClient(token),
		),
	)
}
