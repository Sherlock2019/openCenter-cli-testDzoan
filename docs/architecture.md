# openCenter Architecture

## Overview

openCenter is designed as a modular, extensible CLI tool that transforms declarative YAML configurations into fully-functional GitOps repositories. The architecture emphasizes separation of concerns, testability, and extensibility.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        User Interface                           │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐         │
│  │   CLI        │  │  Interactive │  │   Plugins    │         │
│  │  (Cobra)     │  │    Mode      │  │   System     │         │
│  └──────────────┘  └──────────────┘  └──────────────┘         │
└─────────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌───────▼────────┐  ┌──────▼──────┐
│  Configuration │  │   Validation   │  │   Logging   │
│   Management   │  │     Engine     │  │   System    │
└────────────────┘  └────────────────┘  └─────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌───────▼────────┐  ┌──────▼──────┐
│   GitOps       │  │   Secrets      │  │  Provider   │
│  Scaffolding   │  │  Management    │  │  Adapters   │
└────────────────┘  └────────────────┘  └─────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                ┌───────────▼───────────┐
                │  Infrastructure       │
                │  Provisioning         │
                │  (OpenTofu/Terraform) │
                └───────────────────────┘
```

## Core Components

### 1. CLI Layer (`cmd/`)

The CLI layer provides the user interface and command structure.

#### Command Structure

```
openCenter
├── cluster
│   ├── init          # Initialize new cluster
│   ├── validate      # Validate configuration
│   ├── list          # List clusters
│   ├── select        # Select active cluster
│   ├── current       # Show active cluster
│   ├── info          # Display cluster info
│   ├── update        # Update configuration
│   ├── migrate       # Migrate schema
│   ├── setup         # Setup GitOps
│   ├── bootstrap     # Bootstrap infrastructure
│   ├── render        # Render templates
│   ├── schema        # Generate schema
│   ├── preflight     # Run preflight checks
│   └── destroy       # Destroy cluster
├── sops
│   ├── generate-key  # Generate Age keys
│   ├── rotate-key    # Rotate keys
│   ├── backup-key    # Backup keys
│   ├── validate      # Validate SOPS setup
│   └── secrets-*     # Secrets operations
├── config
│   └── ide           # Generate IDE configs
└── plugins
    ├── list          # List plugins
    ├── install       # Install plugin
    └── remove        # Remove plugin
```

#### Key Design Patterns

**Command Pattern:** Each command is a separate file implementing the Cobra command interface.

**Dependency Injection:** Commands receive dependencies through function parameters rather than global state.

**Error Handling:** Consistent error wrapping with context using `fmt.Errorf`.

### 2. Configuration Management (`internal/config/`)

The configuration system is the heart of openCenter, managing all cluster configuration.

#### Components

```
config/
├── config.go              # Core configuration structures
├── schema.go              # JSON schema generation
├── validator.go           # Validation logic
├── manager.go             # Configuration manager
├── loader.go              # Configuration loading
├── path_resolver.go       # Path resolution
├── cli_config.go          # CLI-specific configuration
├── migrator.go            # Schema migration
├── factory.go             # Configuration factory
└── defaults/              # Default templates
    ├── openstack.yaml
    ├── aws.yaml
    └── kind.yaml
```

#### Configuration Flow

```
┌─────────────┐
│  YAML File  │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Loader    │  ← Reads file, parses YAML
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Validator  │  ← Schema + business rules
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Config    │  ← Validated configuration
│   Struct    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Manager    │  ← Manages lifecycle
└─────────────┘
```

#### Key Features

**Schema-Driven:** Configuration structure is defined by Go structs with YAML tags, automatically generating JSON schema.

**Default Values:** Sensible defaults are applied for all optional fields.

**Validation Layers:**
1. Schema validation (structure, types, constraints)
2. Business rule validation (cross-field dependencies)
3. Provider-specific validation (credentials, connectivity)

**Path Resolution:** Supports both organization-based and legacy directory structures with automatic fallback.

### 3. GitOps Scaffolding (`internal/gitops/`)

Generates complete GitOps repository structures from embedded templates.

#### Components

```
gitops/
├── copy.go                # Template copying logic
├── template.go            # Template rendering
├── embed.go               # Embedded template management
├── gitops-base-dir/       # Base repository structure
│   ├── infrastructure/
│   │   └── clusters/
│   └── apps/
└── templates/             # Cluster-specific templates
    ├── infrastructure/
    └── apps/
```

#### Template System

**Embedded Templates:** Templates are embedded in the binary using Go's `//go:embed` directive.

**Rendering Engine:** Uses Go's `text/template` with Sprig functions for advanced templating.

**Two-Phase Approach:**
1. **Copy Phase:** Base structure is copied to GitOps directory
2. **Render Phase:** Cluster-specific templates are rendered with configuration values

#### Directory Structure

```
gitops-repo/
├── infrastructure/
│   └── clusters/
│       └── <cluster-name>/
│           ├── flux-system/
│           │   ├── gotk-components.yaml
│           │   ├── gotk-sync.yaml
│           │   └── kustomization.yaml
│           ├── opentofu/
│           │   ├── main.tf
│           │   ├── provider.tf
│           │   └── variables.tf
│           └── kustomization.yaml
└── apps/
    └── <cluster-name>/
        ├── cert-manager/
        ├── monitoring/
        ├── networking/
        └── ...
```

### 4. Secrets Management (`internal/sops/`)

Integrates SOPS with Age encryption for secure secrets management.

#### Components

```
sops/
├── keys.go        # Age key management
├── encrypt.go     # Encryption/decryption
├── manager.go     # SOPS manager
├── git.go         # Git integration
├── validator.go   # Validation
└── interfaces.go  # Abstraction layer
```

#### Key Management Flow

```
┌─────────────┐
│  Generate   │  ← Create Age key pair
│  Age Key    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Store     │  ← Save to cluster secrets directory
│   Private   │
│    Key      │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Update    │  ← Add public key to .sops.yaml
│   SOPS      │
│   Config    │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Encrypt    │  ← Encrypt secrets in GitOps repo
│  Secrets    │
└─────────────┘
```

#### Encryption Strategy

**Age-Based:** Uses Age encryption for simplicity and security.

**Organization-Wide:** Keys can be shared across clusters in an organization.

**Git-Friendly:** Encrypted files are YAML-compatible and diff-friendly.

**Selective Encryption:** Only sensitive fields are encrypted using regex patterns.

### 5. Provider Adapters (`internal/cloud/`, `internal/provision/`)

Provider-specific logic for different cloud platforms.

#### Provider Interface

```go
type Provider interface {
    Validate(config Config) error
    TestConnectivity(config Config) error
    Provision(config Config) error
    Destroy(config Config) error
}
```

#### Supported Providers

**OpenStack:**
- Authentication via application credentials
- Network configuration
- Compute resource provisioning
- Floating IP management

**AWS:**
- VPC and subnet configuration
- IAM credential management
- EC2 instance provisioning
- EKS integration (planned)

**Kind:**
- Local cluster creation
- Docker/Podman support
- Custom CNI configuration
- Development workflows

**VMware (Partial):**
- vSphere configuration
- Resource pool management
- Template deployment

### 6. Infrastructure Provisioning (`internal/tofu/`)

Generates and manages OpenTofu/Terraform configurations.

#### Components

```
tofu/
├── provision.go       # Main provisioning logic
└── templates/         # Terraform templates
    ├── main.tf.tmpl
    ├── provider.tf.tmpl
    └── variables.tf.tmpl
```

#### Provisioning Flow

```
┌─────────────┐
│   Config    │  ← Cluster configuration
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Generate   │  ← Render Terraform templates
│  Terraform  │
│   Files     │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Configure  │  ← Setup backend and providers
│   Backend   │
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Execute   │  ← Run terraform init/plan/apply
│  Terraform  │
└─────────────┘
```

### 7. Plugin System (`internal/plugins/`)

Extensible plugin architecture for custom commands and providers.

#### Plugin Discovery

```
$OPENCENTER_PLUGINS_DIR/
├── openCenter-custom-provider
├── openCenter-custom-command
└── openCenter-custom-validator
```

#### Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Execute(args []string) error
}
```

#### Plugin Types

**Command Plugins:** Add new CLI commands
**Provider Plugins:** Add new cloud providers
**Validator Plugins:** Add custom validation rules
**Template Plugins:** Add custom templates

### 8. Validation Engine (`internal/config/validator.go`)

Multi-layered validation system ensuring configuration correctness.

#### Validation Layers

```
┌─────────────────┐
│ Schema          │  ← JSON schema validation
│ Validation      │     (structure, types, constraints)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Business Rule   │  ← Cross-field validation
│ Validation      │     (dependencies, conflicts)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Provider        │  ← Provider-specific validation
│ Validation      │     (credentials, resources)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Connectivity    │  ← Network and API validation
│ Validation      │     (reachability, authentication)
└─────────────────┘
```

#### Validation Rules

**Schema Rules:**
- Required fields present
- Correct data types
- Value constraints (min/max, patterns)
- Enum validation

**Business Rules:**
- Cluster name matches meta.name
- Only one network plugin enabled
- Windows workers disabled when count is 0
- VRRP IP required when Octavia disabled
- AWS credentials required for S3 backend

**Provider Rules:**
- OpenStack: Valid auth URL, credentials, network IDs
- AWS: Valid VPC, subnets, IAM credentials
- VMware: Valid vCenter, datacenter, resource pool

**Connectivity Rules:**
- API endpoints reachable
- Authentication successful
- Required resources exist
- Network connectivity verified

## Data Flow

### Configuration Initialization

```
User Command
    │
    ▼
Generate Defaults ──→ Apply Overrides ──→ Validate ──→ Save
    │                      │                  │           │
    ▼                      ▼                  ▼           ▼
Schema Defaults    CLI Flags/Args    Schema + Rules   YAML File
```

### GitOps Setup

```
Load Config
    │
    ▼
Create Directories ──→ Copy Base ──→ Render Templates ──→ Init Git
    │                     │              │                   │
    ▼                     ▼              ▼                   ▼
Organization      Base Structure   Cluster Manifests    Git Repo
Structure         (embedded)       (from config)        (initialized)
```

### Secrets Management

```
Generate Key
    │
    ▼
Store Key ──→ Update SOPS Config ──→ Encrypt Secrets
    │              │                      │
    ▼              ▼                      ▼
Cluster        .sops.yaml            Encrypted YAML
Secrets Dir    (public key)          (in GitOps repo)
```

## Design Principles

### 1. Configuration as Code

All cluster configuration is declarative and version-controlled. No imperative commands modify cluster state directly.

### 2. GitOps Native

Every cluster has a corresponding GitOps repository. All changes flow through Git.

### 3. Security First

Secrets are encrypted at rest using SOPS. No plaintext secrets in configuration or Git.

### 4. Provider Agnostic

Core logic is independent of cloud providers. Provider-specific code is isolated in adapters.

### 5. Testability

All components are designed for testing with clear interfaces and dependency injection.

### 6. Extensibility

Plugin system allows custom commands, providers, and validators without modifying core code.

### 7. User Experience

Clear error messages, comprehensive validation, and helpful defaults make the tool accessible.

## Technology Choices

### Go Language

**Rationale:**
- Strong typing and compile-time checks
- Excellent standard library
- Cross-platform compilation
- Fast execution
- Good tooling ecosystem

### Cobra CLI Framework

**Rationale:**
- Industry standard for Go CLIs
- Excellent command structure
- Built-in help generation
- Flag parsing and validation
- Subcommand support

### YAML Configuration

**Rationale:**
- Human-readable and writable
- Widely used in Kubernetes ecosystem
- Good tooling support
- Comments support
- Hierarchical structure

### SOPS + Age

**Rationale:**
- Simple and secure encryption
- Git-friendly encrypted files
- No external key management service required
- Selective field encryption
- Active development and support

### OpenTofu

**Rationale:**
- Open-source Terraform alternative
- Compatible with Terraform modules
- Active community
- No licensing concerns
- Multi-provider support

### Mise Build System

**Rationale:**
- Tool version management
- Task automation
- Environment management
- Cross-platform support
- Simple configuration

## Performance Considerations

### Configuration Loading

**Optimization:** Lazy loading of configuration with caching.

**Impact:** Sub-100ms load times for typical configurations.

### Template Rendering

**Optimization:** Embedded templates compiled into binary.

**Impact:** No file I/O for template loading, fast rendering.

### Validation

**Optimization:** Parallel validation of independent rules.

**Impact:** Sub-500ms validation for full configuration.

### GitOps Setup

**Optimization:** Efficient file copying with proper buffering.

**Impact:** Sub-5s setup for complete repository structure.

## Security Architecture

### Threat Model

**Threats:**
- Plaintext secrets in configuration
- Secrets in Git history
- Unauthorized access to clusters
- Man-in-the-middle attacks
- Compromised credentials

**Mitigations:**
- SOPS encryption for all secrets
- Age key-based encryption
- Secure file permissions (0600)
- TLS for all API communication
- Credential validation before use

### Security Boundaries

```
┌─────────────────────────────────────┐
│  User's Local Machine               │
│  ┌───────────────────────────────┐  │
│  │  openCenter CLI               │  │
│  │  - Configuration files        │  │
│  │  - SOPS keys (encrypted)      │  │
│  │  - GitOps repo (local)        │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│  Git Repository (Remote)            │
│  - Encrypted secrets                │
│  - Public configuration             │
│  - Infrastructure manifests         │
└─────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────┐
│  Cloud Provider                     │
│  - Infrastructure resources         │
│  - Kubernetes cluster               │
│  - Managed services                 │
└─────────────────────────────────────┘
```

## Scalability

### Configuration Size

**Current:** Handles configurations up to 10MB efficiently.

**Optimization:** Streaming YAML parser for larger configurations.

### Cluster Count

**Current:** Efficiently manages hundreds of clusters.

**Optimization:** Indexed cluster listing and caching.

### Template Complexity

**Current:** Handles complex templates with nested structures.

**Optimization:** Template compilation and caching.

## Future Architecture Enhancements

### 1. API Server

Add REST API for programmatic access and web UI integration.

### 2. State Management

Implement cluster state tracking and drift detection.

### 3. Multi-Cluster Orchestration

Add support for managing multiple clusters as a fleet.

### 4. Observability

Integrate metrics, tracing, and logging for better visibility.

### 5. Policy Engine

Add policy-as-code for compliance and governance.

## Conclusion

openCenter's architecture is designed for extensibility, maintainability, and user experience. The modular design allows independent development and testing of components while maintaining a cohesive user experience. The configuration-first approach with GitOps integration provides a solid foundation for cluster lifecycle management.
