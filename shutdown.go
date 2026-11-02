package observability

import (
	"context"
	"net/http"
	"time"

	googlegrpc "google.golang.org/grpc"

	"github.com/manovaspace/orbit-observability/internal/infrastructure/shutdown"
)

// ShutdownConfig coordinates graceful process shutdown.
type ShutdownConfig struct {
	GRPCServer *googlegrpc.Server
	HTTPServer *http.Server
	OnShutdown []func(context.Context) error
	Timeout    time.Duration
}

// WaitForSignal blocks until SIGINT/SIGTERM, then drains servers and flushes telemetry.
func WaitForSignal(cfg ShutdownConfig) {
	shutdown.WaitForSignal(state, shutdown.Config{
		GRPCServer: cfg.GRPCServer,
		HTTPServer: cfg.HTTPServer,
		OnShutdown: cfg.OnShutdown,
		Timeout:    cfg.Timeout,
	})
}
