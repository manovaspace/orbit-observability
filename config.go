package observability

import "os"

// ConfigFromEnv builds Config from standard Orbit environment variables.
func ConfigFromEnv(serviceName, serviceVersion string) Config {
	c := Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    envOr("DEPLOYMENT_ENVIRONMENT", "dev"),
		Exporter:       ExporterConsole,
		OTLPEndpoint:   envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:10517"),
		SentryDSN:      os.Getenv("SENTRY_DSN"),
		TenantID:       os.Getenv("TENANT_ID"),
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		c.Exporter = ExporterOTLP
	}
	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
