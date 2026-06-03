```shell
helm repo add open-telemetry https://open-telemetry.github.io/opentelemetry-helm-charts
helm install trace-collector open-telemetry/opentelemetry-collector --values ./values.yaml -n monitoring
```