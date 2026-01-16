---
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
    - {{ .OpenCenter.Services.headlamp.Hostname | default (printf "headlamp.%s" .OpenCenter.Cluster.ClusterFQDN) | quote }}
  rules:
    - backendRefs:
        - name: headlamp
          port: 80
      matches:
        - path:
            type: PathPrefix
            value: /
