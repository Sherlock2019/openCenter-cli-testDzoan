{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-olm-config
  namespace: flux-system
spec:
  interval: 15m
  url: ssh://git@github.com/rpc-environments/5918681-computacenter-united-states-inc.git
  ref:
    branch: main
  secretRef:
    name: flux-system
  include:
    - repository:
        name: opencenter-olm
      fromPath: applications/base/services/olm
      toPath: applications/overlays/{{ .Values.cluster.name }}/services/base/olm/
