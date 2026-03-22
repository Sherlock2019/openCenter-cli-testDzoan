Feature:Configuration-driventemplaterendering
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/gitops-repo"
  @template @alert-proxy @secrets
Scenario:Renderalert-proxysecretswithcustomvalues
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
secrets:
alert_proxy:
core_device_id: "device-123"
account_service_token: "token-456"
core_account_number: "account-789"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @alert-proxy @configuration
Scenario:Renderalert-proxyconfigurationwithcustomimagetag
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
managed-service:
alert-proxy:
enabled:true
image_tag: "v1.2.3"
alert_manager_base_url: "https://alertmanager.example.com"
http_route_fqdn: "alerts.example.com"
secrets:
alert_proxy:
core_device_id: "device-123"
account_service_token: "token-456"
core_account_number: "account-789"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @cert-manager @secrets
Scenario:Rendercert-managerwithcustomAWScredentials
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
admin_email:admin@example.com
cluster_fqdn:test-cluster.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
cert-manager:
enabled:true
region:us-west-2
letsencrypt_server: "https://acme-staging-v02.api.letsencrypt.org/directory"
secrets:
cert_manager:
aws_access_key: "AKIAEXAMPLE123"
aws_secret_access_key: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @cert-manager @letsencrypt
Scenario:Rendercert-managerwithLetsEncryptconfiguration
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
admin_email:devops@example.com
cluster_fqdn:prod.example.com
base_domain:example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
cert-manager:
enabled:true
region:eu-west-1
letsencrypt_server: "https://acme-v02.api.letsencrypt.org/directory"
secrets:
cert_manager:
aws_access_key: "AKIAEXAMPLE456"
aws_secret_access_key: "secretkey789"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @loki @swift
Scenario:RenderLokiwithcustomSwiftcredentials
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
services:
loki:
enabled:true
loki_bucket_name: "my-loki-bucket"
loki_volume_size:50
loki_storage_class: "fast-ssd"
swift_auth_url: "https://keystone.example.com/v3/"
swift_username: "loki-user"
swift_project_name: "my-project"
swift_region: "US-EAST-1"
swift_domain_name: "default"
secrets:
loki:
swift_password: "my-secure-password"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @loki @volume
Scenario:RenderLokiwithvolumeconfiguration
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
storage:
default_storage_class: "csi-cinder-sc-delete"
services:
loki:
enabled:true
loki_volume_size:100
swift_auth_url: "https://keystone.example.com/v3/"
swift_username: "loki"
swift_project_name: "project"
swift_region: "REGION"
swift_domain_name: "default"
secrets:
loki:
swift_password: "password"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @velero @backup
Scenario:RenderVelerowithcustombackupbucket
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
services:
velero:
enabled:true
velero_backup_bucket: "my-backup-bucket"
velero_region: "us-west-1"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @keycloak @oidc
Scenario:RenderKeycloakwithcustomOIDCconfiguration
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
admin_email:admin@example.com
cluster_fqdn:test.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
keycloak:
enabled:true
keycloak_realm: "my-realm"
keycloak_frontend_url: "https://auth.example.com"
keycloak_client_id: "my-client"
hostname: "auth.example.com"
secrets:
keycloak:
client_secret: "secret123"
admin_password: "admin123"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @headlamp @oidc
Scenario:RenderHeadlampwithOIDCintegration
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
cluster_fqdn:test.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
headlamp:
enabled:true
headlamp_oidc_client_id: "headlamp-client"
headlamp_oidc_issuer_url: "https://auth.example.com/realms/my-realm"
hostname: "headlamp.example.com"
secrets:
headlamp:
oidc_client_secret: "headlamp-secret"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @weave-gitops @password
Scenario:RenderWeaveGitOpswithcustompassword
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
cluster_fqdn:test.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
weave-gitops:
enabled:true
hostname: "gitops.example.com"
secrets:
weave_gitops:
password: "mypassword"
password_hash: "$2a$10$abcdefghijklmnopqrstuv"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @grafana @storage
Scenario:RenderGrafanawithcustomstorageconfiguration
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
storage:
default_storage_class: "csi-cinder-sc-delete"
services:
kube-prometheus-stack:
enabled:true
grafana_volume_size:20
grafana_storage_class: "fast-ssd"
prometheus_volume_size:100
prometheus_storage_class: "fast-ssd"
secrets:
grafana:
admin_password: "grafana-admin-pass"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @httproute @hostname
Scenario:HTTPRoutehostnamegenerationfromclusterFQDN
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
cluster_fqdn:prod.k8s.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
keycloak:
enabled:true
headlamp:
enabled:true
weave-gitops:
enabled:true
secrets:
keycloak:
client_secret: "secret"
admin_password: "password"
headlamp:
oidc_client_secret: "secret"
weave_gitops:
password: "password"
password_hash: "hash"
grafana:
admin_password: "password"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @httproute @custom-hostname
Scenario:HTTPRoutewithcustomhostnameoverrides
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
cluster_fqdn:test.example.com
gitops:
git_dir: <<tmp>>/gitops-repo
services:
keycloak:
enabled:true
hostname: "custom-auth.example.com"
headlamp:
enabled:true
hostname: "custom-ui.example.com"
secrets:
keycloak:
client_secret: "secret"
admin_password: "password"
headlamp:
oidc_client_secret: "secret"
grafana:
admin_password: "password"
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @template @secrets @validation @wip
Scenario:Missingrequiredsecretsshouldfailvalidation
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
services:
cert-manager:
enabled:true
  """
WhenIrun "opencenterclustervalidatetest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "secret"
  @template @defaults @fallback
Scenario:Templaterenderingwithdefaultvalues
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
  """
WhenIrun "opencenterclusterselecttest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @integration @full-rendering
Scenario:Fullclusterrenderingwithallnewfields
Givenafile "<<tmp>>/conf/test-integration.yaml"withcontent:
  """
schema_version: "2.0"
opencenter:
meta:
name:test-integration
env:dev
region:us-east-1
organization:test-org
cluster:
cluster_name:test-integration
base_domain:k8s.test.example.com
cluster_fqdn:test-integration.us-east-1.k8s.test.example.com
admin_email:admin@test.example.com
storage:
default_storage_class:csi-cinder-sc-delete
gitops:
git_dir: <<tmp>>/gitops-repo
gitops_base_repo:ssh://git@github.com/opencenter-cloud/opencenter-gitops-base.git
gitops_base_release:v0.2.0
gitops_branch:main
services:
cert-manager:
enabled:true
region:us-east-1
letsencrypt_server:https://acme-staging-v02.api.letsencrypt.org/directory
loki:
enabled:true
loki_bucket_name:test-integration-loki
loki_volume_size:50
loki_storage_class:csi-cinder-sc-delete
swift_auth_url:https://keystone.api.test.example.com/v3/
swift_username:loki-user
swift_project_name:test-project
swift_region:US-EAST-1
swift_domain_name:default
velero:
enabled:true
velero_backup_bucket:test-integration-backups
velero_region:us-east-1
keycloak:
enabled:true
keycloak_realm:opencenter
keycloak_client_id:opencenter
hostname:auth.test-integration.us-east-1.k8s.test.example.com
headlamp:
enabled:true
headlamp_oidc_client_id:opencenter
hostname:headlamp.test-integration.us-east-1.k8s.test.example.com
weave-gitops:
enabled:true
hostname:gitops.test-integration.us-east-1.k8s.test.example.com
kube-prometheus-stack:
enabled:true
grafana_volume_size:20
prometheus_volume_size:100
alertmanager_volume_size:10
managed-service:
alert-proxy:
enabled:true
alert_manager_base_url:https://alertmanager.test-integration.us-east-1.k8s.test.example.com
http_route_fqdn:alerts.test-integration.us-east-1.k8s.test.example.com
image_tag:v1.2.3
secrets:
cert_manager:
aws_access_key:AKIATEST123456789ABC
aws_secret_access_key:wJalrXUtnFEMI/K7MDENG/bPxRfiCYTESTKEY123
loki:
swift_password:test-swift-password-secure-123
keycloak:
client_secret:f8V0we25ajxjm9OMpFz9BsYObGTYKM4Y
admin_password:SecureKeycloakAdminPassword123!
headlamp:
oidc_client_secret:headlamp-oidc-secret-abc123xyz
weave_gitops:
password:WeaveGitOpsPassword123!
password_hash: $2a$10$abcdefghijklmnopqrstuvwxyz1234567890ABCDEFGHIJKLMNOP
grafana:
admin_password:GrafanaAdminPassword123!
alert_proxy:
core_device_id:device-test-integration-12345
account_service_token:token-test-integration-67890
core_account_number:account-test-integration-11111
  """
WhenIrun "opencenterclusterselecttest-integration --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclustersetup --skip-validation --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
Andthedirectory "<<tmp>>/gitops-repo"shouldexist
  @validation @missing-secrets @priority4
Scenario:Missingcert-managersecretsshouldfailvalidation
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
services:
cert-manager:
enabled:true
  """
WhenIrun "opencenterclustervalidatetest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
  @validation @missing-secrets @priority4
Scenario:Missinglokisecretsshouldfailvalidation
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
gitops:
git_dir: <<tmp>>/gitops-repo
services:
loki:
enabled:true
swift_auth_url:https://keystone.example.com/v3/
swift_username:loki
swift_project_name:project
swift_region:REGION
swift_domain_name:default
secrets:
cert_manager:
aws_access_key:test
aws_secret_access_key:test
keycloak:
admin_password:test
grafana:
admin_password:test
weave_gitops:
password_hash:test
  """
WhenIrun "opencenterclustervalidatetest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
  @validation @invalid-email @priority4
Scenario:Invalidadminemailshouldfailvalidation
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
admin_email:invalid-email-format
gitops:
git_dir: <<tmp>>/gitops-repo
secrets:
cert_manager:
aws_access_key:test
aws_secret_access_key:test
keycloak:
admin_password:test
grafana:
admin_password:test
weave_gitops:
password_hash:test
  """
WhenIrun "opencenterclustervalidatetest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
  @validation @invalid-domain @priority4
Scenario:InvalidclusterFQDNshouldfailvalidation
Givenafile "<<tmp>>/conf/test-cluster.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:test-cluster
cluster_fqdn:invaliddomainwithspaces
gitops:
git_dir: <<tmp>>/gitops-repo
secrets:
cert_manager:
aws_access_key:test
aws_secret_access_key:test
keycloak:
admin_password:test
grafana:
admin_password:test
weave_gitops:
password_hash:test
  """
WhenIrun "opencenterclustervalidatetest-cluster --config-dir <<tmp>>/conf"
Thentheexitcodeshouldnotbe0
  @validation @valid-config
Scenario:Validconfigurationshouldpassvalidation
WhenIrun "opencenterclusterinittest-cluster --config-dir <<tmp>>/conf --no-keygen"
Thentheexitcodeshouldbe0
WhenIrun "opencenterclusterinfotest-cluster --validate --config-dir <<tmp>>/conf"
Thentheexitcodeshouldbe0
