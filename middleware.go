package observability

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"google.golang.org/grpc"
)

// HTTPMiddleware wraps an http.Handler with W3C trace propagation.
func HTTPMiddleware(next http.Handler) http.Handler {
	return otelhttp.NewHandler(next, "http")
}

// GRPCServerOptions returns gRPC server options with tracing and internal token auth.
func GRPCServerOptions() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(InternalAuthUnaryServerInterceptor()),
	}
}

// GRPCDialOptions returns gRPC client dial options with tracing.
func GRPCDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}
}
