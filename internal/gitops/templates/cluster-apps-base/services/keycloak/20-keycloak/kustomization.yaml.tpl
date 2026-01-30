apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization

resources:
  - keycloak-cr-patch.yaml
  {{- if and (.OpenCenter.Services.keycloak.MinReplicas) (.OpenCenter.Services.keycloak.MaxReplicas) }}
  - keycloak-hpa.yaml
  {{- end }}
  {{- if .OpenCenter.Services.keycloak.BackupEnabled | default true }}
  - keycloak-backup-cronjob.yaml
  {{- end }}
