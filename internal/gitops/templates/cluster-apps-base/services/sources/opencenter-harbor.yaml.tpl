---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-harbor
  namespace: flux-system
spec:
  interval: 15m
  {{- $service := index .OpenCenter.Services "harbor" }}
  url: {{ $service.Uri | default .OpenCenter.GitOps.GitOpsBaseRepo }}
  ref:
    branch: {{ $service.Branch | default .OpenCenter.GitOps.GitOpsBranch | default "main" }}
  secretRef:
    name: opencenter-base
