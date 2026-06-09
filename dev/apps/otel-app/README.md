# otel-app

Helm wrapper chart for the OTel demo application (`otel-app-chart 0.5.0`) deployed in the `default` namespace.

## What it does

- Deploys the `balman/otel-demo` Go application (image tag `v0.0.5`) — instrumented with the OTel SDK for metrics, traces, and logs.
- Configures the app to export telemetry to [`otel-cluster`](../../addons/otel-cluster/README.md) via OTLP/gRPC at `otel-cluster.monitoring.svc.cluster.local:4317`.
- Annotates the pod with `prometheus.io/scrape: "true"` so the OTel Collector auto-discovers it and scrapes `/metrics` on port `8080`.
- Exposes the app at `http://otel-app.localhost:8080` via ingress.
- Deploys a **Grafana dashboard ConfigMap** (labelled `grafana_dashboard: "1"`) in the same namespace — Grafana's sidecar auto-loads it as the **OTel App - Golden Signals** dashboard, covering request rate, error rate, P50/P99 latency, traffic by route, and log volume.

## Key values

| Value | Default |
|---|---|
| `app.image.repository` | `balman/otel-demo` |
| `app.image.tag` | `v0.0.5` |
| `app.extraEnvs.OTEL_SERVICE_NAME` | `otel-app` |
| `app.extraEnvs.OTEL_EXPORTER_OTLP_ENDPOINT` | `otel-cluster.monitoring.svc.cluster.local:4317` |
| `app.ingress.host` | `otel-app.localhost` |

## Access

| URL | Credentials |
|---|---|
| http://otel-app.localhost:8080 | — |

## Grafana dashboard

The bundled dashboard (`OTel App - Golden Signals`, UID `otel-app`) panels:

| Section | Panels |
|---|---|
| Golden Signals | Request rate, error rate (5xx), P50 latency, P99 latency |
| Traffic | Request rate by route, request rate by status code |
| Errors | Error rate % by route, 5xx error rate |
| Latency | Latency percentiles by route, request duration heatmap |
| Logs | Log volume by level, live log stream |
