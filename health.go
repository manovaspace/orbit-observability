package observability

import (
	"net/http"

	"github.com/manovaspace/orbit-observability/internal/application"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/httpapi"
)

var state = application.NewState()

// Health returns liveness probe payload.
func Health() map[string]string {
	return state.Health()
}

// Readiness returns readiness probe payload.
func Readiness() map[string]string {
	return state.Readiness()
}

// HealthHandler serves GET /healthz style liveness.
func HealthHandler() http.Handler {
	return httpapi.HealthHandler(state)
}

// ReadinessHandler serves GET /readyz style readiness.
func ReadinessHandler() http.Handler {
	return httpapi.ReadinessHandler(state)
}

// MarkNotLive marks the process as not live (for graceful shutdown hooks).
func MarkNotLive() {
	state.MarkNotLive()
}
