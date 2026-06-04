# Dev Environment

## Prerequisites

- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [helm](https://helm.sh/docs/intro/install/)

---

## 1. Create the cluster

```bash
kind create cluster --config kind.yaml
kubectl cluster-info --context kind-gokite-cluster
```

---

## 2. Install ArgoCD

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

helm upgrade --install argocd dev/argocd \
  --namespace argocd \
  --create-namespace \
  --dependency-update
```

Wait for ArgoCD to be ready:

```bash
kubectl wait --for=condition=available deployment/argocd-server \
  --namespace argocd --timeout=120s
```

Retrieve the initial admin password:

```bash
kubectl get secret argocd-initial-admin-secret \
  --namespace argocd \
  -o jsonpath="{.data.password}" | base64 -d && echo
```

Access the UI via ingress (once ingress-nginx is deployed):

```
http://localhost:8080/argocd  (user: admin)
```

Or via port-forward if ingress is not yet available:

```bash
kubectl port-forward svc/argocd-server -n argocd 9090:80
# open http://localhost:9090  (user: admin)
```

---

## 3. Register the Git repository

If the repository is private, register credentials before bootstrapping:

```bash
kubectl create secret generic gokite-repo \
  --namespace argocd \
  --from-literal=type=git \
  --from-literal=url=https://github.com/rbalman/temp-repo.git \
  --from-literal=username=<github-username> \
  --from-literal=password=<github-token>

kubectl label secret gokite-repo \
  --namespace argocd \
  argocd.argoproj.io/secret-type=repository
```

---

## 4. Bootstrap the stack

```bash
kubectl apply -f dev/addons/bootstrap.yaml
kubectl apply -f dev/apps/bootstrap.yaml
```

This creates the ArgoCD Applications which render the Helm wrapper charts under `dev/addons/` and `dev/apps/` and deploy all child applications in sync-wave order:

| Wave | Addon |
|------|-------|
| 1 | ingress, minio |
| 2 | prometheus, grafana |
| 3 | loki, jaeger |
| 4 | otel-cluster, otel-node |

---

## 5. Manage apps

**Disable an addon** — set `enabled: false` in `dev/addons/stack/values.yaml` and push.

**Add a new addon** — add an entry under `apps:` in `dev/addons/stack/values.yaml` and create `dev/addons/<name>/Chart.yaml` + `dev/addons/<name>/values.yaml`.

**Upgrade a chart** — bump `version:` in the relevant `dev/addons/<name>/Chart.yaml` and push. ArgoCD picks it up on the next sync.
