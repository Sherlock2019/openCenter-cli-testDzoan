---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-kube-prometheus-stack
  namespace: flux-system
spec:
  interval: 15m
  {{- $service := index .OpenCenter.Services "kube-prometheus-stack" }}
  url: {{ $service.Uri | default .OpenCenter.GitOps.GitOpsBaseRepo }}
  ref:
    branch: {{ $service.Branch | default .OpenCenter.GitOps.GitOpsBranch | default "main" }}
{{- if not (hasPrefix "https://" .OpenCenter.GitOps.GitOpsBaseRepo) }}
  secretRef:
    name: opencenter-base
{{- end }}
