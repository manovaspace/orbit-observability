package observability

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestHealthHandlerMarksNotLive(t *testing.T) {
	atomic.StoreInt32(&live, 1)
	t.Cleanup(func() { atomic.StoreInt32(&live, 1) })

	rec := httptest.NewRecorder()
	HealthHandler().ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("live code=%d", rec.Code)
	}

	MarkNotLive()
	rec2 := httptest.NewRecorder()
	HealthHandler().ServeHTTP(rec2, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec2.Code != http.StatusServiceUnavailable {
		t.Fatalf("not live code=%d", rec2.Code)
	}
}
