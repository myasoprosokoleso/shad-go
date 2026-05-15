package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "playground/grpc-demo/gen/calculatorpb"
	"playground/grpc-demo/internal/otelinit"
)

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	a := flag.Float64("a", 6, "operand a")
	b := flag.Float64("b", 7, "operand b")
	otelEndpoint := flag.String("otel-endpoint", envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"), "OTLP/gRPC endpoint (host:port)")
	otelSample := flag.Float64("otel-sample", 1.0, "trace sample ratio 0..1")
	serviceName := flag.String("service-name", envOr("OTEL_SERVICE_NAME", "calculator-client"), "OpenTelemetry service.name")
	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
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

	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)
	op := &pb.BinaryOp{A: *a, B: *b}

	call := func(name string, fn func(context.Context, *pb.BinaryOp, ...grpc.CallOption) (*pb.Result, error)) {
		res, err := fn(ctx, op)
		if err != nil {
			fmt.Printf("%-4s %v %v -> ERROR: %v\n", name, *a, *b, err)
			return
		}
		fmt.Printf("%-4s %v %v -> %v\n", name, *a, *b, res.GetValue())
	}

	call("Add", client.Add)
	call("Sub", client.Sub)
	call("Mul", client.Mul)
	call("Div", client.Div)

	zero := &pb.BinaryOp{A: *a, B: 0}
	if _, err := client.Div(ctx, zero); err != nil {
		fmt.Printf("Div  %v 0 -> ERROR (expected): %v\n", *a, err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
