# stack (addons)

ArgoCD app-of-apps chart that renders one `Application` resource per addon and deploys them in sync-wave order.

## What it does

- Iterates over `values.apps` and emits an ArgoCD `Application` for each enabled entry.
- Each `Application` points to `dev/addons/<name>` in this repo (`HEAD` revision) and targets the configured namespace.
- Sets `argocd.argoproj.io/sync-wave` so ArgoCD deploys components in dependency order.
- All applications use automated sync with `prune: true` and `selfHeal: true`, and `CreateNamespace=true`.

## Deployment order (sync waves)

| Wave | Component | Namespace | Chart |
|---|---|---|---|
| 1 | ingress | `ingress-nginx` | [ingress](../ingress/README.md) |
| 1 | minio | `minio` | [minio](../minio/README.md) |
| 2 | prometheus | `monitoring` | [prometheus](../prometheus/README.md) |
| 2 | grafana | `monitoring` | [grafana](../grafana/README.md) |
| 3 | loki | `loki` | [loki](../loki/README.md) |
| 3 | jaeger | `monitoring` | [jaeger](../jaeger/README.md) |
| 4 | otel-cluster | `monitoring` | [otel-cluster](../otel-cluster/README.md) |
| 4 | otel-node | `monitoring` | [otel-node](../otel-node/README.md) |

## Key values

| Value | Default |
|---|---|
| `appRepo` | `https://github.com/rbalman/otel-k8s-config.git` |
| `revision` | `HEAD` |
| `addonsPath` | `dev/addons` |

## Bootstrap

```bash
kubectl apply -f dev/addons/bootstrap.yaml
```
