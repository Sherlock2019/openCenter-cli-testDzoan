locals {
  # this will be the user's name and the DNS zone prefix
  cluster_name                            = "{{ .OpenCenter.Cluster.ClusterName }}"
  # Prefix to add to Openstack resource names
  naming_prefix                           = "${local.cluster_name}-"
  openstack_auth_url                      = "{{ .OpenCenter.Cloud.OpenStack.AuthURL | default "https://keystone.api.dfw3.rackspacecloud.com/v3/" }}"
  openstack_insecure                      = {{ .OpenCenter.Cloud.OpenStack.Insecure | default false }}
  openstack_region                        = "{{ .OpenCenter.Cloud.OpenStack.Region | default "DFW3" }}"
  availability_zone                       = "az1"
  openstack_user_name                     = ""
  openstack_user_password                 = ""
  application_credential_id               = var.os_application_credential_id
  application_credential_secret           = var.os_application_credential_secret
  openstack_project_domain_name           = "rackspace_cloud_domain"
  openstack_user_domain_name              = "rackspace_cloud_domain"
  openstack_tenant_name                   = "33d34083-ef71-464f-9d09-4b545f64baaf"
  floatingip_pool                         = "PUBLICNET"
  router_external_network_id              = "82be3711-cd97-4f7c-8bbd-59f5524a949e"
  # VLAN settings
  vlan_id                                 = ""
  mtu                                     = ""
  network_provider                        = "physnet1"
  #CIDR that the openstack VMs will use for K8s nodes - using default since not in new schema
  subnet_nodes                            = "10.0.4.0/22"
  subnet_nodes_oct                       = join(".", slice(split(".", split("/", local.subnet_nodes)[0]), 0, 3))
  #Leave some IPs free for the VRRP IP and the MetalLB Range
  allocation_pool_start                   = "${local.subnet_nodes_oct}.50"
  allocation_pool_end                     = "10.0.7.254"
  # vrrp_ip Must be an IP from subnet_nodes and will be used as the internal Kubernetes API VIP.
  vrrp_ip                                 = "${local.subnet_nodes_oct}.10"
  #CIDR that will be used by kubernetes pods. Not an openstack network.
  subnet_pods                             = "{{ .OpenCenter.Cluster.Kubernetes.SubnetPods | default "10.42.0.0/16" }}"
  #CIDR that will be used for kubernetes services. Not an openstack network.
  subnet_services                         = "{{ .OpenCenter.Cluster.Kubernetes.SubnetServices | default "10.43.0.0/16" }}"
  # use_octavia set to false to create a floating IP associated with the vrrp_ip port. true will create an octavia LB with a floating IP
  use_octavia                             = false
  loadbalancer_provider                   = "{{ .OpenCenter.Cluster.Kubernetes.LoadbalancerProvider | default "amphora" }}"
  # vrrp_enabled cannot be set to true if use_octavia is true
  vrrp_enabled                            = true
  # Creates a DNS record using the LB floating IP and dns_zone_name
  use_designate                           = false
  # dns_zone_name is the dns zone to create if use_designate is true
  dns_zone_name                           = "{{ .OpenCenter.Cluster.Kubernetes.DNSZoneName | default "dev.attcontroller.com" }}"
  # DNS servers to configure on the nodes
  dns_nameservers                         = ["1.1.1.1","8.8.8.8"]
  ntp_servers                             = ["time.dfw3.rackspace.com","time2.dfw3.rackspace.com"]
  image_id                                = "ec458631-309a-4b7d-846c-cd2ccc601137"
  image_id_windows                        = ""
  k8s_api_port                            = 443
  k8s_api_port_acl                        = {{ if .OpenCenter.Cluster.K8sAPIPortACL }}[{{ range $i, $acl := .OpenCenter.Cluster.K8sAPIPortACL }}{{if $i}}, {{end}}"{{ $acl }}"{{ end }}]{{ else }}["146.20.2.10/32","172.99.99.10/32","134.213.179.10/32","161.47.0.10/32","134.213.178.10/32","119.9.122.10/32","119.9.148.10/32","63.131.145.180/32","78.136.22.232/32"]{{ end }}
  worker_count                            = {{ .OpenCenter.Cluster.Kubernetes.WorkerCount | default 4 }}
  worker_count_windows                    = {{ .OpenCenter.Cluster.Kubernetes.WorkerCountWindows | default 0 }}
  # Enter 1 or 3 masters.
  master_count                            = {{ .OpenCenter.Cluster.Kubernetes.MasterCount | default 3 }}
  ssh_user                                = "ubuntu"
  # these are the ssh public keys that will be able to connect to the cluster's bastion node
  ssh_authorized_keys                     = {{ if .OpenCenter.Cluster.SSHAuthorizedKeys }}[{{ range $i, $key := .OpenCenter.Cluster.SSHAuthorizedKeys }}{{if $i}}, {{end}}"{{ $key }}"{{ end }}]{{ else }}["ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDogzEullM89m//Vd8IGPERto2DotXnUCKGH6II1Vk/klEuDVqXx9kCb981XJKh8mU15bfJVdE4h078q/shK9EIcPMRKSQSMs2LkgF/1yUeVYPNYiIBph6CaqjIxKHy1kYxw3KUTIh8IIl1M4t5fc5c49Gr3QuDpeMN4Z/wrbR1DceIbFDiVxYNeyJWfOdowKgTn4AKh0n1xtg6/XLin3cCstpvfUJUKm0WOcmn3+DHK6cBNqNAMKdtxgnGwlY4MfizJOZE30Y7hwPqXUjOgLgB2vybcdcMpUvw9e8HopogOFQnVwwmlc9/7ZKPCaCKRBEC38IV82CJ6+/eePIMriPF migu4903@MNF0TUDV30"]{{ end }}
  node_worker                             = "wn"
  node_master                             = "cp"
  node_worker_windows                     = "win"
  ub_version                              = "24"
  #FLEX Flavor Settings ==========================
  flavor_bastion                          = "{{ .OpenCenter.Cluster.Kubernetes.FlavorBastion | default "gp.5.2.2" }}"
  flavor_master                           = "{{ .OpenCenter.Cluster.Kubernetes.FlavorMaster | default "gp.5.4.4" }}"
  flavor_worker                           = "{{ .OpenCenter.Cluster.Kubernetes.FlavorWorker | default "gp.5.4.8" }}"

  worker_node_bfv_volume_size             = {{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WorkerNodeBFVSizeWindows | default 100 }}
  worker_node_bfv_destination_type        = "volume"
  worker_node_bfv_source_type             = "image"
  worker_node_bfv_volume_type             = "{{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WorkerNodeBFVTypeWindows | default "Performance" }}"

  # ====================================
  #ca_certificates add CA certificates to server's trusts. Good for trusting internal private Certificate Authorities.
  ca_certificates                         = ""
  openstack_ca                            = ""

  # ====================================
  #Kubespray Settings
  kubespray_version                       = "v2.28.1"
  kubernetes_version                      = "{{ .OpenCenter.Cluster.Kubernetes.Version | default "1.32.8" }}"
  network_plugin                          = "{{ if .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Enabled }}calico{{ else if .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Cilium.Enabled }}cilium{{ else if .OpenCenter.Cluster.Kubernetes.NetworkPlugin.KubeOVN.Enabled }}kube-ovn{{ else }}calico{{ end }}"
  deploy_cluster                          = true
  #kub-vip settings
  kube_vip_enabled                        = true
  #Hardening
  k8s_hardening_enabled                   = true
  kube_pod_security_exemptions_namespaces = ["trivy-temp"]
  kubelet_rotate_server_certificates      = true
  os_hardening_enabled                    = true

  #OIDC Settings
  kube_oidc_auth_enabled                 = {{ .OpenCenter.Cluster.Kubernetes.OIDC.Enabled | default false }}
  kube_oidc_url                          = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCURL | default "https://auth.prosys.dev.dfw3.k8s.opencenter.cloud/realms/opencenter" }}"
  kube_oidc_client_id                    = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCClientID | default "kubernetes" }}"
  # # Optional settings fo OIDC
  kube_oidc_ca_file                      = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCCAFile }}"
  kube_oidc_username_claim               = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCUsernameClaim | default "sub" }}"
  kube_oidc_username_prefix              = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCUsernamePrefix | default "oidc:" }}"
  kube_oidc_groups_claim                 = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCGroupsClaim | default "groups" }}"
  kube_oidc_groups_prefix                = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCGroupsPrefix | default "oidc:" }}"

  #Calico Settings
  cni_iface                               = "{{ .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.CNIIface | default "enp3s0" }}"
  #Interface detection method for Calico nodeAddressAutodetectionV4. Can be "first-found", "interface", "cidr"
  #https://docs.tigera.io/calico/latest/reference/installation/api#operator.tigera.io%2fv1.NodeAddressAutodetection
  calico_interface_autodetect             = "{{ .OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.CalicoInterfaceAutodetect | default "interface" }}"
  calico_interface_autodetect_cidr        = ""
  calico_encapsulation_type               = "VXLAN"
  calico_nat_outgoing                     = true

  # ## Windows settings
  windows_user                            = "{{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WindowsUser | default "Administrator" }}"
  windows_admin_password                  = "{{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WindowsAdminPassword }}"
  worker_node_bfv_size_windows            = {{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WorkerNodeBFVSizeWindows | default 0 }}
  worker_node_bfv_type_windows            = "{{ .OpenCenter.Cluster.Kubernetes.WindowsWorkers.WorkerNodeBFVTypeWindows | default "local" }}"
}

module "openstack-nova" {
  source = "github.com/rackerlabs/openCenter-gitops-base.git//iac/cloud/openstack/openstack-nova?ref=main"
  availability_zone             = local.availability_zone
  application_credential_id     = local.application_credential_id
  application_credential_secret = local.application_credential_secret
  ca_certificates               = local.ca_certificates
  use_octavia                   = local.use_octavia
  use_designate                 = local.use_designate
  dns_nameservers               = local.dns_nameservers
  dns_zone_name                 = local.dns_zone_name
  flavor_bastion                = local.flavor_bastion
  openstack_auth_url            = local.openstack_auth_url
  openstack_ca                  = local.openstack_ca
  openstack_insecure            = local.openstack_insecure
  openstack_region              = local.openstack_region
  openstack_tenant_name         = local.openstack_tenant_name
  openstack_user_name           = local.openstack_user_name
  openstack_password            = local.openstack_user_password
  openstack_project_domain_name = local.openstack_project_domain_name
  openstack_user_domain_name    = local.openstack_user_domain_name
  naming_prefix                 = local.naming_prefix
  ntp_servers                   = local.ntp_servers
  ssh_user                      = local.ssh_user
  floatingip_pool               = local.floatingip_pool
  image_id                      = local.image_id
  image_id_windows              = local.image_id_windows
  router_external_network_id    = local.router_external_network_id
  network_id                    = ""
  vlan_id                       = local.vlan_id
  vrrp_enabled                  = local.vrrp_enabled
  vrrp_ip                       = local.vrrp_ip
  ssh_authorized_keys           = local.ssh_authorized_keys
  subnet_nodes                  = local.subnet_nodes
  subnet_services               = local.subnet_services
  subnet_pods                   = local.subnet_pods
  allocation_pool_start         = local.allocation_pool_start
  allocation_pool_end           = local.allocation_pool_end
  k8s_api_port                  = local.k8s_api_port
  k8s_api_port_acl              = local.k8s_api_port_acl
  size_master = {
    count  = local.master_count
    flavor = local.flavor_master
  }
  size_worker = {
    count  = local.worker_count
    flavor = local.flavor_worker
  }
  # size_worker_windows = {
  #   count  = local.worker_count_windows
  #   flavor = local.flavor_worker_windows
  # }
  node_master                  = local.node_master
  node_worker                  = local.node_worker
  node_worker_windows          = local.node_worker_windows
  ub_version                   = local.ub_version

  worker_node_bfv_volume_size = local.worker_node_bfv_volume_size
  worker_node_bfv_destination_type = local.worker_node_bfv_destination_type
  worker_node_bfv_source_type = local.worker_node_bfv_source_type
  worker_node_bfv_volume_type = local.worker_node_bfv_volume_type
}

module "kubespray-cluster" {
  source = "github.com/rackerlabs/openCenter-gitops-base.git//iac/provider/kubespray?ref=main"
  address_bastion                         = module.openstack-nova.bastion_floating_ip
  cluster_name                            = local.cluster_name
  cni_iface                               = local.cni_iface
  deploy_cluster                          = local.deploy_cluster
  dns_zone_name                           = local.dns_zone_name
  master_nodes                            = module.openstack-nova.master_nodes
  network_plugin                          = local.network_plugin
  k8s_hardening_enabled                   = local.k8s_hardening_enabled
  os_hardening_enabled                    = local.os_hardening_enabled
  ssh_user                                = local.ssh_user
  subnet_nodes                            = local.subnet_nodes
  subnet_pods                             = local.subnet_pods
  subnet_services                         = local.subnet_services
  kubernetes_version                      = local.kubernetes_version
  kubespray_version                       = local.kubespray_version
  kube_vip_enabled                        = local.kube_vip_enabled
  kube_pod_security_exemptions_namespaces = local.kube_pod_security_exemptions_namespaces
  kubelet_rotate_server_certificates      = local.kubelet_rotate_server_certificates
  worker_nodes                            = module.openstack-nova.worker_nodes
  k8s_api_ip                              = module.openstack-nova.k8s_api_ip
  k8s_api_port                            = local.k8s_api_port
  k8s_internal_ip                         = module.openstack-nova.k8s_internal_ip
  vrrp_ip                                 = local.vrrp_ip
  vrrp_enabled                            = local.vrrp_enabled
  windows_nodes                           = module.openstack-nova.windows_nodes
  use_octavia                             = local.use_octavia
  kube_oidc_auth_enabled                  = local.kube_oidc_auth_enabled
  kube_oidc_url                           = local.kube_oidc_url
  kube_oidc_client_id                     = local.kube_oidc_client_id
  # kube_oidc_ca_file                       = local.kube_oidc_ca_file
  kube_oidc_username_claim                = local.kube_oidc_username_claim
  # kube_oidc_username_prefix               = local.kube_oidc_username_prefix
  kube_oidc_groups_claim                  = local.kube_oidc_groups_claim
  # kube_oidc_groups_prefix                 = local.kube_oidc_groups_prefix
}


module "calico" {
  source = "github.com/rackerlabs/openCenter-gitops-base.git//iac/cni/calico?ref=main"

  calico_interface_autodetect      = local.calico_interface_autodetect
  calico_encapsulation_type        = local.calico_encapsulation_type
  calico_nat_outgoing              = local.calico_nat_outgoing
  calico_interface_autodetect_cidr = local.calico_interface_autodetect_cidr == "" ? local.subnet_nodes : local.calico_interface_autodetect_cidr
  cni_iface                        = local.cni_iface
  cluster_name                     = local.cluster_name
  deploy_cluster                   = local.deploy_cluster
  k8s_internal_ip                  = module.openstack-nova.k8s_internal_ip
  k8s_api_port                     = local.k8s_api_port
  subnet_nodes                     = local.subnet_nodes
  subnet_pods                      = local.subnet_pods
  subnet_services                  = local.subnet_services
  windows_dataplane                = length(module.openstack-nova.windows_nodes) > 0 ? "HSN" : "Disabled"
}
