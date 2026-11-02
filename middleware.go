package observability

import (
	"net/http"

	"google.golang.org/grpc"

	internalgrpc "github.com/manovaspace/orbit-observability/internal/infrastructure/grpc"
	"github.com/manovaspace/orbit-observability/internal/infrastructure/httpapi"
)

// HTTPMiddleware wraps an http.Handler with W3C trace propagation.
func HTTPMiddleware(next http.Handler) http.Handler {
	return httpapi.Middleware(next)
}

// GRPCServerOptions returns gRPC server options with tracing and internal token auth.
func GRPCServerOptions() []grpc.ServerOption {
	return internalgrpc.ServerOptions()
}

// GRPCDialOptions returns gRPC client dial options with tracing.
func GRPCDialOptions() []grpc.DialOption {
	return internalgrpc.DialOptions()
}
