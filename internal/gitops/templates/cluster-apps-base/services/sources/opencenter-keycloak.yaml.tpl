---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-keycloak
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockForService "keycloak" }}
