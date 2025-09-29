{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: keycloak
  namespace: keycloak
spec:
  parentRefs:
    - name: rmpk-gateway
      sectionName: keycloak-https
      namespace: rackspace-system
  hostnames:
    - "auth.{{ .Values.cluster.name }}.k8s.opencenter.cloud"
  rules:
    - backendRefs:
        - name: keycloak-service
          port: 8080
