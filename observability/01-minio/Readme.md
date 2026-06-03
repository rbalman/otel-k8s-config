```shell
helm repo add minio https://charts.min.io/
helm repo update

helm upgrade --install minio minio/minio --values ./values.yaml -n minio --create-namespace
```

## Mark: TODO: auto create credentails and bucket