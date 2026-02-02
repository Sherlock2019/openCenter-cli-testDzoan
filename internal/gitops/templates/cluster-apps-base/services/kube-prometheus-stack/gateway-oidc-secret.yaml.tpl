apiVersion: v1
data:
    client-secret: {{ .Secrets.Keycloak.ClientSecret }}
kind: Secret
metadata:
    name: gateway-oidc-secret
    namespace: observability
type: Opaque
