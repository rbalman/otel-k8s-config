# Dev Environment

## Prerequisites

- [kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)
- [helm](https://helm.sh/docs/intro/install/)

---

## 1. Create the cluster

```bash
kind create cluster --name gokite
kubectl cluster-info --context kind-gokite
```

---

## 2. Install ArgoCD

```bash
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update

helm upgrade --install argocd argo/argo-cd \
  --namespace argocd \
  --create-namespace \
  --values dev/argocd/values.yaml
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

Access the UI (port-forward in a separate terminal):

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
# open https://localhost:8080  (user: admin)
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
```

This creates the `addons` ArgoCD Application which renders `dev/addons/stack/` as a Helm chart and deploys all child applications in sync-wave order:

| Wave | App |
|------|-----|
| 1 | minio |
| 2 | prom-stack |
| 3 | loki, jaeger |
| 4 | alloy, otel-collector |

---

## 5. Manage apps

**Disable an app** — set `enabled: false` in `dev/addons/stack/values.yaml` and push.

**Add a new app** — add an entry under `apps:` in `dev/addons/stack/values.yaml` and create `dev/addons/<name>/values.yaml`.
