{{- $service := index .OpenCenter.Services "vsphere-csi" }}
{{- $namespace := $service.Namespace | default "vmware-system-csi" }}
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: vsphere-csi-base
  namespace: flux-system
spec:
  dependsOn:
    - name: sources
      namespace: flux-system
  interval: 15m
  retryInterval: 1m
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: opencenter-vsphere-csi
    namespace: flux-system
  path: applications/base/services/vsphere-csi
  targetNamespace: {{ $namespace }}
  prune: true
  healthChecks:
    - apiVersion: apps/v1
      kind: Deployment
      name: vsphere-csi-controller
      namespace: {{ $namespace }}
  commonMetadata:
    labels:
      app.kubernetes.io/part-of: vsphere-csi
      app.kubernetes.io/managed-by: flux
      opencenter/managed-by: opencenter
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: vsphere-csi-override
  namespace: flux-system
spec:
  interval: 15m
  retryInterval: 1m
  timeout: 5m
  sourceRef:
    kind: GitRepository
    name: flux-system
    namespace: flux-system
  decryption:
    provider: sops
    secretRef:
      name: sops-age
  path: ./applications/overlays/{{ .OpenCenter.Cluster.ClusterName }}/services/vsphere-csi
  targetNamespace: {{ $namespace }}
  prune: true
  commonMetadata:
    labels:
      app.kubernetes.io/part-of: vsphere-csi
      app.kubernetes.io/managed-by: flux
      opencenter/managed-by: opencenter
