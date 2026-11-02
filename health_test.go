package observability

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerIsLive(t *testing.T) {
	rec := httptest.NewRecorder()
	HealthHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("live code=%d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q", got)
	}
	var payload map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "ok" {
		t.Fatalf("health: %v", payload)
	}
}

func TestReadinessHandlerReflectsState(t *testing.T) {
	t.Cleanup(func() { SetReadiness(true, "") })

	rec := httptest.NewRecorder()
	ReadinessHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("ready code=%d", rec.Code)
	}

	SetReadiness(false, "database unavailable")
	rec2 := httptest.NewRecorder()
	ReadinessHandler().ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/readyz", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("not ready code=%d", rec2.Code)
	}
	if got := Readiness()["message"]; got != "database unavailable" {
		t.Fatalf("message = %q", got)
	}
}
