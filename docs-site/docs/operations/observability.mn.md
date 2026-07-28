# Ажиглалт (Observability)

Гурван багана: OpenTelemetry tracing, Prometheus metrics, болон бүтэцлэгдсэн Zap
log-ууд. Production-д metrics/spec эндпойнтууд нь bearer-хамгаалалттай.

## Tracing (OpenTelemetry)

`pkg/observability/tracing.go` нь OTel tracer provider үүсгэдэг. `TracingConfig.
Exporter` нь `stdout` (dev), `otlp` (prod → `OTEL_EXPORTER_OTLP_ENDPOINT` руу
OTLP/gRPC), эсвэл `""` (идэвхгүй)-ийг сонгоно, `SampleRatio` head sampler-тай
(prod-д ихэвчлэн 0.01–0.1).

`TracingMiddleware` нь **эхний** дэлхийн middleware — ингэснээр `RequestIDMiddleware`
нь `trace_id`-г log context-д холбож чадна. pgx pool (`driver_pgx.go`) болон Redis
cache-ийг мөн багажлагдсан (instrumented). `Shutdown` hook нь SIGTERM дээр буфферлэсэн
span-уудыг flush хийдэг.

## Metrics (Prometheus)

`promhttp.Handler()` нь `/metrics`-д mount хийгдсэн. `MetricsMiddleware` нь
`httpRequestsTotal` (CounterVec) болон `httpRequestDuration` (HistogramVec)-ийг
бүртгэдэг. DB pool статистикийг амьд pgxpool-оос
`observability.RegisterDBStatsProvider`-ээр бүртгэнэ.

## Log-ууд (Zap)

`pkg/logger/zap.go`-аар дамжуулан бүтэцлэгдсэн logging; `AccessLogMiddleware` нь
дэлхийн гинжинд байрлаж, `request_id` / `trace_id` агуулсан хүсэлт тус бүрийн access
log гаргадаг.

## Production-ийн хамгаалалт (gating)

`/metrics` болон `/swagger/doc.json` хоёул
`ObservabilityGate(isProduction, OBSERVABILITY_TOKEN)`-ийн ард байрладаг:

- production бус → үргэлж нээлттэй;
- production, хоосон token → бүрэн хаалттай (**404**);
- production, token тохируулсан → `Authorization: Bearer <token>` шаардана,
  `crypto/subtle.ConstantTimeCompare`-ээр харьцуулна; **аливаа таарахгүй нь 404
  буцаана** (401 биш) — эндпойнтын оршин байгааг тагнуулаас нуухын тулд.

`/health` болон `/ready` нь load balancer-үүдэд нээлттэй хэвээр байна.
