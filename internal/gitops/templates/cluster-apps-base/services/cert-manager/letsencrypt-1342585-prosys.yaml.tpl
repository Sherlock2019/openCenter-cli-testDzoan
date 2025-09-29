{/*
This file was generated from overlay template comparison
Environment-specific values are templated with Go template syntax
Original source: dev environment overlay
*/}
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-{{ Values.cluster.name }}
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: {{.Values.services.cert-manager.email }}
    privateKeySecretRef:
      name: letsencrypt-dns01
    solvers:
      - dns01:
          route53:
            region: us-east-1
            accessKeyIDSecretRef:
              name: "{{ .Values.cluster.name }}-aws-credentials-secret"
              key: access-key-id
            secretAccessKeySecretRef:
              name: "{{ .Values.cluster.name }}-aws-credentials-secret"
              key: secret-access-key
        selector:
          dnsZones:
            - "{{ .Values.cluster.name }}.k8s.opencenter.cloud"
