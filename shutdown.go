package observability

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	googlegrpc "google.golang.org/grpc"
)

const defaultShutdownTimeout = 10 * time.Second

// ShutdownConfig coordinates graceful process shutdown.
type ShutdownConfig struct {
	GRPCServer *googlegrpc.Server
	HTTPServer *http.Server
	OnShutdown []func(context.Context) error
	Timeout    time.Duration
}

// WaitForSignal blocks until SIGINT/SIGTERM, then drains servers and flushes telemetry.
func WaitForSignal(cfg ShutdownConfig) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = defaultShutdownTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	SetReadiness(false, "shutting down")
	MarkNotLive()
	Logger().Info("shutting down")

	if cfg.HTTPServer != nil {
		if err := cfg.HTTPServer.Shutdown(ctx); err != nil {
			Logger().Error("http shutdown", "error", err)
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
			Logger().Error("shutdown hook", "error", err)
		}
	}

	if err := Shutdown(ctx); err != nil {
		Logger().Error("observability shutdown", "error", err)
	}
}
