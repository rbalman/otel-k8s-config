package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/baggage"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	promexporter "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	tracer      trace.Tracer
	meter       metric.Meter
	reqDuration metric.Float64Histogram
	reqTotal    metric.Int64Counter
	errTotal    metric.Int64Counter
)

// ---- OTel setup ----

func setupOTel(ctx context.Context) (func(context.Context) error, error) {
	endpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector.monitoring.svc.cluster.local:4317"
	}
	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "demo-app"
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("0.1.0"),
		),
		resource.WithFromEnv(),
	)
	if err != nil {
		return nil, fmt.Errorf("resource: %w", err)
	}

	traceExp, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("trace exporter: %w", err)
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)
	otel.SetTracerProvider(tp)

	promExp, err := promexporter.New()
	if err != nil {
		return nil, fmt.Errorf("prometheus exporter: %w", err)
	}
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(promExp),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		if err := tp.Shutdown(ctx); err != nil {
			return err
		}
		return mp.Shutdown(ctx)
	}, nil
}

// ---- Middleware ----

// baggageMiddleware generates a request_id per request, stores it in OTel
// baggage so it propagates to downstream services and can be correlated
// across traces, metrics, and logs.
func baggageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		m, _ := baggage.NewMember("request_id", requestID)
		bag, _ := baggage.New(m)
		ctx := baggage.ContextWithBaggage(r.Context(), bag)

		// Annotate the otelhttp span (created by the outer wrapper) with request_id.
		trace.SpanFromContext(ctx).SetAttributes(attribute.String("request_id", requestID))

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// statusRecorder captures the HTTP status code written by a handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(code int) {
	sr.status = code
	sr.ResponseWriter.WriteHeader(code)
}

// withMetrics records request duration histogram and counters for a named route.
func withMetrics(route string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		h(sr, r)

		attrs := metric.WithAttributes(
			attribute.String("http.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.status_code", sr.status),
		)
		reqDuration.Record(r.Context(), time.Since(start).Seconds(), attrs)
		reqTotal.Add(r.Context(), 1, attrs)
		if sr.status >= 500 {
			errTotal.Add(r.Context(), 1, attrs)
		}
	}
}

// ---- Logging helper ----

// logCtx logs at the given level enriched with trace_id, span_id, and
// request_id baggage — the three fields Grafana needs to correlate signals.
func logCtx(ctx context.Context, level slog.Level, msg string, args ...any) {
	sc := trace.SpanFromContext(ctx).SpanContext()
	extra := []any{}
	if sc.IsValid() {
		extra = append(extra, "trace_id", sc.TraceID().String(), "span_id", sc.SpanID().String())
	}
	if rid := baggage.FromContext(ctx).Member("request_id"); rid.Key() != "" {
		extra = append(extra, "request_id", rid.Value())
	}
	slog.Log(ctx, level, msg, append(extra, args...)...)
}

// ---- Handlers ----

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"ok"}`)
}

func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"ready"}`)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "hello")
	defer span.End()

	name := r.URL.Query().Get("name")
	if name == "" {
		name = "world"
	}
	span.SetAttributes(attribute.String("name", name))

	// Child span: validate input
	_, vs := tracer.Start(ctx, "validate-input")
	time.Sleep(5 * time.Millisecond)
	vs.End()

	logCtx(ctx, slog.LevelInfo, "hello request", "name", name)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"message":"hello, %s!"}`, name)
}

// itemsHandler simulates a multi-step operation with child spans.
func itemsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "list-items")
	defer span.End()

	// Child span: cache lookup
	_, cacheSpan := tracer.Start(ctx, "cache.get")
	time.Sleep(time.Duration(rand.Intn(30)+5) * time.Millisecond)
	cacheHit := rand.Float64() > 0.5
	cacheSpan.SetAttributes(attribute.Bool("cache.hit", cacheHit))
	cacheSpan.End()

	// Child span: DB query (only on cache miss)
	if !cacheHit {
		_, dbSpan := tracer.Start(ctx, "db.query")
		time.Sleep(time.Duration(rand.Intn(200)+20) * time.Millisecond)
		dbSpan.SetAttributes(attribute.String("db.statement", "SELECT * FROM items LIMIT 5"))
		dbSpan.End()
	}

	// Child span: serialize response
	_, serSpan := tracer.Start(ctx, "serialize")
	time.Sleep(time.Duration(rand.Intn(10)+2) * time.Millisecond)
	serSpan.End()

	span.SetAttributes(attribute.Bool("cache.hit", cacheHit), attribute.Int("items.count", 5))
	logLevel := slog.LevelInfo
	if !cacheHit {
		logLevel = slog.LevelWarn
	}
	logCtx(ctx, logLevel, "items fetched", "cache_hit", cacheHit, "count", 5)

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"items":["alpha","beta","gamma","delta","epsilon"]}`)
}

// errorHandler randomly fails ~60% of the time to produce error metrics and error spans.
func errorHandler(w http.ResponseWriter, r *http.Request) {
	ctx, span := tracer.Start(r.Context(), "flaky-op")
	defer span.End()

	if rand.Float64() < 0.6 {
		err := fmt.Errorf("random failure")
		span.RecordError(err)
		span.SetAttributes(attribute.Bool("error", true))
		logCtx(ctx, slog.LevelError, "operation failed", "reason", err.Error())
		http.Error(w, `{"error":"something went wrong"}`, http.StatusInternalServerError)
		return
	}

	logCtx(ctx, slog.LevelInfo, "operation succeeded")
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, `{"status":"success"}`)
}

// ---- Main ----

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})))

	shutdown, err := setupOTel(ctx)
	if err != nil {
		slog.Error("otel setup failed", "error", err)
		os.Exit(1)
	}
	defer shutdown(context.Background())

	tracer = otel.Tracer("demo-app")
	meter = otel.Meter("demo-app")

	reqDuration, _ = meter.Float64Histogram(
		"http_request_duration_seconds",
		metric.WithDescription("HTTP request latency"),
		metric.WithUnit("s"),
		metric.WithExplicitBucketBoundaries(0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5),
	)
	reqTotal, _ = meter.Int64Counter("http_requests_total",
		metric.WithDescription("Total HTTP requests by route and status"),
	)
	errTotal, _ = meter.Int64Counter("http_errors_total",
		metric.WithDescription("Total HTTP 5xx errors by route"),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)
	mux.HandleFunc("GET /ready", readyHandler)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /api/hello", withMetrics("/api/hello", helloHandler))
	mux.HandleFunc("GET /api/items", withMetrics("/api/items", itemsHandler))
	mux.HandleFunc("GET /api/error", withMetrics("/api/error", errorHandler))

	// otelhttp creates the root HTTP span; baggageMiddleware injects request_id after.
	handler := otelhttp.NewHandler(baggageMiddleware(mux), "http.server",
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + r.URL.Path
		}),
	)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	srv := &http.Server{Addr: addr, Handler: handler}

	slog.Info("server starting", "addr", addr)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
}
