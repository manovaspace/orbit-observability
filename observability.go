package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/trace"

	"github.com/manovaspace/orbit-observability/internal/domain"
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
