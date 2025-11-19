# Kubespray Provider Guide

## Overview

The Kubespray provider in openCenter uses a three-layer architecture to deploy production-ready Kubernetes clusters on OpenStack infrastructure. This approach combines **Terraform** for infrastructure provisioning, **Kubespray** (Ansible-based) for Kubernetes installation, and **Calico** for network configuration.

## Architecture

### High-Level Flow

```mermaid
graph TB
    A[openCenter CLI] -->|Generates| B[main.tf]
    B -->|Defines| C[Infrastructure Module]
    B -->|Defines| D[Kubespray Module]
    B -->|Defines| E[Calico Module]
    
    C -->|Provisions| F[OpenStack Resources]
    F -->|Outputs| G[Node IPs & Network Info]
    
    G -->|Feeds Into| D
    D -->|Runs Ansible| H[Kubernetes Cluster]
    
    G -->|Feeds Into| E
    E -->|Configures| I[CNI & Network Policies]
    
    style A fill:#e1f5ff
    style B fill:#fff4e1
    style C fill:#ffe1e1
    style D fill:#e1ffe1
    style E fill:#f0e1ff
```

### Three-Module Architecture

The generated `main.tf` orchestrates three Terraform modules that work together:

1. **openstack-nova** - Infrastructure provisioning
2. **kubespray-cluster** - Kubernetes installation via Ansible
3. **calico** - CNI configuration and network policies

```mermaid
graph LR
    subgraph "Terraform Orchestration"
        A[main.tf] --> B[locals block]
        A --> C[module: openstack-nova]
        A --> D[module: kubespray-cluster]
        A --> E[module: calico]
    end
    
    subgraph "OpenStack Infrastructure"
        C --> F[Networks & Subnets]
        C --> G[Master Nodes]
        C --> H[Worker Nodes]
        C --> I[Bastion Host]
        C --> J[Floating IPs]
    end
    
    subgraph "Kubernetes Deployment"
        D --> K[Kubespray Ansible]
        K --> L[Control Plane]
        K --> M[Worker Nodes]
        K --> N[Kube-VIP]
    end
    
    subgraph "Network Configuration"
        E --> O[Calico CNI]
        E --> P[IP Pools]
        E --> Q[Network Policies]
    end
    
    C -.->|Outputs| D
    C -.->|Outputs| E
```

## Module Breakdown

### 1. OpenStack Infrastructure Module

**Purpose:** Provisions the underlying compute, network, and storage resources on OpenStack.

**Source:** `github.com/rackerlabs/openCenter-gitops-base.git//install/iac/terraform-openstack`

**Key Responsibilities:**
- Create virtual networks and subnets
- Provision master and worker VMs
- Configure security groups and firewall rules
- Allocate floating IPs for external access
- Set up bastion host for SSH access
- Configure VRRP for HA control plane (optional)
- Create Octavia load balancer (optional)

**Critical Outputs:**
```hcl
output "bastion_floating_ip"    # SSH entry point
output "master_nodes"           # Control plane node details
output "worker_nodes"           # Worker node details
output "k8s_api_ip"            # Kubernetes API endpoint
output "k8s_internal_ip"       # Internal cluster IP
```

**Network Architecture:**
```mermaid
graph TB
    subgraph "OpenStack Network"
        A[External Network<br/>PUBLICNET] -->|Floating IP| B[Router]
        B -->|NAT| C[Internal Network<br/>10.2.188.0/22]
        
        C --> D[Bastion<br/>.50]
        C --> E[Master 1<br/>.51]
        C --> F[Master 2<br/>.52]
        C --> G[Master 3<br/>.53]
        C --> H[Worker 1<br/>.54]
        C --> I[Worker 2<br/>.55]
        
        J[VRRP VIP<br/>.10] -.->|HA API| E
        J -.-> F
        J -.-> G
    end
    
    subgraph "Kubernetes Networks"
        K[Pod Network<br/>10.42.0.0/16]
        L[Service Network<br/>10.43.0.0/16]
    end
    
    style A fill:#e1f5ff
    style J fill:#ffe1e1
    style K fill:#e1ffe1
    style L fill:#f0e1ff
```

### 2. Kubespray Cluster Module

**Purpose:** Installs and configures Kubernetes using Ansible playbooks from the Kubespray project.

**Source:** `github.com/rackerlabs/openCenter-gitops-base.git//install/iac/kubespray`

**Key Responsibilities:**
- Generate Kubespray inventory from Terraform outputs
- Run Ansible playbooks to install Kubernetes
- Configure control plane components (API server, scheduler, controller-manager)
- Set up etcd cluster
- Install and configure kubelet on all nodes
- Deploy Kube-VIP for HA API endpoint
- Apply security hardening (CIS benchmarks)
- Configure OIDC authentication (optional)
- Rotate kubelet certificates

**Deployment Flow:**
```mermaid
sequenceDiagram
    participant TF as Terraform
    participant KB as Kubespray Module
    participant BS as Bastion Host
    participant MS as Master Nodes
    participant WK as Worker Nodes
    
    TF->>KB: Pass node IPs & SSH config
    KB->>KB: Generate Ansible inventory
    KB->>BS: SSH to bastion
    BS->>MS: Run control plane playbook
    MS->>MS: Install etcd, API server, scheduler
    BS->>MS: Configure Kube-VIP
    MS->>MS: Start HA virtual IP
    BS->>WK: Run worker node playbook
    WK->>WK: Install kubelet, kube-proxy
    WK->>MS: Join cluster
    KB->>TF: Return kubeconfig
```

**Key Configuration Options:**
- `kubernetes_version`: K8s version to install (e.g., "1.30.4")
- `kubespray_version`: Kubespray release tag (e.g., "v2.28.1")
- `network_plugin`: CNI plugin ("calico", "cilium", etc.)
- `kube_vip_enabled`: Enable HA virtual IP
- `k8s_hardening_enabled`: Apply CIS hardening
- `os_hardening_enabled`: Apply OS-level hardening

### 3. Calico Network Module

**Purpose:** Configures Calico CNI for pod networking and network policies.

**Source:** `github.com/rackerlabs/openCenter-gitops-base.git//install/iac/calico`

**Key Responsibilities:**
- Configure Calico IP pools for pod networking
- Set up BGP peering (if needed)
- Configure encapsulation (VXLAN, IPIP, or none)
- Enable NAT for outbound traffic
- Configure interface detection for multi-NIC nodes
- Support Windows worker nodes (optional)

**Network Encapsulation:**
```mermaid
graph TB
    subgraph "Calico Configuration"
        A[Interface Detection] -->|Auto or Manual| B[cni_iface: ens3]
        C[Encapsulation Type] --> D[VXLANCrossSubnet]
        E[NAT Outgoing] --> F[Enabled]
        G[IP Pool] --> H[10.42.0.0/16]
    end
    
    subgraph "Pod Communication"
        I[Pod A<br/>Node 1] -->|Same Subnet| J[Pod B<br/>Node 1]
        I -->|Different Subnet<br/>VXLAN Tunnel| K[Pod C<br/>Node 2]
    end
    
    subgraph "External Access"
        L[Pod] -->|NAT| M[Node IP]
        M --> N[External Network]
    end
```

**Encapsulation Modes:**
- **VXLANCrossSubnet**: VXLAN only for cross-subnet traffic (most efficient)
- **VXLAN**: Always use VXLAN encapsulation
- **IPIP**: IP-in-IP encapsulation (older, less efficient)
- **None**: Direct routing (requires BGP or cloud routing)

## Deployment Workflow

### Complete Deployment Sequence

```mermaid
sequenceDiagram
    participant User
    participant CLI as openCenter CLI
    participant TF as Terraform
    participant OS as OpenStack
    participant KB as Kubespray
    participant K8s as Kubernetes
    
    User->>CLI: openCenter cluster init
    CLI->>CLI: Generate config.yaml
    
    User->>CLI: openCenter cluster setup
    CLI->>CLI: Generate main.tf
    
    User->>CLI: openCenter cluster bootstrap
    CLI->>TF: terraform init
    CLI->>TF: terraform apply
    
    TF->>OS: Create networks
    TF->>OS: Create VMs
    OS-->>TF: Return IPs
    
    TF->>KB: Invoke kubespray module
    KB->>KB: Generate inventory
    KB->>K8s: Run Ansible playbooks
    K8s-->>KB: Cluster ready
    
    KB->>TF: Return kubeconfig
    TF-->>CLI: Deployment complete
    CLI-->>User: Cluster ready
```

### Step-by-Step Process

1. **Configuration Generation** (`openCenter cluster init`)
   - Creates `config.yaml` with cluster specifications
   - Sets provider to OpenStack
   - Defines node counts, flavors, and network settings

2. **GitOps Setup** (`openCenter cluster setup`)
   - Generates `main.tf` from templates
   - Creates GitOps repository structure
   - Renders Flux manifests

3. **Infrastructure Provisioning** (`openCenter cluster bootstrap`)
   - Runs `terraform init` to download providers
   - Executes `terraform apply` to create resources
   - Waits for VMs to be ready

4. **Kubernetes Installation** (automatic via Terraform)
   - Kubespray module generates Ansible inventory
   - Runs control plane installation playbook
   - Installs worker nodes
   - Configures networking

5. **Network Configuration** (automatic via Terraform)
   - Calico module applies CNI configuration
   - Sets up IP pools and routing
   - Enables network policies

## Key Configuration Parameters

### Infrastructure Settings

```yaml
# Node configuration
master_count: 3                    # 1 or 3 for HA
worker_count: 2                    # Scale as needed
flavor_master: "gp.0.4.4"         # 4 vCPU, 4GB RAM
flavor_worker: "gp.0.4.8"         # 4 vCPU, 8GB RAM

# Network configuration
subnet_nodes: "10.2.188.0/22"     # Node network
subnet_pods: "10.42.0.0/16"       # Pod network
subnet_services: "10.43.0.0/16"   # Service network
vrrp_ip: "10.2.188.10"            # HA VIP for API

# High availability
use_octavia: false                # Use Octavia LB
vrrp_enabled: true                # Use VRRP for HA
kube_vip_enabled: true            # Enable Kube-VIP
```

### Kubernetes Settings

```yaml
# Version control
kubernetes_version: "1.30.4"
kubespray_version: "v2.28.1"

# Networking
network_plugin: "calico"
cni_iface: "ens3"

# Security
k8s_hardening_enabled: true
os_hardening_enabled: true
kubelet_rotate_server_certificates: true
```

### Calico Settings

```yaml
# Interface detection
calico_interface_autodetect: false
cni_iface: "ens3"

# Encapsulation
calico_encapsulation_type: "VXLANCrossSubnet"
calico_nat_outgoing: true

# IP pool
calico_interface_autodetect_cidr: "10.2.188.0/22"
```

## High Availability Architecture

### Control Plane HA with VRRP

```mermaid
graph TB
    subgraph "External Access"
        A[Floating IP<br/>203.0.113.10]
    end
    
    subgraph "VRRP Virtual IP"
        B[VIP: 10.2.188.10<br/>Port 443]
    end
    
    subgraph "Control Plane Nodes"
        C[Master 1<br/>10.2.188.51<br/>VRRP Priority: 100]
        D[Master 2<br/>10.2.188.52<br/>VRRP Priority: 90]
        E[Master 3<br/>10.2.188.53<br/>VRRP Priority: 80]
    end
    
    A -->|NAT| B
    B -.->|Active| C
    B -.->|Standby| D
    B -.->|Standby| E
    
    C -->|etcd| F[(etcd cluster)]
    D -->|etcd| F
    E -->|etcd| F
    
    style C fill:#e1ffe1
    style D fill:#ffe1e1
    style E fill:#ffe1e1
```

**VRRP (Virtual Router Redundancy Protocol):**
- Creates a virtual IP that floats between master nodes
- Active master holds the VIP
- Automatic failover if active master fails
- All API requests go through the VIP

**Alternative: Octavia Load Balancer:**
```yaml
use_octavia: true
vrrp_enabled: false
```
- Uses OpenStack's native load balancer
- More robust but requires Octavia service
- Distributes load across all masters

## Security Features

### Hardening Options

**Kubernetes Hardening** (`k8s_hardening_enabled: true`):
- CIS Kubernetes Benchmark compliance
- Pod Security Standards enforcement
- Restricted pod security policies
- Audit logging enabled
- Anonymous auth disabled

**OS Hardening** (`os_hardening_enabled: true`):
- CIS OS Benchmark compliance
- Firewall rules (iptables/nftables)
- SSH hardening
- Kernel parameter tuning
- File permission restrictions

### Certificate Management

```yaml
kubelet_rotate_server_certificates: true
```
- Automatic certificate rotation
- Prevents certificate expiration issues
- Uses Kubernetes CSR API

### OIDC Authentication (Optional)

```yaml
kube_oidc_auth_enabled: true
kube_oidc_url: "https://auth.example.com"
kube_oidc_client_id: "kubernetes"
kube_oidc_username_claim: "email"
kube_oidc_groups_claim: "groups"
```

## Troubleshooting

### Common Issues

**1. Ansible Connection Failures**
```bash
# Check bastion connectivity
ssh -i ~/.ssh/id_rsa ubuntu@<bastion-ip>

# Verify node reachability from bastion
ssh ubuntu@<node-ip>
```

**2. VRRP Not Working**
```bash
# Check VRRP status on masters
sudo systemctl status keepalived

# Verify VIP assignment
ip addr show | grep <vrrp-ip>
```

**3. Calico Pods Not Starting**
```bash
# Check Calico status
kubectl get pods -n kube-system | grep calico

# View Calico node logs
kubectl logs -n kube-system <calico-node-pod>

# Verify interface detection
kubectl exec -n kube-system <calico-node-pod> -- ip addr
```

**4. Nodes Not Joining Cluster**
```bash
# Check kubelet status
sudo systemctl status kubelet

# View kubelet logs
sudo journalctl -u kubelet -f

# Verify API server connectivity
curl -k https://<vrrp-ip>:443/healthz
```

### Debug Mode

Enable Ansible verbose output:
```yaml
# In config.yaml
ansible_verbosity: 2  # 0-4, higher = more verbose
```

## Performance Tuning

### Node Sizing Recommendations

| Cluster Size | Master Flavor | Worker Flavor | Master Count |
|--------------|---------------|---------------|--------------|
| Dev/Test     | gp.0.2.2      | gp.0.2.4      | 1            |
| Small Prod   | gp.0.4.4      | gp.0.4.8      | 3            |
| Medium Prod  | gp.0.8.8      | gp.0.8.16     | 3            |
| Large Prod   | gp.0.16.16    | gp.0.16.32    | 3            |

### Network Performance

**Encapsulation Impact:**
- **None** (direct routing): Best performance, requires BGP
- **VXLANCrossSubnet**: Good balance, ~5% overhead
- **VXLAN**: Moderate overhead, ~10-15%
- **IPIP**: Higher overhead, ~15-20%

**MTU Considerations:**
```yaml
mtu: 1450  # Reduce for VXLAN to avoid fragmentation
```

## Migration and Upgrades

### Kubernetes Version Upgrades

```yaml
# Update config.yaml
kubernetes_version: "1.31.0"

# Re-run bootstrap
mise run cluster-bootstrap
```

Kubespray handles rolling upgrades automatically.

### Kubespray Version Upgrades

```yaml
# Update config.yaml
kubespray_version: "v2.29.0"

# Re-run bootstrap
mise run cluster-bootstrap
```

**Note:** Check Kubespray release notes for breaking changes.

## Best Practices

1. **Always use 3 masters** for production clusters
2. **Enable hardening** for security compliance
3. **Use VRRP or Octavia** for HA API endpoint
4. **Size workers appropriately** for workload
5. **Monitor etcd health** regularly
6. **Backup etcd** before upgrades
7. **Test in dev** before production changes
8. **Use GitOps** for all configuration changes
9. **Enable certificate rotation** to avoid expiration
10. **Document custom configurations** in Git

## Related Documentation

- [OpenStack Provider Configuration](../openstack/)
- [Secrets Management](../../secrets.md)
- [Cluster Lifecycle](../../reference/cluster/)
- [Troubleshooting Guide](../../TROUBLESHOOTING.md)

## External References

- [Kubespray Documentation](https://kubespray.io/)
- [Calico Documentation](https://docs.tigera.io/calico/latest/about/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [OpenStack Documentation](https://docs.openstack.org/)
