# Configuration Comment Injection

## Overview

The `cluster template` command supports automatic injection of inline comments into generated YAML configuration files. This feature helps users understand the purpose and valid values for each configuration field without referring to external documentation.

## Table of Contents

- [Overview](#overview)
- [Usage](#usage)
- [Implementation](#implementation)
- [Comment Types](#comment-types)
- [Examples](#examples)
- [Testing](#testing)
- [Extending Comments](#extending-comments)

## Usage

To generate a configuration template with inline comments, use the `--comments` flag:

```bash
# Generate template with comments to stdout
opencenter cluster template --comments

# Save commented template to file
opencenter cluster template --comments --out config-with-comments.yaml

# Generate provider-specific template with comments
opencenter cluster template --provider openstack --comments --out openstack-config.yaml
```

## Implementation

The comment injection feature uses the `yaml.v3` Node API to parse YAML into a node tree, add comments to specific nodes, and re-marshal the YAML with comments preserved.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    cluster template                          │
│                                                              │
│  1. Generate Config struct                                  │
│  2. Marshal to YAML bytes                                   │
│  3. If --comments flag:                                     │
│     a. Parse YAML into yaml.Node tree                       │
│     b. Recursively add comments to nodes                    │
│     c. Re-marshal with comments                             │
│  4. Output to file or stdout                                │
└─────────────────────────────────────────────────────────────┘
```

### Key Functions

#### `addConfigComments(data []byte, provider string) []byte`

Main entry point for comment injection. Parses YAML, adds comments, and re-marshals.

**Parameters:**
- `data`: YAML bytes to add comments to
- `provider`: Cloud provider name (used for provider-specific comments)

**Returns:**
- YAML bytes with comments injected

**Error Handling:**
If parsing or marshaling fails, returns original data with a header comment.

#### `addCommentsToNode(node *yaml.Node, provider string)`

Recursively traverses the YAML node tree and adds comments to specific fields.

**Parameters:**
- `node`: YAML node to process
- `provider`: Cloud provider name

**Behavior:**
- Handles document nodes by processing children
- Only processes mapping nodes (key-value pairs)
- Matches field names and adds appropriate comments
- Recursively processes nested structures

### Comment Types

The implementation supports two types of comments:

1. **Head Comments**: Appear above the field
   ```yaml
   # This is a head comment
   field_name: value
   ```

2. **Line Comments**: Appear at the end of the line
   ```yaml
   field_name: value  # This is a line comment
   ```

## Comment Types

### Schema Version

```yaml
# Configuration schema version (do not modify)
schema_version: "2.0"  # v2.0 schema
```

### Metadata

```yaml
# Cluster metadata and annotations
metadata:
  created_at: 2025-02-02T10:00:00Z  # Timestamp when cluster was created
  updated_at: 2025-02-02T10:00:00Z  # Timestamp of last update
  created_by: admin@example.com  # User who created the cluster
  tags:  # Key-value tags for organization
    environment: production
  annotations:  # Additional metadata annotations
    description: Production cluster
```

### Infrastructure

```yaml
# Infrastructure provider configuration
infrastructure:
  provider: openstack  # Cloud provider: openstack (openstack, aws, talos, kind, baremetal)
  # Cloud provider-specific settings
  cloud:
    # OpenStack provider configuration
    openstack:
      auth_url: https://identity.api.example.com/v3  # OpenStack Identity API endpoint
      region: us-east-1  # OpenStack region
      domain: Default  # OpenStack domain (usually 'Default')
      application_credential_id: ""  # Application credential ID (preferred over password)
      application_credential_secret: ""  # Application credential secret
      image_id: abc123  # Base OS image ID for nodes
```

### Cluster Configuration

```yaml
# Kubernetes cluster configuration
cluster:
  cluster_name: my-cluster  # Kubernetes cluster name
  base_domain: k8s.example.com  # Base DNS domain for cluster
  cluster_fqdn: my-cluster.k8s.example.com  # Full cluster domain name
  admin_email: admin@example.com  # Administrator email address
  ssh_authorized_keys:  # SSH public keys for node access
    - ssh-ed25519 AAAA...
  # Kubernetes version and node configuration
  kubernetes:
    version: 1.33.5  # Kubernetes version (e.g., 1.33.5)
    master_count: 3  # Number of control plane nodes (odd number recommended)
    worker_count: 3  # Number of worker nodes
```

### GitOps Configuration

```yaml
# GitOps repository configuration
gitops:
  git_dir: ./gitops-repo  # Local GitOps repository directory
  git_url: git@github.com:org/repo.git  # Remote GitOps repository URL
  git_branch: main  # Git branch for cluster manifests
  flux_version: v2.2.0  # FluxCD version
```

### Provider-Specific Comments

#### OpenStack

```yaml
# OpenStack provider configuration
openstack:
  auth_url: https://identity.api.example.com/v3  # OpenStack Identity API endpoint
  region: us-east-1  # OpenStack region
  domain: Default  # OpenStack domain (usually 'Default')
  application_credential_id: ""  # Application credential ID (preferred over password)
  application_credential_secret: ""  # Application credential secret
  image_id: abc123  # Base OS image ID for nodes
  networking:  # OpenStack networking configuration
    floating_ip_pool: PUBLICNET
```

#### AWS

```yaml
# AWS provider configuration
aws:
  region: us-east-1  # AWS region (e.g., us-east-1)
  profile: default  # AWS CLI profile name
  vpc_id: vpc-123456  # Existing VPC ID (optional)
  private_subnets:  # Private subnet IDs for worker nodes
    - subnet-abc123
  public_subnets:  # Public subnet IDs for control plane
    - subnet-def456
```

### Secrets

```yaml
# Secrets management configuration (SOPS/Age)
secrets:
  sops_age_key_file: ~/.config/sops/age/keys.txt  # Path to Age encryption key
  sops_age_recipients:  # Age public keys for encryption
    - age1abc123...
```

### OpenTofu

```yaml
# OpenTofu/Terraform configuration
opentofu:
  enabled: true  # Enable infrastructure provisioning with OpenTofu
  version: 1.6.0  # OpenTofu version
  backend:  # Terraform backend configuration
    type: s3
```

## Examples

### Basic Usage

Generate a minimal template with comments:

```bash
opencenter cluster template --comments --minimal --out minimal-commented.yaml
```

Output:
```yaml
# Configuration schema version (do not modify)
schema_version: "2.0"  # v2.0 schema

# OpenCenter cluster configuration
opencenter:
  # Cluster identification
  meta:
    name: example-cluster  # Unique cluster name
    organization: opencenter  # Organization or team name
  
  # Infrastructure provider configuration
  infrastructure:
    provider: openstack  # Cloud provider: openstack (openstack, aws, talos, kind, baremetal)
  
  # Kubernetes cluster configuration
  cluster:
    cluster_name: example-cluster  # Kubernetes cluster name
    # Kubernetes version and node configuration
    kubernetes:
      version: 1.33.5  # Kubernetes version (e.g., 1.33.5)
      master_count: 3  # Number of control plane nodes (odd number recommended)
      worker_count: 2  # Number of worker nodes
  
  # GitOps repository configuration
  gitops:
    git_dir: ./gitops-repo  # Local GitOps repository directory

# OpenTofu/Terraform configuration
opentofu:
  enabled: true  # Enable infrastructure provisioning with OpenTofu

# Secrets management configuration (SOPS/Age)
secrets: {}
```

### Provider-Specific Template

Generate an OpenStack-specific template with comments:

```bash
opencenter cluster template --provider openstack --comments --out openstack.yaml
```

This will include OpenStack-specific comments for fields like `auth_url`, `region`, `application_credential_id`, etc.

### Complete Template

Generate a complete template showing all available fields with comments:

```bash
opencenter cluster template --comments --out complete-config.yaml
```

This generates a comprehensive configuration with comments for every field, useful for:
- Understanding all available options
- Documentation and examples
- IDE autocomplete reference
- Migration from older schema versions

## Testing

The comment injection feature includes comprehensive tests:

### Unit Tests

```bash
# Run all comment-related tests
go test -v ./cmd/ -run TestComment

# Run specific test
go test -v ./cmd/ -run TestAddConfigComments
```

### Test Coverage

Tests verify:
- Comments are present for all major sections
- YAML structure is preserved after comment injection
- Invalid YAML is handled gracefully
- Provider-specific comments are added correctly
- Nil and empty nodes don't cause panics

### Manual Testing

```bash
# Build and test
mise run build

# Generate template with comments
./bin/opencenter cluster template --comments

# Verify comments are present
./bin/opencenter cluster template --comments | grep -c "#"

# Verify YAML is still valid
./bin/opencenter cluster template --comments | yamllint -
```

## Extending Comments

To add comments for new configuration fields:

### 1. Identify the Section

Determine which section the field belongs to (infrastructure, cluster, services, etc.).

### 2. Add Comment Logic

Add a case in the appropriate comment function:

```go
func addClusterComments(key, value *yaml.Node) {
    // ... existing code ...
    
    for i := 0; i < len(value.Content); i += 2 {
        subKey := value.Content[i]
        switch subKey.Value {
        case "new_field":
            subKey.LineComment = "Description of new field"
        // ... other cases ...
        }
    }
}
```

### 3. Add Tests

Add test cases to verify the new comments:

```go
func TestNewFieldComments(t *testing.T) {
    cfg := generateCompleteTemplate("openstack")
    data, _ := yaml.Marshal(&cfg)
    output := addConfigComments(data, "openstack")
    outputStr := string(output)

    if !strings.Contains(outputStr, "Description of new field") {
        t.Error("expected new field comment not found")
    }
}
```

### 4. Update Documentation

Add the new field and its comment to this documentation.

## Best Practices

### Comment Content

- **Be concise**: Keep comments short and to the point
- **Be descriptive**: Explain what the field does, not just repeat the field name
- **Include examples**: Show valid values or formats when helpful
- **Indicate requirements**: Note if a field is required or optional
- **Explain constraints**: Mention valid ranges, formats, or options

### Comment Placement

- **Head comments**: Use for section headers and important fields
- **Line comments**: Use for field descriptions and quick notes
- **Avoid over-commenting**: Don't add comments to every field, focus on non-obvious ones

### Provider-Specific Comments

- Include provider name in comments when relevant
- List valid options for enum-like fields
- Reference provider documentation for complex fields

## Related Documentation

- [Cluster Template Command](../reference/cli/cluster-template.md)
- [Configuration Schema](../reference/config-schema.md)
- [YAML v3 Node API](https://pkg.go.dev/gopkg.in/yaml.v3#Node)

## References

- Implementation: `cmd/cluster_template.go`
- Tests: `cmd/cluster_template_test.go`
- Design: `.kiro/specs/v2-breaking-changes/design.md` (Epic 4.1)
