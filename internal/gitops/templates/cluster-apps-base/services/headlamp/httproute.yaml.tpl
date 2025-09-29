{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
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
    - "headlamp.{{ .Values.cluster.name }}.k8s.opencenter.cloud"
  rules:
    - backendRefs:
        - name: headlamp
          port: 80
