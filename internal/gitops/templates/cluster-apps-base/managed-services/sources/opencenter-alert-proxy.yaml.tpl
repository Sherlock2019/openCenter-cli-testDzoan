---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-alert-proxy
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockDefault }}
