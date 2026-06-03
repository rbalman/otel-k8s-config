```shell
helm repo add argo https://argoproj.github.io/argo-helm
helm repo update
helm upgrade --install argocd argo/argo-cd --version 9.5.17 --values ./values.yaml -n argocd --create-namespace
```

**Access UI:**
```shell
kubectl port-forward svc/argocd-server -n argocd 4040:443
```
Then open https://localhost:4040

**Initial admin password:**
```shell
kubectl get secret argocd-initial-admin-secret -n argocd -o jsonpath="{.data.password}" | base64 -d
```
Username: `admin`
