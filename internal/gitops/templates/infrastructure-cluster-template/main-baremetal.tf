locals {
  # this will be the user's name and the DNS zone prefix
  cluster_name                            = "k8s-qa"

  #CIDR that the openstack VMs will use for K8s nodes
  subnet_nodes                            = "172.26.0.0/24"
  #CIDR that will be used by kubernetes pods. Not an openstack network.
  subnet_pods                             = "10.42.0.0/16"
  #CIDR that will be used for kubernetes services. Not an openstack network.
  subnet_services                         = "10.43.0.0/16"
  # vrrp_enabled cannot be set to true if use_octavia is true
  vrrp_enabled                            = true
  # Creates a DNS record using the LB floating IP and dns_zone_name
  dns_zone_name                           = "k8s-qa.farmcreditfunding.com"

  k8s_api_port                            = 443
  ssh_user                                = "ubuntu"
  use_octavia                             = false

  # ====================================
  #Kubespray Settings
  kubespray_version                       = "v2.28.1"
  kubernetes_version                      = "1.32.8"
  network_plugin                          = "calico"
  deploy_cluster                          = true
  #kub-vip settings
  kube_vip_enabled                        = true
  #Hardening
  k8s_hardening_enabled                   = true
  kube_pod_security_exemptions_namespaces = ["trivy-temp"]
  kubelet_rotate_server_certificates      = true
  os_hardening_enabled                    = true

  #OIDC Settings
  kube_oidc_auth_enabled                 = false
  kube_oidc_url                          = "https://auth.fcc.k8s-qa.ord1.k8s.opencenter.cloud/realms/opencenter"
  kube_oidc_client_id                    = "opencenter"
  # Optional settings fo OIDC
  kube_oidc_ca_file                      = "/etc/ssl/certs/ca-certificates.crt"
  kube_oidc_username_claim               = "preferred_username"
  kube_oidc_username_prefix              = "oidc:"
  kube_oidc_groups_claim                 = "groups"
  kube_oidc_groups_prefix                = "oidc:"

  #Calico Settings
  cni_iface                               = "ens192"
  #Interface detection method for Calico nodeAddressAutodetectionV4. Can be "first-found", "interface", "cidr"
  calico_interface_autodetect             = "interface"
  calico_interface_autodetect_cidr        = ""
  calico_encapsulation_type               = "VXLAN"
  calico_nat_outgoing                     = true

  ######################
  address_bastion                         = "198.101.167.255"
  k8s_api_ip                              = "108.166.24.164"
  windows_dataplane                       = "Disabled"
  vrrp_ip                                 = "172.26.0.5"
  ssh_key_path                            = "/etc/openCenter/1643323-Federal-Farm-Credit/secrets/ssh/k8s-qa-svc01m-ord1"
  master_nodes = [
    {
      id           = "master-0"
      name         = "k8s-qa-ord1-cp0"
      access_ip_v4 = "172.26.0.11"
    },
    {
      id           = "master-1"
      name         = "k8s-qa-ord1-cp1"
      access_ip_v4 = "172.26.0.12"
    },
    {
      id           = "master-2"
      name         = "k8s-qa-ord1-cp2"
      access_ip_v4 = "172.26.0.13"
    }
  ]

  worker_nodes = [
    {
      id           = "worker-0"
      name         = "k8s-qa-ord1-wn0"
      access_ip_v4 = "172.26.0.14"
    },
    {
      id           = "worker-1"
      name         = "k8s-qa-ord1-wn1"
      access_ip_v4 = "172.26.0.15"
    },
    {
      id           = "worker-2"
      name         = "k8s-qa-ord1-wn2"
      access_ip_v4 = "172.26.0.16"
    }
  ]

  windows_nodes = []
}

module "kubespray-cluster" {
  source = "github.com/rackerlabs/openCenter-gitops-base.git//iac/provider/kubespray?ref=main"
  address_bastion                         = local.address_bastion
  cluster_name                            = local.cluster_name
  cni_iface                               = local.cni_iface
  deploy_cluster                          = local.deploy_cluster
  dns_zone_name                           = local.dns_zone_name
  master_nodes                            = local.master_nodes
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
  worker_nodes                            = local.worker_nodes
  k8s_api_ip                              = local.k8s_api_ip
  k8s_api_port                            = local.k8s_api_port
  k8s_internal_ip                         = local.vrrp_ip
  vrrp_ip                                 = local.vrrp_ip
  vrrp_enabled                            = local.vrrp_enabled
  windows_nodes                           = []
  ssh_key_path                            = local.ssh_key_path
  use_octavia                             = local.use_octavia
  kube_oidc_auth_enabled                  = local.kube_oidc_auth_enabled
  kube_oidc_url                           = local.kube_oidc_url
  kube_oidc_client_id                     = local.kube_oidc_client_id
  kube_oidc_ca_file                       = local.kube_oidc_ca_file
  kube_oidc_username_claim                = local.kube_oidc_username_claim
  kube_oidc_username_prefix               = local.kube_oidc_username_prefix
  kube_oidc_groups_claim                  = local.kube_oidc_groups_claim
  kube_oidc_groups_prefix                 = local.kube_oidc_groups_prefix
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
  k8s_internal_ip                  = local.vrrp_ip
  k8s_api_port                     = local.k8s_api_port
  subnet_nodes                     = local.subnet_nodes
  subnet_pods                      = local.subnet_pods
  subnet_services                  = local.subnet_services
  windows_dataplane                = local.windows_dataplane
}
