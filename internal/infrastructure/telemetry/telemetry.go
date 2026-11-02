// Package telemetry configures structured logging, OpenTelemetry traces and
// optional Sentry. It is the only place that talks to those SDKs.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/getsentry/sentry-go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.41.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/manovaspace/orbit-observability/internal/application"
	"github.com/manovaspace/orbit-observability/internal/domain"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/logging"
)

var (
	mu     sync.RWMutex
	logger *slog.Logger
	tracer trace.Tracer
	tp     *sdktrace.TracerProvider
)

// Configure initializes logging, tracing, and optional Sentry.
func Configure(c domain.Config) error {
	if c.ServiceName == "" {
		return fmt.Errorf("observability: ServiceName is required")
	}
	if c.Exporter == "" {
		c.Exporter = domain.ExporterConsole
	}
	if c.OTLPEndpoint == "" {
		c.OTLPEndpoint = "localhost:10517"
	}
	if c.Environment == "" {
		c.Environment = "dev"
	}
	if c.ServiceVersion == "" {
		c.ServiceVersion = "0.0.0"
	}

	resAttrs := []attribute.KeyValue{
		semconv.ServiceName(c.ServiceName),
		semconv.ServiceVersion(c.ServiceVersion),
		attribute.String("deployment.environment", c.Environment),
	}
	if c.TenantID != "" {
		resAttrs = append(resAttrs, attribute.String("tenant.id", c.TenantID))
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(resAttrs...),
	)
	if err != nil {
		return fmt.Errorf("observability: resource: %w", err)
	}

	traceProvider, err := newTracerProvider(c, res)
	if err != nil {
		return err
	}
	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	mu.Lock()
	tp = traceProvider
	mu.Unlock()

	if c.SentryDSN != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              c.SentryDSN,
			Environment:      c.Environment,
			Release:          c.ServiceName + "@" + c.ServiceVersion,
			AttachStacktrace: true,
			BeforeSend:       scrubSentryEvent,
		}); err != nil {
			return fmt.Errorf("observability: sentry: %w", err)
		}
	}

	h := logging.New(logging.Options{
		ServiceName:    c.ServiceName,
		ServiceVersion: c.ServiceVersion,
		Environment:    c.Environment,
		TenantID:       c.TenantID,
	})
	logger = slog.New(h)
	slog.SetDefault(logger)
	tracer = otel.Tracer(c.ServiceName)

	return nil
}

func newTracerProvider(c domain.Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	var exp sdktrace.SpanExporter
	var err error

	switch c.Exporter {
	case domain.ExporterConsole:
		exp, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
	case domain.ExporterOTLP:
		conn, dialErr := grpc.NewClient(c.OTLPEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if dialErr != nil {
			return nil, fmt.Errorf("observability: otlp dial: %w", dialErr)
		}
		exp, err = otlptracegrpc.New(context.Background(), otlptracegrpc.WithGRPCConn(conn))
	default:
		return nil, fmt.Errorf("observability: unknown exporter %q", c.Exporter)
	}
	if err != nil {
		return nil, fmt.Errorf("observability: trace exporter: %w", err)
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
	), nil
}

// Logger returns the configured structured logger.
func Logger() *slog.Logger {
	mu.RLock()
	l := logger
	mu.RUnlock()
	if l == nil {
		return slog.Default()
	}
	return l
}

// Tracer returns the service tracer.
func Tracer() trace.Tracer {
	mu.RLock()
	t := tracer
	mu.RUnlock()
	if t == nil {
		return otel.Tracer("unknown")
	}
	return t
}

// CaptureException sends an error to Sentry when configured.
func CaptureException(err error) {
	if err != nil {
		sentry.CaptureException(err)
	}
}

// Shutdown flushes traces and Sentry buffers.
func Shutdown(ctx context.Context) error {
	mu.RLock()
	provider := tp
	mu.RUnlock()
	if provider != nil {
		if err := provider.Shutdown(ctx); err != nil {
			return fmt.Errorf("observability: tracer shutdown: %w", err)
		}
	}
	sentry.Flush(2 * time.Second)
	return nil
}

func scrubSentryEvent(event *sentry.Event, _ *sentry.EventHint) *sentry.Event {
	sensitive := []string{"otp", "password", "token", "authorization", "secret", "code"}
	for i := range event.Exception {
		event.Exception[i].Value = application.RedactString(event.Exception[i].Value, sensitive)
	}
	return event
}
