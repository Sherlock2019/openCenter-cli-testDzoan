terraform {
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

locals {
  cluster_name                  = "{{ .OpenCenter.Cluster.ClusterName }}"
  naming_prefix                 = "${local.cluster_name}-"
  openstack_auth_url            = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL | default "https://keystone.api.sjc3.rackspacecloud.com/v3/" }}"
  openstack_insecure            = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Insecure | default false }}
  openstack_region              = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region | default "SJC3" }}"
  availability_zone             = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.AvailabilityZone | default "az1" }}"
  openstack_user_name           = ""
  openstack_user_password       = ""
  application_credential_id     = var.os_application_credential_id
  application_credential_secret = var.os_application_credential_secret
  openstack_project_domain_name = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.ProjectDomainName | default "rackspace_cloud_domain" }}"
  openstack_user_domain_name    = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.UserDomainName | default "rackspace_cloud_domain" }}"
  openstack_tenant_name         = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.TenantName | default "f2823901-4194-40c7-9dc4-d56d2105e81a" }}"
  floatingip_pool               = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.FloatingIPPool | default "PUBLICNET" }}"
  router_external_network_id    = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.RouterExternalNetworkID | default "" }}"
  vlan_id                       = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.VLAN.ID | default "" }}"
  mtu                           = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.VLAN.MTU | default "" }}"
  network_provider              = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.VLAN.Provider | default "physnet1" }}"

  subnet_nodes          = "{{ .OpenCenter.Infrastructure.Networking.SubnetNodes | default "10.2.128.0/22" }}"
  subnet_nodes_oct      = join(".", slice(split(".", split("/", local.subnet_nodes)[0]), 0, 3))
  allocation_pool_start = "{{ .OpenCenter.Infrastructure.Networking.AllocationPoolStart | default "${local.subnet_nodes_oct}.50" }}"
  allocation_pool_end   = "{{ .OpenCenter.Infrastructure.Networking.AllocationPoolEnd | default "${local.subnet_nodes_oct}.254" }}"
  vrrp_ip               = "{{ .OpenCenter.Infrastructure.Networking.VRRPIP | default "${local.subnet_nodes_oct}.10" }}"
  subnet_pods           = "{{ .Deployment.Talos.Network.PodSubnet | default .OpenCenter.Cluster.Kubernetes.SubnetPods | default "10.42.0.0/16" }}"
  subnet_services       = "{{ .Deployment.Talos.Network.ServiceSubnet | default .OpenCenter.Cluster.Kubernetes.SubnetServices | default "10.43.0.0/16" }}"
  use_octavia           = {{ .OpenCenter.Infrastructure.Networking.UseOctavia | default false }}
  use_designate         = {{ .OpenCenter.Infrastructure.Networking.UseDesignate | default false }}
  dns_zone_name         = "{{ .OpenCenter.Infrastructure.Networking.DNSZoneName | default "" }}"
  dns_nameservers       = {{ if .OpenCenter.Infrastructure.Networking.DNSNameservers }}[{{ range $i, $dns := .OpenCenter.Infrastructure.Networking.DNSNameservers }}{{if $i}}, {{end}}"{{ $dns }}"{{ end }}]{{ else }}["1.1.1.1","8.8.8.8"]{{ end }}
  ntp_servers           = {{ if .OpenCenter.Infrastructure.Networking.NTPServers }}[{{ range $i, $ntp := .OpenCenter.Infrastructure.Networking.NTPServers }}{{if $i}}, {{end}}"{{ $ntp }}"{{ end }}]{{ else }}["time.cloudflare.com"]{{ end }}

  image_id                  = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.ImageID | default "" }}"
  image_id_windows          = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.ImageIDWindows | default "" }}"
  k8s_api_port              = {{ .OpenCenter.Cluster.Kubernetes.APIPort | default 6443 }}
  k8s_api_port_acl          = {{ if .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.K8sAPIPortACL }}[{{ range $i, $acl := .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.K8sAPIPortACL }}{{if $i}}, {{end}}"{{ $acl }}"{{ end }}]{{ else }}["0.0.0.0/0"]{{ end }}
  worker_count              = {{ .OpenCenter.Infrastructure.Compute.WorkerCount | default 4 }}
  worker_count_windows      = 0
  master_count              = {{ .OpenCenter.Infrastructure.Compute.MasterCount | default 3 }}
  ssh_user                  = "{{ .OpenCenter.Infrastructure.SSH.Username | default .OpenCenter.Infrastructure.SSH.User | default "ubuntu" }}"
  ssh_authorized_keys       = {{ if .OpenCenter.Infrastructure.SSH.AuthorizedKeys }}[{{ range $i, $key := .OpenCenter.Infrastructure.SSH.AuthorizedKeys }}{{if $i}}, {{end}}"{{ $key }}"{{ end }}]{{ else }}[]{{ end }}
  node_worker               = "wn"
  node_master               = "cp"
  node_worker_windows       = "win"
  ub_version                = "{{ .OpenCenter.Infrastructure.OSVersion | default "24" }}"
  flavor_bastion            = "{{ .OpenCenter.Infrastructure.Compute.FlavorBastion | default "" }}"
  flavor_master             = "{{ .OpenCenter.Infrastructure.Compute.FlavorMaster | default "" }}"
  flavor_worker             = "{{ .OpenCenter.Infrastructure.Compute.FlavorWorker | default "" }}"
  flavor_worker_windows     = ""
  worker_node_bfv_volume_size      = {{ .OpenCenter.Infrastructure.Storage.WorkerVolumeSize | default 100 }}
  worker_node_bfv_destination_type = "{{ .OpenCenter.Infrastructure.Storage.WorkerVolumeDestinationType | default "volume" }}"
  worker_node_bfv_source_type      = "{{ .OpenCenter.Infrastructure.Storage.WorkerVolumeSourceType | default "image" }}"
  worker_node_bfv_volume_type      = "{{ .OpenCenter.Infrastructure.Storage.WorkerVolumeType | default "HA-Performance" }}"
  wn_server_group_affinity         = {{ if .OpenCenter.Infrastructure.ServerGroupAffinity }}[{{ range $i, $affinity := .OpenCenter.Infrastructure.ServerGroupAffinity }}{{if $i}}, {{end}}"{{ $affinity }}"{{ end }}]{{ else }}["anti-affinity"]{{ end }}
  additional_block_devices_worker  = {{ if .OpenCenter.Infrastructure.Storage.AdditionalBlockDevices }}[{{ range $i, $device := .OpenCenter.Infrastructure.Storage.AdditionalBlockDevices }}{{if $i}}, {{end}}{{ $device }}{{ end }}]{{ else }}[]{{ end }}
  additional_server_pools_worker   = []
  ca_certificates                  = ""
  openstack_ca                     = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.CA | default "" }}"
  vrrp_enabled                     = "{{ .OpenCenter.Infrastructure.Networking.VRRPEnabled | default true }}"
  loadbalancer_provider            = "{{ .OpenCenter.Infrastructure.Networking.LoadbalancerProvider | default "ovn" }}"
  talos_api_port                   = {{ .Deployment.Talos.Network.TalosAPIPort | default 50000 }}
  talos_install_disk               = "{{ .Deployment.Talos.Install.Disk | default "/dev/sda" }}"
  talos_endpoint                   = {{ if .Deployment.Talos.Endpoint }}"{{ .Deployment.Talos.Endpoint }}"{{ else }}"https://${module.openstack-nova.k8s_api_ip}:6443"{{ end }}
}

module "openstack-nova" {
  source = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Modules.OpenstackNova.Source | default "github.com/opencenter-cloud/openCenter-gitops-base.git//iac/cloud/openstack/openstack-nova?ref=main" }}"

  availability_zone                    = local.availability_zone
  additional_block_devices_worker      = local.additional_block_devices_worker
  additional_server_pools_worker       = local.additional_server_pools_worker
  application_credential_id            = local.application_credential_id
  application_credential_secret        = local.application_credential_secret
  ca_certificates                      = local.ca_certificates
  use_octavia                          = local.use_octavia
  use_designate                        = local.use_designate
  dns_nameservers                      = local.dns_nameservers
  dns_zone_name                        = local.dns_zone_name
  flavor_bastion                       = local.flavor_bastion
  openstack_auth_url                   = local.openstack_auth_url
  openstack_ca                         = local.openstack_ca
  openstack_insecure                   = local.openstack_insecure
  openstack_region                     = local.openstack_region
  openstack_tenant_name                = local.openstack_tenant_name
  openstack_user_name                  = local.openstack_user_name
  openstack_password                   = local.openstack_user_password
  openstack_project_domain_name        = local.openstack_project_domain_name
  openstack_user_domain_name           = local.openstack_user_domain_name
  naming_prefix                        = local.naming_prefix
  ntp_servers                          = local.ntp_servers
  ssh_user                             = local.ssh_user
  floatingip_pool                      = local.floatingip_pool
  image_id                             = local.image_id
  image_id_windows                     = local.image_id_windows
  router_external_network_id           = local.router_external_network_id
  network_id                           = "{{ .OpenCenter.Infrastructure.Cloud.OpenStack.Networking.NetworkID | default "" }}"
  vlan_id                              = local.vlan_id
  vrrp_enabled                         = local.vrrp_enabled
  vrrp_ip                              = local.vrrp_ip
  ssh_authorized_keys                  = local.ssh_authorized_keys
  subnet_nodes                         = local.subnet_nodes
  subnet_services                      = local.subnet_services
  subnet_pods                          = local.subnet_pods
  allocation_pool_start                = local.allocation_pool_start
  allocation_pool_end                  = local.allocation_pool_end
  k8s_api_port                         = local.k8s_api_port
  k8s_api_port_acl                     = local.k8s_api_port_acl
  size_master = {
    count  = local.master_count
    flavor = local.flavor_master
  }
  size_worker = {
    count  = local.worker_count
    flavor = local.flavor_worker
  }
  node_master                          = local.node_master
  node_worker                          = local.node_worker
  node_worker_windows                  = local.node_worker_windows
  ub_version                           = local.ub_version
  worker_node_bfv_volume_size          = local.worker_node_bfv_volume_size
  worker_node_bfv_destination_type     = local.worker_node_bfv_destination_type
  worker_node_bfv_source_type          = local.worker_node_bfv_source_type
  worker_node_bfv_volume_type          = local.worker_node_bfv_volume_type
  wn_server_group_affinity             = local.wn_server_group_affinity
}

resource "local_file" "talos_inventory" {
  filename = "${path.module}/talos/inventory.yaml"
  content = <<-YAML
cluster:
  name: ${local.cluster_name}
  endpoint: ${local.talos_endpoint}
  talos_api_port: {{ .Deployment.Talos.Network.TalosAPIPort | default 50000 }}

control_plane:
%{ for node in module.openstack-nova.master_nodes ~}
  - name: ${node.name}
    talos_api_ip: ${node.access_ip_v4}
    internal_ip: ${node.access_ip_v4}
    install_disk: ${local.talos_install_disk}
    cert_sans:
      - ${module.openstack-nova.k8s_api_ip}
      - ${node.access_ip_v4}
%{ endfor ~}

workers:
%{ for node in module.openstack-nova.worker_nodes ~}
  - name: ${node.name}
    talos_api_ip: ${node.access_ip_v4}
    internal_ip: ${node.access_ip_v4}
    install_disk: ${local.talos_install_disk}
    labels:
      node-role.kubernetes.io/worker: ""
%{ endfor ~}

patch_inputs:
  dns_servers:
%{ for dns in local.dns_nameservers ~}
    - ${dns}
%{ endfor ~}
  ntp_servers:
%{ for ntp in local.ntp_servers ~}
    - ${ntp}
%{ endfor ~}
  pod_subnet: ${local.subnet_pods}
  service_subnet: ${local.subnet_services}
YAML
}
