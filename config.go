package observability

import "github.com/manovaspace/orbit-observability/internal/infrastructure/config"

// ConfigFromEnv builds Config from standard Orbit environment variables.
func ConfigFromEnv(serviceName, serviceVersion string) Config {
	c := config.FromEnv(serviceName, serviceVersion)
	return Config{
		ServiceName:    c.ServiceName,
		ServiceVersion: c.ServiceVersion,
		Environment:    c.Environment,
		TenantID:       c.TenantID,
		Exporter:       Exporter(c.Exporter),
		OTLPEndpoint:   c.OTLPEndpoint,
		SentryDSN:      c.SentryDSN,
	}
}
