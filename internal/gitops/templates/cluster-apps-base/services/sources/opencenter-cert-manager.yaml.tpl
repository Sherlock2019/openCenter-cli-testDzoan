---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-cert-manager
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockForService "cert-manager" }}
