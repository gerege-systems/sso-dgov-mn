# Observability

Three pillars: OpenTelemetry tracing, Prometheus metrics, and structured Zap
logs. In production, the metrics/spec endpoints are bearer-gated.

## Tracing (OpenTelemetry)

`pkg/observability/tracing.go` builds an OTel tracer provider. `TracingConfig.
Exporter` selects `stdout` (dev), `otlp` (prod → OTLP/gRPC to
`OTEL_EXPORTER_OTLP_ENDPOINT`), or `""` (disabled), with a head sampler
`SampleRatio` (prod typically 0.01–0.1).

`TracingMiddleware` is the **first** global middleware so `RequestIDMiddleware`
can bind `trace_id` into the log context. The pgx pool
(`driver_pgx.go`) and the Redis cache are also instrumented. A `Shutdown` hook
flushes buffered spans on SIGTERM.

## Metrics (Prometheus)

`promhttp.Handler()` is mounted at `/metrics`. `MetricsMiddleware` records
`httpRequestsTotal` (CounterVec) and `httpRequestDuration` (HistogramVec). DB
pool stats are registered from the live pgxpool via
`observability.RegisterDBStatsProvider`.

## Logs (Zap)

Structured logging via `pkg/logger/zap.go`; `AccessLogMiddleware` sits in the
global chain and emits per-request access logs carrying the `request_id` /
`trace_id`.

## Production gating

Both `/metrics` and `/swagger/doc.json` sit behind
`ObservabilityGate(isProduction, OBSERVABILITY_TOKEN)`:

- non-production → always open;
- production, empty token → fully closed (**404**);
- production, token set → requires `Authorization: Bearer <token>`, compared with
  `crypto/subtle.ConstantTimeCompare`; **any mismatch returns 404** (not 401) to
  hide the endpoint's existence from reconnaissance.

`/health` and `/ready` stay open for load balancers.
