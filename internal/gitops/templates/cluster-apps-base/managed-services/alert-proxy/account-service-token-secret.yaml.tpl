apiVersion: v1
kind: Secret
metadata:
  name: account-service-token-secret
type: generic
stringData:
  account_service_token: {{ .Secrets.AlertProxy.AccountServiceToken | b64enc }}
