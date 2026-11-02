# orbit-observability

[![CI](https://github.com/manovaspace/orbit-observability/actions/workflows/ci.yml/badge.svg)](https://github.com/manovaspace/orbit-observability/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)

Go library: structured logging (`slog`), OpenTelemetry traces, health helpers, HTTP/gRPC middleware, and optional Sentry for Orbit-style services.

Part of the [Manova / Orbit](https://github.com/manovaspace) open toolkit.

## Install

```bash
go get github.com/manovaspace/orbit-observability@latest
```

Local development (monorepo sibling):

```go
replace github.com/manovaspace/orbit-observability => ../orbit-observability
```

## Quick start

```go
import observability "github.com/manovaspace/orbit-observability"

func main() {
    if err := observability.Configure(observability.ConfigFromEnv("my-service", "0.1.0")); err != nil {
        panic(err)
    }

    log := observability.Logger()
    log.Info("started", "port", 8080)

    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        _ = json.NewEncoder(w).Encode(observability.Health())
    })
    http.ListenAndServe(":8080", observability.HTTPMiddleware(mux))
}
```

Set `OTEL_EXPORTER_OTLP_ENDPOINT` (for example `localhost:10517`) or pass `Exporter: observability.ExporterOTLP` for OTLP trace export.

Graceful shutdown:

```go
observability.WaitForSignal(observability.ShutdownConfig{
    GRPCServer: gs,
    HTTPServer: healthServer,
})
```

## Request-scoped attributes

Services attach request-scoped values such as `org_id` and `subject` to every log
line and span without teaching this library about tenants. Registrations are
applied in order, a panic in one is contained, and all calls are safe for
concurrent use.

```go
type orgKey struct{}

func orgAttrs(ctx context.Context) []slog.Attr {
    org, ok := ctx.Value(orgKey{}).(string)
    if !ok {
        return nil
    }
    return []slog.Attr{slog.String("org_id", org)}
}

func main() {
    observability.RegisterContextAttrs(orgAttrs)

    handler := observability.HTTPMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := context.WithValue(r.Context(), orgKey{}, orgFromRequest(r))
        observability.AnnotateSpan(ctx) // auth has run: add org_id to the span now
        observability.Logger().InfoContext(ctx, "request handled")
    }))
    _ = handler
}
```

The HTTP and gRPC middleware apply the registered attributes to the span at
request start. Because auth usually runs after the tracing middleware, call
`observability.AnnotateSpan(ctx)` once the tenant values exist to re-apply them.

## Layout

```
config.go, health.go, internal_auth.go,   public facade (unchanged API)
middleware.go, observability.go, shutdown.go
internal/domain/                          Config, Exporter, console/OTLP constants
internal/application/                     health state machine and redaction rule
internal/infrastructure/config/           environment contract
internal/infrastructure/logging/          ADR-009 JSON slog handler
internal/infrastructure/telemetry/        OpenTelemetry and Sentry adapters
internal/infrastructure/httpapi/          health handlers and otelhttp middleware
internal/infrastructure/grpc/             internal-token interceptors and gRPC options
internal/infrastructure/shutdown/         signal drain and telemetry flush
```

The domain imports no infrastructure package. The application layer holds pure
state and rules, and the exported facade only wires the adapters, so every
previous name and signature is unchanged.

The library runs no database, no HTTP server and no event bus; tests need no
service.

## Development

```bash
go test ./...
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Security reports: [SECURITY.md](./SECURITY.md).

## License

MIT — see [LICENSE](./LICENSE).
