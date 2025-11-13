locals {
  # this will be the user's name and the DNS zone prefix

  #CIDR that the openstack VMs will use for K8s nodes
  subnet_nodes                            = "172.23.0.0/24"
  subnet_nodes_oct                       = join(".", slice(split(".", split("/", local.subnet_nodes)[0]), 0, 3))
  #Leave some IPs free for the VRRP IP and the MetalLB Range
  allocation_pool_start                   = "${local.subnet_nodes_oct}.50"
  allocation_pool_end                     = "${local.subnet_nodes_oct}.254"
  # vrrp_ip Must be an IP from subnet_nodes and will be used as  the internal Kubernetes API VIP.
  vrrp_ip                                 = "${local.subnet_nodes_oct}.10"
  #CIDR that will be used by kubernetes pods. Not an openstack network.
  subnet_pods                             = "10.42.0.0/16"
  #CIDR that will be used for kubernetes services. Not an openstack network.
  subnet_services                         = "10.43.0.0/16"
  # use_octavia set to false to create a floating IP associated with the vrrp_ip port. true will create an octavia LB with a floating IP
  use_octavia                             = false
  loadbalancer_provider                   = "amphora"
  # vrrp_enabled cannot be set to true if use_octavia is true
  vrrp_enabled                            = true
  # Creates a DNS record using the LB floating IP and dns_zone_name
  use_designate                           = false
  # dns_zone_name is the dns zone to create if use_designate is true
  dns_zone_name                           = "k8s-dev.farmcreditfunding.com"
  # DNS servers to configure on the nodes
  dns_nameservers                         = ["1.1.1.1","8.8.8.8"]
  ntp_servers                             = ["time.dfw3.rackspace.com","time2.dfw3.rackspace.com"]

  k8s_api_port                            = 443
  k8s_api_port_acl                        = ["0.0.0.0/0"]
  worker_count                            = 3
  worker_count_windows                    = 0
  # Enter 1 or 3 masters.
  master_count                            = 3
  ssh_user                                = "ubuntu"
  # these are the ssh public keys that will be able to connect to the cluster's bastion node
  ssh_authorized_keys                     = ["ssh-rsa ..."]
  node_worker                             = "wn"
  node_master                             = "cp"
  node_worker_windows                     = "win"
  ub_version                              = "24"
  #FLEX Flavor Settings ==========================
  flavor_bastion                          = "gp.5.2.2"
  flavor_master                           = "gp.5.4.8"
  flavor_worker                           = "gp.5.4.8"
  flavor_worker_windows                   = "gp.5.4.16"

  worker_node_bfv_volume_size             = 20
  worker_node_bfv_destination_type        = "volume"
  worker_node_bfv_source_type             = "image"
  worker_node_bfv_volume_type             = "Standard"
  additional_block_devices_worker = []
  # additional_block_devices_worker = [
  #   {
  #     source_type           = "blank"
  #     volume_size           = 20
  #     volume_type           = "Performance"
  #     boot_index            = -1
  #     destination_type      = "volume"
  #     delete_on_termination = true
  #     mountpoint            = "/var/lib/longhorn"
  #     filesystem            = "ext4"
  #     label                 = "longhorn-vol"
  #   },
  #   {
  #     source_type           = "blank"
  #     volume_size           = 20
  #     volume_type           = "Standard"
  #     boot_index            = -1
  #     destination_type      = "volume"
  #     delete_on_termination = true
  #     mountpoint            = "/var/lib/mysql"
  #     filesystem            = "ext4"
  #     label                 = "db-vol"
  #   },
  # ]

  # ====================================
  #ca_certificates add CA certificates to server's trusts. Good for trusting internal private Certificate Authorities.
  ca_certificates                         = ""
  openstack_ca                            = ""

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
  kubelet_rotate_server_certificates      = false
  os_hardening_enabled                    = true

  {{- if .OpenCenter.Cluster.Kubernetes.OIDC.Enabled }}
  #OIDC Settings
  kube_oidc_auth_enabled                 = {{ .OpenCenter.Cluster.Kubernetes.OIDC.Enabled | default false }}
  kube_oidc_url                          = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCURL | default "" }}"
  kube_oidc_client_id                    = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCClientID | default "kubernetes" }}"
  # # Optional settings fo OIDC
  kube_oidc_ca_file                      = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCCAFile | default "" }}"
  kube_oidc_username_claim               = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCUsernameClaim | default "sub" }}"
  kube_oidc_username_prefix              = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCUsernamePrefix | default "oidc:" }}"
  kube_oidc_groups_claim                 = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCGroupsClaim | default "groups" }}"
  kube_oidc_groups_prefix                = "{{ .OpenCenter.Cluster.Kubernetes.OIDC.KubeOIDCGroupsPrefix | default "oidc:" }}"
  {{- end }}


  #Calico Settings
  cni_iface                               = "enp3s0"
  #Interface detection method for Calico nodeAddressAutodetectionV4. Can be "first-found", "interface", "cidr"
  #https://docs.tigera.io/calico/latest/reference/installation/api#operator.tigera.io%2fv1.NodeAddressAutodetection
  calico_interface_autodetect             = "interface"
  calico_interface_autodetect_cidr        = ""
  calico_encapsulation_type               = "VXLAN"
  calico_nat_outgoing                     = true

######################
  address_bastion                        = "50.56.158.76" ##Or Public IP NATed to Bastion
  cluster_name                            = "k8s-dev-iad3"
  naming_prefix                           = "${local.cluster_name}-"
  windows_dataplane                       = "Disabled" 

  {{- if .OpenCenter.Cluster.Kubernetes.MasterNodes }}
  master_nodes = [
    {{- range .OpenCenter.Cluster.Kubernetes.MasterNodes }}
    {
      id = "{{ .ID }}"
      name = "{{ .Name }}"
      access_ip_v4 = "{{ .AccessIPv4 }}"
    },
    {{- end }}
  ]
  {{- else }}
  master_nodes = [
    {
    id = "master-0"
    name = "k8s-dev-iad3-cp0"
    access_ip_v4 = "172.23.0.51"
    },
    {
    id = "master-1"
    name = "k8s-dev-iad3-cp1"
    access_ip_v4 = "172.23.0.182"
    },
    {
    id = "master-2"
    name = "k8s-dev-iad3-cp2"
    access_ip_v4 = "172.23.0.168"
    }
  ]
  {{- end }}

  {{- if .OpenCenter.Cluster.Kubernetes.WorkerNodes }}
  worker_nodes = [
    {{- range .OpenCenter.Cluster.Kubernetes.WorkerNodes }}
    {
      id = "{{ .ID }}"
      name = "{{ .Name }}"
      access_ip_v4 = "{{ .AccessIPv4 }}"
    },
    {{- end }}
  ]
  {{- else }}
  worker_nodes = [
    {
    id = "worker-0"
    name = "k8s-dev-iad3-wn0"
    access_ip_v4 = "172.23.0.220"
    },
    {
    id = "worker-1"
    name = "k8s-dev-iad3-wn1"
    access_ip_v4 = "172.23.0.155"
    },
    {
    id = "worker-2"
    name = "k8s-dev-iad3-wn2"
    access_ip_v4 = "172.23.0.212"
    }
  ]
  {{- end }}
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
  k8s_api_ip                              = local.vrrp_ip
  k8s_api_port                            = local.k8s_api_port
  k8s_internal_ip                         = local.vrrp_ip
  vrrp_ip                                 = local.vrrp_ip
  vrrp_enabled                            = local.vrrp_enabled
  windows_nodes                           = []
  use_octavia                             = local.use_octavia
  #kube_oidc_auth_enabled                  = local.kube_oidc_auth_enabled
  #kube_oidc_url                           = local.kube_oidc_url
  #kube_oidc_client_id                     = local.kube_oidc_client_id
  # kube_oidc_ca_file                       = local.kube_oidc_ca_file
  #kube_oidc_username_claim                = local.kube_oidc_username_claim
  # kube_oidc_username_prefix               = local.kube_oidc_username_prefix
  #kube_oidc_groups_claim                  = local.kube_oidc_groups_claim
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
  k8s_internal_ip                  = local.vrrp_ip
  k8s_api_port                     = local.k8s_api_port
  subnet_nodes                     = local.subnet_nodes
  subnet_pods                      = local.subnet_pods
  subnet_services                  = local.subnet_services
  windows_dataplane                = local.windows_dataplane
}

