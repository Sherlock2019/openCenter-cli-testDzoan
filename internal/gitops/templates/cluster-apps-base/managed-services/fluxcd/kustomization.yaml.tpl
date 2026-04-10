---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
{{- $hasManagedServices := false }}
{{- range $name, $service := .OpenCenter.ManagedServices }}
  {{- if $service.Enabled }}
  {{- $hasManagedServices = true }}
  {{- end }}
{{- end }}
{{- if $hasManagedServices }}
  - ./sources.yaml
{{- end }}
{{- if (index .OpenCenter.ManagedServices "alert-proxy").Enabled }}
  - ./alert-proxy.yaml
{{- end }}
