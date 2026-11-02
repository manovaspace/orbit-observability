// Package application holds the health state machine and the redaction rules
// used by the adapters. It imports no telemetry SDK, HTTP server or adapter.
package application

import "sync"

// State is the process health and readiness state.
type State struct {
	mu       sync.RWMutex
	live     bool
	ready    bool
	readyMsg string
}

// NewState returns a live and ready state, matching the previous release.
func NewState() *State {
	return &State{live: true, ready: true}
}

// Health returns the liveness probe payload.
func (s *State) Health() map[string]string {
	s.mu.RLock()
	live := s.live
	s.mu.RUnlock()
	if !live {
		return map[string]string{"status": "not_live"}
	}
	return map[string]string{"status": "ok"}
}

// Readiness returns the readiness probe payload.
func (s *State) Readiness() map[string]string {
	s.mu.RLock()
	ok := s.ready
	msg := s.readyMsg
	s.mu.RUnlock()

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

// MarkNotLive marks the process as not live.
func (s *State) MarkNotLive() {
	s.mu.Lock()
	s.live = false
	s.mu.Unlock()
}

// SetReadiness sets the readiness state and its message.
func (s *State) SetReadiness(ok bool, message string) {
	s.mu.Lock()
	s.ready = ok
	s.readyMsg = message
	s.mu.Unlock()
}
