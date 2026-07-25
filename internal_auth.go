package observability

import (
	"context"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const internalTokenMetadataKey = "x-orbit-internal-token"

// InternalAuthUnaryServerInterceptor rejects RPCs without a matching ORBIT_INTERNAL_TOKEN.
// When ORBIT_INTERNAL_TOKEN is empty and DEPLOYMENT_ENVIRONMENT=dev, auth is skipped.
func InternalAuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		expected := os.Getenv("ORBIT_INTERNAL_TOKEN")
		if expected == "" {
			if os.Getenv("DEPLOYMENT_ENVIRONMENT") == "dev" {
				return handler(ctx, req)
			}
			return nil, status.Error(codes.Unavailable, "ORBIT_INTERNAL_TOKEN not configured")
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing internal token")
		}
		vals := md.Get(internalTokenMetadataKey)
		if len(vals) == 0 || vals[0] != expected {
			return nil, status.Error(codes.Unauthenticated, "invalid internal token")
		}
		return handler(ctx, req)
	}
}

// InternalAuthDialOption attaches ORBIT_INTERNAL_TOKEN to all outgoing RPCs.
func InternalAuthDialOption() grpc.DialOption {
	return grpc.WithUnaryInterceptor(func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if token := os.Getenv("ORBIT_INTERNAL_TOKEN"); token != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, internalTokenMetadataKey, token)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	})
}
