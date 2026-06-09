# stack (apps)

ArgoCD app-of-apps chart that renders one `Application` resource per app and deploys them into the cluster.

## What it does

- Iterates over `values.apps` and emits an ArgoCD `Application` for each enabled entry.
- Each `Application` points to `dev/apps/<name>` in this repo (`HEAD` revision) and targets the configured namespace.
- Sync wave annotation is added only when specified; otherwise ArgoCD uses its default ordering.
- All applications use automated sync with `prune: true` and `selfHeal: true`, and `CreateNamespace=true`.

## Apps managed

| App | Namespace | Chart |
|---|---|---|
| otel-app | `default` | [otel-app](../otel-app/README.md) |

## Key values

| Value | Default |
|---|---|
| `appRepo` | `https://github.com/rbalman/otel-k8s-config.git` |
| `revision` | `HEAD` |
| `appsPath` | `dev/apps` |

## Bootstrap

```bash
kubectl apply -f dev/apps/bootstrap.yaml
```
