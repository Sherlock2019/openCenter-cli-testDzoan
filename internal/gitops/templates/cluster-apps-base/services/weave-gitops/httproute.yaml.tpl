{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
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
    - "gitops.{{ .Values.cluster.name }}.k8s.opencenter.cloud"
  rules:
    - backendRefs:
        - name: weave-gitops
          port: 9001
