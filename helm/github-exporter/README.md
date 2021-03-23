# github-exporter helm chart

# Deployment
```
cd github-exporter/helm

helm upgrade \
github-exporter github-exporter/ \
--install \
--namespace=exporters \
-f helm_vars/nonprod/values.yaml \
--debug --dry-run

```