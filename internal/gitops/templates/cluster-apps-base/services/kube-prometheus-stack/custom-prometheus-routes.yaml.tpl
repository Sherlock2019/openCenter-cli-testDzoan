---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: prometheus-gateway-route
  namespace: observability
spec:
  hostnames:
    - {{ (index .OpenCenter.Services "kube-prometheus-stack").Hostname | default (printf "prometheus.%s" .OpenCenter.Cluster.ClusterFQDN) | quote }}
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
