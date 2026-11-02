package observability_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	observability "github.com/manovaspace/orbit-observability"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/logging"
)

type orgIDKey struct{}

func TestContextAttrsInLogsAndSpans(t *testing.T) {
	observability.RegisterContextAttrs(func(ctx context.Context) []slog.Attr {
		org, ok := ctx.Value(orgIDKey{}).(string)
		if !ok {
			return nil
		}
		return []slog.Attr{
			slog.String("org_id", org),
			slog.String("subject", "user-7"),
		}
	})

	ctx := context.WithValue(context.Background(), orgIDKey{}, "org-9")

	var buf bytes.Buffer
	log := slog.New(logging.New(logging.Options{
		ServiceName:    "test",
		ServiceVersion: "0.1.0",
		Environment:    "dev",
		Writer:         &buf,
	}))
	log.InfoContext(ctx, "request")

	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	attrs, ok := entry["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("attributes: %#v", entry["attributes"])
	}
	if attrs["org_id"] != "org-9" {
		t.Fatalf("log org_id = %#v", attrs["org_id"])
	}
	if attrs["subject"] != "user-7" {
		t.Fatalf("log subject = %#v", attrs["subject"])
	}

	exporter := tracetest.NewInMemoryExporter()
	provider := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	t.Cleanup(func() {
		_ = provider.Shutdown(context.Background())
	})

	spanCtx, span := provider.Tracer("test").Start(ctx, "request")
	observability.AnnotateSpan(spanCtx)
	span.End()

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("spans = %d, want 1", len(spans))
	}
	got := make(map[string]attribute.Value, len(spans[0].Attributes))
	for _, attr := range spans[0].Attributes {
		got[string(attr.Key)] = attr.Value
	}
	if got["org_id"].AsString() != "org-9" {
		t.Fatalf("span org_id = %#v", got["org_id"])
	}
	if got["subject"].AsString() != "user-7" {
		t.Fatalf("span subject = %#v", got["subject"])
	}
}
