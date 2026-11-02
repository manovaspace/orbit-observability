package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/manovaspace/orbit-observability/internal/domain"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/logging"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/telemetry"
)

// Exporter selects trace export backend.
type Exporter string

const (
	ExporterConsole Exporter = "console"
	ExporterOTLP    Exporter = "otlp"
)

// Config holds observability bootstrap options.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	TenantID       string
	Exporter       Exporter
	OTLPEndpoint   string
	SentryDSN      string
}

// Configure initializes logging, tracing, and optional Sentry. Safe to call once at startup.
func Configure(c Config) error {
	return telemetry.Configure(domain.Config{
		ServiceName:    c.ServiceName,
		ServiceVersion: c.ServiceVersion,
		Environment:    c.Environment,
		TenantID:       c.TenantID,
		Exporter:       domain.Exporter(c.Exporter),
		OTLPEndpoint:   c.OTLPEndpoint,
		SentryDSN:      c.SentryDSN,
	})
}

// Logger returns the configured structured logger.
func Logger() *slog.Logger {
	return telemetry.Logger()
}

// RegisterContextAttrs registers a function that contributes request-scoped
// attributes to every log record and, through AnnotateSpan, to the active span.
// Functions run in registration order and every registration is applied. A
// panic in one function is recovered and that function is skipped. Safe for
// concurrent use; registrations usually happen at startup.
func RegisterContextAttrs(fn func(ctx context.Context) []slog.Attr) {
	logging.RegisterContextAttrs(fn)
}

// AnnotateSpan copies the registered context attributes onto the active span in
// ctx. The HTTP and gRPC middleware call it at request start; call it again once
// auth has added tenant data to the context, because auth usually runs after the
// tracing middleware. It is a no-op without a recording span.
func AnnotateSpan(ctx context.Context) {
	telemetry.AnnotateSpan(ctx)
}

// Tracer returns the service tracer.
func Tracer() trace.Tracer {
	return telemetry.Tracer()
}

// SetReadiness marks readiness state for probes.
func SetReadiness(ok bool, message string) {
	state.SetReadiness(ok, message)
}

// CaptureException sends an error to Sentry when configured.
func CaptureException(err error) {
	telemetry.CaptureException(err)
}

// Shutdown flushes traces and Sentry buffers.
func Shutdown(ctx context.Context) error {
	return telemetry.Shutdown(ctx)
}
