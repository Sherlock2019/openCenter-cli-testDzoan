{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: observability
resources:
  - "./custom-alertmanager-routes.yaml"
  - "./custom-prometheus-routes.yaml"
  - "./custom-grafana-routes.yaml"
secretGenerator:
  - name: kube-prometheus-stack-values-override
    type: Opaque
    files: [override.yaml=helm-values/override-values.yaml]
    options:
      disableNameSuffixHash: true
