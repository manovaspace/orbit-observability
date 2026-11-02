// Package domain holds the observability configuration types and constants. It
// imports no telemetry SDK, HTTP server or adapter.
package domain

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
