package observability

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

var live int32 = 1

// Health returns liveness probe payload.
func Health() map[string]string {
	if atomic.LoadInt32(&live) == 0 {
		return map[string]string{"status": "not_live"}
	}
	return map[string]string{"status": "ok"}
}

// Readiness returns readiness probe payload.
func Readiness() map[string]string {
	mu.RLock()
	ok := ready
	msg := readyMsg
	mu.RUnlock()

	status := "ok"
	if !ok {
		status = "not_ready"
	}
	out := map[string]string{"status": status}
	if msg != "" {
		out["message"] = msg
	}
	return out
}

// HealthHandler serves GET /healthz style liveness.
func HealthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := Health()
		w.Header().Set("Content-Type", "application/json")
		if payload["status"] != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
}

// ReadinessHandler serves GET /readyz style readiness.
func ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		payload := Readiness()
		w.Header().Set("Content-Type", "application/json")
		if payload["status"] != "ok" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		_ = json.NewEncoder(w).Encode(payload)
	})
}

// MarkNotLive marks the process as not live (for graceful shutdown hooks).
func MarkNotLive() {
	atomic.StoreInt32(&live, 0)
}
