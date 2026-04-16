apiVersion: "acid.zalan.do/v1"
kind: postgresql
metadata:
  name: postgres-cluster
  namespace: keycloak
spec:
  teamId: "acid"
  numberOfInstances: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}1{{ else }}3{{ end }}
  postgresql:
    version: "17"
    parameters:
      shared_buffers: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"256MB"{{ else }}"2GB"{{ end }}
      max_connections: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"100"{{ else }}"1024"{{ end }}
      log_statement: "all"
  volume:
    size: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}5Gi{{ else }}20Gi{{ end }}
    storageClass: {{ .OpenCenter.Infrastructure.Storage.DefaultStorageClass | default "standard" }}
  databases:
    keycloak: keycloak
  users:
    keycloak:
      - superuser
      - createdb
  resources:
    limits:
      cpu: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"500m"{{ else }}"2"{{ end }}
      memory: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"512Mi"{{ else }}"3000Mi"{{ end }}
    requests:
      cpu: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"100m"{{ else }}"1"{{ end }}
      memory: {{ if eq .OpenCenter.Infrastructure.Provider "kind" }}"256Mi"{{ else }}"1000Mi"{{ end }}
