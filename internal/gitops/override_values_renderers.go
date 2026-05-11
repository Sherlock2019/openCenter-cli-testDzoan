// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gitops

import (
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"
)

func init() {
	RegisterOverrideValuesRenderer("openstack-ccm", templateRenderer(openstackCCMTemplate))
	RegisterOverrideValuesRenderer("openstack-csi", templateRenderer(openstackCSITemplate))
	RegisterOverrideValuesRenderer("vsphere-csi", templateRenderer(vsphereCsiTemplate))
	RegisterOverrideValuesRenderer("velero", templateRenderer(veleroTemplate))
	RegisterOverrideValuesRenderer("loki", templateRenderer(lokiTemplate))
	RegisterOverrideValuesRenderer("tempo", templateRenderer(tempoTemplate))
	RegisterOverrideValuesRenderer("mimir", templateRenderer(mimirTemplate))
	RegisterOverrideValuesRenderer("opentelemetry-kube-stack", staticRenderer(otelTemplate))
	RegisterOverrideValuesRenderer("headlamp", templateRenderer(headlampTemplate))
	RegisterOverrideValuesRenderer("harbor", templateRenderer(harborTemplate))
	RegisterOverrideValuesRenderer("kube-prometheus-stack", templateRenderer(kubePrometheusStackTemplate))
}

// templateRenderer creates a renderer that executes a Go template against the config.
func templateRenderer(tmpl string) OverrideValuesRenderer {
	return func(cfg v2.Config) (string, error) {
		funcMap := sprig.TxtFuncMap()
		t, err := template.New("override-values").Funcs(funcMap).Parse(tmpl)
		if err != nil {
			return "", err
		}
		var buf strings.Builder
		if err := t.Execute(&buf, cfg); err != nil {
			return "", err
		}
		return buf.String(), nil
	}
}

// staticRenderer returns a renderer that always produces the same content.
func staticRenderer(content string) OverrideValuesRenderer {
	return func(_ v2.Config) (string, error) {
		return content, nil
	}
}

// --- Templates (moved from .tpl files) ---

const openstackCCMTemplate = `cloudConfig:
  global:
    auth-url: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL }}
    application-credential-id: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialID }}
    application-credential-secret: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialSecret }}
    domain-name: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Domain | default "default" }}
    region: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
    tenant-name: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.TenantName }}
    tls-insecure: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Insecure | default false }}
  loadBalancer:
    floating-network-id: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.FloatingNetworkID }}
    subnet-id: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.SubnetID }}
`

const openstackCSITemplate = `secret:
  enabled: true
  hostMount: true
  create: true
  filename: cloud.conf
  name: cinder-csi-cloud-config
  data:
    cloud.conf: |-
      [Global]
      auth-url = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL }}
      application-credential-id = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialID }}
      application-credential-secret = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.ApplicationCredentialSecret }}
      domain-name = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Domain }}
      region = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
      tenant-name = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.TenantName }}
      tls-insecure = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Insecure | default false }}
`

const vsphereCsiTemplate = `global:
  config:
    existingSecret: "vsphere-csi"
    global:
      cluster-id: "{{ .OpenCenter.Meta.Name }}"
    csidriver:
      enabled: true
    storageclass:
      enabled: true
      name: "{{ .OpenCenter.Storage.DefaultStorageClass }}"
      storagepolicyname: ""
      expansion: true
      default: true
      reclaimPolicy: Delete
      volumebindingmode: "Immediate"
      datastoreurl: {{ .Secrets.VSphereCsi.Datastoreurl }}
vsphere-cpi:
  enabled: true
  global:
    config:
      existingConfig:
        enabled: true
        type: Secret
        name: "vsphere-cpi-secret"
      secretsInline: false
controller:
  config:
    block-volume-snapshot: true
  replicaCount: 3
  snapshotter:
    image:
      registry: {{ (index .OpenCenter.Services "vsphere-csi").Image.Repository | default "registry.k8s.io" }}
      repository: sig-storage/csi-snapshotter
      tag: {{ (index .OpenCenter.Services "vsphere-csi").Image.Tag | default "v8.2.0" }}
      pullPolicy: IfNotPresent
    args:
      - "--v=4"
      - "--kube-api-qps=100"
      - "--kube-api-burst=100"
      - "--timeout=300s"
      - "--csi-address=$(ADDRESS)"
      - "--leader-election"
      - "--leader-election-lease-duration=120s"
      - "--leader-election-renew-deadline=60s"
      - "--leader-election-retry-period=30s"
snapshot:
  controller:
    enabled: true
    replicaCount: 1
`

const veleroTemplate = `---
credentials:
  extraSecretRef: "cloud-credentials"
configuration:
  features: EnableCSI
  defaultSnapshotMoveData: false
  defaultVolumesToFsBackup: false
  backupStorageLocation:
    - name: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
      provider: community.openstack.org/openstack
      default: true
      bucket: {{ .OpenCenter.Meta.Name }}-velero
      config:
        region: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
  volumeSnapshotLocation: []
initContainers:
  - name: velero-plugin-openstack
    image: lirt/velero-plugin-for-openstack:v0.6.0
    imagePullPolicy: IfNotPresent
    volumeMounts:
      - mountPath: /target
        name: plugins
snapshotsEnabled: true
backupsEnabled: true
deployNodeAgent: false
extraObjects:
  - apiVersion: snapshot.storage.k8s.io/v1
    kind: VolumeSnapshotClass
    metadata:
      name: velero-vsphere-snapshot-class
      labels:
        velero.io/csi-volumesnapshot-class: "true"
    driver: csi.vsphere.vmware.com
    deletionPolicy: Delete
`

const lokiTemplate = `{{- $loki := index .OpenCenter.Services "loki" -}}
{{- $storageType := $loki.StorageType | default "s3" -}}
{{- $bucketName := $loki.BucketName | default (printf "%s-loki" .OpenCenter.Meta.Name) -}}
global:
    dnsService: coredns
loki:
    auth_enabled: true
    schemaConfig:
        configs:
            - from: "2025-01-01"
              store: tsdb
              object_store: {{ $storageType }}
              schema: v13
              index:
                prefix: index_
                period: 24h
    storage:
        bucketNames:
            chunks: {{ $bucketName }}
            ruler: {{ $bucketName }}
            admin: {{ $bucketName }}
        type: {{ $storageType }}
{{- if eq $storageType "swift" }}
        swift:
            auth_version: {{ $loki.SwiftAuthVersion | default 3 }}
            auth_url: {{ $loki.SwiftAuthURL }}
            region_name: {{ $loki.SwiftRegion | default .OpenCenter.Meta.Region }}
            application_credential_id: {{ $loki.SwiftApplicationCredentialID }}
            application_credential_secret: {{ .GetLokiSwiftApplicationCredentialSecret }}
            user_domain_name: {{ $loki.SwiftUserDomainName }}
            domain_name: {{ $loki.SwiftDomainName }}
            container_name: {{ $loki.SwiftContainerName | default $bucketName }}
{{- else }}
        s3:
            s3: null
            endpoint: {{ $loki.S3Endpoint | default (printf "https://swift.api.%s.rackspacecloud.com" .OpenCenter.Meta.Region) }}
            region: {{ $loki.S3Region | default .OpenCenter.Meta.Region }}
            secretAccessKey: {{ .GetLokiS3SecretKey }}
            accessKeyId: {{ .GetLokiS3AccessKey }}
            signatureVersion: null
            s3ForcePathStyle: {{ $loki.S3ForcePathStyle }}
            insecure: {{ $loki.S3Insecure }}
            http_config: {}
            backoff_config: {}
            disable_dualstack: false
{{- end }}
    storage_config:
        tsdb_shipper:
            active_index_directory: /var/loki/index
            cache_location: /var/loki/index-cache
    serviceMonitor:
        enabled: true
write:
    replicas: 3
    persistence:
        enabled: true
        size: 100Gi
    podAntiAffinityPreset: soft
read:
    replicas: 3
    persistence:
        enabled: true
        size: 50Gi
    podAntiAffinityPreset: soft
backend:
    replicas: 3
    persistence:
        enabled: true
        size: 50Gi
    podAntiAffinityPreset: soft
gateway:
    replicas: 2
    ingress:
        enabled: false
chunksCache:
    enabled: true
    memcached:
        replicaCount: 3
resultsCache:
    enabled: true
    memcached:
        replicaCount: 3
`

const tempoTemplate = `{{- $tempo := index .OpenCenter.Services "tempo" -}}
{{- $storageType := $tempo.StorageType | default "s3" -}}
{{- $bucketName := $tempo.BucketName | default (printf "%s-tempo" .OpenCenter.Meta.Name) -}}
global:
    storageClass: {{ $tempo.StorageClass | default .OpenCenter.Infrastructure.Storage.DefaultStorageClass }}
storage:
    trace:
        backend: {{ $storageType }}
{{- if eq $storageType "swift" }}
        swift:
            auth_version: {{ $tempo.SwiftAuthVersion | default 3 }}
            auth_url: {{ $tempo.SwiftAuthURL }}
            region: {{ $tempo.SwiftRegion | default .OpenCenter.Meta.Region }}
            application_credential_id: {{ $tempo.SwiftApplicationCredentialID }}
            application_credential_secret: {{ .GetTempoSwiftApplicationCredentialSecret }}
            user_domain_name: {{ $tempo.SwiftUserDomainName }}
            domain_name: {{ $tempo.SwiftDomainName }}
            container_name: {{ $tempo.SwiftContainerName | default $bucketName }}
{{- else }}
        s3:
            bucket: {{ $bucketName }}
            endpoint: {{ $tempo.S3Endpoint | default (printf "swift.api.%s.rackspacecloud.com" .OpenCenter.Meta.Region) }}
            access_key: {{ .GetTempoS3AccessKey }}
            secret_key: {{ .GetTempoS3SecretKey }}
            region: {{ $tempo.S3Region | default .OpenCenter.Meta.Region }}
            forcepathstyle: {{ $tempo.S3ForcePathStyle }}
            insecure: {{ $tempo.S3Insecure }}
{{- end }}
multitenancyEnabled: true
ingester:
    persistence:
        size: {{ $tempo.VolumeSize | default 50 }}Gi
`

const mimirTemplate = `global:
    dnsService: coredns
alertmanager:
    enabled: false
metaMonitoring:
    dashboards:
        enabled: true
    serviceMonitor:
        enabled: true
    prometheusRule:
        enabled: true
        mimirAlerts: true
        mimirRules: true
kafka:
    enabled: false
mimir:
    structuredConfig:
        blocks_storage:
            backend: s3
            s3:
                bucket_name: {{ .OpenCenter.Cluster.ClusterName }}-mimir
                endpoint: swift.api.{{ .OpenCenter.Meta.Region }}.rackspacecloud.com
                access_key_id: {{ .Secrets.Global.AWS.Application.AccessKey | default "PLACEHOLDER-MIMIR-ACCESS-KEY" }}
                secret_access_key: {{ .Secrets.Global.AWS.Application.SecretAccessKey | default "PLACEHOLDER-MIMIR-SECRET-KEY" }}
        ingest_storage:
            kafka:
                address: kafka-cluster-kafka-brokers.kafka-system.svc.cluster.local:9092
                topic: mimir-ingest
                auto_create_topic_enabled: true
                auto_create_topic_default_partitions: 1000
        limits:
            ingestion_rate: 100000
            ingestion_burst_size: 500000
            max_global_series_per_user: 2000000
            compactor_blocks_retention_period: 14d
compactor:
    persistentVolume:
        storageClassName: {{ .OpenCenter.Storage.DefaultStorageClass }}
        size: 20Gi
distributor:
    replicas: 2
ingester:
    persistentVolume:
        storageClassName: {{ .OpenCenter.Storage.DefaultStorageClass }}
        size: 15Gi
    replicas: 3
    topologySpreadConstraints: {}
    affinity:
        podAntiAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                        - key: app.kubernetes.io/component
                          operator: In
                          values:
                            - ingester
                  topologyKey: kubernetes.io/hostname
    zoneAwareReplication:
        topologyKey: kubernetes.io/hostname
admin-cache:
    enabled: true
    replicas: 2
chunks-cache:
    enabled: true
    replicas: 2
    allocatedMemory: 500
index-cache:
    enabled: true
    replicas: 3
metadata-cache:
    enabled: true
results-cache:
    enabled: true
minio:
    enabled: false
overrides_exporter:
    replicas: 1
querier:
    replicas: 1
query_frontend:
    replicas: 2
ruler:
    enabled: false
store_gateway:
    persistentVolume:
        storageClassName: {{ .OpenCenter.Storage.DefaultStorageClass }}
        size: 15Gi
    replicas: 3
    topologySpreadConstraints: {}
    affinity:
        podAntiAffinity:
            requiredDuringSchedulingIgnoredDuringExecution:
                - labelSelector:
                    matchExpressions:
                        - key: app.kubernetes.io/component
                          operator: In
                          values:
                            - store-gateway
                  topologyKey: kubernetes.io/hostname
    zoneAwareReplication:
        topologyKey: kubernetes.io/hostname
gateway:
    replicas: 2
`

const otelTemplate = `collectors:
  daemon:
    config:
      exporters:
        otlphttp/loki:
          endpoint: http://observability-loki-gateway.observability.svc.cluster.local/otlp
          headers:
            X-Scope-OrgID: "default"
          compression: gzip
          timeout: 30s
          retry_on_failure:
            enabled: true
            initial_interval: 1s
            max_interval: 10s
            max_elapsed_time: 0s
          sending_queue:
            enabled: true
            num_consumers: 10
            queue_size: 2000
        otlp/tempo:
          endpoint: observability-tempo-distributor.observability.svc.cluster.local:4317
          headers:
            X-Scope-OrgID: "default"
          tls:
            insecure: true
          compression: gzip
          timeout: 30s
          retry_on_failure:
            enabled: true
            initial_interval: 1s
            max_interval: 10s
            max_elapsed_time: 0s
          sending_queue:
            enabled: true
            num_consumers: 10
            queue_size: 2000
`

const headlampTemplate = `config:
    oidc:
        enabled: true
        externalSecret:
            enabled: false
        secret:
            create: true
        clientID: opencenter
        clientSecret: {{ .Secrets.Headlamp.OIDCClientSecret }}
        issuerURL: https://{{ (index .OpenCenter.Services "keycloak").Hostname | default (printf "auth.%s" .OpenCenter.Cluster.ClusterFQDN) }}/realms/opencenter
        scopes: openid profile email groups
        callbackURL: https://{{ (index .OpenCenter.Services "headlamp").Hostname | default (printf "headlamp.%s" .OpenCenter.Cluster.ClusterFQDN) }}/oidc-callback
    pluginsDir: /build/plugins
initContainers:
    - command:
        - /bin/sh
        - -c
        - mkdir -p /build/plugins && cp -r /plugins/* /build/plugins/ && chown -R 100:101 /build
      image: ghcr.io/headlamp-k8s/headlamp-plugin-flux:latest
      imagePullPolicy: Always
      name: headlamp-plugins
      securityContext:
        runAsNonRoot: false
        privileged: false
        runAsUser: 0
        runAsGroup: 0
      volumeMounts:
        - mountPath: /build/plugins
          name: headlamp-plugins
volumeMounts:
    - mountPath: /build/plugins
      name: headlamp-plugins
volumes:
    - name: headlamp-plugins
      emptyDir: {}
`

const harborTemplate = `{{- $harbor := index .OpenCenter.Services "harbor" -}}
externalURL: https://{{ $harbor.Hostname | default (printf "harbor.%s" .OpenCenter.Cluster.ClusterFQDN) }}
logLevel: info
expose:
    type: clusterIP
persistence:
    enabled: true
    resourcePolicy: keep
    persistentVolumeClaim:
        registry:
            size: 100Gi
        jobservice:
            jobLog:
                size: 100Gi
        database:
            size: 100Gi
        redis:
            size: 100Gi
        trivy:
            size: 100Gi
    imageChartStorage:
        type: s3
        s3:
            region: {{ .OpenCenter.Meta.Region | upper }}
            bucket: {{ $harbor.S3Bucket | default (printf "%s-harbor" .OpenCenter.Cluster.ClusterName) }}
            accesskey: {{ .Secrets.Global.AWS.Application.AccessKey | default "PLACEHOLDER-HARBOR-ACCESS-KEY" }}
            secretkey: {{ .Secrets.Global.AWS.Application.SecretAccessKey | default "PLACEHOLDER-HARBOR-SECRET-KEY" }}
            regionendpoint: swift.api.{{ .OpenCenter.Meta.Region }}.rackspacecloud.com
            v4auth: true
            secure: true
            rootdirectory: images
harborAdminPassword: {{ $harbor.AdminPassword | default "PLACEHOLDER-HARBOR-ADMIN-PASSWORD" }}
metrics:
    enabled: true
    serviceMonitor:
        enabled: true
cache:
    enabled: true
    expireHours: 24
portal:
    replicas: 1
core:
    replicas: 1
jobservice:
    replicas: 1
registry:
    replicas: 1
    credentials:
        username: harbor-registry
        password: PLACEHOLDER-HARBOR-REGISTRY-PASSWORD
        htpasswdString: PLACEHOLDER-HARBOR-HTPASSWD
trivy:
    replicas: 1
database:
    internal:
        password: PLACEHOLDER-HARBOR-DATABASE-PASSWORD
exporter:
    replicas: 1
`

const kubePrometheusStackTemplate = `---
alertmanager:
  enabled: true
  alertmanagerSpec:
    externalUrl: https://{{ (index .OpenCenter.Services "kube-prometheus-stack").Hostname | default (printf "alertmanager.%s" .OpenCenter.Cluster.ClusterFQDN) }}
    storage:
      volumeClaimTemplate:
        spec:
          accessModes: ["ReadWriteOnce"]
          resources:
            requests:
              storage: 50Gi
  config:
    global:
      resolve_timeout: 5m
    inhibit_rules:
      - source_matchers: [severity = critical]
        target_matchers: [severity =~ warning|info]
        equal: [namespace, alertname]
      - source_matchers: [severity = warning]
        target_matchers: [severity = info]
        equal: [namespace, alertname]
      - source_matchers: [alertname = InfoInhibitor]
        target_matchers: [severity = info]
        equal: [namespace]
      - target_matchers: [alertname = InfoInhibitor]
    route:
      group_by: [namespace, alertname]
      group_wait: 30s
      group_interval: 60s
      repeat_interval: 12h
      routes:
        - receiver: "null"
          matchers: [alertname = "Watchdog"]
        - receiver: warning_alerts_receiver
          continue: false
          matchers: [severity =~ "warning"]
        - receiver: alert_proxy_receiver
          continue: false
          matchers: [severity =~ "critical"]
    receivers:
      - name: "null"
      - name: warning_alerts_receiver
        msteamsv2_configs:
          - send_resolved: true
            webhook_url: {{ (index .OpenCenter.Services "kube-prometheus-stack").WebhookURL }}
      - name: alert_proxy_receiver
        webhook_configs:
          - url: http://rackspace-alert-proxy.rackspace.svc.cluster.local/alert/process
            send_resolved: true
prometheus:
  enabled: true
  prometheusSpec:
    externalUrl: https://{{ (index .OpenCenter.Services "kube-prometheus-stack").Hostname | default (printf "prometheus.%s" .OpenCenter.Cluster.ClusterFQDN) }}
    externalLabels:
      cluster: k8s-dev
      region: {{ .OpenCenter.Meta.Region }}
      customer: {{ .OpenCenter.Meta.Organization }}
    serviceMonitorSelectorNilUsesHelmValues: false
    podMonitorSelectorNilUsesHelmValues: false
    ruleSelectorNilUsesHelmValues: false
grafana:
  enabled: true
  admin:
    existingSecret: "grafana-admin-password"
    userKey: admin-user
    passwordKey: admin-password
  persistence:
    enabled: true
    type: sts
    accessModes:
      - ReadWriteOnce
    size: 50Gi
    finalizers:
      - kubernetes.io/pvc-protection
  datasources:
    datasources.yaml:
      apiVersion: 1
      datasources:
        - name: Loki
          uid: loki-default
          type: loki
          access: proxy
          url: http://observability-loki-gateway.observability.svc.cluster.local
          isDefault: false
          jsonData:
            httpHeaderName1: "X-Scope-OrgID"
            maxLines: 1000
          secureJsonData:
            httpHeaderValue1: "default"
          editable: true
        - name: Tempo
          uid: tempo-default
          type: tempo
          access: proxy
          url: http://observability-tempo-query-frontend.observability.svc.cluster.local:3200
          isDefault: false
          jsonData:
            httpHeaderName1: x-scope-orgid
            maxLines: 1000
            pdcInjected: false
            tracesToLogsV2:
              customQuery: false
              datasourceUid: loki-default
              filterBySpanID: true
              filterByTraceID: true
            tracesToMetrics:
              datasourceUid: prometheus
          secureJsonData:
            httpHeaderValue1: "default"
          editable: true
`
