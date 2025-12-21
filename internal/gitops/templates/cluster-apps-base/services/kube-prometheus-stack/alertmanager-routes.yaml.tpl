---
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: SecurityPolicy
metadata:
  name: alertmanager-oidc
  namespace: observability
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: alertmanager-gateway-route
  oidc:
    provider:
      issuer: "https://auth.{{ .OpenCenter.Cluster.ClusterName }}.{{ .OpenCenter.Meta.Region }}.k8s.opencenter.cloud/realms/opencenter"
    clientID: "opencenter"  
    clientSecret:
      name: "gateway-oidc-secret" 
    redirectURL: "https://alertmanager.{{ .OpenCenter.Cluster.ClusterName }}.{{ .OpenCenter.Meta.Region }}.k8s.opencenter.cloud/oauth2/callback"
    scopes:
      - openid
      - profile
      - email
      - roles
    logoutPath: "/logout"
