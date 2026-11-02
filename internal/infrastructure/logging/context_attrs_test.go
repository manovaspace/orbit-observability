package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"sync"
	"testing"
)

func resetContextAttrs() {
	contextAttrsMu.Lock()
	contextAttrsFns = nil
	contextAttrsMu.Unlock()
}

func newTestLogger(w io.Writer) *slog.Logger {
	return slog.New(New(Options{
		ServiceName:    "test",
		ServiceVersion: "0.1.0",
		Environment:    "dev",
		Writer:         w,
	}))
}

func decodeLogEntry(t *testing.T, buf *bytes.Buffer) map[string]any {
	t.Helper()
	var entry map[string]any
	if err := json.Unmarshal(buf.Bytes(), &entry); err != nil {
		t.Fatalf("invalid json %q: %v", buf.String(), err)
	}
	return entry
}

func logAttributes(t *testing.T, entry map[string]any) map[string]any {
	t.Helper()
	attrs, ok := entry["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("attributes: %#v", entry["attributes"])
	}
	return attrs
}

func TestNoContextAttrsLeavesOutputUnchanged(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	var buf bytes.Buffer
	newTestLogger(&buf).InfoContext(context.Background(), "event", "channel", "email")

	attrs := logAttributes(t, decodeLogEntry(t, &buf))
	if len(attrs) != 1 || attrs["channel"] != "email" {
		t.Fatalf("attributes = %#v, want only channel", attrs)
	}
}

func TestRegisteredContextAttrsAppearInLogOutput(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	RegisterContextAttrs(func(context.Context) []slog.Attr {
		return []slog.Attr{
			slog.String("org_id", "org-42"),
			slog.String("subject", "user-7"),
		}
	})

	var buf bytes.Buffer
	newTestLogger(&buf).InfoContext(context.Background(), "event")

	attrs := logAttributes(t, decodeLogEntry(t, &buf))
	if attrs["org_id"] != "org-42" {
		t.Fatalf("org_id = %#v", attrs["org_id"])
	}
	if attrs["subject"] != "user-7" {
		t.Fatalf("subject = %#v", attrs["subject"])
	}
}

func TestMultipleContextAttrsRegistrationsApplyInOrder(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	RegisterContextAttrs(func(context.Context) []slog.Attr {
		return []slog.Attr{slog.String("first", "1")}
	})
	RegisterContextAttrs(func(context.Context) []slog.Attr {
		return []slog.Attr{slog.String("second", "2")}
	})

	attrs := ContextAttrs(context.Background())
	if len(attrs) != 2 || attrs[0].Key != "first" || attrs[1].Key != "second" {
		t.Fatalf("ordered attrs = %#v", attrs)
	}

	var buf bytes.Buffer
	newTestLogger(&buf).InfoContext(context.Background(), "event")

	logged := logAttributes(t, decodeLogEntry(t, &buf))
	if logged["first"] != "1" || logged["second"] != "2" {
		t.Fatalf("attributes = %#v", logged)
	}
}

func TestContextAttrsReadFromRecordContext(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	type orgKey struct{}
	RegisterContextAttrs(func(ctx context.Context) []slog.Attr {
		org, ok := ctx.Value(orgKey{}).(string)
		if !ok {
			return nil
		}
		return []slog.Attr{slog.String("org_id", org)}
	})

	var buf bytes.Buffer
	ctx := context.WithValue(context.Background(), orgKey{}, "org-from-request")
	newTestLogger(&buf).InfoContext(ctx, "event")

	attrs := logAttributes(t, decodeLogEntry(t, &buf))
	if attrs["org_id"] != "org-from-request" {
		t.Fatalf("org_id = %#v", attrs["org_id"])
	}
}

func TestPanicInContextAttrsIsContained(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	RegisterContextAttrs(func(context.Context) []slog.Attr {
		panic("boom")
	})
	RegisterContextAttrs(func(context.Context) []slog.Attr {
		return []slog.Attr{slog.String("ok", "yes")}
	})

	var buf bytes.Buffer
	newTestLogger(&buf).InfoContext(context.Background(), "event")

	attrs := logAttributes(t, decodeLogEntry(t, &buf))
	if attrs["ok"] != "yes" {
		t.Fatalf("attributes = %#v, want ok=yes", attrs)
	}
}

func TestConcurrentLoggingAndRegistration(t *testing.T) {
	resetContextAttrs()
	t.Cleanup(resetContextAttrs)

	logger := newTestLogger(io.Discard)

	var wg sync.WaitGroup
	for i := range 50 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			RegisterContextAttrs(func(context.Context) []slog.Attr {
				return []slog.Attr{slog.Int("worker", i)}
			})
		}()
		go func() {
			defer wg.Done()
			for j := range 50 {
				logger.InfoContext(context.Background(), "event", "n", j)
			}
		}()
	}
	wg.Wait()
}
