# Service Template Automation - Quick Summary

**Full Report:** [service-template-automation-report.md](./service-template-automation-report.md)

## Critical Findings

### Top 5 Hardcoded Values Blocking Automation

1. **Gateway Name/Namespace** (`rmpk-gateway` / `rackspace-system`)
   - Appears in: 10+ templates
   - Impact: Blocks multi-tenant deployments
   - Priority: HIGH

2. **MetalLB IP Ranges** (`172.23.0.6-172.23.0.8`)
   - Appears in: `metallb/ipaddresspool.yaml`
   - Impact: Requires manual editing per cluster
   - Priority: HIGH

3. **OIDC Configuration** (client_id: `opencenter`)
   - Appears in: 3+ SecurityPolicy resources
   - Impact: Scattered auth configuration
   - Priority: HIGH

4. **GitOps Secret Name** (`opencenter-base`)
   - Appears in: 20+ GitRepository sources
   - Impact: Limits organization flexibility
   - Priority: MEDIUM

5. **Certificate Issuer** (`letsencrypt-k8s-dev`)
   - Appears in: Gateway annotations
   - Impact: Can't switch between staging/prod
   - Priority: HIGH

## Services Needing Configuration Types

| Service | Current Config | Missing Fields | Priority |
|---------|---------------|----------------|----------|
| MetalLB | BaseServiceCfg only | IP pools, L2 config | HIGH |
| Gateway | BaseServiceCfg only | Name, namespace, listeners | HIGH |
| Harbor | BaseServiceCfg only | Storage, database, admin | MEDIUM |
| Longhorn | BaseServiceCfg only | Replicas, backup target | MEDIUM |
| OpenTelemetry | BaseServiceCfg only | Collectors, exporters | LOW |
| Cert-Manager | Partial | Region, DNS zones | HIGH |
| VSphere CSI | Partial | Storage classes | MEDIUM |
| Keycloak | Partial | Database, SMTP | LOW |

## Recommended New Config Sections

### Proposed Defaults (Backward Compatible)

```yaml
opencenter:
  # NEW: Gateway configuration
  gateway:
    name: rmpk-gateway                    # Default: rmpk-gateway
    namespace: rackspace-system           # Default: rackspace-system
    class_name: eg                        # Default: eg (Envoy Gateway)
    default_issuer: letsencrypt-prod      # Default: letsencrypt-{cluster_name}
  
  # NEW: OIDC configuration
  oidc:
    enabled: true                         # Default: true
    client_id: opencenter                 # Default: opencenter
    secret_name: gateway-oidc-secret      # Default: gateway-oidc-secret
    scopes:                               # Default: [openid, profile, email, roles]
      - openid
      - profile
      - email
      - roles
    logout_path: /logout                  # Default: /logout
  
  # ENHANCED: GitOps configuration
  gitops:
    secret_name: opencenter-base          # Default: opencenter-base (NEW)
    git_ops_base_repo: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git  # Default
    git_ops_base_release: ""              # Default: "" (use branch)
    git_ops_branch: main                  # Default: main
```

### Service-Specific Defaults

```yaml
services:
  # MetalLB - IP Address Pools
  metallb:
    enabled: false                        # Default: false
    namespace: metallb-system             # Default: metallb-system
    ip_address_pools:                     # Default: [] (must be configured)
      - name: default-pool                # Example configuration
        addresses:
          - 172.23.0.6-172.23.0.8
        auto_assign: true                 # Default: true
  
  # Cert-Manager - Enhanced Configuration
  cert-manager:
    enabled: true                         # Default: true
    namespace: cert-manager               # Default: cert-manager
    email: ""                             # Default: opencenter.cluster.admin_email
    region: ""                            # Default: opencenter.meta.region (for Route53)
    letsencrypt_server: https://acme-v02.api.letsencrypt.org/directory  # Default: production
    dns_zones:                            # Default: [opencenter.cluster.cluster_fqdn]
      - "*.example.com"
  
  # Gateway - Listener Configuration
  gateway:
    enabled: true                         # Default: true
    namespace: rackspace-system           # Default: rackspace-system
    gateway_name: rmpk-gateway            # Default: rmpk-gateway
    gateway_class: eg                     # Default: eg
    listeners:                            # Default: auto-generated from enabled services
      - name: keycloak-https
        port: 443
        protocol: HTTPS
        hostname: ""                      # Default: auth.{opencenter.cluster.cluster_fqdn}
        tls_secret_name: keycloak-tls     # Default: {service}-tls
      - name: grafana-https
        port: 443
        protocol: HTTPS
        hostname: ""                      # Default: grafana.{opencenter.cluster.cluster_fqdn}
        tls_secret_name: grafana-tls
  
  # VSphere CSI - Storage Classes
  vsphere-csi:
    enabled: false                        # Default: false (provider-specific)
    namespace: vmware-system-csi          # Default: vmware-system-csi
    storage_classes:                      # Default: [] (must be configured)
      - name: default-retain              # Example configuration
        datastore_url: "ds:///vmfs/volumes/datastore1"
        reclaim_policy: Retain            # Default: Retain
        allow_expansion: true             # Default: true
        volume_binding_mode: Immediate    # Default: Immediate
  
  # Harbor - Container Registry
  harbor:
    enabled: false                        # Default: false
    namespace: harbor                     # Default: harbor
    hostname: ""                          # Default: harbor.{opencenter.cluster.cluster_fqdn}
    storage_type: filesystem              # Default: filesystem
    registry_volume_size: 100             # Default: 100 (GB)
    database_type: internal               # Default: internal
  
  # Longhorn - Distributed Storage
  longhorn:
    enabled: false                        # Default: false
    namespace: longhorn-system            # Default: longhorn-system
    hostname: ""                          # Default: longhorn.{opencenter.cluster.cluster_fqdn}
    default_replica_count: 3              # Default: 3
    default_data_path: /var/lib/longhorn  # Default: /var/lib/longhorn
    storage_over_provisioning_percentage: 200  # Default: 200
    storage_minimal_available_percentage: 25   # Default: 25
  
  # Loki - Log Aggregation (Already Well-Defined)
  loki:
    enabled: true                         # Default: true
    namespace: observability              # Default: observability
    hostname: ""                          # Default: loki.{opencenter.cluster.cluster_fqdn}
    loki_storage_type: swift              # Default: swift
    loki_volume_size: 50                  # Default: 50 (GB)
    swift_auth_version: 3                 # Default: 3
  
  # Tempo - Distributed Tracing (Already Well-Defined)
  tempo:
    enabled: true                         # Default: true
    namespace: observability              # Default: observability
    hostname: ""                          # Default: tempo.{opencenter.cluster.cluster_fqdn}
    storage_type: swift                   # Default: swift
    volume_size: 50                       # Default: 50 (GB)
  
  # Kube-Prometheus-Stack (Already Well-Defined)
  kube-prometheus-stack:
    enabled: true                         # Default: true
    namespace: observability              # Default: observability
    hostname: ""                          # Default: prometheus.{opencenter.cluster.cluster_fqdn}
    grafana_volume_size: 10               # Default: 10 (GB)
    prometheus_volume_size: 50            # Default: 50 (GB)
    alertmanager_volume_size: 10          # Default: 10 (GB)
  
  # Keycloak - Identity Management (Already Partially Defined)
  keycloak:
    enabled: true                         # Default: true
    namespace: keycloak                   # Default: keycloak
    hostname: ""                          # Default: auth.{opencenter.cluster.cluster_fqdn}
    keycloak_realm: ""                    # Default: opencenter.meta.organization
    keycloak_client_id: opencenter        # Default: opencenter
  
  # Headlamp - Kubernetes Dashboard (Already Partially Defined)
  headlamp:
    enabled: true                         # Default: true
    namespace: headlamp                   # Default: headlamp
    hostname: ""                          # Default: headlamp.{opencenter.cluster.cluster_fqdn}
  
  # Velero - Backup and Restore (Already Partially Defined)
  velero:
    enabled: false                        # Default: false
    namespace: velero                     # Default: velero
    velero_region: ""                     # Default: opencenter.meta.region
  
  # OpenTelemetry Kube Stack
  opentelemetry-kube-stack:
    enabled: false                        # Default: false
    namespace: observability              # Default: observability
    collector_mode: deployment            # Default: deployment
    collector_replicas: 1                 # Default: 1
```

### Variable Substitution Strategy

**Priority Order:**
1. User-specified value in cluster config
2. Environment-specific default (dev/staging/prod)
3. Provider-specific default (OpenStack/AWS/vSphere)
4. Global default value

**Variable Substitution Patterns (v2 Schema):**

| Variable | v2 Schema Path | Example Value |
|----------|----------------|---------------|
| `{cluster_name}` | `opencenter.meta.name` | `prod-k8s-cluster` |
| `{cluster_fqdn}` | `opencenter.cluster.cluster_fqdn` | `prod.acme-corp.com` |
| `{base_domain}` | `opencenter.cluster.base_domain` | `acme-corp.com` |
| `{organization}` | `opencenter.meta.organization` | `acme-corp` |
| `{env}` | `opencenter.meta.env` | `production` |
| `{region}` | `opencenter.meta.region` | `sjc3` |
| `{admin_email}` | `opencenter.cluster.admin_email` | `ops@acme-corp.com` |
| `{provider}` | `opencenter.infrastructure.provider` | `openstack` |
| `{k8s_version}` | `opencenter.cluster.kubernetes.version` | `1.31.4` |
| `{api_port}` | `opencenter.cluster.kubernetes.api_port` | `6443` |
| `{vrrp_ip}` | `opencenter.infrastructure.networking.vrrp_ip` | `10.2.128.5` |

**Reference Syntax (v2):**

Templates can use Go template syntax with v2 schema paths:

```yaml
# Example: Gateway hostname using cluster FQDN
hostname: "{{ .OpenCenter.Cluster.ClusterFQDN }}"
# Resolves to: prod.acme-corp.com

# Example: Service hostname with subdomain
hostname: "grafana.{{ .OpenCenter.Cluster.ClusterFQDN }}"
# Resolves to: grafana.prod.acme-corp.com

# Example: Keycloak issuer URL
issuer: "https://auth.{{ .OpenCenter.Cluster.ClusterFQDN }}/realms/{{ .OpenCenter.Meta.Organization }}"
# Resolves to: https://auth.prod.acme-corp.com/realms/acme-corp

# Example: Cert-manager email
email: "{{ .OpenCenter.Cluster.AdminEmail }}"
# Resolves to: ops@acme-corp.com

# Example: API server endpoint
api_endpoint: "{{ .OpenCenter.Infrastructure.Networking.VrrpIP }}:{{ .OpenCenter.Cluster.Kubernetes.APIPort }}"
# Resolves to: 10.2.128.5:6443
```

**Template Context Structure (v2):**

```go
type TemplateContext struct {
    OpenCenter struct {
        Meta struct {
            Name         string  // opencenter.meta.name
            Organization string  // opencenter.meta.organization
            Env          string  // opencenter.meta.env
            Region       string  // opencenter.meta.region
            Status       string  // opencenter.meta.status
        }
        Cluster struct {
            ClusterName  string  // opencenter.cluster.cluster_name
            BaseDomain   string  // opencenter.cluster.base_domain
            ClusterFQDN  string  // opencenter.cluster.cluster_fqdn
            AdminEmail   string  // opencenter.cluster.admin_email
            Kubernetes struct {
                Version  string  // opencenter.cluster.kubernetes.version
                APIPort  int     // opencenter.cluster.kubernetes.api_port
            }
        }
        Infrastructure struct {
            Provider string  // opencenter.infrastructure.provider
            Networking struct {
                VrrpIP      string  // opencenter.infrastructure.networking.vrrp_ip
                DNSZoneName string  // opencenter.infrastructure.networking.dns_zone_name
            }
        }
        Services map[string]interface{}  // opencenter.services.*
        GitOps struct {
            GitDir            string  // opencenter.gitops.git_dir
            GitURL            string  // opencenter.gitops.git_url
            GitOpsBaseRepo    string  // opencenter.gitops.git_ops_base_repo
            GitOpsBaseRelease string  // opencenter.gitops.git_ops_base_release
            GitOpsBranch      string  // opencenter.gitops.git_ops_branch
        }
    }
    Secrets map[string]interface{}  // Secrets from SOPS
}
```

## Quick Wins (Immediate Implementation)

1. **Add Gateway Config Section**
   - Files to modify: 10+ templates
   - Effort: 2-3 hours
   - Impact: Enables multi-tenant deployments

2. **Add MetalLB Config Type**
   - Files to modify: `types_services.go`, `metallb/ipaddresspool.yaml`
   - Effort: 1 hour
   - Impact: Eliminates manual IP configuration

3. **Add OIDC Config Section**
   - Files to modify: 3 SecurityPolicy templates
   - Effort: 1 hour
   - Impact: Centralized authentication

4. **Add Cert-Manager Region Field**
   - Files to modify: `types_services.go`, `letsencrypt-issuer.yaml.tpl`
   - Effort: 30 minutes
   - Impact: Fixes Route53 DNS validation

## Implementation Approach

### Phase 1: Add with Defaults (Non-Breaking)
- Add new config fields
- Templates use config with fallback to current hardcoded values
- Existing clusters continue working

### Phase 2: Deprecation Warnings
- Warn when using default values
- Provide migration documentation
- Add migration tool

### Phase 3: Remove Hardcoded Values
- Make fields required
- Remove fallbacks
- Major version bump

## Expected Benefits

- **80% reduction** in post-deployment manual configuration
- **Zero-touch deployment** for standard configurations
- **Multi-tenant support** with organization-specific conventions
- **Environment flexibility** (dev/staging/prod variations)
- **Faster onboarding** for new clusters

## Next Actions

1. Review full report: `docs/dev/service-template-automation-report.md`
2. Prioritize configuration additions
3. Create implementation tickets
4. Update schema generator
5. Implement with backward compatibility
