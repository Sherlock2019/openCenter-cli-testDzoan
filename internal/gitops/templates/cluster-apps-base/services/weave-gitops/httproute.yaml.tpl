apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: gitops
  namespace: flux-system
spec:
  parentRefs:
    - name: rmpk-gateway
      sectionName: gitops-https
      namespace: rackspace-system
  hostnames:
    - {{ (index .OpenCenter.Services "weave-gitops").Hostname | default (printf "gitops.%s" .OpenCenter.Cluster.ClusterFQDN) | quote }}
  rules:
    - backendRefs:
        - name: weave-gitops
          port: 9001
