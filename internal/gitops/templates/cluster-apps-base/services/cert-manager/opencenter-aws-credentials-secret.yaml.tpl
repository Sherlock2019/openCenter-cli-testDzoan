apiVersion: v1
data:
  access-key-id: {{ .GetCertManagerAWSAccessKey }}
  secret-access-key: {{ .GetCertManagerAWSSecretKey }}
kind: Secret
metadata:
  name: opencenter-aws-credentials-secret
