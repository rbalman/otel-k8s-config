**Deploy via ArgoCD (recommended):**
```shell
kubectl apply -f application.yaml
```

**Deploy manually via Helm:**
```shell
helm repo add minio https://charts.min.io/
helm repo update
helm upgrade --install minio minio/minio --version 5.4.0 --values ./values.yaml -n minio --create-namespace
```

## Mark: TODO: auto create credentials and buckets