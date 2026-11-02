package application_test

import (
	"testing"

	"github.com/manovaspace/orbit-observability/internal/application"
)

func TestStateStartsLiveAndReady(t *testing.T) {
	state := application.NewState()
	if got := state.Health()["status"]; got != "ok" {
		t.Fatalf("health = %q, want ok", got)
	}
	if got := state.Readiness()["status"]; got != "ok" {
		t.Fatalf("readiness = %q, want ok", got)
	}
}

func TestMarkNotLiveIsPermanent(t *testing.T) {
	state := application.NewState()
	state.MarkNotLive()
	if got := state.Health()["status"]; got != "not_live" {
		t.Fatalf("health = %q, want not_live", got)
	}
}

func TestSetReadinessCarriesMessage(t *testing.T) {
	state := application.NewState()
	state.SetReadiness(false, "shutting down")
	payload := state.Readiness()
	if payload["status"] != "not_ready" || payload["message"] != "shutting down" {
		t.Fatalf("readiness = %v", payload)
	}
	state.SetReadiness(true, "")
	payload = state.Readiness()
	if payload["status"] != "ok" {
		t.Fatalf("readiness = %v", payload)
	}
	if _, ok := payload["message"]; ok {
		t.Fatalf("empty message must be omitted: %v", payload)
	}
}

func TestRedactStringMasksKnownKeys(t *testing.T) {
	got := application.RedactString("otp=123456 token: abc password=secret", []string{"otp", "token", "password"})
	want := "otp=[REDACTED] token: [REDACTED] password=[REDACTED]"
	if got != want {
		t.Fatalf("RedactString = %q, want %q", got, want)
	}
}

func TestIndexFold(t *testing.T) {
	if got := application.IndexFold("OTP=123", "otp="); got != 0 {
		t.Fatalf("IndexFold = %d, want 0", got)
	}
	if got := application.IndexFold("abc", "zzz"); got != -1 {
		t.Fatalf("IndexFold = %d, want -1", got)
	}
}
