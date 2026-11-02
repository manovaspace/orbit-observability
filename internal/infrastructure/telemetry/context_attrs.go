package telemetry

import (
	"context"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/manovaspace/orbit-observability/internal/infrastructure/logging"
)

// AnnotateSpan copies the registered context attributes onto the active span in
// ctx. It is a no-op when ctx carries no recording span.
func AnnotateSpan(ctx context.Context) {
	if ctx == nil {
		return
	}
	span := trace.SpanFromContext(ctx)
	if !span.IsRecording() {
		return
	}

	attrs := logging.ContextAttrs(ctx)
	if len(attrs) == 0 {
		return
	}
	kvs := make([]attribute.KeyValue, 0, len(attrs))
	appendSpanAttrs(&kvs, attrs)
	span.SetAttributes(kvs...)
}

func appendSpanAttrs(dst *[]attribute.KeyValue, attrs []slog.Attr) {
	for _, a := range attrs {
		v := a.Value.Resolve()
		switch v.Kind() {
		case slog.KindGroup:
			appendSpanAttrs(dst, v.Group())
		case slog.KindString:
			*dst = append(*dst, attribute.String(a.Key, v.String()))
		case slog.KindInt64:
			*dst = append(*dst, attribute.Int64(a.Key, v.Int64()))
		case slog.KindUint64:
			*dst = append(*dst, attribute.Int64(a.Key, int64(v.Uint64())))
		case slog.KindFloat64:
			*dst = append(*dst, attribute.Float64(a.Key, v.Float64()))
		case slog.KindBool:
			*dst = append(*dst, attribute.Bool(a.Key, v.Bool()))
		case slog.KindDuration:
			*dst = append(*dst, attribute.Int64(a.Key, int64(v.Duration())))
		case slog.KindTime:
			*dst = append(*dst, attribute.String(a.Key, v.Time().UTC().Format(time.RFC3339Nano)))
		default:
			*dst = append(*dst, attribute.String(a.Key, v.String()))
		}
	}
}
