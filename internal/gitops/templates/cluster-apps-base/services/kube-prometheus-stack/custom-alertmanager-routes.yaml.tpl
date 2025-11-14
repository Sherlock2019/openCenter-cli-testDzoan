---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: alertmanager-gateway-route
  namespace: observability
spec:
  hostnames:
    - {{ (index .OpenCenter.Services "kube-prometheus-stack").Hostname | default (printf "alertmanager.%s" .OpenCenter.Cluster.ClusterFQDN) | quote }}
  parentRefs:
    - group: gateway.networking.k8s.io
      kind: Gateway
      name: rmpk-gateway
      namespace: rackspace-system
      sectionName: alertmanager-https
  rules:
    - backendRefs:
        - group: ""
          kind: Service
          name: observability-kube-prometh-alertmanager
          port: 9093
          weight: 1
      matches:
        - path:
            type: PathPrefix
            value: /
