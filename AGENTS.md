---
read_with:
  - handbook/docs/orbit/architecture/observability.md
  - handbook/docs/orbit/decisions/009-structured-logging.md
  - handbook/docs/orbit/decisions/010-observability-stack.md
---

# orbit-observability

Structured logging, OpenTelemetry, health checks, and Sentry hooks for Go services. Status: `beta`. MIT — `github.com/manovaspace/orbit-observability`.

## Commands

```bash
go test ./...
```

## Architecture

- `internal/domain` — `Config`, `Exporter` and the console/OTLP constants. It imports no telemetry SDK, gRPC, HTTP server or adapter.
- `internal/application` — the health/readiness state machine and the Sentry redaction rule.
- `internal/infrastructure/config` — the `DEPLOYMENT_ENVIRONMENT`, `OTEL_EXPORTER_OTLP_ENDPOINT`, `SENTRY_DSN` and `TENANT_ID` contract.
- `internal/infrastructure/logging` — the ADR-009 JSON `slog` handler.
- `internal/infrastructure/telemetry` — the only package that talks to OpenTelemetry and Sentry.
- `internal/infrastructure/httpapi` — `/healthz` and `/readyz` handlers plus `otelhttp` middleware.
- `internal/infrastructure/grpc` — `ORBIT_INTERNAL_TOKEN` interceptors and gRPC options.
- `internal/infrastructure/shutdown` — SIGINT/SIGTERM drain and telemetry flush.
- Root files — the public facade. Every exported function and type keeps its previous name and signature.
- `.golangci.yml` `depguard` rules enforce both boundaries. Do not weaken them to land a change.

The library has no database, no HTTP server of its own and publishes no events. Its `net/http` handlers and middleware are consumed by services that own the server; they stay `net/http` because changing them would change the exported contract (Fiber v3 does not apply).

## Docs

- Public README in this repo
- Staff handbook (private): ADR-009 structured logging, observability contracts
