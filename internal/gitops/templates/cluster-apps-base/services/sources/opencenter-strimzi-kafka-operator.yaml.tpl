---
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-strimzi-kafka-operator
  namespace: flux-system
spec:
  interval: 15m
{{ sourceAuthBlockForService "strimzi-kafka-operator" }}
