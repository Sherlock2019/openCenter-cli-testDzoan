# File Formats Reference

**Document Type:** Reference  
**Audience:** All users  
**Purpose:** Precise specifications for all file formats used in openCenter-cli

This document provides complete specifications for configuration files, schemas, templates, and other file formats used throughout openCenter-cli. All specifications are based on actual implementation.

---

## Table of Contents

1. [Cluster Configuration Files](#cluster-configuration-files)
2. [CLI Configuration File](#cli-configuration-file)
3. [JSON Schema Files](#json-schema-files)
4. [SOPS Configuration](#sops-configuration)
5. [GitOps Repository Structure](#gitops-repository-structure)
6. [Template Files](#template-files)
7. [SSH Key Files](#ssh-key-files)
8. [State Files](#state-files)

---

## Cluster Configuration Files

### File Location

Cluster configurations are stored in organization-based directory structures:

```
~/.config/openCenter/clusters/<organization>/<cluster>/.{cluster}-config.yaml
```

**Alternative locations** (backward compatibility):
- `~/.config/openCenter/clusters/<organization>/.{cluster}-config.yaml`
- `~/.config/openCenter/{cluster}.yaml` (legacy flat structure)

### File Format

**Format:** YAML  
**Encoding:** UTF-8  
**Schema Version:** 1.0.0  
**Extension:** `.yaml`

### Structure Overview

```yaml
schema_version: "1.0.0"
opencenter:
  meta: {}
  secrets: {}
  infrastructure: {}
  cluster: {}
  gitops: {}
  storage: {}
  talos: {}
  managed-service: {}
  services: {}
opentofu: {}
secrets: {}
networking: {}
deployment: {}
metadata: {}
overrides: {}
```

### Top-Level Sections

#### schema_version

**Type:** String  
**Required:** No  
**Default:** "1.0.0"  
**Pattern:** `^\d+\.\d+\.\d+$`

Semantic version of the configuration schema. Used for migration and compatibility checking.

**Example:**
```yaml
schema_version: "1.0.0"
```

#### opencenter

**Type:** Object  
**Required:** Yes

Main configuration section containing all cluster-specific settings.

**Subsections:**
- `meta` - Cluster metadata (name, environment, region, organization)
- `secrets` - Secrets backend configuration (Barbican)
- `infrastructure` - Cloud provider and infrastructure settings
- `cluster` - Kubernetes cluster configuration
- `gitops` - GitOps repository settings
- `storage` - Storage class and volume configuration
- `talos` - Talos Linux configuration (optional)
- `managed-service` - Rackspace-managed services
- `services` - Self-managed Kubernetes services

#### opentofu

**Type:** Object  
**Required:** Yes

OpenTofu/Terraform configuration for infrastructure provisioning.

**Fields:**
- `enabled` (boolean) - Enable OpenTofu provisioning
- `path` (string) - Path to OpenTofu working directory
- `backend` (object) - State backend configuration

#### secrets

**Type:** Object  
**Required:** Yes

Global secrets configuration including SOPS keys, SSH keys, and service-specific secrets.

#### metadata

**Type:** Object  
**Required:** No

Configuration lifecycle metadata.

**Fields:**
- `created_at` (string, RFC3339) - Creation timestamp
- `created_by` (string) - Creator username
- `updated_at` (string, RFC3339) - Last update timestamp
- `tags` (object) - Custom tags for categorization
- `annotations` (object) - Custom annotations

### opencenter.meta Section

Cluster metadata and identification.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Cluster name (3-63 chars, lowercase alphanumeric with hyphens) |
| `env` | string | No | Environment (dev, staging, prod) |
| `region` | string | Yes | Cloud provider region |
| `status` | string | No | Cluster status (pending, running, success, failed) |
| `stage` | string | No | Deployment stage (init, preflight, setup, bootstrap, etc.) |
| `organization` | string | Yes | Organization name for multi-tenancy |

**Example:**
```yaml
opencenter:
  meta:
    name: my-cluster
    env: production
    region: sjc3
    status: running
    organization: myorg
```

**Validation Rules:**
- `name`: Pattern `^[a-z0-9][a-z0-9-]*[a-z0-9]$`, length 3-63
- `region`: Must match provider's available regions
- `organization`: Pattern `^[a-z0-9][a-z0-9-]*[a-z0-9]$`

### opencenter.infrastructure Section

Infrastructure provider and cloud configuration.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `provider` | string | Yes | Provider type (openstack, aws, vmware, kind, baremetal) |
| `cloud` | object | Yes | Provider-specific cloud configuration |
| `ssh_user` | string | Yes | SSH username for node access |
| `os_version` | string | Yes | Operating system version |
| `server_group_affinity` | array | No | Server group affinity policies |
| `node_naming` | object | Yes | Node naming conventions |

**Example:**
```yaml
opencenter:
  infrastructure:
    provider: openstack
    ssh_user: ubuntu
    os_version: "24"
    server_group_affinity:
      - anti-affinity
    node_naming:
      worker: wn
      master: cp
      worker_windows: win
    cloud:
      openstack:
        auth_url: https://identity.example.com/v3
        region: RegionOne
        # ... additional fields
```

### opencenter.infrastructure.cloud.openstack Section

OpenStack-specific configuration.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `auth_url` | string | Yes | Keystone authentication URL |
| `region` | string | Yes | OpenStack region name |
| `application_credential_id` | string | Recommended | Application credential ID |
| `application_credential_secret` | string | Recommended | Application credential secret |
| `tenant_name` | string | Yes | Project/tenant name |
| `availability_zone` | string | Yes | Availability zone for instances |
| `project_domain_name` | string | Yes | Project domain name |
| `user_domain_name` | string | Yes | User domain name |
| `image_id` | string | Yes | Glance image UUID for Linux nodes |
| `image_id_windows` | string | No | Glance image UUID for Windows nodes |
| `insecure` | boolean | No | Skip TLS verification (not recommended) |
| `ca` | string | No | Custom CA certificate |
| `networking` | object | Yes | Networking configuration |

**Networking Subfields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `floating_ip_pool` | string | Yes | External network name for floating IPs |
| `floating_network_id` | string | No | External network UUID |
| `network_id` | string | No | Internal network UUID |
| `router_external_network_id` | string | Yes | Router external network UUID |
| `subnet_id` | string | No | Subnet UUID |
| `designate.dns_zone_name` | string | No | Designate DNS zone name |
| `vlan.id` | string | No | VLAN ID |
| `vlan.mtu` | integer | No | VLAN MTU size |
| `vlan.provider` | string | No | VLAN provider network |

**Example:**
```yaml
cloud:
  openstack:
    auth_url: https://identity.api.rackspacecloud.com/v3
    region: DFW3
    application_credential_id: abc123...
    application_credential_secret: secret123...
    tenant_name: my-project
    availability_zone: az1
    project_domain_name: rackspace_cloud_domain
    user_domain_name: rackspace_cloud_domain
    image_id: 799dcf97-3656-4361-8187-13ab1b295e33
    networking:
      floating_ip_pool: PUBLICNET
      router_external_network_id: 723f8fa2-dbf7-4cec-8d5f-017e62c12f79
      designate:
        dns_zone_name: example.com
```

### opencenter.cluster Section

Kubernetes cluster configuration.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `cluster_name` | string | Yes | Cluster name (must match meta.name) |
| `base_domain` | string | No | Base domain for cluster services |
| `cluster_fqdn` | string | No | Fully qualified domain name |
| `admin_email` | string | No | Administrator email for certificates |
| `ssh_authorized_keys` | array | Yes | SSH public keys for node access |
| `kubernetes` | object | Yes | Kubernetes-specific settings |
| `networking` | object | Yes | Cluster networking configuration |

**kubernetes Subfields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `version` | string | Yes | Kubernetes version (e.g., "1.33.5") |
| `kubespray_version` | string | Yes | Kubespray version tag |
| `api_port` | integer | Yes | Kubernetes API server port |
| `kube_vip_enabled` | boolean | No | Enable KubeVIP for HA |
| `master_count` | integer | Yes | Number of control plane nodes (1-9) |
| `worker_count` | integer | Yes | Number of worker nodes (0-100) |
| `worker_count_windows` | integer | No | Number of Windows workers (0-50) |
| `flavor_bastion` | string | Yes | Instance flavor for bastion |
| `flavor_master` | string | Yes | Instance flavor for masters |
| `flavor_worker` | string | Yes | Instance flavor for workers |
| `subnet_pods` | string | Yes | Pod network CIDR |
| `subnet_services` | string | Yes | Service network CIDR |
| `loadbalancer_provider` | string | Yes | LB provider (ovn, octavia, metallb, none) |
| `network_plugin` | object | Yes | CNI configuration |
| `oidc` | object | No | OIDC authentication settings |
| `security` | object | Yes | Security configuration |

**Example:**
```yaml
cluster:
  cluster_name: my-cluster
  base_domain: k8s.opencenter.cloud
  cluster_fqdn: my-cluster.sjc3.k8s.opencenter.cloud
  admin_email: admin@example.com
  ssh_authorized_keys:
    - ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... user@host
  kubernetes:
    version: "1.33.5"
    kubespray_version: v2.29.1
    api_port: 443
    kube_vip_enabled: true
    master_count: 3
    worker_count: 2
    flavor_bastion: gp.0.2.2
    flavor_master: gp.0.4.8
    flavor_worker: gp.0.4.16
    subnet_pods: 10.42.0.0/16
    subnet_services: 10.43.0.0/16
    loadbalancer_provider: ovn
```

### Network Plugin Configuration

Only one CNI plugin should be enabled at a time.

**Calico Configuration:**

```yaml
network_plugin:
  calico:
    enabled: true
    cni_iface: enp3s0
    calico_interface_autodetect: interface
    autodetect_cidr: ""
    encapsulation_type: VXLAN
    nat_outgoing: true
```

**Cilium Configuration:**

```yaml
network_plugin:
  cilium:
    enabled: true
    operator_enabled: true
    kubeProxyReplacement: true
```

**Kube-OVN Configuration:**

```yaml
network_plugin:
  kube-ovn:
    enabled: true
    cilium_integration: true
```

### Additional Server Pools

Define additional worker node pools with custom configurations.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Unique pool name (lowercase, alphanumeric, hyphens) |
| `worker_count` | integer | Yes | Number of nodes (0-100) |
| `flavor_worker` | string | Yes | Instance flavor |
| `node_worker` | string | Yes | Node suffix identifier |
| `server_group_affinity` | string | No | Affinity policy |
| `image_id` | string | No | Custom image UUID |
| `subnet_id` | string | No | Custom subnet UUID |

**Example:**
```yaml
kubernetes:
  additional_server_pools_worker:
    - name: gpu-pool
      worker_count: 2
      flavor_worker: gp.5.8.32
      node_worker: gpu
      server_group_affinity: anti-affinity
      image_id: abc123...
```

### opencenter.gitops Section

GitOps repository configuration.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `git_dir` | string | Yes | Local directory path for GitOps repo |
| `git_url` | string | No | Remote Git repository URL |
| `git_ssh_key` | string | No | Path to SSH private key |
| `git_ssh_pub` | string | No | Path to SSH public key |
| `git_branch` | string | No | Git branch (default: main) |
| `gitops_base_repo` | string | No | Base template repository URL |
| `gitops_base_release` | string | No | Base template release tag |
| `gitops_branch` | string | No | Base template branch |
| `flux` | object | No | FluxCD settings |

**flux Subfields:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `interval` | string | 15m | Reconciliation interval |
| `prune` | boolean | true | Enable resource pruning |

**Example:**
```yaml
gitops:
  git_dir: ./gitops-repo
  git_url: ssh://git@github.com/myorg/my-cluster-gitops.git
  git_branch: main
  gitops_base_repo: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
  gitops_base_release: v0.1.0
  flux:
    interval: 15m
    prune: true
```

### opencenter.services Section

Self-managed Kubernetes services configuration. Each service has a common structure.

**Common Service Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable/disable service |
| `status` | string | No | Deployment status |
| `release` | string | No | Release version tag |
| `branch` | string | No | Git branch (mutually exclusive with release) |
| `uri` | string | No | Custom Git repository URI |

**Available Services:**

- `calico` - Calico CNI service
- `cert-manager` - TLS certificate management
- `etcd-backup` - Automated etcd backups
- `external-snapshotter` - Volume snapshot controller
- `fluxcd` - GitOps continuous delivery
- `gateway` - Gateway API implementation
- `gateway-api` - Gateway API CRDs
- `headlamp` - Kubernetes dashboard
- `keycloak` - Identity and access management
- `kube-prometheus-stack` - Monitoring stack
- `kyverno` - Policy engine
- `loki` - Log aggregation
- `olm` - Operator Lifecycle Manager
- `openstack-ccm` - OpenStack cloud controller
- `openstack-csi` - OpenStack storage driver
- `postgres-operator` - PostgreSQL operator
- `rbac-manager` - RBAC management
- `sources` - Flux source controller
- `velero` - Backup and restore
- `vsphere-csi` - vSphere storage driver
- `weave-gitops` - GitOps UI

### Service-Specific Configuration Examples

**cert-manager:**

```yaml
services:
  cert-manager:
    enabled: true
    email: admin@example.com
    region: us-east-1
    letsencrypt_server: https://acme-v02.api.letsencrypt.org/directory
```

**kube-prometheus-stack:**

```yaml
services:
  kube-prometheus-stack:
    enabled: true
    prometheus_volume_size: 50
    prometheus_storage_class: csi-cinder-sc-delete
    grafana_volume_size: 10
    grafana_storage_class: csi-cinder-sc-delete
```

**loki:**

```yaml
services:
  loki:
    enabled: true
    volume_size: 20
    storage_class: csi-cinder-sc-delete
    bucket_name: my-cluster-loki
    swift_auth_url: https://keystone.api.dfw3.rackspacecloud.com/v3/
    swift_region: DFW3
    swift_domain_name: Default
```

### secrets Section

Global secrets configuration.

**Structure:**

```yaml
secrets:
  sops_age_key_file: /path/to/age/key.txt
  ssh_key:
    private: /path/to/ssh/key
    public: /path/to/ssh/key.pub
    cypher: ed25519
  global:
    aws:
      infrastructure:
        access_key: AKIA...
        secret_access_key: secret...
        region: us-east-1
      application:
        access_key: ""
        secret_access_key: ""
        region: ""
  cert_manager:
    aws_access_key: AKIA...
    aws_secret_access_key: secret...
  loki:
    swift_password: password...
  keycloak:
    client_secret: secret...
    admin_password: password...
  grafana:
    admin_password: password...
```

**Security Note:** All secrets should be encrypted with SOPS before committing to Git. Use environment variables for sensitive values:

```yaml
secrets:
  cert_manager:
    aws_access_key: ${CERT_MANAGER_AWS_KEY}
    aws_secret_access_key: ${CERT_MANAGER_AWS_SECRET}
```

### opentofu Section

OpenTofu/Terraform state backend configuration.

**Fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `enabled` | boolean | Yes | Enable OpenTofu provisioning |
| `path` | string | Yes | Working directory path |
| `backend` | object | Yes | State backend configuration |

**Backend Types:**

**Local Backend:**

```yaml
opentofu:
  enabled: true
  path: opentofu
  backend:
    type: local
    local:
      path: ./terraform.tfstate
```

**S3 Backend:**

```yaml
opentofu:
  enabled: true
  path: opentofu
  backend:
    type: s3
    s3:
      bucket: my-tfstate-bucket
      key: clusters/my-cluster/terraform.tfstate
      region: us-east-1
      endpoint: https://s3.amazonaws.com
      encrypt: true
```

---

## CLI Configuration File

### File Location

```
~/.config/openCenter/config.yaml
```

### File Format

**Format:** YAML  
**Encoding:** UTF-8  
**Extension:** `.yaml`

### Structure

```yaml
version: "1.0.0"
logging:
  level: warn
  format: text
  output: stderr
  file: ""
behavior:
  dry_run: false
  verbose: false
  auto_approve: false
  color: true
paths:
  clusters_dir: ~/.config/openCenter/clusters
  plugins_dir: ~/.config/openCenter/plugins
  cache_dir: ~/.cache/openCenter
defaults:
  provider: openstack
  region: sjc3
  organization: opencenter
active_cluster: ""
```

**Field Descriptions:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `version` | string | 1.0.0 | CLI config version |
| `logging.level` | string | warn | Log level (debug, info, warn, error) |
| `logging.format` | string | text | Log format (text, json) |
| `logging.output` | string | stderr | Output destination |
| `behavior.dry_run` | boolean | false | Enable dry-run mode globally |
| `behavior.verbose` | boolean | false | Enable verbose output |
| `behavior.auto_approve` | boolean | false | Skip confirmation prompts |
| `behavior.color` | boolean | true | Enable colored output |
| `paths.clusters_dir` | string | (auto) | Clusters directory path |
| `active_cluster` | string | "" | Currently active cluster |

---

## JSON Schema Files

### File Location

```
schema/cluster.schema.json
```

Generated via: `openCenter cluster schema --output schema/cluster.schema.json`

### File Format

**Format:** JSON  
**Encoding:** UTF-8  
**Schema Standard:** JSON Schema Draft 2020-12  
**Extension:** `.json`

### Structure

```json
{
  "$id": "https://opencenter.cloud/schemas/cluster-config.json",
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "description": "Complete schema for openCenter cluster configuration",
  "properties": {
    "schema_version": {},
    "opencenter": {},
    "opentofu": {},
    "secrets": {},
    "metadata": {}
  },
  "required": ["opencenter", "opentofu", "secrets"]
}
```

### Usage

**IDE Integration:**

Add to cluster configuration file:

```yaml
# yaml-language-server: $schema=https://opencenter.cloud/schemas/cluster-config.json
schema_version: "1.0.0"
opencenter:
  # ...
```

**Validation:**

```bash
mise run validate my-cluster
```

---

## SOPS Configuration

### File Location

```
<gitops-repo>/.sops.yaml
<gitops-repo>/overlays/<cluster>/.sops.yaml
```

### File Format

**Format:** YAML  
**Encoding:** UTF-8  
**Extension:** `.yaml`

### Structure

```yaml
creation_rules:
  - path_regex: .*\.(yaml|yml)$
    age: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
    encrypted_regex: '^(data|stringData|password|token|key|secret|credentials)'
  
  - path_regex: secrets/openstack-credentials\.yaml$
    age: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
  
  - path_regex: customer-managed/services/.*/secret\.yaml$
    age: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

**Field Descriptions:**

| Field | Type | Description |
|-------|------|-------------|
| `path_regex` | string | Regex pattern matching file paths |
| `age` | string | Age public key for encryption |
| `encrypted_regex` | string | Regex matching fields to encrypt |
| `pgp` | array | PGP key fingerprints (alternative to age) |
| `kms` | array | Cloud KMS keys (AWS, GCP, Azure) |

### Encrypted File Format

SOPS-encrypted files contain both encrypted data and metadata:

```yaml
apiVersion: v1
kind: Secret
metadata:
    name: my-secret
type: Opaque
data:
    password: ENC[AES256_GCM,data:encrypted_value,iv:...,tag:...,type:str]
sops:
    age:
        - recipient: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
          enc: |
            -----BEGIN AGE ENCRYPTED FILE-----
            ...
            -----END AGE ENCRYPTED FILE-----
    lastmodified: "2024-01-15T10:30:00Z"
    mac: ENC[AES256_GCM,data:...,iv:...,tag:...,type:str]
    version: 3.8.1
```

---

## GitOps Repository Structure

### Directory Layout

```
<gitops-repo>/
├── .sops.yaml                          # SOPS encryption configuration
├── README.md                           # Repository documentation
├── <cluster>/                          # Cluster-specific directory
│   ├── infrastructure/
│   │   └── clusters/
│   │       └── <cluster>/
│   │           ├── flux-system/
│   │           │   ├── gotk-components.yaml
│   │           │   ├── gotk-sync.yaml (encrypted)
│   │           │   └── kustomization.yaml
│   │           ├── secrets/
│   │           │   └── openstack-credentials.yaml (encrypted)
│   │           ├── managed-services/
│   │           │   ├── sources/
│   │           │   │   └── base-repo.yaml (encrypted)
│   │           │   └── kustomization.yaml
│   │           └── customer-managed/
│   │               ├── services/
│   │               │   ├── cert-manager/
│   │               │   ├── kube-prometheus-stack/
│   │               │   └── ...
│   │               └── kustomization.yaml
│   └── secrets/
│       ├── age/
│       │   └── keys/
│       │       └── <cluster>-key.txt
│       └── ssh/
│           ├── <cluster>
│           └── <cluster>.pub
```

### Key Files

#### flux-system/gotk-sync.yaml

**Purpose:** FluxCD GitRepository and Kustomization resources  
**Encryption:** Required (contains Git credentials)  
**Format:** YAML

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 1m0s
  ref:
    branch: main
  secretRef:
    name: flux-system
  url: ssh://git@github.com/myorg/my-cluster-gitops
---
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: flux-system
  namespace: flux-system
spec:
  interval: 15m0s
  path: ./my-cluster/infrastructure/clusters/my-cluster
  prune: true
  sourceRef:
    kind: GitRepository
    name: flux-system
```

#### secrets/openstack-credentials.yaml

**Purpose:** OpenStack cloud credentials for CCM and CSI  
**Encryption:** Required  
**Format:** YAML

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: openstack-credentials
  namespace: kube-system
type: Opaque
stringData:
  clouds.yaml: |
    clouds:
      openstack:
        auth:
          auth_url: https://identity.example.com/v3
          application_credential_id: abc123...
          application_credential_secret: secret123...
        region_name: RegionOne
```

#### managed-services/sources/base-repo.yaml

**Purpose:** FluxCD source for base GitOps repository  
**Encryption:** Required (may contain credentials)  
**Format:** YAML

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: openCenter-gitops-base
  namespace: flux-system
spec:
  interval: 15m
  ref:
    tag: v0.1.0
  url: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
```

---

## Template Files

### File Location

```
internal/gitops/templates/
internal/gitops/gitops-base-dir/
```

### File Format

**Format:** Go text/template  
**Encoding:** UTF-8  
**Extension:** `.yaml`, `.tmpl`

### Template Syntax

Templates use Go's `text/template` syntax with Sprig functions.

**Basic Syntax:**

```yaml
# Variable interpolation
cluster_name: {{ .Config.OpenCenter.Cluster.ClusterName }}

# Conditionals
{{- if .Config.OpenCenter.Infrastructure.Provider | eq "openstack" }}
provider: openstack
{{- end }}

# Loops
{{- range .Config.OpenCenter.Cluster.SSHAuthorizedKeys }}
  - {{ . }}
{{- end }}

# Sprig functions
region: {{ .Config.OpenCenter.Meta.Region | upper }}
timestamp: {{ now | date "2006-01-02T15:04:05Z07:00" }}
```

### Available Template Context

**Top-Level Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `.Config` | Config | Complete cluster configuration |
| `.ClusterName` | string | Cluster name (shortcut) |
| `.Organization` | string | Organization name |
| `.Region` | string | Cloud region |
| `.Provider` | string | Infrastructure provider |

**Nested Access:**

```yaml
# OpenCenter configuration
{{ .Config.OpenCenter.Meta.Name }}
{{ .Config.OpenCenter.Infrastructure.Provider }}
{{ .Config.OpenCenter.Cluster.Kubernetes.Version }}

# GitOps configuration
{{ .Config.OpenCenter.GitOps.GitURL }}
{{ .Config.OpenCenter.GitOps.GitBranch }}

# Secrets (use with caution)
{{ .Config.Secrets.SopsAgeKeyFile }}
```

### Common Sprig Functions

**String Functions:**

```yaml
# Case conversion
{{ .ClusterName | upper }}
{{ .ClusterName | lower }}
{{ .ClusterName | title }}

# String manipulation
{{ .ClusterName | replace "-" "_" }}
{{ .ClusterName | trim }}
{{ .ClusterName | trunc 10 }}

# Encoding
{{ .Secret | b64enc }}
{{ .EncodedSecret | b64dec }}
```

**Logic Functions:**

```yaml
# Conditionals
{{ if eq .Provider "openstack" }}openstack{{ else }}other{{ end }}
{{ if ne .Count 0 }}enabled{{ end }}
{{ if and .Enabled .Ready }}active{{ end }}

# Default values
{{ .OptionalField | default "default-value" }}
{{ .EmptyString | empty }}
```

**List Functions:**

```yaml
# List operations
{{ .List | first }}
{{ .List | last }}
{{ .List | join "," }}
{{ .List | sortAlpha }}
```

### Template Example

**File:** `infrastructure/clusters/cluster-template.yaml`

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: cluster-info
  namespace: kube-system
data:
  cluster-name: {{ .Config.OpenCenter.Cluster.ClusterName }}
  cluster-fqdn: {{ .Config.OpenCenter.Cluster.ClusterFQDN }}
  provider: {{ .Config.OpenCenter.Infrastructure.Provider }}
  region: {{ .Config.OpenCenter.Meta.Region }}
  kubernetes-version: {{ .Config.OpenCenter.Cluster.Kubernetes.Version }}
  created-at: {{ now | date "2006-01-02T15:04:05Z07:00" }}
  
  {{- if .Config.OpenCenter.Infrastructure.Provider | eq "openstack" }}
  openstack-region: {{ .Config.OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
  openstack-az: {{ .Config.OpenCenter.Infrastructure.Cloud.OpenStack.AvailabilityZone }}
  {{- end }}
  
  master-count: {{ .Config.OpenCenter.Cluster.Kubernetes.MasterCount | quote }}
  worker-count: {{ .Config.OpenCenter.Cluster.Kubernetes.WorkerCount | quote }}
  
  ssh-keys: |
    {{- range .Config.OpenCenter.Cluster.SSHAuthorizedKeys }}
    {{ . }}
    {{- end }}
```

---

## SSH Key Files

### File Location

```
~/.config/openCenter/clusters/<organization>/<cluster>/secrets/ssh/<cluster>
~/.config/openCenter/clusters/<organization>/<cluster>/secrets/ssh/<cluster>.pub
```

### File Format

**Private Key:**
- **Format:** OpenSSH private key format
- **Encoding:** PEM (ASCII armor)
- **Extension:** None (no extension)
- **Permissions:** 0600 (read/write owner only)

**Public Key:**
- **Format:** OpenSSH public key format
- **Encoding:** ASCII
- **Extension:** `.pub`
- **Permissions:** 0644 (readable by all)

### Supported Key Types

| Type | Cypher | Key Size | Recommended |
|------|--------|----------|-------------|
| Ed25519 | ed25519 | 256-bit | ✅ Yes (default) |
| RSA | rsa | 4096-bit | ⚠️ Acceptable |
| ECDSA | ecdsa-sha2-nistp256 | 256-bit | ⚠️ Acceptable |

### Private Key Format

```
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACBK8...
-----END OPENSSH PRIVATE KEY-----
```

### Public Key Format

```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEr... user@hostname
```

**Components:**
1. Key type (`ssh-ed25519`, `ssh-rsa`, `ecdsa-sha2-nistp256`)
2. Base64-encoded public key
3. Comment (typically `user@hostname`)

---

## Age Key Files

### File Location

```
~/.config/openCenter/clusters/<organization>/<cluster>/secrets/age/keys/<cluster>-key.txt
```

### File Format

**Format:** Age private key format  
**Encoding:** ASCII (Bech32)  
**Extension:** `.txt`  
**Permissions:** 0600 (read/write owner only)

### Structure

```
# created: 2024-01-15T10:30:00Z
# public key: age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
AGE-SECRET-KEY-1XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

**Lines:**
1. Creation timestamp (comment)
2. Corresponding public key (comment)
3. Private key (Bech32-encoded)

### Public Key Derivation

Public keys are derived from private keys and use Bech32 encoding:

```
age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

**Format:** `age1` prefix + 58 characters (Bech32)

---

## State Files

### OpenTofu State Files

#### File Location

**Local Backend:**
```
<gitops-repo>/terraform.tfstate
```

**S3 Backend:**
```
s3://<bucket>/<key-path>/terraform.tfstate
```

#### File Format

**Format:** JSON  
**Encoding:** UTF-8  
**Extension:** `.tfstate`

#### Structure

```json
{
  "version": 4,
  "terraform_version": "1.6.0",
  "serial": 1,
  "lineage": "abc123-def456-...",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "openstack_compute_instance_v2",
      "name": "master",
      "provider": "provider[\"registry.opentofu.org/terraform-provider-openstack/openstack\"]",
      "instances": []
    }
  ]
}
```

**Security Note:** State files contain sensitive data. Always use:
- Remote backends (S3, Swift) with encryption
- Access controls and authentication
- State locking to prevent concurrent modifications

### Cluster State Tracking

#### File Location

```
~/.config/openCenter/clusters/<organization>/<cluster>/.state
```

#### File Format

**Format:** JSON  
**Encoding:** UTF-8  
**Extension:** `.state` (hidden file)

#### Structure

```json
{
  "cluster_name": "my-cluster",
  "organization": "myorg",
  "current_stage": "bootstrap",
  "status": "running",
  "last_updated": "2024-01-15T10:30:00Z",
  "stages": {
    "init": {
      "status": "completed",
      "timestamp": "2024-01-15T09:00:00Z"
    },
    "validate": {
      "status": "completed",
      "timestamp": "2024-01-15T09:15:00Z"
    },
    "preflight": {
      "status": "completed",
      "timestamp": "2024-01-15T09:30:00Z"
    },
    "setup": {
      "status": "completed",
      "timestamp": "2024-01-15T10:00:00Z"
    },
    "bootstrap": {
      "status": "in_progress",
      "timestamp": "2024-01-15T10:15:00Z"
    }
  }
}
```

---

## Validation Rules

### Cluster Name Validation

**Pattern:** `^[a-z0-9][a-z0-9-]*[a-z0-9]$`  
**Length:** 3-63 characters  
**Rules:**
- Must start with alphanumeric character
- Must end with alphanumeric character
- May contain hyphens in the middle
- Lowercase only
- No path separators (`/`, `\`)
- No relative path components (`.`, `..`)

**Valid Examples:**
- `my-cluster`
- `prod-k8s-01`
- `dev-cluster-2024`

**Invalid Examples:**
- `My-Cluster` (uppercase)
- `-my-cluster` (starts with hyphen)
- `my-cluster-` (ends with hyphen)
- `my_cluster` (underscore not allowed)
- `my/cluster` (path separator)

### CIDR Validation

**Pattern:** `^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`  
**Rules:**
- Valid IPv4 address
- Valid subnet mask (0-32)
- No overlapping with reserved ranges (unless intentional)

**Examples:**
- `10.42.0.0/16` (pod network)
- `10.43.0.0/16` (service network)
- `192.168.1.0/24` (node network)

### UUID Validation

**Pattern:** `^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`  
**Format:** 8-4-4-4-12 hexadecimal groups  
**Case:** Lowercase

**Example:**
```
799dcf97-3656-4361-8187-13ab1b295e33
```

### SSH Public Key Validation

**Pattern:** `^(ssh-rsa|ssh-ed25519|ecdsa-sha2-nistp256|ecdsa-sha2-nistp384|ecdsa-sha2-nistp521) `  
**Minimum Length:** 100 characters  
**Format:** `<type> <base64-key> [comment]`

**Example:**
```
ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIEr8... user@hostname
```

### URL Validation

**Git URL Pattern:** `^(https?://|git@|ssh://)`  
**HTTP URL Pattern:** `^https?://`

**Valid Git URLs:**
- `https://github.com/myorg/repo.git`
- `git@github.com:myorg/repo.git`
- `ssh://git@github.com/myorg/repo.git`

### Email Validation

**Format:** RFC 5322 compliant  
**Pattern:** Standard email format validation

**Example:**
```
admin@example.com
```

---

## Environment Variable Expansion

Configuration files support environment variable expansion using `${VAR}` or `$VAR` syntax.

### Syntax

```yaml
# Curly brace syntax (recommended)
password: ${DATABASE_PASSWORD}

# Direct reference
api_key: $API_KEY

# With default value (shell-style, not supported)
# Use empty string check in validation instead
token: ${AUTH_TOKEN}
```

### Use Cases

**Secrets Management:**

```yaml
secrets:
  cert_manager:
    aws_access_key: ${CERT_MANAGER_AWS_KEY}
    aws_secret_access_key: ${CERT_MANAGER_AWS_SECRET}
  
  keycloak:
    admin_password: ${KEYCLOAK_ADMIN_PASSWORD}
    client_secret: ${KEYCLOAK_CLIENT_SECRET}
```

**Dynamic Configuration:**

```yaml
opencenter:
  infrastructure:
    cloud:
      openstack:
        auth_url: ${OPENSTACK_AUTH_URL}
        region: ${OPENSTACK_REGION}
```

### Security Best Practices

1. **Never commit plaintext secrets** to Git
2. **Use environment variables** for sensitive values
3. **Encrypt with SOPS** before committing
4. **Use secret management tools** (Vault, AWS Secrets Manager)
5. **Rotate credentials regularly**
6. **Audit secret access** via logging

---

## File Permissions

### Security-Sensitive Files

| File Type | Permissions | Owner | Description |
|-----------|-------------|-------|-------------|
| Private SSH keys | 0600 | User | Read/write by owner only |
| Age private keys | 0600 | User | Read/write by owner only |
| SOPS config | 0644 | User | Readable by all (no secrets) |
| Cluster config | 0644 | User | Readable by all (encrypted secrets) |
| CLI config | 0644 | User | Readable by all |
| State files | 0644 | User | Readable by all |
| Public SSH keys | 0644 | User | Readable by all |
| GitOps manifests | 0644 | User | Readable by all |

### Directory Permissions

| Directory | Permissions | Description |
|-----------|-------------|-------------|
| `~/.config/openCenter/` | 0755 | Config root directory |
| `clusters/<org>/<cluster>/` | 0755 | Cluster directory |
| `secrets/age/keys/` | 0700 | Age keys directory (restricted) |
| `secrets/ssh/` | 0700 | SSH keys directory (restricted) |
| `plugins/` | 0755 | Plugins directory |
| `cache/` | 0755 | Cache directory |

---

## Character Encoding

All text files use **UTF-8 encoding** without BOM (Byte Order Mark).

### Line Endings

- **Unix/Linux/macOS:** LF (`\n`)
- **Windows:** CRLF (`\r\n`) automatically converted to LF
- **Git:** Configured via `.gitattributes` for consistent line endings

### Special Characters

**Allowed in Configuration:**
- UTF-8 characters in comments and string values
- YAML escape sequences (`\n`, `\t`, `\"`, etc.)
- Unicode characters in descriptions and metadata

**Restricted in Identifiers:**
- Cluster names: ASCII lowercase alphanumeric and hyphens only
- Organization names: ASCII lowercase alphanumeric and hyphens only
- Field names: ASCII as per YAML specification

---

## File Size Limits

### Practical Limits

| File Type | Typical Size | Maximum Size | Notes |
|-----------|--------------|--------------|-------|
| Cluster config | 5-50 KB | 1 MB | Larger configs indicate complexity issues |
| CLI config | 1-5 KB | 100 KB | Should remain small |
| JSON schema | 50-200 KB | 10 MB | Generated file |
| SOPS config | 1-5 KB | 100 KB | Simple rules only |
| Template files | 1-10 KB | 1 MB | Per template |
| State files | 10-500 KB | 100 MB | Grows with infrastructure |
| SSH private key | 1-4 KB | 10 KB | Depends on key type |
| Age private key | <1 KB | 10 KB | Fixed size |

### Performance Considerations

- **Large configurations** (>100 KB) may indicate over-complexity
- **Template rendering** slows with large data structures
- **State files** should be stored remotely for large infrastructures
- **Git repositories** perform better with smaller files

---

## Backward Compatibility

### Configuration Migration

openCenter-cli automatically migrates configurations between schema versions.

**Migration Process:**

1. Detect schema version from `schema_version` field
2. Compare with current schema version
3. Apply migration transformations
4. Update `schema_version` field
5. Preserve metadata and comments where possible

**Supported Migrations:**

- Schema 0.x → 1.0.0 (initial release migration)
- Future versions will maintain backward compatibility

### Legacy File Locations

**Supported for backward compatibility:**

```
# Legacy flat structure
~/.config/openCenter/{cluster}.yaml

# Legacy cluster directory
~/.config/openCenter/clusters/{cluster}/.{cluster}-config.yaml

# Current organization-based structure
~/.config/openCenter/clusters/{organization}/{cluster}/.{cluster}-config.yaml
```

**Resolution Order:**

1. Organization-based path (if organization specified)
2. Search all organizations (if no organization specified)
3. Legacy cluster directory
4. Legacy flat file

---

## See Also

- [Glossary](glossary.md) - Terminology reference
- [CLI Reference](../reference/cli.md) - Command-line interface
- [Configuration Guide](../guides/configuration.md) - Configuration best practices
- [SOPS Documentation](https://github.com/mozilla/sops) - External SOPS reference
- [JSON Schema Specification](https://json-schema.org/) - JSON Schema standard
- [YAML Specification](https://yaml.org/spec/) - YAML format specification
- [Age Encryption](https://age-encryption.org/) - Age encryption tool
