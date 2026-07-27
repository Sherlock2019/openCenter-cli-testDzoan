---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-keycloak-config
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockCustomerRepository }}
  include:
    - repository:
        name: opencenter-keycloak
      fromPath: applications/base/services/keycloak
      toPath: applications/overlays/{{ .OpenCenter.Cluster.ClusterName }}/services/base/keycloak/
