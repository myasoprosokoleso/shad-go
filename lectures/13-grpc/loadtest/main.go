// loadtest fires gRPC requests at a Calculator server at a target RPS,
// then prints throughput, latency percentiles, and per-code error counts.
//
// Tracing is opt-in: pass -trace to attach a TracerProvider. With many requests
// you usually want a small sample ratio (e.g. -trace-sample 0.01) so the UI
// stays useful and the SDK doesn't dominate latency.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "playground/grpc-demo/gen/calculatorpb"
	"playground/grpc-demo/internal/otelinit"
)

type methodFn func(context.Context, *pb.BinaryOp, ...grpc.CallOption) (*pb.Result, error)

func pickMethod(client pb.CalculatorClient, name string) (methodFn, error) {
	switch name {
	case "add":
		return client.Add, nil
	case "sub":
		return client.Sub, nil
	case "mul":
		return client.Mul, nil
	case "div":
		return client.Div, nil
	}
	return nil, fmt.Errorf("unknown method %q (want add|sub|mul|div)", name)
}

type sample struct {
	latency time.Duration
	errCode string
}

func main() {
	addr := flag.String("addr", "localhost:50051", "server address")
	rps := flag.Int("rps", 100, "target requests per second")
	dur := flag.Duration("duration", 10*time.Second, "test duration")
	conc := flag.Int("concurrency", 16, "number of worker goroutines")
	method := flag.String("method", "add", "RPC method: add|sub|mul|div")
	a := flag.Float64("a", 6, "operand a")
	b := flag.Float64("b", 7, "operand b")
	timeout := flag.Duration("timeout", 2*time.Second, "per-request timeout")
	traceOn := flag.Bool("trace", true, "enable OpenTelemetry tracing")
	traceSample := flag.Float64("trace-sample", 0.01, "trace sample ratio when -trace is set")
	otelEndpoint := flag.String("otel-endpoint", envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"), "OTLP/gRPC endpoint")
	serviceName := flag.String("service-name", envOr("OTEL_SERVICE_NAME", "calculator-loadtest"), "OpenTelemetry service.name")
	flag.Parse()

	if *rps <= 0 || *conc <= 0 || *dur <= 0 {
		log.Fatalf("rps, concurrency and duration must be > 0")
	}

	if *traceOn {
		shutdown, err := otelinit.Setup(context.Background(), otelinit.Config{
			ServiceName: *serviceName,
			Endpoint:    *otelEndpoint,
			SampleRatio: *traceSample,
		})
		if err != nil {
			log.Fatalf("otel setup: %v", err)
		}
		defer func() {
			ctx, c := context.WithTimeout(context.Background(), 5*time.Second)
			defer c()
			_ = shutdown(ctx)
		}()
	}

	dialOpts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	if *traceOn {
		dialOpts = append(dialOpts, grpc.WithStatsHandler(otelgrpc.NewClientHandler()))
	}
	conn, err := grpc.NewClient(*addr, dialOpts...)
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()

	client := pb.NewCalculatorClient(conn)
	call, err := pickMethod(client, *method)
	if err != nil {
		log.Fatal(err)
	}

	limiter := rate.NewLimiter(rate.Limit(*rps), 1)

	ctx, cancel := context.WithTimeout(context.Background(), *dur)
	defer cancel()

	go func() {
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigs:
			cancel()
		case <-ctx.Done():
		}
	}()

	op := &pb.BinaryOp{A: *a, B: *b}

	ticks := make(chan struct{}, *conc)
	samples := make(chan sample, *conc*4)

	var sent, ok, fail int64

	var wg sync.WaitGroup
	wg.Add(*conc)
	for i := 0; i < *conc; i++ {
		go func() {
			defer wg.Done()
			for range ticks {
				atomic.AddInt64(&sent, 1)
				reqCtx, reqCancel := context.WithTimeout(context.Background(), *timeout)
				start := time.Now()
				_, err := call(reqCtx, op)
				lat := time.Since(start)
				reqCancel()

				s := sample{latency: lat}
				if err != nil {
					atomic.AddInt64(&fail, 1)
					s.errCode = status.Code(err).String()
				} else {
					atomic.AddInt64(&ok, 1)
				}
				samples <- s
			}
		}()
	}

	var (
		latencies []time.Duration
		errCounts = map[string]int{}
		collectWG sync.WaitGroup
	)
	collectWG.Add(1)
	go func() {
		defer collectWG.Done()
		for s := range samples {
			latencies = append(latencies, s.latency)
			if s.errCode != "" {
				errCounts[s.errCode]++
			}
		}
	}()

	startWall := time.Now()
	fmt.Printf("firing %d rps for %s at %s (method=%s, concurrency=%d, trace=%v)\n",
		*rps, *dur, *addr, *method, *conc, *traceOn)

producer:
	for {
		if err := limiter.Wait(ctx); err != nil {
			break producer
		}
		select {
		case ticks <- struct{}{}:
		case <-ctx.Done():
			break producer
		}
	}

	close(ticks)
	wg.Wait()
	close(samples)
	collectWG.Wait()
	elapsed := time.Since(startWall)

	printReport(elapsed, sent, ok, fail, errCounts, latencies)
}

func printReport(elapsed time.Duration, sent, ok, fail int64, errCounts map[string]int, lats []time.Duration) {
	fmt.Println()
	fmt.Println("==== results ====")
	fmt.Printf("elapsed:     %s\n", elapsed.Round(time.Millisecond))
	fmt.Printf("sent:        %d\n", sent)
	fmt.Printf("success:     %d\n", ok)
	fmt.Printf("errors:      %d\n", fail)
	if elapsed > 0 {
		fmt.Printf("actual rps:  %.1f\n", float64(sent)/elapsed.Seconds())
	}

	if len(errCounts) > 0 {
		fmt.Println("errors by code:")
		keys := make([]string, 0, len(errCounts))
		for k := range errCounts {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("  %-20s %d\n", k, errCounts[k])
		}
	}

	if len(lats) == 0 {
		return
	}
	sort.Slice(lats, func(i, j int) bool { return lats[i] < lats[j] })
	fmt.Println("latency:")
	fmt.Printf("  min:  %s\n", lats[0].Round(time.Microsecond))
	fmt.Printf("  p50:  %s\n", percentile(lats, 0.50).Round(time.Microsecond))
	fmt.Printf("  p90:  %s\n", percentile(lats, 0.90).Round(time.Microsecond))
	fmt.Printf("  p95:  %s\n", percentile(lats, 0.95).Round(time.Microsecond))
	fmt.Printf("  p99:  %s\n", percentile(lats, 0.99).Round(time.Microsecond))
	fmt.Printf("  max:  %s\n", lats[len(lats)-1].Round(time.Microsecond))
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
