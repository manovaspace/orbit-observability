package loghandler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"

	"go.opentelemetry.io/otel/trace"
)

const schemaVersion = "1"

// Options configures the ADR-009 JSON log handler.
type Options struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	TenantID       string
	Writer         io.Writer
}

// Handler emits OpenTelemetry-aligned JSON log lines.
type Handler struct {
	opts   Options
	attrs  []slog.Attr
	groups []string
	mu     sync.Mutex
}

// New creates an ADR-009 structured log handler.
func New(opts Options) *Handler {
	if opts.Writer == nil {
		opts.Writer = os.Stdout
	}
	return &Handler{opts: opts}
}

func (h *Handler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	entry := map[string]any{
		"schema_version":  schemaVersion,
		"timestamp":       r.Time.UTC().Format(time.RFC3339Nano),
		"severity_text":   r.Level.String(),
		"severity_number": severityNumber(r.Level),
		"body":            r.Message,
		"resource": h.resource(),
		"attributes": map[string]any{},
	}

	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		entry["trace_id"] = sc.TraceID().String()
		entry["span_id"] = sc.SpanID().String()
	}

	attrs := make(map[string]any)
	r.Attrs(func(a slog.Attr) bool {
		flattenAttr(a, attrs)
		return true
	})
	for _, a := range h.attrs {
		flattenAttr(a, attrs)
	}
	if cid, ok := attrs["correlation_id"]; ok {
		entry["correlation_id"] = cid
		delete(attrs, "correlation_id")
	}
	entry["attributes"] = attrs

	h.mu.Lock()
	defer h.mu.Unlock()
	enc := json.NewEncoder(h.opts.Writer)
	return enc.Encode(entry)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		opts:   h.opts,
		attrs:  append(append([]slog.Attr{}, h.attrs...), attrs...),
		groups: h.groups,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		opts:   h.opts,
		attrs:  h.attrs,
		groups: append(append([]string{}, h.groups...), name),
	}
}

func (h *Handler) resource() map[string]string {
	res := map[string]string{
		"service.name":           h.opts.ServiceName,
		"service.version":        h.opts.ServiceVersion,
		"deployment.environment": h.opts.Environment,
	}
	if h.opts.TenantID != "" {
		res["tenant.id"] = h.opts.TenantID
	}
	return res
}

func severityNumber(level slog.Level) int {
	switch {
	case level < slog.LevelInfo:
		return 5
	case level < slog.LevelWarn:
		return 9
	case level < slog.LevelError:
		return 13
	default:
		return 17
	}
}

func flattenAttr(a slog.Attr, out map[string]any) {
	v := a.Value.Any()
	switch a.Value.Kind() {
	case slog.KindGroup:
		for _, ga := range a.Value.Group() {
			flattenAttr(ga, out)
		}
	default:
		out[a.Key] = v
	}
}
