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

## Development

```bash
go test ./...
```

## Contributing

See [CONTRIBUTING.md](./CONTRIBUTING.md). Security reports: [SECURITY.md](./SECURITY.md).

## License

MIT — see [LICENSE](./LICENSE).
