package observability

import (
	"google.golang.org/grpc"

	internalgrpc "github.com/manovaspace/orbit-observability/internal/infrastructure/grpc"
)

// InternalAuthUnaryServerInterceptor rejects RPCs without a matching ORBIT_INTERNAL_TOKEN.
// When ORBIT_INTERNAL_TOKEN is empty and DEPLOYMENT_ENVIRONMENT=dev, auth is skipped.
func InternalAuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return internalgrpc.InternalAuthUnaryServerInterceptor()
}

// InternalAuthDialOption attaches ORBIT_INTERNAL_TOKEN to all outgoing RPCs.
func InternalAuthDialOption() grpc.DialOption {
	return internalgrpc.InternalAuthDialOption()
}
