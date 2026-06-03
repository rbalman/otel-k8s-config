Otel operator

```shell
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
##Mark: Defaults to memory backend
helm install jaeger jaegertracing/jaeger --version 4.8.0 -n monitoring
```