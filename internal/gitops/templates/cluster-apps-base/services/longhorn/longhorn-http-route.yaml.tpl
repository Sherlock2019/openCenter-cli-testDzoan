---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: longhorn-gateway-route
  namespace: longhorn-system
spec:
  hostnames:
  - {{ .OpenCenter.Services.longhorn.Hostname | default (printf "longhorn.%s" .OpenCenter.Cluster.ClusterFQDN) | quote }}
  parentRefs:
  - group: gateway.networking.k8s.io
  kind: Gateway
  name: rmpk-gateway
  namespace: rackspace-system
  sectionName: longhorn-https
  rules:
  - backendRefs:
    - group: ""
  kind: Service
  name: longhorn-frontend
  port: 80
  weight: 1
  matches:
    - path:
    type: PathPrefix
    value: /
