Feature:Configurationvalidationrules
Background:
Givenanemptydirectory "<<tmp>>/conf"
Andanemptydirectory "<<tmp>>/repo-bad"
  @validation @missing_git_dir
Scenario:missingopencenter.gitops.git_dir ->error
Givenafile "<<tmp>>/conf/mgd.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:mgd
gitops:
git_dir: ""
  """
WhenIrun "opencenterclusterinfomgd --validate"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "Config.OpenCenter.GitOps.GitDir:fieldisrequired"
  @validation @opentofu_s3_requires_creds @wip
Scenario:OpenTofuS3backendrequirescredentials ->errorthenpass
Givenafile "<<tmp>>/conf/s3.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:s3
gitops:
git_dir: "<<tmp>>/repo-bad"
opentofu:
enabled:true
backend:
type:s3
s3:
bucket:b
key:k
region:us-east-1
  """
WhenIrun "opencenterclusterinfos3 --validate"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "opencenter.cluster.aws_access_key"
Andstderrshouldcontain "opencenter.cluster.aws_secret_access_key"
  @validation @s3_with_creds_ok
Scenario:OpenTofuS3backendwithcredentials ->ok
Givenafile "<<tmp>>/conf/s3ok.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:s3ok
aws_access_key:AKIA...
aws_secret_access_key:secret
gitops:
git_dir: "<<tmp>>/repo-bad"
opentofu:
enabled:true
backend:
type:s3
s3:
bucket:b
key:k
region:us-east-1
  """
WhenIrun "opencenterclusterinfos3ok --validate"
Thentheexitcodeshouldbe0
  @validation @prosys_cluster_validation
Scenario:prosys-dev-dfw3clusterconfigurationvalidation
Givenafile "<<tmp>>/conf/prosys-dev-dfw3.yaml"withcontent:
  """
schema_version: "2.0"
opencenter:
meta:
name:prosys-dev-dfw3
organization:opencenter
env:dev
region:dfw3
cluster:
cluster_name:prosys-dev-dfw3
base_domain:dev.attcontroller.com
cluster_fqdn:prosys-dev-dfw3.dev.attcontroller.com
admin_email:ops@example.com
kubernetes:
version: "1.32.8"
api_port:6443
subnet_pods: "10.42.0.0/16"
subnet_services: "10.43.0.0/16"
network_plugin:
calico:
enabled:true
infrastructure:
provider:openstack
os_version: "24"
ssh:
authorized_keys:
  - "ssh-ed25519AAAAC3NzaC1lZDI1NTE5AAAAIExamplePublicKeyDataHereuser@example.com"
networking:
subnet_nodes: "10.0.4.0/22"
allocation_pool_start: "10.0.4.50"
allocation_pool_end: "10.0.7.254"
vrrp_enabled:true
vrrp_ip: "10.0.4.10"
loadbalancer_provider:octavia
dns_zone_name:dev.attcontroller.com
dns_nameservers:
  - "1.1.1.1"
  - "8.8.8.8"
ntp_servers:
  - "time.dfw3.rackspace.com"
  - "time2.dfw3.rackspace.com"
compute:
flavor_bastion:gp.5.2.2
flavor_master:gp.5.4.8
flavor_worker:gp.5.4.8
master_count:3
worker_count:4
storage:
default_storage_class:csi-cinder-sc-delete
worker_volume_size:100
worker_volume_destination_type:volume
worker_volume_source_type:image
worker_volume_type:Performance
cloud:
openstack:
auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
region:dfw3
project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
image_id: "ec458631-309a-4b7d-846c-cd2ccc601137"
network_id: "12345678-1234-1234-1234-123456789012"
availability_zones:
  -az1
gitops:
git_url: "ssh://git@example.com/opencenter/prosys-dev-dfw3.git"
git_branch:main
flux_interval: "15m"
flux_prune:true
deployment:
method:kubespray
opentofu:
backend:
type:local
local:
path:terraform.tfstate
secrets:
global:
openstack_auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
openstack_project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
sops:
enabled:true
age_key_file: <<tmp>>/sops/age/keys/prosys-dev-dfw3-key.txt
  """
WhenIrun "opencenterclustervalidateprosys-dev-dfw3"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Validationsuccessful"
  @validation @prosys_cluster_debug_config
Scenario:prosys-dev-dfw3clusterdebugconfiggeneration
Givenafile "<<tmp>>/conf/prosys-dev-dfw3.yaml"withcontent:
  """
schema_version: "2.0"
opencenter:
meta:
name:prosys-dev-dfw3
organization:opencenter
env:dev
region:dfw3
cluster:
cluster_name:prosys-dev-dfw3
base_domain:dev.attcontroller.com
cluster_fqdn:prosys-dev-dfw3.dev.attcontroller.com
admin_email:ops@example.com
kubernetes:
version: "1.32.8"
api_port:6443
subnet_pods: "10.42.0.0/16"
subnet_services: "10.43.0.0/16"
network_plugin:
calico:
enabled:true
infrastructure:
provider:openstack
os_version: "24"
ssh:
authorized_keys:
  - "ssh-ed25519AAAAC3NzaC1lZDI1NTE5AAAAIExamplePublicKeyDataHereuser@example.com"
networking:
subnet_nodes: "10.0.4.0/22"
allocation_pool_start: "10.0.4.50"
allocation_pool_end: "10.0.7.254"
vrrp_enabled:true
vrrp_ip: "10.0.4.10"
loadbalancer_provider:octavia
dns_zone_name:dev.attcontroller.com
dns_nameservers:
  - "1.1.1.1"
  - "8.8.8.8"
ntp_servers:
  - "time.dfw3.rackspace.com"
  - "time2.dfw3.rackspace.com"
compute:
flavor_bastion:gp.5.2.2
flavor_master:gp.5.4.8
flavor_worker:gp.5.4.8
master_count:3
worker_count:4
storage:
default_storage_class:csi-cinder-sc-delete
worker_volume_size:100
worker_volume_destination_type:volume
worker_volume_source_type:image
worker_volume_type:Performance
cloud:
openstack:
auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
region:dfw3
project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
image_id: "ec458631-309a-4b7d-846c-cd2ccc601137"
network_id: "12345678-1234-1234-1234-123456789012"
availability_zones:
  -az1
gitops:
git_url: "ssh://git@example.com/opencenter/prosys-dev-dfw3.git"
git_branch:main
flux_interval: "15m"
flux_prune:true
deployment:
method:kubespray
opentofu:
backend:
type:local
local:
path:terraform.tfstate
secrets:
global:
openstack_auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
openstack_project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
sops:
enabled:true
age_key_file: <<tmp>>/sops/age/keys/prosys-dev-dfw3-key.txt
  """
WhenIrun "opencenterclustervalidateprosys-dev-dfw3 --generate-debug-config --output-dir <<tmp>>"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Debugconfigsavedto"
Andstdoutshouldcontain "Validationsuccessful"
Andafile "<<tmp>>/.opencenter-v2.yaml"shouldexist
  @validation @prosys_cluster_vrrp_validation
Scenario:prosys-dev-dfw3clusterVRRPvalidationwithnetworkingsection
Givenafile "<<tmp>>/conf/prosys-dev-dfw3.yaml"withcontent:
  """
schema_version: "2.0"
opencenter:
meta:
name:prosys-dev-dfw3
organization:opencenter
env:dev
region:dfw3
cluster:
cluster_name:prosys-dev-dfw3
base_domain:dev.attcontroller.com
cluster_fqdn:prosys-dev-dfw3.dev.attcontroller.com
admin_email:ops@example.com
kubernetes:
version: "1.32.8"
api_port:6443
subnet_pods: "10.42.0.0/16"
subnet_services: "10.43.0.0/16"
network_plugin:
calico:
enabled:true
infrastructure:
provider:openstack
os_version: "24"
ssh:
authorized_keys:
  - "ssh-ed25519AAAAC3NzaC1lZDI1NTE5AAAAIExamplePublicKeyDataHereuser@example.com"
networking:
subnet_nodes: "10.0.4.0/22"
allocation_pool_start: "10.0.4.50"
allocation_pool_end: "10.0.7.254"
vrrp_enabled:true
vrrp_ip: "10.0.4.10"
loadbalancer_provider:octavia
dns_zone_name:dev.attcontroller.com
dns_nameservers:
  - "1.1.1.1"
  - "8.8.8.8"
ntp_servers:
  - "time.dfw3.rackspace.com"
  - "time2.dfw3.rackspace.com"
compute:
flavor_bastion:gp.5.2.2
flavor_master:gp.5.4.8
flavor_worker:gp.5.4.8
master_count:3
worker_count:4
storage:
default_storage_class:csi-cinder-sc-delete
worker_volume_size:100
worker_volume_destination_type:volume
worker_volume_source_type:image
worker_volume_type:Performance
cloud:
openstack:
auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
region:dfw3
project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
image_id: "ec458631-309a-4b7d-846c-cd2ccc601137"
network_id: "12345678-1234-1234-1234-123456789012"
availability_zones:
  -az1
gitops:
git_url: "ssh://git@example.com/opencenter/prosys-dev-dfw3.git"
git_branch:main
flux_interval: "15m"
flux_prune:true
deployment:
method:kubespray
opentofu:
backend:
type:local
local:
path:terraform.tfstate
secrets:
global:
openstack_auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
openstack_project_id: "33d34083-ef71-464f-9d09-4b545f64baaf"
sops:
enabled:true
age_key_file: <<tmp>>/sops/age/keys/prosys-dev-dfw3-key.txt
  """
WhenIrun "opencenterclustervalidateprosys-dev-dfw3"
Thentheexitcodeshouldbe0
Andstdoutshouldcontain "Validationsuccessful"
  @validation @prosys_cluster_vrrp_missing_ip @priority4 @wip
Scenario:prosys-dev-dfw3clusterVRRPvalidationfailswhenIPmissing
Givenafile "<<tmp>>/conf/prosys-dev-dfw3.yaml"withcontent:
  """
opencenter:
cluster:
cluster_name:prosys-dev-dfw3
domain:dev.attcontroller.com
gitops:
git_dir: <<tmp>>/prosys-gitops-repo
infrastructure:
cloud:
openstack:
application_credential_id: "12345678-1234-1234-1234-123456789012"
application_credential_secret: "test-app-cred-secret"
auth_url: "https://keystone.api.dfw3.rackspacecloud.com/v3/"
region: "DFW3"
domain: "Default"
networking:
floating_network_id: "12345678-1234-1234-1234-123456789012"
provider:openstack
opentofu:
enabled:true
backend:
type:local
local:
path:terraform.tfstate
secrets:
sops_age_key_file: <<tmp>>/sops/age/keys/prosys-dev-dfw3-key.txt
global:
openstack:
application_credential_id: "12345678-1234-1234-1234-123456789012"
application_credential_secret: "test-app-cred-secret"
networking:
use_octavia:false
vrrp_enabled:true
vrrp_ip: ""
  """
WhenIrun "opencenterclustervalidateprosys-dev-dfw3"
Thentheexitcodeshouldnotbe0
Andstderrshouldcontain "vrrp_ipmustbesetwhenuse_octaviaisfalse"
Andstderrshouldcontain "opencenter.infrastructure.cloud.openstack.regionmustbesetwhenproviderisopenstack"
Andstderrshouldcontain "opencenter.secrets.barbican.auth_urlmustbesetwhensecretsbackendisbarbican"
