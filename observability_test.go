package observability_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	observability "github.com/manovaspace/orbit-observability"
	"github.com/manovaspace/orbit-observability/internal/loghandler"
)

func TestADR009LogEnvelope(t *testing.T) {
	var buf bytes.Buffer
	h := loghandler.New(loghandler.Options{
		ServiceName:    "test",
		ServiceVersion: "0.1.0",
		Environment:    "dev",
		Writer:         &buf,
	})
	log := slog.New(h)
	log.InfoContext(context.Background(), "otp_requested", "channel", "email", "delivery.id", "abc")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if entry["schema_version"] != "1" {
		t.Fatalf("schema_version: %v", entry["schema_version"])
	}
	if entry["body"] != "otp_requested" {
		t.Fatalf("body: %v", entry["body"])
	}
	body := buf.String()
	if strings.Contains(body, "123456") {
		t.Fatal("must not log raw otp in test")
	}
}

func TestConfigureAndHealth(t *testing.T) {
	if err := observability.Configure(observability.Config{
		ServiceName: "test-svc",
		Exporter:    observability.ExporterConsole,
	}); err != nil {
		t.Fatal(err)
	}
	h := observability.Health()
	if h["status"] != "ok" {
		t.Fatalf("health: %v", h)
	}
}

func TestConfigFromEnvOTLP(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "collector:4317")
	t.Setenv("DEPLOYMENT_ENVIRONMENT", "staging")
	t.Setenv("TENANT_ID", "tenant-1")

	c := observability.ConfigFromEnv("svc", "1.0.0")
	if c.Exporter != observability.ExporterOTLP {
		t.Fatalf("exporter: %v", c.Exporter)
	}
	if c.OTLPEndpoint != "collector:4317" {
		t.Fatalf("endpoint: %v", c.OTLPEndpoint)
	}
	if c.Environment != "staging" {
		t.Fatalf("environment: %v", c.Environment)
	}
	if c.TenantID != "tenant-1" {
		t.Fatalf("tenant: %v", c.TenantID)
	}
}

func TestTenantIDInLogResource(t *testing.T) {
	var buf bytes.Buffer
	h := loghandler.New(loghandler.Options{
		ServiceName:    "test",
		ServiceVersion: "0.1.0",
		Environment:    "dev",
		TenantID:       "client-uuid",
		Writer:         &buf,
	})
	log := slog.New(h)
	log.Info("event")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	res, ok := entry["resource"].(map[string]any)
	if !ok {
		t.Fatal("missing resource")
	}
	if res["tenant.id"] != "client-uuid" {
		t.Fatalf("tenant.id: %v", res["tenant.id"])
	}
}
