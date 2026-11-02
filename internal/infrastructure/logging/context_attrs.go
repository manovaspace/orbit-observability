package logging

import (
	"context"
	"log/slog"
	"sync"
)

var (
	contextAttrsMu  sync.RWMutex
	contextAttrsFns []func(context.Context) []slog.Attr
)

// RegisterContextAttrs adds fn to the functions consulted for every log
// record. Functions run in registration order and their attributes are appended
// to the record attributes. A panic in fn is recovered and fn is skipped.
func RegisterContextAttrs(fn func(context.Context) []slog.Attr) {
	if fn == nil {
		return
	}
	contextAttrsMu.Lock()
	contextAttrsFns = append(contextAttrsFns, fn)
	contextAttrsMu.Unlock()
}

// ContextAttrs returns the attributes produced by every registered function for
// ctx, in registration order. It is safe for concurrent use.
func ContextAttrs(ctx context.Context) []slog.Attr {
	contextAttrsMu.RLock()
	fns := make([]func(context.Context) []slog.Attr, len(contextAttrsFns))
	copy(fns, contextAttrsFns)
	contextAttrsMu.RUnlock()

	if len(fns) == 0 {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	var attrs []slog.Attr
	for _, fn := range fns {
		attrs = append(attrs, callContextAttrs(ctx, fn)...)
	}
	return attrs
}

func callContextAttrs(ctx context.Context, fn func(context.Context) []slog.Attr) (attrs []slog.Attr) {
	defer func() {
		if recover() != nil {
			attrs = nil
		}
	}()
	return fn(ctx)
}
