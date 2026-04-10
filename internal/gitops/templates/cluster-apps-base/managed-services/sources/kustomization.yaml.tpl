---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
namespace: flux-system
resources:
{{- $managedServices := .OpenCenter.ManagedServices }}
{{- if (index $managedServices "alert-proxy").Enabled }}
  - "./opencenter-alert-proxy.yaml"
{{- end }}
