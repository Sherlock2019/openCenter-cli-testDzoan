---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-harbor
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockForService "harbor" }}
