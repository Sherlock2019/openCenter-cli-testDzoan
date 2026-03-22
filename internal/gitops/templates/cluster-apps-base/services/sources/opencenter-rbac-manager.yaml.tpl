---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-rbac-manager
  namespace: flux-system
spec:
  interval: 15m
  {{- $service := index .OpenCenter.Services "rbac-manager" }}
  url: {{ $service.Uri | default .OpenCenter.GitOps.GitOpsBaseRepo }}
  ref:
    branch: {{ $service.Branch | default .OpenCenter.GitOps.GitOpsBranch | default "main" }}
  secretRef:
    name: opencenter-base
