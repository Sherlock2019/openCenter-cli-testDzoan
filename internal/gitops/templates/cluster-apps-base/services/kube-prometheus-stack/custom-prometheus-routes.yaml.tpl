{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: prometheus-gateway-route
  namespace: observability
spec:
  hostnames:
    - "prometheus.{{ .Values.cluster.name }}.k8s.opencenter.cloud"
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: rmpk-gateway
      namespace: rackspace-system
      sectionName: prometheus-https
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: observability-kube-prometh-prometheus
          port: 9090
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
