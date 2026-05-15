package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "playground/grpc-demo/gen/calculatorpb"
	"playground/grpc-demo/internal/otelinit"
)

type calcServer struct {
	pb.UnimplementedCalculatorServer
}

func (s *calcServer) Add(_ context.Context, op *pb.BinaryOp) (*pb.Result, error) {
	return &pb.Result{Value: op.GetA() + op.GetB()}, nil
}

func (s *calcServer) Sub(_ context.Context, op *pb.BinaryOp) (*pb.Result, error) {
	return &pb.Result{Value: op.GetA() - op.GetB()}, nil
}

func (s *calcServer) Mul(_ context.Context, op *pb.BinaryOp) (*pb.Result, error) {
	return &pb.Result{Value: op.GetA() * op.GetB()}, nil
}

func (s *calcServer) Div(_ context.Context, op *pb.BinaryOp) (*pb.Result, error) {
	if op.GetB() == 0 {
		return nil, status.Error(codes.InvalidArgument, "division by zero")
	}
	return &pb.Result{Value: op.GetA() / op.GetB()}, nil
}

func main() {
	addr := flag.String("addr", ":50051", "listen address")
	otelEndpoint := flag.String("otel-endpoint", envOr("OTEL_EXPORTER_OTLP_ENDPOINT", ""), "OTLP/gRPC endpoint (host:port)")
	otelSample := flag.Float64("otel-sample", 1.0, "trace sample ratio 0..1")
	serviceName := flag.String("service-name", envOr("OTEL_SERVICE_NAME", "calculator-server"), "OpenTelemetry service.name")
	flag.Parse()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	shutdownOtel, err := otelinit.Setup(ctx, otelinit.Config{
		ServiceName: *serviceName,
		Endpoint:    *otelEndpoint,
		SampleRatio: *otelSample,
	})
	if err != nil {
		log.Fatalf("otel setup: %v", err)
	}
	defer func() {
		shutdownCtx, c := context.WithTimeout(context.Background(), 5*time.Second)
		defer c()
		_ = shutdownOtel(shutdownCtx)
	}()

	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("listen %s: %v", *addr, err)
	}

	srv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	pb.RegisterCalculatorServer(srv, &calcServer{})
	reflection.Register(srv)

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs
		log.Println("shutting down")
		srv.GracefulStop()
	}()

	log.Printf("calculator gRPC server listening on %s (otel=%s, sample=%.2f)", lis.Addr(), *otelEndpoint, *otelSample)
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
