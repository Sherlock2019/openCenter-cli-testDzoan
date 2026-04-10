---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: alert-proxy
  namespace: flux-system
spec:
  interval: 1h
  url: {{ (index .OpenCenter.ManagedServices "alert-proxy").Uri | default (index .OpenCenter.ManagedServices "alert-proxy").GitOpsSourceRepo }}
  ref:
    branch: {{ (index .OpenCenter.ManagedServices "alert-proxy").Branch | default (index .OpenCenter.ManagedServices "alert-proxy").GitOpsSourceBranch | default "main" }}
