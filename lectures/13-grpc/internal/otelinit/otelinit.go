// Package otelinit wires up an OTLP/gRPC trace exporter and a TracerProvider.
//
// One call to Setup at startup, defer the returned shutdown — that's it.
// The OTel global TracerProvider is set so otelgrpc and any library code
// using otel.Tracer(...) Just Works.
package otelinit

import (
	"context"
	"fmt"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.40.0"
)

// Config controls the exporter and sampling.
//
// Endpoint may be empty — OTEL_EXPORTER_OTLP_ENDPOINT is honored by the
// exporter as a fallback, matching the OTel spec.
type Config struct {
	ServiceName string
	Endpoint    string  // host:port, OTLP/gRPC; empty = use env
	SampleRatio float64 // 0..1; 1 = always, 0 = never
}

// Setup installs a global TracerProvider and returns a shutdown func.
// Call shutdown on exit so buffered spans are flushed.
func Setup(ctx context.Context, cfg Config) (func(context.Context) error, error) {
	opts := []otlptracegrpc.Option{otlptracegrpc.WithInsecure()}
	if cfg.Endpoint != "" {
		opts = append(opts, otlptracegrpc.WithEndpoint(cfg.Endpoint))
	}

	exp, err := otlptracegrpc.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("otlp exporter: %w", err)
	}

	res, err := resource.Merge(resource.Default(), resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		semconv.ServiceInstanceID(instanceID()),
	))
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	sampler := sdktrace.ParentBased(sdktrace.TraceIDRatioBased(cfg.SampleRatio))

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exp,
			sdktrace.WithMaxQueueSize(4096),
			sdktrace.WithMaxExportBatchSize(512),
			sdktrace.WithBatchTimeout(2*time.Second),
		),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	return tp.Shutdown, nil
}

func instanceID() string {
	if h, err := os.Hostname(); err == nil && h != "" {
		return h
	}
	return fmt.Sprintf("pid-%d", os.Getpid())
}
