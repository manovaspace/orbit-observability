// Package shutdown drains servers and flushes telemetry on SIGINT/SIGTERM.
package shutdown

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	googlegrpc "google.golang.org/grpc"

	"github.com/manovaspace/orbit-observability/internal/application"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/telemetry"
)

const defaultShutdownTimeout = 10 * time.Second

// Config coordinates graceful process shutdown.
type Config struct {
	GRPCServer *googlegrpc.Server
	HTTPServer *http.Server
	OnShutdown []func(context.Context) error
	Timeout    time.Duration
}

// WaitForSignal blocks until SIGINT/SIGTERM, then drains servers and flushes
// telemetry.
func WaitForSignal(state *application.State, cfg Config) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	state.SetReadiness(false, "shutting down")
	state.MarkNotLive()
	telemetry.Logger().Info("shutting down")

	if cfg.HTTPServer != nil {
		if err := cfg.HTTPServer.Shutdown(ctx); err != nil {
			telemetry.Logger().Error("http shutdown", "error", err)
		}
	}

	if cfg.GRPCServer != nil {
		stopped := make(chan struct{})
		go func() {
			cfg.GRPCServer.GracefulStop()
			close(stopped)
		}()
		select {
		case <-stopped:
		case <-ctx.Done():
			cfg.GRPCServer.Stop()
		}
	}

	for _, fn := range cfg.OnShutdown {
		if fn == nil {
			continue
		}
		if err := fn(ctx); err != nil {
			telemetry.Logger().Error("shutdown hook", "error", err)
		}
	}

	if err := telemetry.Shutdown(ctx); err != nil {
		telemetry.Logger().Error("observability shutdown", "error", err)
	}
}
