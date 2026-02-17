# Design Document: TODO Completion Phase 1

## Overview

This design addresses 12 outstanding TODO items (4-15) in the openCenter-cli codebase, organized into 8 logical epics. The implementation focuses on completing deferred work across multiple subsystems: Terraform state integration for drift detection, OpenStack provider enhancements, secrets management automation, configuration validation, audit logging, drift detection callbacks, and Keycloak backup automation.

The design prioritizes Epic 1 (Terraform State Integration) as the foundational piece that enables accurate drift detection. All other epics build upon or complement this foundation.

## Architecture

### System Context

The openCenter-cli manages Kubernetes cluster lifecycle through a GitOps workflow:

1. User creates cluster configuration (YAML)
2. CLI renders Terraform configuration (`main.tf`)
3. Terraform provisions infrastructure and saves state (`terraform.tfstate`)
4. CLI detects drift by comparing Terraform state (desired) with cloud provider APIs (actual)

The Terraform state file is the authoritative source of truth for desired infrastructure state, not the YAML configuration. This is because:
- YAML is high-level (node counts, flavors)
- Terraform translates to low-level resources (security groups, load balancers, volumes)
- Terraform state contains actual resource IDs and computed values
- Terraform state includes resource relationships (volume attachments, floating IP associations)

### Component Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    Drift Detection System                       │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  cmd/cluster_drift.go                                    │   │
│  │  - CLI commands (detect, reconcile, schedule)            │   │
│  │  - User interaction and output formatting                │   │
│  └──────────────────────────────────────────────────────────┘   │
│                          │                                      │
│                          ▼                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Terraform State Reader (NEW)                            │   │
│  │  - Parse terraform.tfstate JSON                          │   │
│  │  - Map Terraform resources to InfrastructureState        │   │
│  │  - Extract attributes and relationships                  │   │
│  └──────────────────────────────────────────────────────────┘   │
│                          │                                      │
│                          ▼                                      │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  internal/cloud/factory.go                               │   │
│  │  - CloudProvider interface                               │   │
│  │  - InfrastructureState types                             │   │
│  │  - DriftReport types                                     │   │
│  └──────────────────────────────────────────────────────────┘   │
│                          │                                      │
│           ┌──────────────┴──────────────┐                       │
│           ▼                             ▼                       │
│  ┌─────────────────────┐       ┌─────────────────────┐          │
│  │ OpenStack Provider  │       │   AWS Provider      │          │
│  │ (ENHANCED)          │       │                     │          │
│  │ - Security Groups   │       │ - EC2 instances     │          │
│  │ - Load Balancers    │       │ - VPCs              │          │
│  │ - Volumes           │       │ - Security Groups   │          │
│  │ - Floating IPs      │       │ - ELBs              │          │
│  └─────────────────────┘       └─────────────────────┘          │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```


### Data Flow

```
User Configuration (YAML)
    ↓
Terraform Rendering (opencenter cluster setup)
    ↓
Infrastructure Provisioning (terraform apply)
    ↓
Terraform State File (terraform.tfstate)
    ↓
Drift Detection (opencenter cluster drift detect)
    ├─ Read Terraform State → Desired State
    ├─ Query Cloud Provider API → Actual State
    └─ Compare → Drift Report
    ↓
Drift Reconciliation (opencenter cluster drift reconcile)
    ├─ Apply changes via cloud provider APIs
    └─ Update resources to match desired state
```

## Components and Interfaces

### Epic 1: Terraform State Integration

#### Component: TerraformStateReader

**Purpose:** Parse Terraform state files and convert resources to InfrastructureState types.

**Location:** `internal/terraform/state_reader.go` (new package)

**Interface:**
```go
package terraform

type StateReader interface {
    // ReadState reads and parses a Terraform state file
    ReadState(path string) (*State, error)
    
    // ConvertToInfrastructureState converts Terraform state to InfrastructureState
    ConvertToInfrastructureState(state *State) (*cloud.InfrastructureState, error)
}

type State struct {
    Version          int                 `json:"version"`
    TerraformVersion string              `json:"terraform_version"`
    Resources        []Resource          `json:"resources"`
}

type Resource struct {
    Type      string                   `json:"type"`
    Name      string                   `json:"name"`
    Provider  string                   `json:"provider"`
    Instances []ResourceInstance       `json:"instances"`
}

type ResourceInstance struct {
    SchemaVersion int                    `json:"schema_version"`
    Attributes    map[string]interface{} `json:"attributes"`
}
```

**Resource Type Mapping:**

| Terraform Resource Type | InfrastructureState Type | Attributes Extracted |
|------------------------|-------------------------|---------------------|
| `openstack_compute_instance_v2` | `Server` | id, name, flavor_id, image_id, status, network, metadata |
| `openstack_networking_network_v2` | `Network` | id, name, tags |
| `openstack_networking_subnet_v2` | `Subnet` (nested in Network) | id, name, cidr |
| `openstack_networking_secgroup_v2` | `SecurityGroup` | id, name, description |
| `openstack_networking_secgroup_rule_v2` | `SecurityRule` (nested in SecurityGroup) | direction, protocol, port_range_min, port_range_max, remote_ip_prefix |
| `openstack_lb_loadbalancer_v2` | `LoadBalancer` | id, name, vip_address, vip_port_id, provisioning_status |
| `openstack_blockstorage_volume_v3` | `Volume` | id, name, size, status, volume_type, attachments |
| `openstack_networking_floatingip_v2` | `FloatingIP` | id, floating_ip, fixed_ip, port_id, status |

**Implementation Strategy:**

1. Read JSON file using `encoding/json`
2. Unmarshal into `State` struct
3. Iterate through `Resources` array
4. For each resource, extract `type` and `instances[0].attributes`
5. Map to appropriate InfrastructureState type based on resource type
6. Handle nested resources (subnets in networks, rules in security groups)
7. Preserve resource relationships (volume attachments, floating IP associations)


#### Integration with Drift Detection

**Current Implementation (cmd/cluster_drift.go):**
```go
// buildDesiredState constructs desired state from configuration
desiredState := buildDesiredState(cfg)
```

**New Implementation:**
```go
// buildDesiredState reads Terraform state and converts to InfrastructureState
func buildDesiredState(cfg config.Config) (*cloud.InfrastructureState, error) {
    // Locate Terraform state file
    statePath := filepath.Join(cfg.GitDir, "infrastructure", "clusters", 
                               cfg.OpenCenter.Cluster.ClusterName, "terraform.tfstate")
    
    // Read and parse state
    reader := terraform.NewStateReader()
    state, err := reader.ReadState(statePath)
    if err != nil {
        return nil, fmt.Errorf("failed to read terraform state: %w", err)
    }
    
    // Convert to InfrastructureState
    infraState, err := reader.ConvertToInfrastructureState(state)
    if err != nil {
        return nil, fmt.Errorf("failed to convert terraform state: %w", err)
    }
    
    return infraState, nil
}
```

### Epic 2: OpenStack Provider Enhancements

#### Component: OpenStack Provider Extensions

**Purpose:** Retrieve comprehensive infrastructure state from OpenStack APIs for all resource types.

**Location:** `internal/cloud/openstack/provider.go` (existing, enhanced)

**New Methods:**

```go
// listSecurityGroups retrieves all security groups for the cluster
func (p *Provider) listSecurityGroups(ctx context.Context, clusterName string) ([]cloud.SecurityGroup, error)

// listLoadBalancers retrieves all load balancers for the cluster
func (p *Provider) listLoadBalancers(ctx context.Context, clusterName string) ([]cloud.LoadBalancer, error)

// listVolumes retrieves all volumes for the cluster
func (p *Provider) listVolumes(ctx context.Context, clusterName string) ([]cloud.Volume, error)

// listFloatingIPs retrieves all floating IPs for the cluster
func (p *Provider) listFloatingIPs(ctx context.Context, clusterName string) ([]cloud.FloatingIP, error)
```

**Security Groups Implementation:**

Uses `gophercloud/openstack/networking/v2/extensions/security/groups` and `rules` packages:

```go
func (p *Provider) listSecurityGroups(ctx context.Context, clusterName string) ([]cloud.SecurityGroup, error) {
    client, err := p.getNetworkClient()
    if err != nil {
        return nil, err
    }
    
    // List security groups with cluster tag
    opts := groups.ListOpts{
        Tags: clusterName,
    }
    
    allPages, err := groups.List(client, opts).AllPages()
    if err != nil {
        return nil, fmt.Errorf("failed to list security groups: %w", err)
    }
    
    sgList, err := groups.ExtractGroups(allPages)
    if err != nil {
        return nil, fmt.Errorf("failed to extract security groups: %w", err)
    }
    
    // Convert to cloud.SecurityGroup
    result := make([]cloud.SecurityGroup, 0, len(sgList))
    for _, sg := range sgList {
        rules := make([]cloud.SecurityRule, 0, len(sg.Rules))
        for _, rule := range sg.Rules {
            rules = append(rules, cloud.SecurityRule{
                ID:          rule.ID,
                Direction:   rule.Direction,
                Protocol:    rule.Protocol,
                PortRange:   fmt.Sprintf("%d-%d", rule.PortRangeMin, rule.PortRangeMax),
                RemoteIP:    rule.RemoteIPPrefix,
                Description: rule.Description,
            })
        }
        
        result = append(result, cloud.SecurityGroup{
            ID:    sg.ID,
            Name:  sg.Name,
            Rules: rules,
        })
    }
    
    return result, nil
}
```

**Load Balancers Implementation:**

Uses `gophercloud/openstack/loadbalancer/v2/loadbalancers`, `listeners`, `pools`, and `members` packages:

```go
func (p *Provider) listLoadBalancers(ctx context.Context, clusterName string) ([]cloud.LoadBalancer, error) {
    client, err := p.getLoadBalancerClient()
    if err != nil {
        return nil, err
    }
    
    // List load balancers with cluster tag
    opts := loadbalancers.ListOpts{
        Tags: []string{clusterName},
    }
    
    allPages, err := loadbalancers.List(client, opts).AllPages()
    if err != nil {
        return nil, fmt.Errorf("failed to list load balancers: %w", err)
    }
    
    lbList, err := loadbalancers.ExtractLoadBalancers(allPages)
    if err != nil {
        return nil, fmt.Errorf("failed to extract load balancers: %w", err)
    }
    
    // Convert to cloud.LoadBalancer
    result := make([]cloud.LoadBalancer, 0, len(lbList))
    for _, lb := range lbList {
        // Get members for this load balancer
        members, err := p.getLoadBalancerMembers(ctx, lb.ID)
        if err != nil {
            return nil, fmt.Errorf("failed to get members for LB %s: %w", lb.ID, err)
        }
        
        result = append(result, cloud.LoadBalancer{
            ID:       lb.ID,
            Name:     lb.Name,
            VIP:      lb.VipAddress,
            Members:  members,
            Protocol: "TCP", // Default, should be extracted from listener
            Port:     6443,  // Default, should be extracted from listener
        })
    }
    
    return result, nil
}
```


**Volumes Implementation:**

Uses `gophercloud/openstack/blockstorage/v3/volumes` package:

```go
func (p *Provider) listVolumes(ctx context.Context, clusterName string) ([]cloud.Volume, error) {
    client, err := p.getBlockStorageClient()
    if err != nil {
        return nil, err
    }
    
    // List volumes with cluster metadata
    opts := volumes.ListOpts{
        Metadata: map[string]string{
            "cluster": clusterName,
        },
    }
    
    allPages, err := volumes.List(client, opts).AllPages()
    if err != nil {
        return nil, fmt.Errorf("failed to list volumes: %w", err)
    }
    
    volumeList, err := volumes.ExtractVolumes(allPages)
    if err != nil {
        return nil, fmt.Errorf("failed to extract volumes: %w", err)
    }
    
    // Convert to cloud.Volume
    result := make([]cloud.Volume, 0, len(volumeList))
    for _, vol := range volumeList {
        attachedTo := ""
        if len(vol.Attachments) > 0 {
            attachedTo = vol.Attachments[0].ServerID
        }
        
        result = append(result, cloud.Volume{
            ID:         vol.ID,
            Name:       vol.Name,
            Size:       vol.Size,
            Status:     vol.Status,
            AttachedTo: attachedTo,
        })
    }
    
    return result, nil
}
```

**Floating IPs Implementation:**

Uses `gophercloud/openstack/networking/v2/extensions/layer3/floatingips` package:

```go
func (p *Provider) listFloatingIPs(ctx context.Context, clusterName string) ([]cloud.FloatingIP, error) {
    client, err := p.getNetworkClient()
    if err != nil {
        return nil, err
    }
    
    // List floating IPs with cluster tag
    opts := floatingips.ListOpts{
        Tags: clusterName,
    }
    
    allPages, err := floatingips.List(client, opts).AllPages()
    if err != nil {
        return nil, fmt.Errorf("failed to list floating IPs: %w", err)
    }
    
    fipList, err := floatingips.ExtractFloatingIPs(allPages)
    if err != nil {
        return nil, fmt.Errorf("failed to extract floating IPs: %w", err)
    }
    
    // Convert to cloud.FloatingIP
    result := make([]cloud.FloatingIP, 0, len(fipList))
    for _, fip := range fipList {
        attachedTo := ""
        if fip.PortID != "" {
            // Resolve port to instance ID
            attachedTo, _ = p.resolvePortToInstance(ctx, fip.PortID)
        }
        
        result = append(result, cloud.FloatingIP{
            ID:         fip.ID,
            Address:    fip.FloatingIP,
            Status:     fip.Status,
            AttachedTo: attachedTo,
        })
    }
    
    return result, nil
}
```

**Integration with GetCurrentState:**

```go
func (p *Provider) GetCurrentState(ctx context.Context, cfg config.Config) (*cloud.InfrastructureState, error) {
    state := &cloud.InfrastructureState{}
    clusterName := cfg.OpenCenter.Cluster.ClusterName
    
    // Existing: Retrieve servers and networks
    state.Servers, _ = p.listServers(ctx, clusterName)
    state.Networks, _ = p.listNetworks(ctx, clusterName)
    
    // NEW: Retrieve security groups, load balancers, volumes, floating IPs
    state.SecurityGroups, _ = p.listSecurityGroups(ctx, clusterName)
    state.LoadBalancers, _ = p.listLoadBalancers(ctx, clusterName)
    state.Volumes, _ = p.listVolumes(ctx, clusterName)
    state.FloatingIPs, _ = p.listFloatingIPs(ctx, clusterName)
    
    return state, nil
}
```

### Epic 3: Secrets Management Enhancements

#### Component: Manifest Scanner

**Purpose:** Automatically discover SOPS-encrypted secrets in overlay directories.

**Location:** `internal/secrets/scanner.go` (new file)

**Interface:**
```go
package secrets

type ManifestScanner interface {
    // ScanManifests recursively scans a directory for YAML files containing secrets
    ScanManifests(overlayDir string, serviceFilter []string) (map[string][]string, error)
}

type Scanner struct {
    // Configuration
}

func NewScanner() *Scanner {
    return &Scanner{}
}

// ScanManifests scans overlay directory for secret manifests
func (s *Scanner) ScanManifests(overlayDir string, serviceFilter []string) (map[string][]string, error) {
    // Returns map[serviceName][]manifestPaths
}
```

**Implementation Strategy:**

1. Use `filepath.Walk` to recursively traverse overlay directory
2. For each YAML file, check if it contains:
   - `kind: Secret` in the YAML
   - SOPS encryption markers (`sops:`, `ENC[AES256_GCM,`)
3. Extract service name from directory path:
   - `applications/overlays/<cluster>/services/keycloak/secret.yaml` → `keycloak`
   - `applications/overlays/<cluster>/managed-services/app1/secret.yaml` → `app1`
4. Skip common ignore patterns: `.git/`, `node_modules/`, `*.swp`, `*~`
5. Handle symlinks by following them once (no recursive symlink following)
6. Apply service filter if provided


#### Component: Organization Detector

**Purpose:** Automatically determine organization from cluster name.

**Location:** `internal/config/org_detector.go` (new file)

**Interface:**
```go
package config

type OrganizationDetector interface {
    // DetermineOrganization finds the organization for a given cluster
    DetermineOrganization(clusterName string) (string, error)
}

type OrgDetector struct {
    configDir string // ~/.config/opencenter/clusters/
}

func NewOrgDetector(configDir string) *OrgDetector {
    return &OrgDetector{configDir: configDir}
}

// DetermineOrganization searches for cluster in organization directories
func (d *OrgDetector) DetermineOrganization(clusterName string) (string, error) {
    // Returns organization ID/name
}
```

**Implementation Strategy:**

1. Scan `~/.config/opencenter/clusters/*/` for directories
2. For each organization directory, check if `<cluster>/<cluster>-config.yaml` exists
3. If found, return organization name (directory name)
4. If multiple matches, return error listing all matches
5. If no matches, return error with available clusters
6. Cache results for duration of command execution

**Directory Structure:**
```
~/.config/opencenter/clusters/
├── 1861184-Metro-Bank-PLC/
│   ├── prod-cluster/
│   │   └── .prod-cluster-config.yaml
│   └── staging-cluster/
│       └── .staging-cluster-config.yaml
└── 2345678-Acme-Corp/
    └── dev-cluster/
        └── .dev-cluster-config.yaml
```

### Epic 4: Key Revocation User Context

#### Component: User Context Extractor

**Purpose:** Extract user identity for audit trails in key revocation operations.

**Location:** `internal/secrets/user_context.go` (new file)

**Interface:**
```go
package secrets

type UserContext struct {
    Email          string
    ServiceAccount string
    Source         string // "git", "env", "flag", "system"
}

type UserContextExtractor interface {
    // ExtractUserContext determines the current user from execution context
    ExtractUserContext(flagUser string) (*UserContext, error)
}

type ContextExtractor struct{}

func NewContextExtractor() *ContextExtractor {
    return &ContextExtractor{}
}

// ExtractUserContext determines user from multiple sources
func (e *ContextExtractor) ExtractUserContext(flagUser string) (*UserContext, error) {
    // Priority: --user flag > OPENCENTER_USER env > OPENCENTER_SERVICE_ACCOUNT env > git user.email > "system"
}
```

**Implementation Strategy:**

1. Check `--user` flag (highest priority, for admin operations)
2. Check `OPENCENTER_USER` environment variable (CI/CD mode)
3. Check `OPENCENTER_SERVICE_ACCOUNT` environment variable (service account mode)
4. Check Git `user.email` configuration (interactive mode)
5. Default to "system" if no context available
6. Validate user identifier as email or service account name (alphanumeric + hyphens)

**Git User Email Extraction:**
```go
func (e *ContextExtractor) getGitUserEmail() (string, error) {
    cmd := exec.Command("git", "config", "user.email")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return strings.TrimSpace(string(output)), nil
}
```

#### Component: Key Ownership Checker

**Purpose:** Verify if a user owns a key before allowing revocation.

**Location:** `internal/secrets/ownership.go` (new file)

**Interface:**
```go
package secrets

type KeyOwnershipChecker interface {
    // IsKeyOwnedByUser checks if a user owns a specific key
    IsKeyOwnedByUser(keyID string, userEmail string) (bool, error)
}

type OwnershipChecker struct {
    registryPath string // Path to key registry
}

func NewOwnershipChecker(registryPath string) *OwnershipChecker {
    return &OwnershipChecker{registryPath: registryPath}
}

// IsKeyOwnedByUser checks key ownership
func (c *OwnershipChecker) IsKeyOwnedByUser(keyID string, userEmail string) (bool, error) {
    // Check key registry for ownership records
}
```

**Implementation Strategy:**

1. Read key registry from `~/.config/opencenter/clusters/<org>/secrets/key-registry.json`
2. Check if key's `CreatedBy` field matches user email
3. Check if user email is in key's `UsedBy` list
4. If no registry exists, fall back to checking `.sops.yaml` recipient list
5. Match Age public keys against user's registered keys
6. Return false if no ownership information found (deny by default)
7. Allow `--force` flag to bypass ownership checks (admin override)

**Key Registry Format:**
```json
{
  "keys": [
    {
      "id": "age1abc123...",
      "created_at": "2026-02-01T10:00:00Z",
      "created_by": "admin@example.com",
      "used_by": ["admin@example.com", "ops@example.com"],
      "revoked": false
    }
  ]
}
```


### Epic 5: Configuration Validation

#### Component: ConfigStructureValidator

**Purpose:** Comprehensive validation of Config struct before saving.

**Location:** `internal/config/structure_validator.go` (new file)

**Interface:**
```go
package config

type ConfigStructureValidator interface {
    // Validate performs comprehensive validation of Config struct
    Validate(cfg *Config) error
}

type StructureValidator struct {
    schema *jsonschema.Schema
}

func NewStructureValidator(schema *jsonschema.Schema) *StructureValidator {
    return &StructureValidator{schema: schema}
}

// Validate checks all required fields and cross-field constraints
func (v *StructureValidator) Validate(cfg *Config) error {
    // Returns aggregated validation errors
}
```

**Validation Rules:**

1. **Required Fields:**
   - `OpenCenter.Cluster.ClusterName` must be non-empty
   - `OpenCenter.Infrastructure.Provider` must be valid (openstack, aws, vmware, kind)
   - `OpenCenter.Cluster.Kubernetes.Version` must be valid semver

2. **Cross-Field Constraints:**
   - `Cluster.WorkerCount` ≥ `Cluster.MasterCount`
   - If `Networking.UseOctavia` is true, `Infrastructure.Provider` must be "openstack"
   - If `Storage.StorageClass` is "longhorn", `Services.Longhorn.Enabled` must be true

3. **Provider-Specific Requirements:**
   - OpenStack: `Infrastructure.OpenStack.NetworkID` must be set
   - AWS: `Infrastructure.AWS.Region` must be valid AWS region
   - VMware: `Infrastructure.VMware.Datacenter` must be set

4. **Service Dependencies:**
   - If `Services.Keycloak.Enabled`, then `Services.CertManager.Enabled` must be true
   - If `Services.Harbor.Enabled`, then `Services.CertManager.Enabled` must be true

**Implementation Strategy:**

1. Use JSON Schema for structural validation (types, required fields, enums)
2. Implement custom validators for cross-field constraints
3. Aggregate all validation errors (don't stop at first error)
4. Return detailed error messages with field paths
5. Support `--skip-validation` flag with warning and audit log entry

**Error Aggregation:**
```go
type ValidationErrors struct {
    Errors []ValidationError
}

type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationErrors) Error() string {
    var sb strings.Builder
    sb.WriteString(fmt.Sprintf("validation failed with %d errors:\n", len(e.Errors)))
    for _, err := range e.Errors {
        sb.WriteString(fmt.Sprintf("  - %s: %s\n", err.Field, err.Message))
    }
    return sb.String()
}
```

**Integration with Config Manager:**
```go
func (m *Manager) Save(cfg *Config) error {
    // Validate before saving
    if !m.skipValidation {
        validator := NewStructureValidator(m.schema)
        if err := validator.Validate(cfg); err != nil {
            return fmt.Errorf("configuration validation failed: %w", err)
        }
    } else {
        // Log skip event to audit log
        m.auditLogger.LogEvent("config_validation_skipped", cfg.OpenCenter.Cluster.ClusterName)
    }
    
    // Proceed with saving
    return m.writeConfig(cfg)
}
```

### Epic 6: Audit Logging Implementation

#### Component: Audit Log Signer

**Purpose:** Generate cryptographic signatures for audit events.

**Location:** `internal/security/audit_signer.go` (new file)

**Interface:**
```go
package security

type AuditSigner interface {
    // Sign generates a signature for an audit event
    Sign(event *AuditEvent) (string, error)
    
    // Verify verifies a signature for an audit event
    Verify(event *AuditEvent, signature string) error
}

type Ed25519Signer struct {
    privateKey ed25519.PrivateKey
    publicKey  ed25519.PublicKey
}

func NewEd25519Signer(keyPath string) (*Ed25519Signer, error) {
    // Load or generate Ed25519 key pair
}

// Sign creates a signature over the event data
func (s *Ed25519Signer) Sign(event *AuditEvent) (string, error) {
    // Serialize event to JSON (excluding signature field)
    // Sign with Ed25519
    // Return base64-encoded signature
}

// Verify checks if a signature is valid for an event
func (s *Ed25519Signer) Verify(event *AuditEvent, signature string) error {
    // Serialize event to JSON (excluding signature field)
    // Decode base64 signature
    // Verify with Ed25519
}
```

**Key Management:**

1. Signing key stored at `~/.config/opencenter/audit/signing-key`
2. Public key stored at `~/.config/opencenter/audit/signing-key.pub`
3. Generate new key pair on first use if not exists
4. Key permissions: 0600 (owner read/write only)
5. Key rotation: Manual process (not automated)

**Signature Format:**
```json
{
  "timestamp": "2026-02-17T10:30:00Z",
  "event_type": "key_revocation",
  "user": "admin@example.com",
  "cluster": "prod-cluster",
  "details": {...},
  "signature": "base64-encoded-ed25519-signature"
}
```


#### Component: Audit Log Verifier

**Purpose:** Verify integrity of audit logs by checking signatures.

**Location:** `internal/security/audit_verifier.go` (new file)

**Interface:**
```go
package security

type AuditVerifier interface {
    // VerifyLog verifies all signatures in an audit log file
    VerifyLog(logPath string) (*VerificationReport, error)
}

type Verifier struct {
    signer AuditSigner
}

func NewVerifier(signer AuditSigner) *Verifier {
    return &Verifier{signer: signer}
}

type VerificationReport struct {
    TotalEvents    int
    VerifiedEvents int
    FailedEvents   []FailedEvent
}

type FailedEvent struct {
    Index     int
    Timestamp string
    Error     string
}

// VerifyLog reads and verifies all events in a log file
func (v *Verifier) VerifyLog(logPath string) (*VerificationReport, error) {
    // Read log file line by line
    // Parse each JSON event
    // Verify signature
    // Aggregate results
}
```

**Implementation Strategy:**

1. Read audit log file line by line (one JSON event per line)
2. For each event, extract signature field
3. Verify signature using Ed25519Signer
4. Track verified and failed events
5. Return detailed report with failed event indices and timestamps
6. Support `--repair` flag to re-sign events with missing/invalid signatures

#### Component: Audit Log Parser

**Purpose:** Parse and filter audit log events.

**Location:** `internal/security/audit_parser.go` (new file)

**Interface:**
```go
package security

type AuditParser interface {
    // ParseLog reads and parses an audit log file
    ParseLog(logPath string, filters *LogFilters) ([]*AuditEvent, error)
}

type Parser struct{}

func NewParser() *Parser {
    return &Parser{}
}

type LogFilters struct {
    Since     time.Time
    EventType string
    User      string
}

type AuditEvent struct {
    Timestamp string                 `json:"timestamp"`
    EventType string                 `json:"event_type"`
    User      string                 `json:"user"`
    Cluster   string                 `json:"cluster"`
    Details   map[string]interface{} `json:"details"`
    Signature string                 `json:"signature"`
}

// ParseLog reads and filters audit events
func (p *Parser) ParseLog(logPath string, filters *LogFilters) ([]*AuditEvent, error) {
    // Read log file
    // Parse JSON events
    // Apply filters
    // Return sorted events
}
```

**Implementation Strategy:**

1. Read audit log file line by line
2. Parse each line as JSON into AuditEvent struct
3. Apply filters:
   - `--since`: Filter events after timestamp
   - `--event-type`: Filter by event type
   - `--user`: Filter by user
4. Handle corrupted lines (invalid JSON) by logging warning and skipping
5. Return events sorted by timestamp
6. Return empty list (not error) if log file doesn't exist

### Epic 7: Drift Detection Callbacks

#### Component: HTTP Callback Client

**Purpose:** Send drift reports to external systems via HTTP POST.

**Location:** `internal/drift/callback.go` (new file)

**Interface:**
```go
package drift

type CallbackClient interface {
    // SendDriftReport sends a drift report to a callback URL
    SendDriftReport(url string, report *cloud.DriftReport, authToken string) error
}

type HTTPCallbackClient struct {
    client  *http.Client
    timeout time.Duration
}

func NewHTTPCallbackClient(timeout time.Duration) *HTTPCallbackClient {
    return &HTTPCallbackClient{
        client:  &http.Client{Timeout: timeout},
        timeout: timeout,
    }
}

// SendDriftReport sends HTTP POST with drift report JSON
func (c *HTTPCallbackClient) SendDriftReport(url string, report *cloud.DriftReport, authToken string) error {
    // Marshal report to JSON
    // Create HTTP POST request
    // Set headers (Content-Type, Authorization)
    // Send request with timeout
    // Handle response
}
```

**Implementation Strategy:**

1. Marshal DriftReport to JSON
2. Create HTTP POST request to callback URL
3. Set headers:
   - `Content-Type: application/json`
   - `Authorization: <authToken>` (if provided via `--callback-auth`)
4. Set 10-second timeout
5. Send request and check response status
6. Log success (2xx status) or failure (non-2xx or network error)
7. Don't block drift detection on callback failure
8. Support `--callback-retry` flag for retry logic (up to 3 retries with exponential backoff)

**Retry Logic:**
```go
func (c *HTTPCallbackClient) sendWithRetry(url string, data []byte, authToken string, maxRetries int) error {
    var lastErr error
    for attempt := 0; attempt <= maxRetries; attempt++ {
        if attempt > 0 {
            // Exponential backoff: 1s, 2s, 4s
            time.Sleep(time.Duration(1<<uint(attempt-1)) * time.Second)
        }
        
        err := c.send(url, data, authToken)
        if err == nil {
            return nil
        }
        lastErr = err
    }
    return fmt.Errorf("failed after %d retries: %w", maxRetries, lastErr)
}
```

**Integration with Drift Detection:**
```go
// In cmd/cluster_drift.go
if callbackURL != "" {
    callbackClient := drift.NewHTTPCallbackClient(10 * time.Second)
    if err := callbackClient.SendDriftReport(callbackURL, report, callbackAuth); err != nil {
        fmt.Fprintf(cmd.ErrOrStderr(), "Warning: failed to send drift report to callback: %v\n", err)
        // Continue execution, don't fail drift detection
    }
}
```


### Epic 8: Keycloak Backup Automation

#### Component: Backup Upload Script

**Purpose:** Upload Keycloak realm backups to object storage (S3 or Swift).

**Location:** `internal/gitops/templates/services/keycloak/backup-script.sh.tpl` (new template)

**Script Structure:**
```bash
#!/bin/bash
set -euo pipefail

# Configuration from environment variables
BACKUP_FILE="${BACKUP_FILE:-/backup/realm-export.json}"
STORAGE_BACKEND="${STORAGE_BACKEND:-s3}"  # s3 or swift
BACKUP_RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"

# Detect storage backend and upload
if [ "$STORAGE_BACKEND" = "s3" ]; then
    upload_to_s3
elif [ "$STORAGE_BACKEND" = "swift" ]; then
    upload_to_swift
else
    echo "Error: Unknown storage backend: $STORAGE_BACKEND"
    exit 1
fi
```

**S3 Upload Implementation:**
```bash
upload_to_s3() {
    # Read S3 configuration from environment
    S3_BUCKET="${S3_BUCKET:?S3_BUCKET not set}"
    S3_REGION="${S3_REGION:-us-east-1}"
    AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:?AWS_ACCESS_KEY_ID not set}"
    AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:?AWS_SECRET_ACCESS_KEY not set}"
    
    # Generate object key with timestamp
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    OBJECT_KEY="keycloak-backups/${CLUSTER_NAME}/${TIMESTAMP}-realm-export.json"
    
    # Upload with retry logic
    for attempt in {1..3}; do
        if aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET}/${OBJECT_KEY}" \
            --region "$S3_REGION" \
            --metadata "cluster=${CLUSTER_NAME},timestamp=${TIMESTAMP},keycloak-version=${KEYCLOAK_VERSION}"; then
            echo "Successfully uploaded backup to s3://${S3_BUCKET}/${OBJECT_KEY}"
            
            # Set lifecycle policy if retention is configured
            if [ -n "$BACKUP_RETENTION_DAYS" ]; then
                set_s3_lifecycle_policy
            fi
            
            return 0
        fi
        
        echo "Upload attempt $attempt failed, retrying..."
        sleep $((2 ** attempt))
    done
    
    echo "Error: Failed to upload backup after 3 attempts"
    exit 1
}

set_s3_lifecycle_policy() {
    # Create lifecycle policy to delete old backups
    cat > /tmp/lifecycle-policy.json <<EOF
{
  "Rules": [{
    "Id": "DeleteOldBackups",
    "Status": "Enabled",
    "Prefix": "keycloak-backups/${CLUSTER_NAME}/",
    "Expiration": {
      "Days": ${BACKUP_RETENTION_DAYS}
    }
  }]
}
EOF
    
    aws s3api put-bucket-lifecycle-configuration \
        --bucket "$S3_BUCKET" \
        --lifecycle-configuration file:///tmp/lifecycle-policy.json
}
```

**Swift Upload Implementation:**
```bash
upload_to_swift() {
    # Read Swift configuration from environment
    OS_AUTH_URL="${OS_AUTH_URL:?OS_AUTH_URL not set}"
    OS_USERNAME="${OS_USERNAME:?OS_USERNAME not set}"
    OS_PASSWORD="${OS_PASSWORD:?OS_PASSWORD not set}"
    OS_PROJECT_NAME="${OS_PROJECT_NAME:?OS_PROJECT_NAME not set}"
    SWIFT_CONTAINER="${SWIFT_CONTAINER:?SWIFT_CONTAINER not set}"
    
    # Generate object name with timestamp
    TIMESTAMP=$(date +%Y%m%d-%H%M%S)
    OBJECT_NAME="${CLUSTER_NAME}/${TIMESTAMP}-realm-export.json"
    
    # Upload with retry logic
    for attempt in {1..3}; do
        if swift upload "$SWIFT_CONTAINER" "$BACKUP_FILE" \
            --object-name "$OBJECT_NAME" \
            --header "X-Object-Meta-Cluster:${CLUSTER_NAME}" \
            --header "X-Object-Meta-Timestamp:${TIMESTAMP}" \
            --header "X-Object-Meta-Keycloak-Version:${KEYCLOAK_VERSION}"; then
            echo "Successfully uploaded backup to ${SWIFT_CONTAINER}/${OBJECT_NAME}"
            
            # Set delete-after header if retention is configured
            if [ -n "$BACKUP_RETENTION_DAYS" ]; then
                RETENTION_SECONDS=$((BACKUP_RETENTION_DAYS * 86400))
                swift post "$SWIFT_CONTAINER" "$OBJECT_NAME" \
                    --header "X-Delete-After:${RETENTION_SECONDS}"
            fi
            
            return 0
        fi
        
        echo "Upload attempt $attempt failed, retrying..."
        sleep $((2 ** attempt))
    done
    
    echo "Error: Failed to upload backup after 3 attempts"
    exit 1
}
```

**CronJob Template:**
```yaml
# internal/gitops/templates/services/keycloak/backup-cronjob.yaml.tpl
apiVersion: batch/v1
kind: CronJob
metadata:
  name: keycloak-backup
  namespace: keycloak
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  successfulJobsHistoryLimit: 3
  failedJobsHistoryLimit: 3
  jobTemplate:
    spec:
      template:
        spec:
          restartPolicy: OnFailure
          containers:
          - name: backup
            image: quay.io/keycloak/keycloak:{{ .KeycloakVersion }}
            command:
            - /bin/bash
            - -c
            - |
              # Export realm
              /opt/keycloak/bin/kc.sh export --dir /backup --realm {{ .RealmName }}
              
              # Upload to object storage
              /scripts/backup-upload.sh
            env:
            - name: CLUSTER_NAME
              value: "{{ .ClusterName }}"
            - name: KEYCLOAK_VERSION
              value: "{{ .KeycloakVersion }}"
            - name: STORAGE_BACKEND
              valueFrom:
                configMapKeyRef:
                  name: keycloak-backup-config
                  key: storage-backend
            - name: BACKUP_RETENTION_DAYS
              valueFrom:
                configMapKeyRef:
                  name: keycloak-backup-config
                  key: retention-days
            # S3 credentials (if using S3)
            - name: AWS_ACCESS_KEY_ID
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-s3
                  key: access-key-id
                  optional: true
            - name: AWS_SECRET_ACCESS_KEY
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-s3
                  key: secret-access-key
                  optional: true
            - name: S3_BUCKET
              valueFrom:
                configMapKeyRef:
                  name: keycloak-backup-config
                  key: s3-bucket
                  optional: true
            - name: S3_REGION
              valueFrom:
                configMapKeyRef:
                  name: keycloak-backup-config
                  key: s3-region
                  optional: true
            # Swift credentials (if using Swift)
            - name: OS_AUTH_URL
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-swift
                  key: auth-url
                  optional: true
            - name: OS_USERNAME
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-swift
                  key: username
                  optional: true
            - name: OS_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-swift
                  key: password
                  optional: true
            - name: OS_PROJECT_NAME
              valueFrom:
                secretKeyRef:
                  name: keycloak-backup-swift
                  key: project-name
                  optional: true
            - name: SWIFT_CONTAINER
              valueFrom:
                configMapKeyRef:
                  name: keycloak-backup-config
                  key: swift-container
                  optional: true
            volumeMounts:
            - name: backup
              mountPath: /backup
            - name: scripts
              mountPath: /scripts
          volumes:
          - name: backup
            emptyDir: {}
          - name: scripts
            configMap:
              name: keycloak-backup-scripts
              defaultMode: 0755
```


## Data Models

### Terraform State Structures

```go
// internal/terraform/state.go

// State represents the complete Terraform state file structure
type State struct {
    Version          int        `json:"version"`
    TerraformVersion string     `json:"terraform_version"`
    Resources        []Resource `json:"resources"`
}

// Resource represents a single Terraform resource
type Resource struct {
    Type      string             `json:"type"`
    Name      string             `json:"name"`
    Provider  string             `json:"provider"`
    Instances []ResourceInstance `json:"instances"`
}

// ResourceInstance represents a single instance of a resource
type ResourceInstance struct {
    SchemaVersion int                    `json:"schema_version"`
    Attributes    map[string]interface{} `json:"attributes"`
}
```

### Secrets Management Structures

```go
// internal/secrets/types.go

// ManifestMap maps service names to manifest file paths
type ManifestMap map[string][]string

// UserContext represents the user performing an operation
type UserContext struct {
    Email          string
    ServiceAccount string
    Source         string // "git", "env", "flag", "system"
}

// KeyRegistry tracks key ownership and lifecycle
type KeyRegistry struct {
    Keys []KeyEntry `json:"keys"`
}

// KeyEntry represents a single key in the registry
type KeyEntry struct {
    ID        string    `json:"id"`
    CreatedAt time.Time `json:"created_at"`
    CreatedBy string    `json:"created_by"`
    UsedBy    []string  `json:"used_by"`
    Revoked   bool      `json:"revoked"`
    RevokedAt time.Time `json:"revoked_at,omitempty"`
    RevokedBy string    `json:"revoked_by,omitempty"`
}
```

### Audit Logging Structures

```go
// internal/security/audit.go

// AuditEvent represents a single audit log entry
type AuditEvent struct {
    Timestamp string                 `json:"timestamp"`
    EventType string                 `json:"event_type"`
    User      string                 `json:"user"`
    Cluster   string                 `json:"cluster"`
    Details   map[string]interface{} `json:"details"`
    Signature string                 `json:"signature"`
}

// VerificationReport summarizes audit log verification
type VerificationReport struct {
    TotalEvents    int
    VerifiedEvents int
    FailedEvents   []FailedEvent
}

// FailedEvent represents a verification failure
type FailedEvent struct {
    Index     int
    Timestamp string
    Error     string
}
```

### Drift Detection Structures

```go
// internal/drift/types.go

// CallbackRequest represents the payload sent to callback URLs
type CallbackRequest struct {
    ClusterName     string              `json:"cluster_name"`
    DetectedAt      string              `json:"detected_at"`
    OverallSeverity string              `json:"overall_severity"`
    DriftCount      int                 `json:"drift_count"`
    Drifts          []cloud.DriftItem   `json:"drifts"`
}
```

## Correctness Properties

A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.

### Epic 1: Terraform State Integration

**Property 1: Terraform state file path construction**
*For any* cluster name and git directory, the constructed Terraform state file path should follow the pattern `<git_dir>/infrastructure/clusters/<cluster>/terraform.tfstate`
**Validates: Requirements 1.2**

**Property 2: Valid Terraform state parsing**
*For any* valid Terraform state JSON file, parsing should succeed and extract the resources array
**Validates: Requirements 1.3, 1.4**

**Property 3: Terraform and OpenTofu compatibility**
*For any* valid state file, parsing should succeed regardless of whether it was created by Terraform or OpenTofu
**Validates: Requirements 1.7**

**Property 4: Resource type mapping completeness**
*For any* Terraform resource with a known OpenStack type, the system should map it to the corresponding InfrastructureState type
**Validates: Requirements 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8**

**Property 5: Unknown resource type handling**
*For any* Terraform resource with an unknown type, the system should skip it and log a warning without failing
**Validates: Requirements 2.9**

**Property 6: Resource relationship preservation**
*For any* set of related Terraform resources (e.g., volume and attachment), the relationships should be preserved in the InfrastructureState
**Validates: Requirements 2.10**

**Property 7: Attribute extraction completeness**
*For any* Terraform resource of a known type, all expected attributes should be extracted into the corresponding InfrastructureState object
**Validates: Requirements 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8**

**Property 8: Missing attribute default handling**
*For any* Terraform resource with missing optional attributes, the system should use sensible defaults without failing
**Validates: Requirements 3.9**

**Property 9: Nested attribute extraction**
*For any* Terraform resource with nested attributes, the system should correctly extract values from nested paths
**Validates: Requirements 3.10**

**Property 10: buildDesiredState integration**
*For any* cluster with a valid Terraform state file, buildDesiredState should return a complete InfrastructureState with all resource types populated
**Validates: Requirements 4.1, 4.2, 4.3, 4.4, 4.5**


### Epic 2: OpenStack Provider Enhancements

**Property 11: Security groups retrieval**
*For any* OpenStack cluster, GetCurrentState should retrieve all security groups with their rules
**Validates: Requirements 5.1, 5.2, 5.3, 5.4, 5.5**

**Property 12: Load balancers retrieval**
*For any* OpenStack cluster, GetCurrentState should retrieve all load balancers with their listeners, pools, and members
**Validates: Requirements 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8**

**Property 13: Volumes retrieval**
*For any* OpenStack cluster, GetCurrentState should retrieve all volumes with their attachments
**Validates: Requirements 7.1, 7.2, 7.3, 7.4, 7.5, 7.6**

**Property 14: Floating IPs retrieval**
*For any* OpenStack cluster, GetCurrentState should retrieve all floating IPs with their associations
**Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7**

### Epic 3: Secrets Management Enhancements

**Property 15: Manifest scanning completeness**
*For any* overlay directory, ScanManifests should find all YAML files containing Secret resources or SOPS encryption markers
**Validates: Requirements 9.1, 9.2, 9.3, 9.4**

**Property 16: Service name extraction**
*For any* manifest file path in an overlay directory, the service name should be correctly extracted from the directory structure
**Validates: Requirements 9.3**

**Property 17: Manifest scanning filtering**
*For any* service filter list, ScanManifests should only return manifests for the specified services
**Validates: Requirements 9.8**

**Property 18: Organization detection uniqueness**
*For any* cluster name that exists in exactly one organization, DetermineOrganization should return that organization
**Validates: Requirements 10.1, 10.2, 10.3, 10.4**

**Property 19: Organization detection ambiguity handling**
*For any* cluster name that exists in multiple organizations, DetermineOrganization should return an error listing all matches
**Validates: Requirements 10.5**

### Epic 4: Key Revocation User Context

**Property 20: User context extraction priority**
*For any* execution context, ExtractUserContext should prioritize sources in order: --user flag, OPENCENTER_USER env, OPENCENTER_SERVICE_ACCOUNT env, git user.email, "system"
**Validates: Requirements 11.1, 11.2, 11.3, 11.4, 11.5**

**Property 21: User identifier validation**
*For any* extracted user context, the identifier should be a valid email address or service account name
**Validates: Requirements 11.7**

**Property 22: Key ownership verification**
*For any* key and user, IsKeyOwnedByUser should return true if and only if the user is in the key's CreatedBy or UsedBy fields
**Validates: Requirements 12.1, 12.2, 12.3, 12.4**

**Property 23: Key ownership fallback**
*For any* key without registry information, IsKeyOwnedByUser should fall back to checking .sops.yaml recipient list
**Validates: Requirements 12.5, 12.6**

### Epic 5: Configuration Validation

**Property 24: Required field validation**
*For any* Config struct, Validate should fail if any required field is missing or empty
**Validates: Requirements 13.1, 13.2, 13.3**

**Property 25: Cross-field constraint validation**
*For any* Config struct, Validate should fail if cross-field constraints are violated (e.g., worker count < master count)
**Validates: Requirements 13.4**

**Property 26: Provider-specific validation**
*For any* Config struct with a specific provider, Validate should check provider-specific requirements
**Validates: Requirements 13.5**

**Property 27: Service dependency validation**
*For any* Config struct with enabled services, Validate should verify all service dependencies are satisfied
**Validates: Requirements 13.6**

**Property 28: Validation error aggregation**
*For any* Config struct with multiple validation failures, Validate should return all errors, not just the first
**Validates: Requirements 13.7**

### Epic 6: Audit Logging Implementation

**Property 29: Signature generation determinism**
*For any* audit event, signing it twice should produce the same signature
**Validates: Requirements 14.1, 14.4, 14.5, 14.6**

**Property 30: Signature verification round-trip**
*For any* audit event, signing then verifying should succeed
**Validates: Requirements 15.1, 15.2, 15.3**

**Property 31: Signature tampering detection**
*For any* audit event, modifying the event data after signing should cause verification to fail
**Validates: Requirements 15.4**

**Property 32: Audit log parsing completeness**
*For any* valid audit log file, parseAuditLog should successfully parse all non-corrupted events
**Validates: Requirements 16.1, 16.2, 16.3**

**Property 33: Audit log filtering correctness**
*For any* audit log and filter criteria, parseAuditLog should return only events matching the filter
**Validates: Requirements 16.4, 16.5, 16.6**

### Epic 7: Drift Detection Callbacks

**Property 34: Callback request format**
*For any* drift report, the callback request should contain cluster name, drift count, drift items, detection timestamp, and severity
**Validates: Requirements 17.1, 17.2, 17.3**

**Property 35: Callback authentication**
*For any* callback request with auth token, the Authorization header should be set correctly
**Validates: Requirements 17.5**

**Property 36: Callback timeout enforcement**
*For any* callback request, the HTTP client should enforce a 10-second timeout
**Validates: Requirements 17.8**

**Property 37: Callback retry behavior**
*For any* failed callback request with retry enabled, the system should retry up to 3 times with exponential backoff
**Validates: Requirements 17.9**

### Epic 8: Keycloak Backup Automation

**Property 38: Backup upload retry**
*For any* backup upload failure, the script should retry up to 3 times with exponential backoff
**Validates: Requirements 18.6, 19.6**

**Property 39: Backup metadata inclusion**
*For any* uploaded backup, the object metadata should include backup timestamp and Keycloak version
**Validates: Requirements 18.8, 19.8**

**Property 40: Storage backend auto-detection**
*For any* set of environment variables, the script should correctly detect whether to use S3 or Swift
**Validates: Requirements 19.9**


## Error Handling

### Terraform State Reading Errors

**Missing State File:**
- Error: `failed to read terraform state: no such file or directory`
- Cause: Cluster has not been provisioned yet
- Remediation: Run `opencenter cluster bootstrap <cluster>` to provision infrastructure

**Corrupted State File:**
- Error: `failed to parse terraform state: invalid JSON at line X`
- Cause: State file is corrupted or manually edited
- Remediation: Restore from backup or re-run `terraform apply`

**Large State File:**
- Warning: `terraform state file is large (>10MB), parsing may take time`
- Cause: Cluster has many resources
- Remediation: None required, informational only

### OpenStack API Errors

**Authentication Failure:**
- Error: `failed to authenticate with OpenStack: invalid credentials`
- Cause: OpenStack credentials are invalid or expired
- Remediation: Update credentials in cluster configuration

**API Rate Limiting:**
- Error: `OpenStack API rate limit exceeded, retrying in X seconds`
- Cause: Too many API requests in short time
- Remediation: Automatic retry with exponential backoff

**Resource Not Found:**
- Warning: `resource X not found in OpenStack, may have been deleted manually`
- Cause: Resource was deleted outside of Terraform
- Remediation: Run drift reconciliation to recreate resource

### Secrets Management Errors

**Manifest Scan Permission Denied:**
- Warning: `failed to read manifest file X: permission denied, skipping`
- Cause: Insufficient file permissions
- Remediation: Fix file permissions or run with appropriate user

**Organization Detection Ambiguity:**
- Error: `cluster 'prod' found in multiple organizations: [org1, org2]`
- Cause: Multiple organizations have clusters with the same name
- Remediation: Use `--organization` flag to specify which organization

**Key Ownership Verification Failure:**
- Error: `user 'user@example.com' does not own key 'age1abc123'`
- Cause: User attempting to revoke a key they don't own
- Remediation: Use `--force` flag (admin only) or contact key owner

### Configuration Validation Errors

**Required Field Missing:**
- Error: `validation failed: field 'OpenCenter.Cluster.ClusterName' is required`
- Cause: Required configuration field is empty
- Remediation: Set the required field in configuration

**Cross-Field Constraint Violation:**
- Error: `validation failed: worker_count (2) must be >= master_count (3)`
- Cause: Configuration violates business rules
- Remediation: Adjust configuration to satisfy constraints

**Provider-Specific Requirement Missing:**
- Error: `validation failed: OpenStack provider requires 'Infrastructure.OpenStack.NetworkID'`
- Cause: Provider-specific required field is missing
- Remediation: Set the required provider-specific field

### Audit Logging Errors

**Signing Key Not Found:**
- Error: `failed to load signing key: no such file or directory`
- Cause: Signing key has not been generated yet
- Remediation: Automatic key generation on first use

**Signature Verification Failure:**
- Error: `audit log verification failed: event at index 42 has invalid signature`
- Cause: Audit log has been tampered with
- Remediation: Investigate security incident, use `--repair` to re-sign (admin only)

**Corrupted Audit Log:**
- Warning: `failed to parse audit event at line 123: invalid JSON, skipping`
- Cause: Audit log line is corrupted
- Remediation: Manual investigation required

### Drift Detection Callback Errors

**Callback URL Unreachable:**
- Warning: `failed to send drift report to callback: connection refused`
- Cause: Callback URL is unreachable
- Remediation: Check callback URL and network connectivity

**Callback Authentication Failure:**
- Warning: `callback returned 401 Unauthorized`
- Cause: Invalid or missing authentication token
- Remediation: Check `--callback-auth` token

**Callback Timeout:**
- Warning: `callback request timed out after 10 seconds`
- Cause: Callback endpoint is slow to respond
- Remediation: Increase timeout or optimize callback endpoint

### Backup Upload Errors

**S3 Credentials Invalid:**
- Error: `S3 upload failed: invalid credentials`
- Cause: AWS credentials are invalid or expired
- Remediation: Update S3 credentials in Kubernetes secret

**Swift Authentication Failure:**
- Error: `Swift upload failed: authentication failed`
- Cause: OpenStack Swift credentials are invalid
- Remediation: Update Swift credentials in Kubernetes secret

**Storage Quota Exceeded:**
- Error: `upload failed: storage quota exceeded`
- Cause: Object storage bucket/container is full
- Remediation: Increase quota or delete old backups

## Testing Strategy

### Unit Testing

Unit tests verify specific examples, edge cases, and error conditions for individual components.

**Terraform State Reader:**
- Test parsing valid Terraform state JSON
- Test parsing OpenTofu state JSON
- Test handling missing state file
- Test handling corrupted JSON
- Test resource type mapping for each supported type
- Test attribute extraction for each resource type
- Test handling missing attributes
- Test handling nested attributes

**OpenStack Provider:**
- Test security groups retrieval with mock API
- Test load balancers retrieval with mock API
- Test volumes retrieval with mock API
- Test floating IPs retrieval with mock API
- Test error handling for API failures
- Test filtering by cluster tag

**Secrets Management:**
- Test manifest scanning with sample directory structure
- Test service name extraction from paths
- Test filtering by service list
- Test organization detection with sample config directories
- Test handling ambiguous cluster names
- Test user context extraction from different sources
- Test key ownership verification

**Configuration Validation:**
- Test required field validation
- Test cross-field constraint validation
- Test provider-specific validation
- Test service dependency validation
- Test error aggregation

**Audit Logging:**
- Test signature generation
- Test signature verification
- Test log parsing
- Test log filtering
- Test handling corrupted events

**Drift Callbacks:**
- Test HTTP POST request formatting
- Test authentication header setting
- Test timeout enforcement
- Test retry logic

**Backup Upload:**
- Test S3 upload with mock AWS SDK
- Test Swift upload with mock Swift client
- Test retry logic
- Test metadata setting

### Property-Based Testing

Property tests verify universal properties across all inputs using randomized test data. Each test should run a minimum of 100 iterations.

**Configuration:**
- Use `gopter` library for Go property-based testing
- Minimum 100 iterations per property test
- Tag each test with feature name and property number

**Test Tagging Format:**
```go
// Feature: todo-completion-phase-1, Property 1: Terraform state file path construction
func TestProperty1_TerraformStateFilePath(t *testing.T) {
    // Property test implementation
}
```

**Property Test Coverage:**
- All 40 correctness properties must have corresponding property tests
- Each property test must reference its design document property
- Property tests should generate random valid inputs
- Property tests should verify the property holds for all generated inputs

### Integration Testing

Integration tests verify end-to-end workflows across multiple components.

**Drift Detection Integration:**
- Test complete drift detection workflow: read Terraform state → query OpenStack → compare → report
- Test drift reconciliation workflow: detect drift → apply changes → verify fixed
- Test drift callback workflow: detect drift → send HTTP POST → verify received

**Secrets Management Integration:**
- Test complete secrets scanning workflow: scan directory → extract service names → filter
- Test organization detection workflow: search directories → find cluster → return organization
- Test key revocation workflow: extract user → verify ownership → revoke key

**Configuration Validation Integration:**
- Test complete validation workflow: load config → validate → save
- Test validation with multiple errors
- Test validation skip with audit logging

**Audit Logging Integration:**
- Test complete audit workflow: log event → sign → write to file
- Test verification workflow: read log → parse events → verify signatures
- Test parsing workflow: read log → filter events → return results

**Backup Upload Integration:**
- Test complete backup workflow: export realm → upload to S3 → verify uploaded
- Test complete backup workflow: export realm → upload to Swift → verify uploaded
- Test retry workflow: simulate failure → retry → succeed

### Test Coverage Goals

- Unit test coverage: ≥80% for all new code
- Property test coverage: 100% of correctness properties
- Integration test coverage: All major workflows
- Edge case coverage: All error conditions and boundary cases

