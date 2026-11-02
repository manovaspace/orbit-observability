// Package httpapi serves the health probes and wraps an http.Handler with W3C
// trace propagation.
package httpapi

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/manovaspace/orbit-observability/internal/application"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/telemetry"
)

// HealthHandler serves GET /healthz style liveness.
func HealthHandler(state *application.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := state.Health()
		w.Header().Set("Content-Type", "application/json")
		if payload["status"] != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
}

// ReadinessHandler serves GET /readyz style readiness.
func ReadinessHandler(state *application.State) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := state.Readiness()
		w.Header().Set("Content-Type", "application/json")
		if payload["status"] != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
}

// Middleware wraps an http.Handler with W3C trace propagation and copies the
// registered context attributes onto the request span.
func Middleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		telemetry.AnnotateSpan(r.Context())
		next.ServeHTTP(w, r)
	}), "http")
}
