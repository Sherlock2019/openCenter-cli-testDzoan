apiVersion: v1
kind: Secret
metadata:
  name: alert-manager-url-secret
type: Opaque
data:
  alert_manager_url: {{ ((index .OpenCenter.ManagedServices "alert-proxy").AlertManagerBaseURL | default "") | b64enc }}
