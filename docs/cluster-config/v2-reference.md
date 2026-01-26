# v2 Cluster Configuration Reference

## Table of Contents

- [Overview](#overview)
- [Schema Version](#schema-version)
- [Configuration Domains](#configuration-domains)
  - [Meta Domain](#meta-domain)
  - [Cluster Domain](#cluster-domain)
  - [Infrastructure Domain](#infrastructure-domain)
  - [Deployment Domain](#deployment-domain)
  - [Services Domain](#services-domain)
- [Reference Resolution](#reference-resolution)
- [Provider-Region Defaults](#provider-region-defaults)
- [Validation Rules](#validation-rules)
- [Error Codes](#error-codes)
- [Complete Examples](#complete-examples)

## Overview

The v2 cluster configuration schema provides a hierarchical, domain-driven approach to Kubernetes cluster configuration. It eliminates duplication, establishes clear ownership boundaries, isolates provider-specific settings, and supports advanced deployment methods like Kamaji hosted control planes.

**Key Principles:**

- **Single Source of Truth**: Each setting defined exactly once at the appropriate hierarchy level
- **Domain Separation**: Clear boundaries between Meta, Cluster, Infrastructure, Deployment, and Services
- **Provider Isolation**: Provider-specific settings under `infrastructure.cloud.<provider>`
- **Deployment Abstraction**: Deployment method (how) separated from infrastructure provider (where)
- **Reference Resolution**: Explicit `${path.to.value}` syntax for shared resources
- **Context-Aware Defaults**: Provider-region registry supplies intelligent defaults

**Current Implementation Status**: OpenStack is the only production-ready provider. AWS, GCP, Azure, and VMware are included in the schema for architectural completeness and future extensibility but are not currently scheduled or planned for implementation.

## Schema Version

All v2 configurations MUST include the `schema_version` field at the root level:

```yaml
schema_version: "2.0"
```

**Valid Values:**
- `"1.0"`: v1 schema (deprecated, see [Migration Guide](migration-guide.md))
- `"2.0"`: v2 schema (current)

**Detection Logic:**
- If `schema_version` is missing, defaults to v1 for backward compatibility
- If `schema_version: "2.0"`, enforces v2 validation rules
- If `schema_version` is invalid, configuration is rejected


## Configuration Domains

### Meta Domain

**Purpose**: Cluster identity and organizational context

**Location**: `opencenter.meta`

**Responsibility**: Defines cluster metadata for identification, organization, and environment classification.

**Owner**: Cluster operator

#### Fields

```yaml
opencenter:
  meta:
    name: my-cluster                    # Cluster identifier (required)
    organization: acme-corp             # Organization name (required)
    env: production                     # Environment (optional)
    region: sjc3                        # Deployment region (required)
    status: active                      # Cluster status (optional)
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `name` | string | Yes | Unique cluster identifier. Must be DNS-1123 compliant (lowercase, alphanumeric, hyphens). | 3-63 characters |
| `organization` | string | Yes | Organization name for multi-tenant deployments. | Any string |
| `env` | string | No | Environment designation for lifecycle management. | `dev`, `staging`, `production`, or empty |
| `region` | string | Yes | Geographic region for deployment. Must match provider region. | Provider-specific |
| `status` | string | No | Current cluster operational status. | `active`, `inactive`, `maintenance`, or empty |

#### Validation Rules

- `name` must be unique within organization
- `name` must match `cluster.cluster_name`
- `region` must exist in provider-region registry
- `organization` is used for configuration path resolution

#### Example

```yaml
opencenter:
  meta:
    name: prod-k8s-cluster
    organization: acme-corp
    env: production
    region: sjc3
    status: active
```


### Cluster Domain

**Purpose**: Kubernetes configuration (provider-agnostic)

**Location**: `opencenter.cluster`

**Responsibility**: Defines Kubernetes-specific settings that are independent of infrastructure provider.

**Owner**: Kubernetes administrator

#### Fields

```yaml
opencenter:
  cluster:
    cluster_name: prod-k8s-cluster      # Must match meta.name (required)
    base_domain: acme-corp.com          # Base DNS domain (required)
    cluster_fqdn: prod.acme-corp.com    # Cluster FQDN (required)
    admin_email: ops@acme-corp.com      # Administrator email (required)
    kubernetes:                         # Kubernetes configuration (required)
      version: "1.31.4"                 # Kubernetes version (required)
      api_port: 6443                    # API server port (required)
      kube_vip_enabled: true            # Enable kube-vip for HA (optional)
      subnet_pods: "10.233.64.0/18"     # Pod network CIDR (required)
      subnet_services: "10.233.0.0/18"  # Service network CIDR (required)
      network_plugin:                   # CNI plugin (required, mutually exclusive)
        calico:
          enabled: true
          version: "v3.27.0"
          cni_iface: "enp3s0"
          calico_interface_autodetect: "interface"
        cilium:
          enabled: false
        kube-ovn:
          enabled: false
      storage_plugin:                   # CSI plugin (required, mutually exclusive)
        cinder_csi:
          enabled: true
          version: "v1.28.0"
        aws_ebs_csi:
          enabled: false
        vsphere_csi:
          enabled: false
      security:                         # Security configuration (optional)
        pod_security_standard: "restricted"
        audit_logging_enabled: true
      oidc:                             # OIDC authentication (optional)
        enabled: false
        kube_oidc_url: ""
        kube_oidc_client_id: "kubernetes"
        kube_oidc_username_claim: "sub"
        kube_oidc_groups_claim: "groups"
```

#### Kubernetes Configuration

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `version` | string | Yes | Kubernetes version to deploy. | Semantic version (e.g., "1.31.4") |
| `api_port` | integer | Yes | API server port. | 1-65535 (default: 6443) |
| `kube_vip_enabled` | boolean | No | Enable kube-vip for control plane HA. | `true`, `false` (default: `false`) |
| `subnet_pods` | string | Yes | Pod network CIDR (CNI-managed). | Valid CIDR notation |
| `subnet_services` | string | Yes | Service network CIDR (CNI-managed). | Valid CIDR notation |

#### Network Plugin Configuration

**Rule**: Only ONE network plugin can be enabled per cluster.

**Calico:**

```yaml
network_plugin:
  calico:
    enabled: true
    version: "v3.27.0"
    cni_iface: "enp3s0"                 # Network interface name
    calico_interface_autodetect: "interface"  # Detection method
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `enabled` | boolean | Yes | Enable Calico CNI. | `true`, `false` |
| `version` | string | No | Calico version. | Semantic version |
| `cni_iface` | string | No | Network interface for BGP. | Interface name (e.g., "enp3s0", "eth0") |
| `calico_interface_autodetect` | string | No | Interface detection method. | `interface`, `can-reach`, `skip-interface`, `cidr` |

**Cilium:**

```yaml
network_plugin:
  cilium:
    enabled: true
    version: "v1.14.0"
    operator_enabled: true
    kubeProxyReplacement: true
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `enabled` | boolean | Yes | Enable Cilium CNI. | `true`, `false` |
| `version` | string | No | Cilium version. | Semantic version |
| `operator_enabled` | boolean | No | Enable Cilium operator. | `true`, `false` (default: `true`) |
| `kubeProxyReplacement` | boolean | No | Replace kube-proxy with eBPF. | `true`, `false` (default: `false`) |

**Kube-OVN:**

```yaml
network_plugin:
  kube-ovn:
    enabled: true
    version: "v1.12.0"
    cilium_integration: false
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `enabled` | boolean | Yes | Enable Kube-OVN CNI. | `true`, `false` |
| `version` | string | No | Kube-OVN version. | Semantic version |
| `cilium_integration` | boolean | No | Enable Cilium integration. | `true`, `false` (default: `false`) |

#### Storage Plugin Configuration

**Rule**: Only ONE storage plugin can be enabled per cluster.

**Cinder CSI (OpenStack):**

```yaml
storage_plugin:
  cinder_csi:
    enabled: true
    version: "v1.28.0"
```

**AWS EBS CSI:**

```yaml
storage_plugin:
  aws_ebs_csi:
    enabled: true
    version: "v1.25.0"
```

**vSphere CSI:**

```yaml
storage_plugin:
  vsphere_csi:
    enabled: true
    version: "v3.0.0"
```

#### OIDC Configuration

```yaml
oidc:
  enabled: true
  kube_oidc_url: "https://keycloak.acme-corp.com/realms/kubernetes"
  kube_oidc_client_id: "kubernetes"
  kube_oidc_ca_file: "/etc/kubernetes/pki/oidc-ca.crt"
  kube_oidc_username_claim: "sub"
  kube_oidc_username_prefix: "oidc:"
  kube_oidc_groups_claim: "groups"
  kube_oidc_groups_prefix: "oidc:"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable OIDC authentication. |
| `kube_oidc_url` | string | Yes (if enabled) | OIDC provider URL. |
| `kube_oidc_client_id` | string | Yes (if enabled) | OIDC client ID. |
| `kube_oidc_ca_file` | string | No | Path to OIDC CA certificate. |
| `kube_oidc_username_claim` | string | No | JWT claim for username (default: "sub"). |
| `kube_oidc_username_prefix` | string | No | Prefix for usernames (default: "oidc:"). |
| `kube_oidc_groups_claim` | string | No | JWT claim for groups (default: "groups"). |
| `kube_oidc_groups_prefix` | string | No | Prefix for groups (default: "oidc:"). |

#### Validation Rules

- `cluster_name` must match `meta.name`
- `subnet_pods` and `subnet_services` must not overlap
- Only one network plugin can have `enabled: true`
- Only one storage plugin can have `enabled: true`
- Storage plugin must be compatible with infrastructure provider
- OIDC URL must be valid HTTPS URL when OIDC is enabled


### Infrastructure Domain

**Purpose**: Physical resources (provider-agnostic core + provider-specific extensions)

**Location**: `opencenter.infrastructure`

**Responsibility**: Defines infrastructure resources including networking, compute, storage, and provider-specific settings.

**Owner**: Infrastructure team

#### Core Infrastructure Fields

```yaml
opencenter:
  infrastructure:
    provider: openstack                 # Infrastructure provider (required)
    os_version: "ubuntu-22.04"          # Operating system version (required)
    ssh:                                # SSH configuration (required)
      user: "ubuntu"
      key_path: "~/.ssh/cluster-key"
      authorized_keys:
        - "ssh-rsa AAAAB3..."
    networking:                         # Networking configuration (required)
      # ... see Networking section
    compute:                            # Compute configuration (required)
      # ... see Compute section
    storage:                            # Storage configuration (required)
      # ... see Storage section
    cloud:                              # Provider-specific configuration (required)
      openstack:                        # OpenStack settings
        # ... see Provider Configuration section
      aws:                              # AWS settings (not populated if provider != aws)
        # ... see Provider Configuration section
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `provider` | string | Yes | Infrastructure provider. | `openstack`, `aws`, `gcp`, `azure`, `baremetal`, `vsphere` |
| `os_version` | string | Yes | Operating system version for nodes. | Provider-specific (e.g., "ubuntu-22.04", "rhel-9") |

#### Networking Configuration

**Location**: `infrastructure.networking`

**Purpose**: Physical network configuration (infrastructure-level, not CNI-managed)

```yaml
opencenter:
  infrastructure:
    networking:
      # Network topology
      subnet_nodes: "10.2.128.0/22"             # Node network CIDR (required)
      allocation_pool_start: "10.2.128.10"      # DHCP pool start (required)
      allocation_pool_end: "10.2.131.254"       # DHCP pool end (required)
      
      # High availability
      vrrp_ip: "10.2.128.5"                     # Virtual IP for HA (required if vrrp_enabled)
      vrrp_enabled: true                        # Enable VRRP (optional)
      
      # Load balancing
      use_octavia: true                         # Use Octavia LB (OpenStack-specific)
      loadbalancer_provider: "octavia"          # LB provider (required)
      
      # DNS
      use_designate: true                       # Use Designate DNS (OpenStack-specific)
      dns_zone_name: "acme-corp.com"            # DNS zone (required)
      dns_nameservers:                          # DNS servers (required)
        - "8.8.8.8"
        - "8.8.4.4"
      
      # Time synchronization
      ntp_servers:                              # NTP servers (required)
        - "time.google.com"
      
      # Security
      security:
        firewall_enabled: true
        allowed_cidr_blocks:
          - "10.0.0.0/8"
      
      # VLAN (optional)
      vlan:
        enabled: false
        vlan_id: 100
```

| Field | Type | Required | Description | Validation |
|-------|------|----------|-------------|------------|
| `subnet_nodes` | string | Yes | Node network CIDR. | Valid CIDR notation |
| `allocation_pool_start` | string | Yes | DHCP allocation pool start IP. | Must be within `subnet_nodes` |
| `allocation_pool_end` | string | Yes | DHCP allocation pool end IP. | Must be within `subnet_nodes` |
| `vrrp_ip` | string | Conditional | Virtual IP for control plane HA. | Required if `vrrp_enabled: true` |
| `vrrp_enabled` | boolean | No | Enable VRRP for HA. | Default: `false` |
| `loadbalancer_provider` | string | Yes | Load balancer provider. | `ovn`, `octavia`, `metallb`, `cloud-native` |
| `dns_zone_name` | string | Yes | DNS zone name. | Valid FQDN |
| `dns_nameservers` | array | Yes | DNS server IPs. | Valid IPv4 addresses |
| `ntp_servers` | array | Yes | NTP server addresses. | Valid FQDN or IPv4 |

**VRRP IP Consolidation**: In v2, VRRP IP has a single location: `infrastructure.networking.vrrp_ip`. This eliminates duplication from v1 where it appeared in multiple locations.

#### Compute Configuration

**Location**: `infrastructure.compute`

**Purpose**: Instance types and node counts

```yaml
opencenter:
  infrastructure:
    compute:
      # Instance flavors
      flavor_bastion: "gp.0.2.2"                # Bastion instance type (required)
      flavor_master: "gp.0.4.4"                 # Control plane instance type (required if master_count > 0)
      flavor_worker: "gp.0.4.8"                 # Worker instance type (required if worker_count > 0)
      flavor_worker_windows: "gp.0.8.16"        # Windows worker instance type (optional)
      
      # Node counts
      master_count: 3                           # Control plane nodes (required)
      worker_count: 3                           # Worker nodes (required)
      worker_count_windows: 0                   # Windows workers (optional)
      
      # Additional worker pools
      additional_server_pools_worker:
        - name: "high-memory"
          worker_count: 2
          flavor_worker: "gp.0.8.64"
          boot_volume:
            size: 200
            type: "HA-Standard"
          labels:
            workload: "memory-intensive"
          taints:
            - key: "workload"
              value: "memory-intensive"
              effect: "NoSchedule"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `flavor_bastion` | string | Yes | Instance flavor for bastion host. |
| `flavor_master` | string | Conditional | Instance flavor for control plane. Required if `master_count > 0`. |
| `flavor_worker` | string | Conditional | Instance flavor for workers. Required if `worker_count > 0`. |
| `master_count` | integer | Yes | Number of control plane nodes (0-9). Must be 0 for Kamaji deployment. |
| `worker_count` | integer | Yes | Number of worker nodes (0-100). |
| `additional_server_pools_worker` | array | No | Additional worker pools with custom configurations. |

**Worker Pool Configuration:**

```yaml
additional_server_pools_worker:
  - name: "pool-name"                   # Pool identifier (required)
    worker_count: 2                     # Number of nodes (required)
    flavor_worker: "gp.0.8.64"          # Instance flavor (optional, inherits from compute.flavor_worker)
    boot_volume:                        # Boot volume config (optional)
      size: 200
      type: "HA-Standard"
    additional_volumes:                 # Additional volumes (optional)
      - device_name: "/dev/vdb"
        size: 500
        type: "HA-Standard"
    labels:                             # Kubernetes labels (optional)
      workload: "memory-intensive"
    taints:                             # Kubernetes taints (optional)
      - key: "workload"
        value: "memory-intensive"
        effect: "NoSchedule"
```

#### Storage Configuration

**Location**: `infrastructure.storage`

**Purpose**: Boot volumes, additional storage, and storage classes

```yaml
opencenter:
  infrastructure:
    storage:
      # Default storage class
      default_storage_class: "csi-cinder-sc-delete"  # Default StorageClass (required)
      
      # Worker boot volumes
      worker_volume_size: 100                   # Boot volume size in GB (required)
      worker_volume_destination_type: "volume"  # Destination type (required)
      worker_volume_source_type: "image"        # Source type (required)
      worker_volume_type: "HA-Standard"         # Volume type (required)
      worker_volume_delete_on_termination: true # Delete on termination (optional)
      
      # Master boot volumes
      master_volume_size: 100                   # Boot volume size in GB (optional)
      master_volume_destination_type: "volume"  # Destination type (optional)
      master_volume_source_type: "image"        # Source type (optional)
      master_volume_type: "HA-Standard"         # Volume type (optional)
      master_volume_delete_on_termination: true # Delete on termination (optional)
      
      # Additional block devices
      additional_block_devices:
        - device_name: "/dev/vdb"
          volume_size: 500
          volume_type: "HA-Standard"
          delete_on_termination: false
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `default_storage_class` | string | Yes | Default Kubernetes StorageClass. | Must match CSI driver StorageClass |
| `worker_volume_size` | integer | Yes | Worker boot volume size (GB). | Minimum: 1 |
| `worker_volume_destination_type` | string | Yes | Boot volume destination type. | `volume`, `local` |
| `worker_volume_source_type` | string | Yes | Boot volume source type. | `image`, `volume`, `snapshot` |
| `worker_volume_type` | string | Yes | Volume type (provider-specific). | Provider-specific |
| `additional_block_devices` | array | No | Additional volumes to attach. | Array of volume configs |


#### Provider Configuration

**Location**: `infrastructure.cloud.<provider>`

**Purpose**: Provider-specific settings isolated from generic configuration

**Rule**: Only the active provider's section should be populated. Multiple provider sections populated simultaneously will be rejected.

**OpenStack Configuration:**

```yaml
opencenter:
  infrastructure:
    provider: openstack
    cloud:
      openstack:
        # Authentication
        auth_url: "https://keystone.sjc3.rackspace.com/v3/"  # Keystone URL (required)
        insecure: false                                      # Skip TLS verification (optional)
        region: "sjc3"                                       # OpenStack region (required)
        application_credential_id: ""                        # App credential ID (required)
        application_credential_secret: ""                    # App credential secret (required)
        domain: "Default"                                    # Domain name (optional)
        tenant_name: "my-project"                            # Project/tenant (required)
        
        # Networking
        floating_network_id: "network-uuid"                  # External network (required)
        subnet_id: "subnet-uuid"                             # Subnet ID (optional)
        
        # Images
        image_id: "799dcf97-3656-4361-8187-13ab1b295e33"     # Base image ID (required)
        
        # Availability
        availability_zones:                                  # AZs for node distribution (optional)
          - "az1"
          - "az2"
          - "az3"
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `auth_url` | string | Yes | Keystone authentication URL. |
| `region` | string | Yes | OpenStack region name. |
| `application_credential_id` | string | Yes | Application credential ID for authentication. |
| `application_credential_secret` | string | Yes | Application credential secret. |
| `tenant_name` | string | Yes | Project/tenant name or ID. |
| `floating_network_id` | string | Yes | External network UUID for floating IPs. |
| `image_id` | string | Yes | Base image UUID for nodes. |

**AWS Configuration (Reference Only):**

```yaml
opencenter:
  infrastructure:
    provider: aws
    cloud:
      aws:
        # Authentication
        profile: "default"                      # AWS CLI profile (optional)
        region: "us-east-1"                     # AWS region (required)
        
        # Networking
        vpc_id: "vpc-12345678"                  # VPC ID (required)
        private_subnets:                        # Private subnet IDs (required)
          - "subnet-11111111"
          - "subnet-22222222"
        public_subnets:                         # Public subnet IDs (required)
          - "subnet-33333333"
          - "subnet-44444444"
        
        # Images
        ami_id: "ami-0c55b159cbfafe1f0"         # AMI ID (required)
```

**GCP Configuration (Reference Only):**

```yaml
opencenter:
  infrastructure:
    provider: gcp
    cloud:
      gcp:
        # Authentication
        project_id: "my-project"                # GCP project ID (required)
        region: "us-central1"                   # GCP region (required)
        
        # Networking
        network: "default"                      # VPC network (required)
        subnetwork: "default"                   # Subnetwork (required)
        
        # Images
        image_family: "ubuntu-2204-lts"         # Image family (required)
```

#### Validation Rules

- Only one provider section can be populated
- Provider-specific required fields must be present
- `allocation_pool_start` and `allocation_pool_end` must be within `subnet_nodes`
- `vrrp_ip` must be within `subnet_nodes` if specified
- Storage class must match CSI driver capabilities
- Flavors must be valid for the provider


### Deployment Domain

**Purpose**: Kubernetes installation method and auto-deploy configuration

**Location**: `deployment` (root level, not under `opencenter`)

**Responsibility**: Defines how Kubernetes is deployed (deployment method) and whether to auto-deploy.

**Owner**: Platform team

#### Core Deployment Fields

```yaml
deployment:
  auto_deploy: true                     # Enable automatic deployment (optional)
  method: kubespray                     # Deployment method (required)
  kubespray:                            # Kubespray configuration (if method=kubespray)
    # ... see Kubespray section
  talos:                                # Talos configuration (if method=talos)
    # ... see Talos section
  kamaji:                               # Kamaji configuration (if method=kamaji)
    # ... see Kamaji section
```

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `auto_deploy` | boolean | No | Automatically deploy after setup. | `true`, `false` (default: `false`) |
| `method` | string | Yes | Deployment method. | `kubespray`, `talos`, `kamaji`, `eks`, `gke`, `aks`, `cluster-api` |

#### Kubespray Deployment

**Purpose**: Deploy Kubernetes using Kubespray (Ansible-based)

```yaml
deployment:
  method: kubespray
  kubespray:
    version: "v2.29.1"                  # Kubespray version (required)
    modules:                            # Kubespray modules (optional)
      metallb:
        enabled: true
      ingress_nginx:
        enabled: true
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Kubespray version (semantic version). |
| `modules` | map | No | Kubespray module configurations. |

**Compatibility**: Kubespray supports OpenStack, AWS, GCP, bare metal, and VMware providers.

#### Talos Deployment

**Purpose**: Deploy Kubernetes using Talos Linux

```yaml
deployment:
  method: talos
  talos:
    version: "v1.6.0"                   # Talos version (required)
    install_disk: "/dev/sda"            # Installation disk (required)
    control_plane_vip: "10.2.128.5"     # Control plane VIP (required)
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Talos version (semantic version). |
| `install_disk` | string | Yes | Disk for Talos installation. |
| `control_plane_vip` | string | Yes | Virtual IP for control plane. |

**Compatibility**: Talos supports OpenStack, AWS, GCP, bare metal, and VMware providers.

#### Kamaji Deployment

**Purpose**: Deploy Kubernetes with hosted control plane using Kamaji

**Key Feature**: Kamaji enables hosted control planes with mixed OS worker pools (Ubuntu, Windows, Talos).

```yaml
deployment:
  method: kamaji
  kamaji:
    enabled: true
    version: "v1.0.0"                   # Kamaji version (required)
    control_plane:                      # Control plane configuration (required)
      replicas: 3                       # Control plane replicas (required, must be odd)
      datastore: etcd                   # Datastore type (required)
      etcd:                             # Etcd configuration (if datastore=etcd)
        storage_class: "csi-cinder-sc-delete"
        storage_size: "10Gi"
      postgresql:                       # PostgreSQL configuration (if datastore=postgresql)
        host: "postgres.example.com"
        port: 5432
        database: "kamaji"
        username: "kamaji"
        password: ""                    # From secrets
      service_type: "LoadBalancer"      # Service type (required)
      api_server_port: 6443             # API server port (required)
      resources:                        # Resource requests/limits (optional)
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"
    cluster_api:                        # Cluster API configuration (required)
      version: "v1.6.0"                 # CAPI version (required)
      providers:                        # CAPI providers (required)
        infrastructure: openstack       # Must match infrastructure.provider
        bootstrap: kubeadm              # Bootstrap provider
        control_plane: kubeadm          # Control plane provider
      openstack:                        # OpenStack CAPI config (if infrastructure=openstack)
        cloud_name: "openstack"
        clouds_yaml_secret: "cloud-config"
    worker_pools:                       # Worker pools (required, min 1)
      - name: "ubuntu-workers"
        os: ubuntu                      # OS type (required)
        count: 3                        # Node count (required)
        flavor: "gp.0.4.8"              # Instance flavor (required)
        image: "ubuntu-22.04"           # Image name (required)
        bootstrap_provider: kubeadm     # Bootstrap provider (required)
        boot_volume:                    # Boot volume config (required)
          size: 100
          type: "HA-Standard"
        additional_volumes:             # Additional volumes (optional)
          - device_name: "/dev/vdb"
            size: 500
            type: "HA-Standard"
        labels:                         # Kubernetes labels (optional)
          workload: "general"
        taints: []                      # Kubernetes taints (optional)
        autoscaling:                    # Autoscaling config (optional)
          enabled: false
          min_replicas: 3
          max_replicas: 10
      - name: "talos-workers"
        os: talos                       # Talos OS
        count: 2
        flavor: "gp.0.4.8"
        image: "talos-v1.6.0"
        bootstrap_provider: talos       # Must be "talos" for Talos OS
        talos_version: "v1.6.0"         # Required for Talos OS
        boot_volume:
          size: 50
          type: "HA-Standard"
        talos_config:                   # Talos-specific config (optional)
          install_disk: "/dev/sda"
```

**Kamaji Control Plane:**

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `replicas` | integer | Yes | Control plane replicas (must be odd for HA). | 1, 3, 5, 7 |
| `datastore` | string | Yes | Datastore type for etcd. | `etcd`, `postgresql`, `mysql` |
| `service_type` | string | Yes | Kubernetes service type for API server. | `LoadBalancer`, `NodePort` |
| `api_server_port` | integer | Yes | API server port. | 1-65535 (default: 6443) |

**Kamaji Worker Pools:**

| Field | Type | Required | Description | Valid Values |
|-------|------|----------|-------------|--------------|
| `name` | string | Yes | Worker pool identifier. | DNS-1123 compliant |
| `os` | string | Yes | Operating system. | `ubuntu`, `windows`, `talos` |
| `count` | integer | Yes | Number of nodes. | Minimum: 1 |
| `flavor` | string | Yes | Instance flavor. | Provider-specific |
| `image` | string | Yes | Image name or ID. | Provider-specific |
| `bootstrap_provider` | string | Yes | Bootstrap provider. | `kubeadm` (ubuntu/windows), `talos` (talos) |
| `talos_version` | string | Conditional | Talos version. | Required if `os: talos` |

**Kamaji Constraints:**

- `infrastructure.compute.master_count` MUST be 0 (no self-hosted control plane)
- `infrastructure.networking.vrrp_enabled` MUST be false (no VRRP with hosted control plane)
- `cluster.kubernetes.kube_vip_enabled` MUST be false (no kube-vip with hosted control plane)
- At least one worker pool MUST be defined
- `cluster_api.providers.infrastructure` MUST match `infrastructure.provider`
- Worker pool `bootstrap_provider` MUST match OS (ubuntu/windows→kubeadm, talos→talos)

**Compatibility**: Kamaji supports OpenStack, AWS, GCP, and VMware providers.

#### Validation Rules

- Deployment method must be compatible with infrastructure provider
- Kamaji-specific constraints must be enforced when `method: kamaji`
- Worker pool OS and bootstrap provider must be compatible
- Control plane replicas must be odd number for HA


### Services Domain

**Purpose**: Self-hosted platform workloads and managed services

**Location**: `opencenter.services` (self-hosted), `opencenter.managed_services` (external)

**Responsibility**: Defines which services are enabled and their configurations.

**Owner**: Platform team

#### Service Configuration

Services are configured using a polymorphic map structure where each service has a base configuration plus service-specific fields.

**Base Service Configuration:**

```yaml
opencenter:
  services:
    <service-name>:
      enabled: true                     # Enable service (required)
      status: "active"                  # Service status (optional)
      namespace: "service-namespace"    # Kubernetes namespace (optional)
      hostname: "service.example.com"   # Service hostname (optional)
      image_repository: "custom-repo"   # Custom image repository (optional)
      image_tag: "v1.0.0"               # Custom image tag (optional)
      release: "v1.0.0"                 # Service release version (optional)
      branch: "main"                    # Git branch for GitOps (optional)
      uri: "https://service.example.com"  # Service URI (optional)
      gitops_source_type: "git"         # GitOps source type (optional)
      gitops_source_url: ""             # GitOps source URL (optional)
      gitops_source_ref: "main"         # GitOps source ref (optional)
```

#### Common Services

**Calico (CNI):**

```yaml
opencenter:
  services:
    calico:
      enabled: true
      release: "v3.27.0"
      calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:6443"
```

**Cert-Manager (Certificate Management):**

```yaml
opencenter:
  services:
    cert-manager:
      enabled: true
      release: "v1.14.0"
      email: "ops@acme-corp.com"
      dns_provider: "route53"           # Auto-selected based on infrastructure
      region: "us-east-1"
```

**Service Provider Polymorphism**: cert-manager DNS provider is auto-selected based on infrastructure:
- AWS → `route53`
- OpenStack + Designate → `designate`
- OpenStack without Designate → `cloudflare`

**FluxCD (GitOps):**

```yaml
opencenter:
  services:
    fluxcd:
      enabled: true
      release: "v2.2.0"
      interval: "15m"
      prune: true
```

**Kube-Prometheus-Stack (Monitoring):**

```yaml
opencenter:
  services:
    kube-prometheus-stack:
      enabled: true
      release: "v0.72.0"
      grafana_admin_password: ""        # From secrets
      retention: "30d"
```

**Loki (Logging):**

```yaml
opencenter:
  services:
    loki:
      enabled: true
      release: "v2.9.0"
      storage_type: "s3"
      bucket_name: "loki-logs"
      s3_endpoint: "https://s3.amazonaws.com"
      retention_period: "30d"
```

**Tempo (Tracing):**

```yaml
opencenter:
  services:
    tempo:
      enabled: true
      release: "v2.3.0"
      storage_type: "s3"
      bucket_name: "tempo-traces"
      s3_endpoint: "https://s3.amazonaws.com"
```

**Keycloak (Identity Management):**

```yaml
opencenter:
  services:
    keycloak:
      enabled: true
      release: "v23.0.0"
      admin_username: "admin"
      admin_password: ""                # From secrets
      database_vendor: "postgres"
```

**Velero (Backup):**

```yaml
opencenter:
  services:
    velero:
      enabled: true
      release: "v1.12.0"
      provider: "aws"
      bucket_name: "cluster-backups"
      backup_schedule: "0 2 * * *"      # Daily at 2 AM
```

#### CSI Driver Services

CSI drivers are deployed as services based on `cluster.kubernetes.storage_plugin` selection:

**Cinder CSI (OpenStack):**

```yaml
opencenter:
  services:
    cinder-csi:
      enabled: true                     # Auto-enabled if storage_plugin.cinder_csi.enabled
      release: "v1.28.0"
```

**AWS EBS CSI:**

```yaml
opencenter:
  services:
    aws-ebs-csi:
      enabled: true                     # Auto-enabled if storage_plugin.aws_ebs_csi.enabled
      release: "v1.25.0"
```

#### Service Dependencies

Some services require other services to be enabled:

- `weave-gitops` requires `fluxcd`
- `headlamp` requires `keycloak` (when OIDC is enabled)
- `kube-prometheus-stack` requires `prometheus-operator`
- `loki` requires `promtail` or `fluent-bit`

**Validation**: The system validates service dependencies and rejects configurations with missing dependencies.

#### Required Secrets

Services may require secrets to be configured:

- `cert-manager` with `route53` requires `aws_access_key` and `aws_secret_access_key`
- `loki` with `s3` storage requires `aws_access_key` and `aws_secret_access_key`
- `velero` requires provider-specific credentials
- `keycloak` requires `admin_password`

**Validation**: The system validates required secrets are configured for enabled services.


## Reference Resolution

**Purpose**: Share values across configuration using explicit reference syntax

**Syntax**: `${path.to.value}`

**Resolution Order**: References are resolved before hydration (default application)

### Reference Syntax

References use dot notation to navigate the configuration hierarchy:

```yaml
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "10.2.128.5"
  
  services:
    calico:
      # Reference to infrastructure.networking.vrrp_ip
      calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:6443"
```

**Result after resolution:**

```yaml
opencenter:
  services:
    calico:
      calico_kube_api_server: "10.2.128.5:6443"
```

### Valid Reference Paths

References can point to any field in the configuration:

```yaml
# Reference to meta fields
cluster_name: "${meta.name}"

# Reference to infrastructure fields
api_endpoint: "${infrastructure.networking.vrrp_ip}:${cluster.kubernetes.api_port}"

# Reference to provider-specific fields
image_id: "${infrastructure.cloud.openstack.image_id}"

# Reference to deployment fields
k8s_version: "${cluster.kubernetes.version}"
```

### Reference Resolution Process

1. **Parse Configuration**: Load YAML into Go structs
2. **Build Dependency Graph**: Identify all references and their dependencies
3. **Detect Cycles**: Check for circular dependencies
4. **Topological Sort**: Determine resolution order
5. **Resolve References**: Replace `${...}` with actual values
6. **Apply Defaults**: Apply provider-region defaults (hydration)
7. **Validate**: Validate resolved configuration

### Circular Reference Detection

The system detects circular dependencies and rejects configurations:

**Invalid Configuration (Circular Reference):**

```yaml
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "${cluster.kubernetes.api_endpoint}"
  
  cluster:
    kubernetes:
      api_endpoint: "${infrastructure.networking.vrrp_ip}:6443"
```

**Error:**

```
E003: Circular reference detected: infrastructure.networking.vrrp_ip -> cluster.kubernetes.api_endpoint -> infrastructure.networking.vrrp_ip
```

### Reference Validation

References are validated during resolution:

**Invalid Reference (Path Not Found):**

```yaml
opencenter:
  services:
    calico:
      calico_kube_api_server: "${infrastructure.networking.nonexistent_field}:6443"
```

**Error:**

```
E003: services.calico.calico_kube_api_server: reference ${infrastructure.networking.nonexistent_field} not found
```

### Common Reference Patterns

**API Endpoint:**

```yaml
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "10.2.128.5"
  
  cluster:
    kubernetes:
      api_port: 6443
  
  services:
    calico:
      calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:${cluster.kubernetes.api_port}"
```

**DNS Zone:**

```yaml
opencenter:
  infrastructure:
    networking:
      dns_zone_name: "acme-corp.com"
  
  cluster:
    base_domain: "${infrastructure.networking.dns_zone_name}"
    cluster_fqdn: "${meta.name}.${infrastructure.networking.dns_zone_name}"
```

**Image ID:**

```yaml
opencenter:
  infrastructure:
    cloud:
      openstack:
        image_id: "799dcf97-3656-4361-8187-13ab1b295e33"
  
  deployment:
    kamaji:
      worker_pools:
        - name: "ubuntu-workers"
          image: "${infrastructure.cloud.openstack.image_id}"
```

### Best Practices

1. **Use References for Shared Values**: Define values once, reference everywhere
2. **Avoid Deep Nesting**: Keep reference paths readable
3. **Document References**: Comment complex reference chains
4. **Test Resolution**: Validate references resolve correctly
5. **Avoid Circular Dependencies**: Design configuration hierarchy to prevent cycles


## Provider-Region Defaults

**Purpose**: Supply context-aware defaults based on infrastructure provider and region

**Registry**: Hardcoded defaults for known provider-region combinations

**Precedence**: `explicit config > CLI config > provider-region > provider > global`

### Default Application

Defaults are applied during hydration (after reference resolution, before validation):

1. **Explicit Configuration**: User-specified values (highest precedence)
2. **CLI Configuration**: Values from CLI config file
3. **Provider-Region Defaults**: Region-specific defaults from registry
4. **Provider Defaults**: Provider-level defaults
5. **Global Defaults**: System-wide defaults (lowest precedence)

**Rule**: Defaults only populate empty fields. Explicit values are never overridden.

### OpenStack Provider-Region Defaults

#### sjc3 (San Jose 3)

```yaml
infrastructure:
  cloud:
    openstack:
      region: "sjc3"
      # Applied defaults:
      image_id: "799dcf97-3656-4361-8187-13ab1b295e33"  # Ubuntu 22.04
      availability_zones:
        - "az1"
        - "az2"
        - "az3"
  networking:
    # Applied defaults:
    ntp_servers:
      - "time.sjc3.rackspace.com"
    dns_nameservers:
      - "8.8.8.8"
      - "8.8.4.4"
  storage:
    # Applied defaults:
    default_storage_class: "csi-cinder-sc-delete"
    worker_volume_type: "HA-Standard"
```

#### dfw3 (Dallas Fort Worth 3)

```yaml
infrastructure:
  cloud:
    openstack:
      region: "dfw3"
      # Applied defaults:
      image_id: "a1b2c3d4-5678-90ab-cdef-1234567890ab"  # Ubuntu 22.04
      availability_zones:
        - "az1"
        - "az2"
        - "az3"
  networking:
    # Applied defaults:
    ntp_servers:
      - "time.dfw3.rackspace.com"
```

#### iad3 (Ashburn 3)

```yaml
infrastructure:
  cloud:
    openstack:
      region: "iad3"
      # Applied defaults:
      image_id: "e5f6g7h8-9012-34ij-klmn-5678901234op"  # Ubuntu 22.04
      availability_zones:
        - "az1"
        - "az2"
        - "az3"
  networking:
    # Applied defaults:
    ntp_servers:
      - "time.iad3.rackspace.com"
```

### AWS Provider-Region Defaults (Reference Only)

#### us-east-1

```yaml
infrastructure:
  cloud:
    aws:
      region: "us-east-1"
      # Applied defaults:
      ami_id: "ami-0c55b159cbfafe1f0"  # Ubuntu 22.04
      availability_zones:
        - "us-east-1a"
        - "us-east-1b"
        - "us-east-1c"
```

#### us-west-2

```yaml
infrastructure:
  cloud:
    aws:
      region: "us-west-2"
      # Applied defaults:
      ami_id: "ami-0d1cd67c26f5fca19"  # Ubuntu 22.04
      availability_zones:
        - "us-west-2a"
        - "us-west-2b"
        - "us-west-2c"
```

#### eu-west-1

```yaml
infrastructure:
  cloud:
    aws:
      region: "eu-west-1"
      # Applied defaults:
      ami_id: "ami-0dad359ff462124ca"  # Ubuntu 22.04
      availability_zones:
        - "eu-west-1a"
        - "eu-west-1b"
        - "eu-west-1c"
```

### GCP Provider-Region Defaults (Reference Only)

#### us-central1

```yaml
infrastructure:
  cloud:
    gcp:
      region: "us-central1"
      # Applied defaults:
      image_family: "ubuntu-2204-lts"
      zones:
        - "us-central1-a"
        - "us-central1-b"
        - "us-central1-c"
```

#### europe-west1

```yaml
infrastructure:
  cloud:
    gcp:
      region: "europe-west1"
      # Applied defaults:
      image_family: "ubuntu-2204-lts"
      zones:
        - "europe-west1-b"
        - "europe-west1-c"
        - "europe-west1-d"
```

### Flavor Defaults

Provider-region defaults include instance flavor recommendations:

**OpenStack (sjc3):**

```yaml
infrastructure:
  compute:
    # Applied defaults:
    flavor_bastion: "gp.0.2.2"    # 2 vCPU, 2 GB RAM
    flavor_master: "gp.0.4.4"     # 4 vCPU, 4 GB RAM
    flavor_worker: "gp.0.4.8"     # 4 vCPU, 8 GB RAM
```

**AWS (us-east-1):**

```yaml
infrastructure:
  compute:
    # Applied defaults:
    flavor_bastion: "t3.small"    # 2 vCPU, 2 GB RAM
    flavor_master: "t3.medium"    # 2 vCPU, 4 GB RAM
    flavor_worker: "t3.large"     # 2 vCPU, 8 GB RAM
```

### Overriding Defaults

**CLI Configuration Override:**

Create `~/.config/opencenter/cli-config.yaml`:

```yaml
provider_region_overrides:
  openstack:
    sjc3:
      image_id: "custom-image-uuid"
      ntp_servers:
        - "custom-ntp.example.com"
```

**Cluster Configuration Override:**

Specify explicit values in cluster config:

```yaml
opencenter:
  infrastructure:
    cloud:
      openstack:
        image_id: "custom-image-uuid"  # Overrides provider-region default
    networking:
      ntp_servers:
        - "custom-ntp.example.com"     # Overrides provider-region default
```

### Viewing Applied Defaults

Export effective configuration to see which defaults were applied:

```bash
opencenter cluster config export-effective --config mycluster.yaml
```

**Output includes comments indicating default sources:**

```yaml
opencenter:
  infrastructure:
    cloud:
      openstack:
        image_id: "799dcf97-3656-4361-8187-13ab1b295e33"  # default: provider-region (sjc3)
        availability_zones:  # default: provider-region (sjc3)
          - "az1"
          - "az2"
          - "az3"
    networking:
      ntp_servers:  # default: provider-region (sjc3)
        - "time.sjc3.rackspace.com"
```


## Validation Rules

The v2 configuration system performs multi-layered validation:

### 1. Schema Validation

**Purpose**: Verify required fields, data types, and enum values

**Rules:**

- All required fields must be present
- Field values must match expected data types
- Enum fields must use valid values
- String fields must meet length constraints
- Integer fields must be within valid ranges

**Example Violations:**

```yaml
# Missing required field
opencenter:
  meta:
    # ERROR: name is required
    organization: "acme-corp"

# Invalid enum value
opencenter:
  infrastructure:
    provider: "unknown"  # ERROR: must be openstack, aws, gcp, azure, baremetal, or vsphere

# Invalid data type
opencenter:
  cluster:
    kubernetes:
      api_port: "6443"  # ERROR: must be integer, not string
```

### 2. CIDR Validation

**Purpose**: Verify network CIDR notation and IP address ranges

**Rules:**

- CIDR notation must be valid (e.g., `10.2.128.0/22`)
- Prefix length must be 0-32 for IPv4
- Allocation pool IPs must be within subnet range
- VRRP IP must be within subnet range
- Pod and service subnets must not overlap

**Example Violations:**

```yaml
# Invalid CIDR notation
opencenter:
  infrastructure:
    networking:
      subnet_nodes: "10.2.128.0/33"  # ERROR: prefix must be 0-32

# Allocation pool outside subnet
opencenter:
  infrastructure:
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.132.1"  # ERROR: outside subnet range
```

### 3. Cross-Field Validation

**Purpose**: Verify relationships between fields

**Rules:**

- `cluster.cluster_name` must match `meta.name`
- `allocation_pool_start` must be less than `allocation_pool_end`
- `master_count` must be 0 when `deployment.method: kamaji`
- `vrrp_enabled` must be false when `deployment.method: kamaji`
- Only one network plugin can have `enabled: true`
- Only one storage plugin can have `enabled: true`

**Example Violations:**

```yaml
# Cluster name mismatch
opencenter:
  meta:
    name: "prod-cluster"
  cluster:
    cluster_name: "test-cluster"  # ERROR: must match meta.name

# Multiple network plugins enabled
opencenter:
  cluster:
    kubernetes:
      network_plugin:
        calico:
          enabled: true
        cilium:
          enabled: true  # ERROR: only one plugin can be enabled
```

### 4. Provider-Specific Validation

**Purpose**: Verify provider-specific required fields and constraints

**Rules:**

- Provider-specific required fields must be present
- Only active provider section should be populated
- Provider-specific constraints must be met
- Flavors must be valid for provider
- Images must exist in provider region

**Example Violations:**

```yaml
# Missing OpenStack required field
opencenter:
  infrastructure:
    provider: openstack
    cloud:
      openstack:
        auth_url: "https://keystone.example.com/v3/"
        # ERROR: application_credential_id is required

# Multiple provider sections populated
opencenter:
  infrastructure:
    provider: openstack
    cloud:
      openstack:
        auth_url: "https://keystone.example.com/v3/"
      aws:
        region: "us-east-1"  # ERROR: only openstack section should be populated
```

### 5. Deployment Method Validation

**Purpose**: Verify deployment method compatibility and constraints

**Rules:**

- Deployment method must be compatible with infrastructure provider
- Kamaji-specific constraints must be enforced
- Worker pool OS and bootstrap provider must be compatible
- Control plane replicas must be odd number for HA

**Example Violations:**

```yaml
# Incompatible deployment method
opencenter:
  infrastructure:
    provider: openstack
deployment:
  method: eks  # ERROR: EKS only supported on AWS

# Kamaji with non-zero master count
opencenter:
  infrastructure:
    compute:
      master_count: 3  # ERROR: must be 0 for Kamaji
deployment:
  method: kamaji

# Invalid control plane replicas
deployment:
  method: kamaji
  kamaji:
    control_plane:
      replicas: 4  # ERROR: must be odd number (1, 3, 5, 7)
```

### 6. Service Dependency Validation

**Purpose**: Verify service dependencies are satisfied

**Rules:**

- `weave-gitops` requires `fluxcd`
- `headlamp` requires `keycloak` (when OIDC enabled)
- `kube-prometheus-stack` requires `prometheus-operator`
- CSI driver service must match storage plugin selection

**Example Violations:**

```yaml
# Missing service dependency
opencenter:
  services:
    weave-gitops:
      enabled: true
    fluxcd:
      enabled: false  # ERROR: weave-gitops requires fluxcd
```

### 7. Required Secrets Validation

**Purpose**: Verify required secrets are configured for enabled services

**Rules:**

- `cert-manager` with `route53` requires AWS credentials
- `loki` with `s3` storage requires AWS credentials
- `velero` requires provider-specific credentials
- `keycloak` requires admin password

**Example Violations:**

```yaml
# Missing required secret
opencenter:
  services:
    cert-manager:
      enabled: true
      dns_provider: "route53"
      # ERROR: aws_access_key and aws_secret_access_key required
```

### 8. Reference Validation

**Purpose**: Verify all references resolve and no circular dependencies exist

**Rules:**

- All referenced paths must exist
- No circular reference dependencies
- References must resolve before hydration

**Example Violations:**

```yaml
# Reference to non-existent path
opencenter:
  services:
    calico:
      calico_kube_api_server: "${infrastructure.networking.nonexistent}:6443"
      # ERROR: reference path not found

# Circular reference
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "${cluster.kubernetes.api_endpoint}"
  cluster:
    kubernetes:
      api_endpoint: "${infrastructure.networking.vrrp_ip}:6443"
      # ERROR: circular reference detected
```


## Error Codes

The v2 validation system uses structured error codes for clear error reporting:

### E001: Schema Validation Error

**Trigger**: Missing required fields, invalid data types, invalid enum values

**Examples:**

```
E001: infrastructure.networking.vrrp_ip: required field is missing
E001: infrastructure.provider: invalid value "unknown", must be one of: openstack, aws, gcp, azure, baremetal, vsphere
E001: cluster.kubernetes.api_port: invalid type, expected integer, got string
```

**Resolution**: Ensure all required fields are present and have correct data types.

### E002: CIDR Validation Error

**Trigger**: Invalid CIDR notation, IP addresses outside subnet range

**Examples:**

```
E002: infrastructure.networking.subnet_nodes: invalid CIDR notation "10.2.128.0/33"
E002: infrastructure.networking.allocation_pool_start: IP 10.2.129.1 is outside subnet 10.2.128.0/22
E002: infrastructure.networking.vrrp_ip: IP 10.3.0.1 is outside subnet 10.2.128.0/22
```

**Resolution**: Verify CIDR notation is valid and IP addresses are within subnet ranges.

### E003: Reference Resolution Error

**Trigger**: Reference path doesn't exist, circular dependencies

**Examples:**

```
E003: services.calico.calico_kube_api_server: reference ${infrastructure.networking.vrrp_ip} not found
E003: circular reference detected: infrastructure.networking.vrrp_ip -> cluster.kubernetes.api_endpoint -> infrastructure.networking.vrrp_ip
```

**Resolution**: Ensure referenced paths exist and there are no circular dependencies.

### E004: Provider Validation Error

**Trigger**: Provider-specific required fields missing, invalid provider configuration

**Examples:**

```
E004: infrastructure.cloud.openstack.auth_url: required field for OpenStack provider
E004: infrastructure.cloud.openstack.image_id: image not found in region sjc3
E004: infrastructure.cloud.openstack.floating_network_id: network not found
```

**Resolution**: Verify all provider-specific required fields are present and valid.

### E005: Service Dependency Error

**Trigger**: Service enabled but dependency not enabled

**Examples:**

```
E005: services.weave-gitops: requires services.fluxcd to be enabled
E005: services.headlamp: requires services.keycloak when OIDC is enabled
E005: services.kube-prometheus-stack: requires services.prometheus-operator to be enabled
```

**Resolution**: Enable required service dependencies or disable dependent services.

### E006: Secret Configuration Error

**Trigger**: Service enabled but required secrets not configured

**Examples:**

```
E006: services.cert-manager: requires secrets.cert_manager.aws_access_key when using route53 DNS provider
E006: services.loki: requires secrets.loki.swift_password when using swift storage
E006: services.keycloak: requires secrets.keycloak.admin_password
```

**Resolution**: Configure required secrets for enabled services.

### E007: Provider-Service Compatibility Error

**Trigger**: Incompatible service provider for infrastructure provider

**Examples:**

```
E007: services.cert-manager: route53 DNS provider is not compatible with OpenStack infrastructure
E007: services.cert-manager: suggested providers for OpenStack: designate, cloudflare
E007: services.velero: aws provider is not compatible with OpenStack infrastructure
```

**Resolution**: Use compatible service providers or change infrastructure provider.

### E008: Deployment Method Compatibility Error

**Trigger**: Deployment method not supported for infrastructure provider

**Examples:**

```
E008: deployment.method: eks is not supported on OpenStack infrastructure
E008: supported deployment methods for OpenStack: kubespray, talos, kamaji, cluster-api
E008: deployment.method: gke is not supported on AWS infrastructure
```

**Resolution**: Use compatible deployment method for infrastructure provider.

### E009: Kamaji Control Plane Error

**Trigger**: Invalid Kamaji control plane configuration

**Examples:**

```
E009: deployment.kamaji.control_plane.replicas: must be odd number (1, 3, 5, 7), got 4
E009: deployment.kamaji.control_plane.etcd: storage_class required when datastore is etcd
E009: deployment.kamaji.control_plane.postgresql: host required when datastore is postgresql
```

**Resolution**: Fix Kamaji control plane configuration according to requirements.

### E010: Worker Pool Configuration Error

**Trigger**: Invalid worker pool configuration

**Examples:**

```
E010: deployment.kamaji.worker_pools[0]: bootstrap_provider must be "talos" when os is "talos"
E010: deployment.kamaji.worker_pools[1]: talos_version required when os is "talos"
E010: deployment.kamaji.worker_pools[2]: bootstrap_provider must be "kubeadm" when os is "ubuntu"
```

**Resolution**: Ensure worker pool OS and bootstrap provider are compatible.

### E011: Cluster API Provider Mismatch

**Trigger**: CAPI infrastructure provider doesn't match infrastructure provider

**Examples:**

```
E011: deployment.kamaji.cluster_api.providers.infrastructure: must match infrastructure.provider (expected "openstack", got "aws")
E011: deployment.kamaji.cluster_api.providers.infrastructure: must match infrastructure.provider (expected "aws", got "openstack")
```

**Resolution**: Ensure CAPI infrastructure provider matches infrastructure.provider.

### E012: Autoscaling Configuration Error

**Trigger**: Invalid autoscaling configuration

**Examples:**

```
E012: deployment.kamaji.worker_pools[0].autoscaling: count (5) must be between min_replicas (3) and max_replicas (4)
E012: deployment.kamaji.worker_pools[1].autoscaling: min_replicas (5) must be less than or equal to max_replicas (3)
```

**Resolution**: Ensure autoscaling configuration is valid (min <= count <= max).

### E013: Mixed OS Worker Pool Error

**Trigger**: Mixed OS worker pools without Kamaji or CAPI

**Examples:**

```
E013: cluster.kubernetes.additional_server_pools_worker: mixed OS worker pools require deployment method "kamaji" or "cluster-api"
E013: infrastructure.compute.additional_server_pools_worker: Windows workers require deployment method "kamaji" or "cluster-api"
```

**Resolution**: Use Kamaji or Cluster API deployment method for mixed OS worker pools.

### Error Aggregation

The validator collects all errors before returning, allowing you to fix multiple issues in one iteration:

```bash
opencenter cluster validate --config mycluster.yaml

# Output:
# Validation failed with 3 errors:
# 
# E001: infrastructure.networking.vrrp_ip: required field is missing
# E002: infrastructure.networking.subnet_nodes: invalid CIDR notation "10.2.128.0/33"
# E005: services.weave-gitops: requires services.fluxcd to be enabled
```


## Complete Examples

### Minimal OpenStack + Kubespray

**Purpose**: Simplest production-ready configuration

```yaml
schema_version: "2.0"

opencenter:
  meta:
    name: minimal-cluster
    organization: acme-corp
    region: sjc3
  
  cluster:
    cluster_name: minimal-cluster
    base_domain: acme-corp.com
    cluster_fqdn: minimal.acme-corp.com
    admin_email: ops@acme-corp.com
    kubernetes:
      version: "1.31.4"
      api_port: 6443
      subnet_pods: "10.233.64.0/18"
      subnet_services: "10.233.0.0/18"
      network_plugin:
        calico:
          enabled: true
      storage_plugin:
        cinder_csi:
          enabled: true
  
  infrastructure:
    provider: openstack
    os_version: "ubuntu-22.04"
    ssh:
      user: "ubuntu"
      key_path: "~/.ssh/cluster-key"
      authorized_keys:
        - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC..."
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.128.10"
      allocation_pool_end: "10.2.131.254"
      vrrp_ip: "10.2.128.5"
      vrrp_enabled: true
      loadbalancer_provider: "octavia"
      dns_zone_name: "acme-corp.com"
      dns_nameservers:
        - "8.8.8.8"
      ntp_servers:
        - "time.google.com"
    compute:
      flavor_bastion: "gp.0.2.2"
      flavor_master: "gp.0.4.4"
      flavor_worker: "gp.0.4.8"
      master_count: 3
      worker_count: 3
    storage:
      default_storage_class: "csi-cinder-sc-delete"
      worker_volume_size: 100
      worker_volume_destination_type: "volume"
      worker_volume_source_type: "image"
      worker_volume_type: "HA-Standard"
    cloud:
      openstack:
        auth_url: "https://keystone.sjc3.rackspace.com/v3/"
        region: "sjc3"
        application_credential_id: "app-cred-id"
        application_credential_secret: "app-cred-secret"
        tenant_name: "my-project"
        floating_network_id: "network-uuid"
  
  services:
    calico:
      enabled: true
    cinder-csi:
      enabled: true
    fluxcd:
      enabled: true

deployment:
  auto_deploy: true
  method: kubespray
  kubespray:
    version: "v2.29.1"

secrets:
  sops_age_key_file: "~/.config/opencenter/clusters/acme-corp/secrets/age/minimal-cluster-key.txt"
```

### Production OpenStack + Kamaji with Mixed OS

**Purpose**: Advanced configuration with hosted control plane and mixed OS worker pools

```yaml
schema_version: "2.0"

opencenter:
  meta:
    name: prod-kamaji-cluster
    organization: acme-corp
    env: production
    region: sjc3
    status: active
  
  cluster:
    cluster_name: prod-kamaji-cluster
    base_domain: acme-corp.com
    cluster_fqdn: prod-kamaji.acme-corp.com
    admin_email: ops@acme-corp.com
    kubernetes:
      version: "1.31.4"
      api_port: 6443
      kube_vip_enabled: false  # Not used with Kamaji
      subnet_pods: "10.233.64.0/18"
      subnet_services: "10.233.0.0/18"
      network_plugin:
        cilium:
          enabled: true
          operator_enabled: true
          kubeProxyReplacement: true
      storage_plugin:
        cinder_csi:
          enabled: true
      oidc:
        enabled: true
        kube_oidc_url: "https://keycloak.acme-corp.com/realms/kubernetes"
        kube_oidc_client_id: "kubernetes"
  
  infrastructure:
    provider: openstack
    os_version: "ubuntu-22.04"
    ssh:
      user: "ubuntu"
      key_path: "~/.ssh/prod-cluster-key"
      authorized_keys:
        - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC..."
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.128.10"
      allocation_pool_end: "10.2.131.254"
      vrrp_enabled: false  # Not used with Kamaji
      use_octavia: true
      loadbalancer_provider: "octavia"
      use_designate: true
      dns_zone_name: "acme-corp.com"
      dns_nameservers:
        - "8.8.8.8"
        - "8.8.4.4"
      ntp_servers:
        - "time.sjc3.rackspace.com"
    compute:
      flavor_bastion: "gp.0.2.2"
      master_count: 0  # Kamaji uses hosted control plane
      worker_count: 0  # Workers defined in deployment.kamaji.worker_pools
    storage:
      default_storage_class: "csi-cinder-sc-delete"
      worker_volume_size: 100
      worker_volume_destination_type: "volume"
      worker_volume_source_type: "image"
      worker_volume_type: "HA-Standard"
    cloud:
      openstack:
        auth_url: "https://keystone.sjc3.rackspace.com/v3/"
        region: "sjc3"
        application_credential_id: "app-cred-id"
        application_credential_secret: "app-cred-secret"
        tenant_name: "production-project"
        floating_network_id: "network-uuid"
        image_id: "799dcf97-3656-4361-8187-13ab1b295e33"
        availability_zones:
          - "az1"
          - "az2"
          - "az3"
  
  services:
    cilium:
      enabled: true
    cinder-csi:
      enabled: true
    fluxcd:
      enabled: true
    cert-manager:
      enabled: true
      email: "ops@acme-corp.com"
      dns_provider: "designate"
    kube-prometheus-stack:
      enabled: true
    loki:
      enabled: true
      storage_type: "s3"
      bucket_name: "prod-loki-logs"
    keycloak:
      enabled: true
    velero:
      enabled: true
      provider: "openstack"
      bucket_name: "prod-cluster-backups"

deployment:
  auto_deploy: true
  method: kamaji
  kamaji:
    enabled: true
    version: "v1.0.0"
    control_plane:
      replicas: 3
      datastore: etcd
      etcd:
        storage_class: "csi-cinder-sc-delete"
        storage_size: "10Gi"
      service_type: "LoadBalancer"
      api_server_port: 6443
      resources:
        requests:
          cpu: "500m"
          memory: "1Gi"
        limits:
          cpu: "2000m"
          memory: "4Gi"
    cluster_api:
      version: "v1.6.0"
      providers:
        infrastructure: openstack
        bootstrap: kubeadm
        control_plane: kubeadm
      openstack:
        cloud_name: "openstack"
        clouds_yaml_secret: "cloud-config"
    worker_pools:
      - name: "ubuntu-general"
        os: ubuntu
        count: 5
        flavor: "gp.0.4.8"
        image: "${infrastructure.cloud.openstack.image_id}"
        bootstrap_provider: kubeadm
        boot_volume:
          size: 100
          type: "HA-Standard"
        labels:
          workload: "general"
        autoscaling:
          enabled: true
          min_replicas: 3
          max_replicas: 10
      
      - name: "ubuntu-high-memory"
        os: ubuntu
        count: 2
        flavor: "gp.0.8.64"
        image: "${infrastructure.cloud.openstack.image_id}"
        bootstrap_provider: kubeadm
        boot_volume:
          size: 200
          type: "HA-Standard"
        additional_volumes:
          - device_name: "/dev/vdb"
            size: 500
            type: "HA-Standard"
        labels:
          workload: "memory-intensive"
        taints:
          - key: "workload"
            value: "memory-intensive"
            effect: "NoSchedule"
      
      - name: "talos-workers"
        os: talos
        count: 3
        flavor: "gp.0.4.8"
        image: "talos-v1.6.0"
        bootstrap_provider: talos
        talos_version: "v1.6.0"
        boot_volume:
          size: 50
          type: "HA-Standard"
        labels:
          workload: "secure"
        talos_config:
          install_disk: "/dev/sda"

secrets:
  sops_age_key_file: "~/.config/opencenter/clusters/acme-corp/secrets/age/prod-kamaji-cluster-key.txt"
```

### Reference Resolution Example

**Purpose**: Demonstrate reference resolution across domains

```yaml
schema_version: "2.0"

opencenter:
  meta:
    name: reference-demo
    organization: acme-corp
    region: sjc3
  
  cluster:
    cluster_name: "${meta.name}"  # Reference to meta.name
    base_domain: "acme-corp.com"
    cluster_fqdn: "${meta.name}.${cluster.base_domain}"  # Nested reference
    admin_email: "ops@acme-corp.com"
    kubernetes:
      version: "1.31.4"
      api_port: 6443
      subnet_pods: "10.233.64.0/18"
      subnet_services: "10.233.0.0/18"
      network_plugin:
        calico:
          enabled: true
      storage_plugin:
        cinder_csi:
          enabled: true
  
  infrastructure:
    provider: openstack
    os_version: "ubuntu-22.04"
    ssh:
      user: "ubuntu"
      key_path: "~/.ssh/${meta.name}-key"  # Reference in path
      authorized_keys:
        - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC..."
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.128.10"
      allocation_pool_end: "10.2.131.254"
      vrrp_ip: "10.2.128.5"
      vrrp_enabled: true
      loadbalancer_provider: "octavia"
      dns_zone_name: "${cluster.base_domain}"  # Reference to cluster domain
      dns_nameservers:
        - "8.8.8.8"
      ntp_servers:
        - "time.google.com"
    compute:
      flavor_bastion: "gp.0.2.2"
      flavor_master: "gp.0.4.4"
      flavor_worker: "gp.0.4.8"
      master_count: 3
      worker_count: 3
    storage:
      default_storage_class: "csi-cinder-sc-delete"
      worker_volume_size: 100
      worker_volume_destination_type: "volume"
      worker_volume_source_type: "image"
      worker_volume_type: "HA-Standard"
    cloud:
      openstack:
        auth_url: "https://keystone.sjc3.rackspace.com/v3/"
        region: "${meta.region}"  # Reference to meta.region
        application_credential_id: "app-cred-id"
        application_credential_secret: "app-cred-secret"
        tenant_name: "my-project"
        floating_network_id: "network-uuid"
  
  services:
    calico:
      enabled: true
      # Reference to infrastructure networking
      calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:${cluster.kubernetes.api_port}"
    cinder-csi:
      enabled: true
    fluxcd:
      enabled: true

deployment:
  auto_deploy: true
  method: kubespray
  kubespray:
    version: "v2.29.1"

secrets:
  sops_age_key_file: "~/.config/opencenter/clusters/${meta.organization}/secrets/age/${meta.name}-key.txt"
```

## Additional Resources

- [Migration Guide](migration-guide.md) - Complete v1 to v2 migration guide
- [v2 Configuration Examples](examples/v2/) - Additional example configurations
- [JSON Schema Documentation](../reference/json-schema.md) - Schema documentation and IDE integration
- [Validation Error Codes](validation-errors.md) - Detailed error code reference
- [Provider Documentation](../providers/) - Provider-specific documentation

---

**Last Updated**: January 2026  
**Schema Version**: v2.0  
**CLI Version**: 2.0.0+

