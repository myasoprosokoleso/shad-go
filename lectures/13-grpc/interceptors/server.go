// Package interceptors demonstrates unary server interceptors:
// logging, auth and panic recovery. Production code usually uses
// go-grpc-middleware/v2, but the shape of the function is the same.
package interceptors

import (
	"context"
	"log"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingUnary logs every RPC: method, duration, status code.
func LoggingUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("rpc %s took=%s code=%s", info.FullMethod, time.Since(start), status.Code(err))
	return resp, err
}

// AuthUnary checks a bearer token from request metadata.
func AuthUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	tokens := md.Get("authorization")
	if len(tokens) == 0 || tokens[0] != "Bearer secret" {
		return nil, status.Error(codes.Unauthenticated, "missing or bad token")
	}
	return handler(ctx, req)
}

// RecoveryUnary turns panics into Internal errors instead of killing the server.
func RecoveryUnary(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic in %s: %v\n%s", info.FullMethod, r, debug.Stack())
			err = status.Errorf(codes.Internal, "panic: %v", r)
		}
	}()
	return handler(ctx, req)
}

// RegisterChain shows how to chain server interceptors at server construction.
func RegisterChain() *grpc.Server {
	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			RecoveryUnary, // outermost: first to defer, last to recover
			LoggingUnary,
			AuthUnary, // innermost: closest to handler
		),
	)
}
