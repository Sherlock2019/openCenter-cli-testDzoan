apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: headlamp
  namespace: headlamp
spec:
  parentRefs:
  - name: rmpk-gateway
  sectionName: headlamp-https
  namespace: rackspace-system
  hostnames:
  - "{{ .OpenCenter.Services.headlamp.Hostname | default (printf "headlamp.%s" .OpenCenter.Cluster.ClusterFQDN) }}"
  rules:
  - backendRefs:
    - name: headlamp
  port: 80
---
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: headlamp-http-redirect
  namespace: headlamp
spec:
  parentRefs:
  - name: rmpk-gateway
  namespace: rackspace-system
  sectionName: headlamp-http
  hostnames:
  - "{{ .OpenCenter.Services.headlamp.Hostname | default (printf "headlamp.%s" .OpenCenter.Cluster.ClusterFQDN) }}"
  rules:
  - filters:
    - type: RequestRedirect
  requestRedirect:
    scheme: https
    statusCode: 301 # Permanent redirect
