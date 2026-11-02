// Package config reads the Orbit environment contract.
package config

import (
	"os"

	"github.com/manovaspace/orbit-observability/internal/domain"
)

// FromEnv builds a domain.Config from standard Orbit environment variables.
func FromEnv(serviceName, serviceVersion string) domain.Config {
	c := domain.Config{
		ServiceName:    serviceName,
		ServiceVersion: serviceVersion,
		Environment:    envOr("DEPLOYMENT_ENVIRONMENT", "dev"),
		Exporter:       domain.ExporterConsole,
		OTLPEndpoint:   envOr("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:10517"),
		SentryDSN:      os.Getenv("SENTRY_DSN"),
		TenantID:       os.Getenv("TENANT_ID"),
	}
	if os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		c.Exporter = domain.ExporterOTLP
	}
	return c
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
