```shell
helm upgrade --install loki grafana-community/loki --values ./values.yaml --version 7.0.0  -n loki --create-namespace
```

##Mark: Todo: auto link the trace_id in grafana ui