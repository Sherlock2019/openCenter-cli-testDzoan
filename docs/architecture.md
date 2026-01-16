# openCenter Architecture

## Overview

openCenter is designed as a modular, extensible CLI tool that transforms declarative YAML configurations into fully-functional GitOps repositories. The architecture emphasizes separation of concerns, testability, and extensibility through clean abstractions and well-defined interfaces.

**Architecture Status:** The system is undergoing a comprehensive refactor to introduce modular components with feature flags for gradual adoption. The new architecture is production-ready and available via feature flags.

## High-Level Architecture

The refactored architecture introduces clean separation between core components with well-defined interfaces:

```mermaid
graph TB
    CLI[CLI Commands] --> CB[Configuration Builder]
    CLI --> TM[Template Manager]
    CLI --> FF[Feature Flags]
    
    CB --> CV[Config Validator]
    CB --> CM[Config Migrator]
    CB --> Meta[Metadata Manager]
    
    TM --> TE[Template Engine]
    TM --> TR[Template Registry]
    TM --> TC[Template Composer]
    
    TE --> Cache[Template Cache]
    TR --> ES[Embedded Templates]
    TR --> SP[Service Plugins]
    
    GG[GitOps Generator] --> Pipeline[Generation Pipeline]
    Pipeline --> WS[Workspace Manager]
    Pipeline --> Stages[Generation Stages]
    
    SR[Service Registry] --> SP
    SR --> DM[Dependency Manager]
    SR --> LC[Lifecycle Hooks]
    
    subgraph "Storage Layer"
        FS[File System]
        ES
        CS[Configuration Store]
    end
    
    subgraph "Error Handling"
        EH[Error Handler]
        EA[Error Aggregator]
        EC[Error Context]
    end
    
    WS --> FS
    CB --> CS
    
    CV --> EH
    TE --> EH
    GG --> EH
    
    style FF fill:#f9f,stroke:#333,stroke-width:2px
    style Pipeline fill:#bbf,stroke:#333,stroke-width:2px
    style SR fill:#bfb,stroke:#333,stroke-width:2px
```

### Legacy vs. Refactored Architecture

The system supports both legacy and refactored implementations via feature flags:

| Component | Legacy | Refactored | Feature Flag |
|-----------|--------|------------|--------------|
| Template Engine | Direct rendering | Abstracted engine with caching | `OPENCENTER_USE_NEW_TEMPLATE_ENGINE` |
| GitOps Generation | Monolithic copy | Pipeline-based stages | `OPENCENTER_USE_PIPELINE_GENERATOR` |
| Configuration Builder | Reflection-based | Type-safe fluent API | `OPENCENTER_USE_NEW_CONFIG_BUILDER` |
| Service Management | Hardcoded | Plugin-based registry | `OPENCENTER_USE_SERVICE_REGISTRY` |

## Modular Architecture Components

### 1. Template Engine Abstraction (`internal/template/`)

The template engine provides a unified interface for all template operations with caching and validation.

#### Architecture

```mermaid
graph LR
    Client[Client Code] --> Engine[Template Engine Interface]
    Engine --> Go[Go Template Engine]
    Engine --> Cache[Template Cache]
    Go --> Funcs[Function Registry]
    Go --> Validator[Template Validator]
    Cache --> Store[In-Memory Store]
```

#### Key Interfaces

```go
type TemplateEngine interface {
    Render(ctx context.Context, templatePath string, data interface{}) ([]byte, error)
    ValidateTemplate(templatePath string) error
    RegisterFunction(name string, fn interface{})
    SetCacheEnabled(enabled bool)
    ClearCache()
}
```

#### Features

- **Caching**: Parsed templates are cached for performance
- **Validation**: Syntax validation before rendering
- **Custom Functions**: Extensible function registry (Sprig + custom)
- **Error Context**: Detailed error messages with line numbers
- **Multiple Formats**: Support for Go templates, Helm, Jinja2 (planned)

#### Component Interactions

1. **Template Loading**: Templates loaded from embedded filesystem or disk
2. **Parsing**: Templates parsed and validated on first use
3. **Caching**: Parsed templates stored in memory cache
4. **Rendering**: Templates rendered with configuration data
5. **Error Handling**: Detailed errors with source context

### 2. Template Registry System (`internal/template/`)

Centralized template management with metadata and dependency resolution.

#### Architecture

```mermaid
graph TB
    Registry[Template Registry] --> Meta[Metadata Store]
    Registry --> Deps[Dependency Resolver]
    Registry --> Filter[Template Filter]
    
    Meta --> Templates[Template Definitions]
    Deps --> Graph[Dependency Graph]
    Filter --> Provider[Provider Filter]
    Filter --> Service[Service Filter]
    
    Templates --> Base[Base Templates]
    Templates --> Overlay[Overlay Templates]
    Templates --> Service[Service Templates]
```

#### Key Interfaces

```go
type TemplateRegistry interface {
    RegisterTemplate(template TemplateDefinition) error
    GetTemplate(name string) (TemplateDefinition, error)
    GetTemplatesForProvider(provider string) []TemplateDefinition
    GetTemplatesForService(service string) []TemplateDefinition
    ResolveTemplateDependencies(templates []string) ([]TemplateDefinition, error)
}
```

#### Template Types

- **Infrastructure**: Provider-specific infrastructure templates
- **Service**: Service-specific configuration templates
- **Base**: Foundation templates for composition
- **Overlay**: Patches and extensions for base templates

#### Features

- **Metadata Management**: Rich metadata for each template
- **Dependency Resolution**: Automatic dependency ordering
- **Provider Filtering**: Select templates by cloud provider
- **Service Filtering**: Filter by enabled services
- **Versioning**: Template version compatibility checks

### 3. Configuration Builder Pattern (`internal/config/`)

Type-safe, fluent API for constructing cluster configurations.

#### Architecture

```mermaid
graph LR
    Builder[Config Builder] --> Validator[Validators]
    Builder --> Paths[Type-Safe Paths]
    Builder --> Errors[Error Aggregator]
    
    Validator --> Schema[Schema Validator]
    Validator --> Business[Business Rules]
    Validator --> Provider[Provider Validator]
    
    Paths --> Fields[Field Definitions]
    Errors --> Context[Error Context]
```

#### Key Interfaces

```go
type ConfigBuilder interface {
    WithProvider(provider string) ConfigBuilder
    WithOrganization(org string) ConfigBuilder
    WithClusterName(name string) ConfigBuilder
    WithKubernetesVersion(version string) ConfigBuilder
    WithNodeCounts(masters, workers int) ConfigBuilder
    WithNetworking(config NetworkingConfig) ConfigBuilder
    WithServices(services ...string) ConfigBuilder
    WithOverride(path string, value interface{}) ConfigBuilder
    Build() (Config, error)
    Validate() []ValidationError
}
```

#### Features

- **Fluent API**: Method chaining for readable configuration
- **Type Safety**: Compile-time type checking for configuration paths
- **Validation**: Comprehensive validation with error aggregation
- **Conditional Logic**: Provider-specific configuration options
- **Error Aggregation**: Collect all errors before failing

#### Usage Example

```go
config, err := NewFluentConfigBuilder().
    WithProvider("openstack").
    WithOrganization("my-org").
    WithClusterName("prod-cluster").
    WithKubernetesVersion("1.28.0").
    WithNodeCounts(3, 5).
    WithServices("cert-manager", "monitoring").
    Build()
```

### 4. GitOps Generation Pipeline (`internal/gitops/`)

Pipeline-based GitOps repository generation with staged execution and rollback.

#### Architecture

```mermaid
graph TB
    Generator[GitOps Generator] --> Pipeline[Generation Pipeline]
    Pipeline --> Stage1[Base Structure Stage]
    Pipeline --> Stage2[Infrastructure Stage]
    Pipeline --> Stage3[Service Stage]
    Pipeline --> Stage4[Configuration Stage]
    Pipeline --> Stage5[Validation Stage]
    
    Pipeline --> WS[Workspace Manager]
    WS --> CP[Checkpoints]
    WS --> RB[Rollback Handler]
    WS --> Atomic[Atomic Operations]
    
    Stage1 --> FS[File System]
    Stage2 --> FS
    Stage3 --> FS
    Stage4 --> FS
    
    RB --> CP
```

#### Key Interfaces

```go
type GitOpsGenerator interface {
    Generate(ctx context.Context, config Config) error
    GenerateDryRun(ctx context.Context, config Config) (*GenerationPlan, error)
    Rollback(ctx context.Context, checkpointID string) error
}

type GenerationStage interface {
    Name() string
    Execute(ctx context.Context, workspace *GitOpsWorkspace) error
    Rollback(ctx context.Context, workspace *GitOpsWorkspace) error
    Validate(ctx context.Context, workspace *GitOpsWorkspace) error
}
```

#### Generation Stages

1. **Base Structure**: Create directory layout and base files
2. **Infrastructure**: Generate provider-specific infrastructure templates
3. **Service**: Generate enabled service configurations
4. **Configuration**: Create cluster-specific configurations
5. **Validation**: Verify repository completeness and correctness

#### Features

- **Staged Execution**: Discrete, independent stages
- **Automatic Rollback**: Failed stages trigger rollback of previous stages
- **Checkpointing**: Capture workspace state at any point
- **Dry-Run Mode**: Preview changes without filesystem modifications
- **Atomic Operations**: All-or-nothing file writes
- **Progress Reporting**: Real-time progress updates

### 5. Service Registry and Plugin System (`internal/services/`)

Modular service management with dynamic loading and lifecycle hooks.

#### Architecture

```mermaid
graph TB
    Registry[Service Registry] --> Plugins[Service Plugins]
    Registry --> Deps[Dependency Manager]
    Registry --> Lifecycle[Lifecycle Manager]
    
    Plugins --> Builtin[Built-in Services]
    Plugins --> Custom[Custom Plugins]
    
    Deps --> Graph[Dependency Graph]
    Deps --> Circular[Circular Detection]
    
    Lifecycle --> PreInstall[Pre-Install Hooks]
    Lifecycle --> PostInstall[Post-Install Hooks]
    Lifecycle --> PreUpdate[Pre-Update Hooks]
    Lifecycle --> PostUpdate[Post-Update Hooks]
```

#### Key Interfaces

```go
type ServiceRegistry interface {
    RegisterService(service ServiceDefinition) error
    GetService(name string) (ServiceDefinition, error)
    GetEnabledServices(config Config) []ServiceDefinition
    ResolveDependencies(services []string) ([]ServiceDefinition, error)
    ValidateDependencies(services []string) error
}

type ServicePlugin interface {
    Name() string
    Type() ServiceType
    Validate(config Config) error
    Render(ctx context.Context, config Config, workspace *GitOpsWorkspace) error
    Status(config Config) ServiceStatus
}
```

#### Features

- **Dynamic Loading**: Load service plugins from manifests
- **Dependency Resolution**: Automatic dependency ordering
- **Circular Detection**: Detect and reject circular dependencies
- **Lifecycle Hooks**: Pre/post hooks for install, update, remove
- **Status Reporting**: Service health and status information
- **Plugin Isolation**: Services isolated from core code

### 6. Configuration Migration System (`internal/config/`)

Versioned schema migration with validation and rollback support.

#### Architecture

```mermaid
graph LR
    Manager[Migration Manager] --> Versions[Version Registry]
    Manager --> Migrations[Migration Definitions]
    Manager --> Validator[Migration Validator]
    
    Migrations --> Transform[Transformation Logic]
    Migrations --> Rollback[Rollback Logic]
    
    Validator --> Path[Path Validator]
    Validator --> Schema[Schema Validator]
```

#### Key Interfaces

```go
type MigrationManager interface {
    MigrateConfig(config Config, targetVersion string) (Config, error)
    GetCurrentVersion() string
    GetSupportedVersions() []string
    ValidateMigrationPath(fromVersion, toVersion string) error
}
```

#### Features

- **Versioned Transformations**: Schema migrations between versions
- **Automatic Detection**: Detect configuration version automatically
- **Path Validation**: Ensure valid migration paths exist
- **Value Preservation**: Preserve all user-specified values
- **Dry-Run Support**: Preview migration changes
- **Rollback Capability**: Revert migrations if needed

### 7. Enhanced Error Handling (`internal/util/errors/`)

Structured error handling with context and aggregation.

#### Architecture

```mermaid
graph TB
    Error[OpenCenter Error] --> Type[Error Type]
    Error --> Context[Error Context]
    Error --> Suggestions[Suggestions]
    
    Aggregator[Error Aggregator] --> Errors[Error Collection]
    Aggregator --> Report[Error Reporter]
    
    Context --> Operation[Operation Info]
    Context --> Location[File/Line Info]
    Context --> Component[Component Info]
```

#### Error Types

- **Validation**: Configuration validation errors
- **Template**: Template rendering errors
- **Configuration**: Configuration loading/parsing errors
- **Service**: Service registration/execution errors
- **Generation**: GitOps generation errors
- **System**: System-level errors

#### Features

- **Typed Errors**: Structured error types for different categories
- **Error Context**: Rich context (file, line, operation, component)
- **Error Aggregation**: Collect multiple errors before failing
- **Suggestions**: Actionable suggestions for error resolution
- **Error Wrapping**: Preserve error chains with context

### 8. Template Composition System (`internal/template/`)

Compose complex templates from reusable components.

#### Architecture

```mermaid
graph TB
    Composition[Template Composition] --> Base[Base Template]
    Composition --> Overlays[Overlay System]
    Composition --> Patches[Patch System]
    
    Overlays --> Priority[Priority Ordering]
    Overlays --> Conditions[Render Conditions]
    
    Patches --> Add[Add Operations]
    Patches --> Remove[Remove Operations]
    Patches --> Replace[Replace Operations]
```

#### Features

- **Base Templates**: Foundation templates for extension
- **Overlays**: Layer additional configuration on base templates
- **Patches**: Targeted modifications (add, remove, replace)
- **Priority Ordering**: Deterministic overlay application
- **Conditional Rendering**: Apply overlays based on conditions
- **Conflict Resolution**: Clear error messages for conflicts

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
│   ├── ide           # Generate IDE configs
│   └── features      # Manage feature flags
└── plugins
    ├── list          # List plugins
    ├── install       # Install plugin
    └── remove        # Remove plugin
```

#### Feature Flag Management

The `config features` command manages the new modular architecture:

```bash
# List all feature flags and their status
opencenter config features list

# Enable specific feature flag
opencenter config features enable new-template-engine

# Disable specific feature flag
opencenter config features disable new-template-engine

# Enable all new features
opencenter config features enable-all

# Disable all new features (use legacy)
opencenter config features disable-all
```

#### Available Feature Flags

- `new-template-engine`: Use refactored template engine with caching
- `pipeline-generator`: Use pipeline-based GitOps generation
- `new-config-builder`: Use type-safe configuration builder
- `service-registry`: Use plugin-based service registry

#### Key Design Patterns

**Command Pattern:** Each command is a separate file implementing the Cobra command interface.

**Dependency Injection:** Commands receive dependencies through function parameters rather than global state.

**Error Handling:** Consistent error wrapping with context using `fmt.Errorf`.

### 2. Configuration Management (`internal/config/`)

The configuration system is the heart of openCenter, managing all cluster configuration with enhanced type safety and validation.

#### Components

```
config/
├── config.go              # Core configuration structures
├── metadata.go            # Configuration metadata
├── comparison.go          # Configuration comparison
├── schema.go              # JSON schema generation
├── validator.go           # Validation logic
├── enhanced_validator.go  # Enhanced validation with suggestions
├── suggestions.go         # Validation suggestions
├── manager.go             # Configuration manager
├── loader.go              # Configuration loading
├── path_resolver.go       # Path resolution
├── paths.go               # Type-safe path definitions
├── builder.go             # Fluent configuration builder
├── cli_config.go          # CLI-specific configuration
├── migrator.go            # Schema migration
├── migration.go           # Migration definitions
├── versions.go            # Version management
├── feature_flags.go       # Feature flag system
├── factory.go             # Configuration factory
├── flags/                 # CLI flag processing
│   ├── parser.go          # Flag parser
│   ├── path_parser.go     # Path-based flag parsing
│   ├── reflection_engine.go # Reflection-based updates
│   ├── configuration_merger.go # Configuration merging
│   ├── *_handler.go       # Specific flag handlers
│   └── *_property_test.go # Property-based tests
└── defaults/              # Default templates
    ├── openstack.yaml
    ├── aws.yaml
    ├── baremetal.yaml
    ├── kind.yaml
    └── talos.yaml
```

#### Configuration Flow (Refactored)

```mermaid
graph TB
    YAML[YAML File] --> Loader[Loader]
    Loader --> Parser[Parser]
    Parser --> Migrator[Migrator]
    Migrator --> Builder[Config Builder]
    Builder --> Validator[Enhanced Validator]
    Validator --> Suggestions[Suggestion Engine]
    Validator --> Config[Validated Config]
    Config --> Manager[Config Manager]
    Manager --> Metadata[Metadata Store]
    
    style Builder fill:#bbf,stroke:#333,stroke-width:2px
    style Validator fill:#bfb,stroke:#333,stroke-width:2px
```

#### Enhanced Features

**Type-Safe Builder**: Fluent API with compile-time type checking
```go
config := NewFluentConfigBuilder().
    WithProvider("openstack").
    WithOrganization("my-org").
    WithClusterName("prod").
    Build()
```

**Configuration Metadata**: Track creation, updates, and changes
```go
type ConfigMetadata struct {
    CreatedAt    time.Time
    UpdatedAt    time.Time
    CreatedBy    string
    Tags         map[string]string
    Annotations  map[string]string
}
```

**Configuration Comparison**: Detect and report configuration changes
```go
diff := CompareConfigs(oldConfig, newConfig)
// Returns detailed diff with field-level changes
```

**Schema Versioning**: Automatic migration between schema versions
```go
migratedConfig, err := migrator.MigrateConfig(config, "v2.0.0")
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

Generates complete GitOps repository structures with pipeline-based execution and rollback capabilities.

#### Components

```
gitops/
├── generator.go           # Pipeline-based generator
├── pipeline.go            # Generation pipeline
├── workspace.go           # Workspace management
├── checkpoint.go          # Checkpoint system
├── atomic.go              # Atomic operations
├── rollback.go            # Rollback functionality
├── dryrun.go              # Dry-run mode
├── dryrun_writer.go       # Dry-run file writer
├── progress.go            # Progress reporting
├── copy.go                # Legacy template copying
├── template.go            # Template rendering (legacy)
├── embed.go               # Embedded template management
├── legacy_compat.go       # Legacy compatibility layer
├── stages/                # Generation stages
│   ├── base_stage.go      # Base structure stage
│   ├── init_stage.go      # Initialization stage
│   ├── infrastructure_stage.go # Infrastructure stage
│   ├── service_stage.go   # Service stage
│   ├── config_stage.go    # Configuration stage
│   └── validation_stage.go # Validation stage
├── gitops-base-dir/       # Base repository structure
│   ├── infrastructure/
│   │   └── clusters/
│   └── apps/
└── templates/             # Cluster-specific templates
    ├── infrastructure/
    └── apps/
```

#### Pipeline-Based Generation Flow

```mermaid
graph TB
    Start[Start Generation] --> WS[Create Workspace]
    WS --> CP1[Checkpoint: Initial]
    CP1 --> S1[Stage 1: Base Structure]
    S1 --> CP2[Checkpoint: Base]
    CP2 --> S2[Stage 2: Infrastructure]
    S2 --> CP3[Checkpoint: Infrastructure]
    CP3 --> S3[Stage 3: Services]
    S3 --> CP4[Checkpoint: Services]
    CP4 --> S4[Stage 4: Configuration]
    S4 --> CP5[Checkpoint: Configuration]
    CP5 --> S5[Stage 5: Validation]
    S5 --> Success[Generation Complete]
    
    S1 --> |Error| RB1[Rollback to Initial]
    S2 --> |Error| RB2[Rollback to Base]
    S3 --> |Error| RB3[Rollback to Infrastructure]
    S4 --> |Error| RB4[Rollback to Services]
    S5 --> |Error| RB5[Rollback to Configuration]
    
    style Success fill:#bfb,stroke:#333,stroke-width:2px
    style RB1 fill:#fbb,stroke:#333,stroke-width:2px
    style RB2 fill:#fbb,stroke:#333,stroke-width:2px
    style RB3 fill:#fbb,stroke:#333,stroke-width:2px
    style RB4 fill:#fbb,stroke:#333,stroke-width:2px
    style RB5 fill:#fbb,stroke:#333,stroke-width:2px
```

#### Generation Stages

**Stage 1: Base Structure**
- Create directory layout
- Copy base repository structure
- Initialize Git repository

**Stage 2: Infrastructure**
- Generate provider-specific templates
- Create OpenTofu/Terraform configurations
- Setup infrastructure kustomizations

**Stage 3: Services**
- Generate enabled service configurations
- Resolve service dependencies
- Create service kustomizations

**Stage 4: Configuration**
- Generate cluster-specific configurations
- Apply configuration overlays
- Create Flux/ArgoCD sync configurations

**Stage 5: Validation**
- Verify repository structure
- Validate generated manifests
- Check for missing dependencies

#### Workspace Management

**Checkpointing**: Capture workspace state at each stage
```go
checkpoint := workspace.CreateCheckpoint("after-infrastructure")
// Later, if needed:
workspace.RestoreCheckpoint(checkpoint.ID)
```

**Atomic Operations**: All-or-nothing file writes
```go
atomic.WriteFile(path, content) // Writes to temp, then renames
```

**Rollback**: Automatic rollback on stage failure
```go
if err := stage.Execute(ctx, workspace); err != nil {
    workspace.Rollback(lastCheckpoint)
    return err
}
```

#### Template System (Refactored)

**Template Engine Abstraction**: Unified interface for template operations
```go
engine := NewGoTemplateEngine()
engine.SetCacheEnabled(true)
result, err := engine.Render(ctx, templatePath, data)
```

**Template Registry**: Centralized template management
```go
registry := NewTemplateRegistry()
registry.RegisterTemplate(templateDef)
templates := registry.GetTemplatesForProvider("openstack")
```

**Template Composition**: Build complex templates from components
```go
composition := NewTemplateComposition().
    WithBase("cluster-base.yaml").
    WithOverlay("openstack-overlay.yaml").
    WithPatch(patch)
result := composition.Render(ctx, data)
```

#### Two-Phase Approach (Legacy)

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

## Component Interactions

### Configuration Building Flow

```mermaid
sequenceDiagram
    participant User
    participant CLI
    participant Builder
    participant Validator
    participant Migrator
    participant Store
    
    User->>CLI: opencenter cluster init
    CLI->>Builder: NewFluentConfigBuilder()
    Builder->>Builder: Apply defaults
    CLI->>Builder: WithProvider("openstack")
    CLI->>Builder: WithOrganization("my-org")
    CLI->>Builder: Build()
    Builder->>Validator: Validate(config)
    Validator->>Validator: Schema validation
    Validator->>Validator: Business rules
    Validator->>Validator: Provider validation
    Validator-->>Builder: Validation results
    Builder->>Migrator: Check version
    Migrator->>Migrator: Migrate if needed
    Migrator-->>Builder: Migrated config
    Builder-->>CLI: Config
    CLI->>Store: Save(config)
    Store-->>User: Success
```

### Template Rendering Flow

```mermaid
sequenceDiagram
    participant Client
    participant Engine
    participant Cache
    participant Registry
    participant Composer
    
    Client->>Engine: Render(template, data)
    Engine->>Cache: Get(template)
    alt Cache Hit
        Cache-->>Engine: Parsed template
    else Cache Miss
        Engine->>Registry: GetTemplate(name)
        Registry-->>Engine: Template definition
        Engine->>Engine: Parse template
        Engine->>Cache: Store(template)
    end
    Engine->>Composer: Apply overlays
    Composer->>Composer: Resolve dependencies
    Composer->>Composer: Apply patches
    Composer-->>Engine: Composed template
    Engine->>Engine: Execute template
    Engine-->>Client: Rendered output
```

### GitOps Generation Flow

```mermaid
sequenceDiagram
    participant CLI
    participant Generator
    participant Pipeline
    participant Workspace
    participant Stages
    
    CLI->>Generator: Generate(config)
    Generator->>Workspace: Create()
    Workspace-->>Generator: Workspace
    Generator->>Pipeline: Execute(workspace)
    
    loop For each stage
        Pipeline->>Workspace: CreateCheckpoint()
        Pipeline->>Stages: Execute(workspace)
        alt Stage Success
            Stages-->>Pipeline: Success
        else Stage Failure
            Stages-->>Pipeline: Error
            Pipeline->>Workspace: Rollback(checkpoint)
            Pipeline-->>Generator: Error
            Generator-->>CLI: Error
        end
    end
    
    Pipeline->>Stages: Validate(workspace)
    Stages-->>Pipeline: Valid
    Pipeline-->>Generator: Success
    Generator-->>CLI: Success
```

### Service Dependency Resolution

```mermaid
sequenceDiagram
    participant Config
    participant Registry
    participant Resolver
    participant Graph
    
    Config->>Registry: GetEnabledServices(config)
    Registry->>Resolver: ResolveDependencies(services)
    Resolver->>Graph: BuildGraph(services)
    Graph->>Graph: Topological sort
    Graph->>Graph: Detect cycles
    alt Circular Dependency
        Graph-->>Resolver: Error
        Resolver-->>Registry: Error
        Registry-->>Config: Error
    else Valid Graph
        Graph-->>Resolver: Ordered services
        Resolver-->>Registry: Ordered services
        Registry-->>Config: Services with deps
    end
```

### Error Handling Flow

```mermaid
sequenceDiagram
    participant Component
    participant ErrorHandler
    participant Aggregator
    participant Logger
    participant User
    
    Component->>ErrorHandler: Handle(error)
    ErrorHandler->>ErrorHandler: Add context
    ErrorHandler->>ErrorHandler: Generate suggestions
    ErrorHandler->>Aggregator: Add(error)
    Aggregator->>Aggregator: Collect errors
    Aggregator->>Logger: Log(errors)
    Aggregator->>User: Report(errors)
    User->>User: Review suggestions
```

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

### Configuration Initialization (Refactored)

```mermaid
graph LR
    User[User Command] --> Defaults[Generate Defaults]
    Defaults --> Builder[Config Builder]
    Builder --> Overrides[Apply Overrides]
    Overrides --> Migrator[Schema Migrator]
    Migrator --> Validator[Enhanced Validator]
    Validator --> Suggestions[Generate Suggestions]
    Validator --> Metadata[Add Metadata]
    Metadata --> Save[Save to File]
    
    style Builder fill:#bbf,stroke:#333,stroke-width:2px
    style Validator fill:#bfb,stroke:#333,stroke-width:2px
```

**Flow Steps:**
1. User executes `opencenter cluster init`
2. System generates defaults from provider template
3. Config builder applies CLI flags and arguments
4. Schema migrator checks and updates version if needed
5. Enhanced validator runs multi-layer validation
6. Suggestion engine provides actionable guidance for errors
7. Metadata manager adds timestamps and tracking info
8. Configuration saved to YAML file

### GitOps Setup (Refactored)

```mermaid
graph TB
    Config[Load Config] --> Workspace[Create Workspace]
    Workspace --> Checkpoint1[Checkpoint: Initial]
    Checkpoint1 --> Base[Stage 1: Base Structure]
    Base --> Checkpoint2[Checkpoint: Base]
    Checkpoint2 --> Infra[Stage 2: Infrastructure]
    Infra --> Checkpoint3[Checkpoint: Infra]
    Checkpoint3 --> Services[Stage 3: Services]
    Services --> Checkpoint4[Checkpoint: Services]
    Checkpoint4 --> ConfigStage[Stage 4: Configuration]
    ConfigStage --> Checkpoint5[Checkpoint: Config]
    Checkpoint5 --> Validate[Stage 5: Validation]
    Validate --> Git[Initialize Git]
    Git --> Success[Complete]
    
    Base -.->|Error| Rollback1[Rollback to Initial]
    Infra -.->|Error| Rollback2[Rollback to Base]
    Services -.->|Error| Rollback3[Rollback to Infra]
    ConfigStage -.->|Error| Rollback4[Rollback to Services]
    Validate -.->|Error| Rollback5[Rollback to Config]
    
    style Success fill:#bfb,stroke:#333,stroke-width:2px
    style Rollback1 fill:#fbb,stroke:#333,stroke-width:2px
    style Rollback2 fill:#fbb,stroke:#333,stroke-width:2px
    style Rollback3 fill:#fbb,stroke:#333,stroke-width:2px
    style Rollback4 fill:#fbb,stroke:#333,stroke-width:2px
    style Rollback5 fill:#fbb,stroke:#333,stroke-width:2px
```

**Flow Steps:**
1. Load and validate cluster configuration
2. Create isolated workspace for generation
3. Execute generation stages with checkpointing:
   - **Base Structure**: Directory layout and base files
   - **Infrastructure**: Provider-specific templates
   - **Services**: Enabled service configurations
   - **Configuration**: Cluster-specific configs
   - **Validation**: Verify completeness
4. Initialize Git repository
5. On any error, automatically rollback to last checkpoint

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

### 8. Modular Architecture (New)

Clean separation of concerns with well-defined interfaces between components. Each module can be developed, tested, and deployed independently.

### 9. Gradual Migration (New)

Feature flags enable gradual adoption of new systems without breaking existing workflows. Users can opt-in to new features at their own pace.

### 10. Backward Compatibility (New)

Legacy compatibility layers ensure existing configurations and workflows continue to work. No breaking changes during refactoring.

### 11. Error Aggregation (New)

Collect and report all errors together with actionable suggestions, rather than failing on first error.

### 12. Rollback Capability (New)

All operations support rollback to previous state. Failed operations leave no partial artifacts.

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

### Gopter (Property-Based Testing)

**Rationale:**
- Generative testing for core invariants
- Catches edge cases missed by unit tests
- Validates universal properties
- Complements example-based testing
- Excellent Go integration

## Refactoring Strategy

### Feature Flag System

The refactoring uses feature flags to enable gradual adoption without breaking changes:

```bash
# Enable individual features
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
export OPENCENTER_USE_PIPELINE_GENERATOR=true
export OPENCENTER_USE_NEW_CONFIG_BUILDER=true
export OPENCENTER_USE_SERVICE_REGISTRY=true

# Or enable all new features at once
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
```

### Migration Phases

**Phase 1: Foundation (Complete)**
- Core error handling system
- Template engine abstraction
- Testing framework with property-based tests

**Phase 2: Configuration System (Complete)**
- Enhanced configuration types with metadata
- Fluent configuration builder
- Configuration migration system
- Enhanced validation with suggestions

**Phase 3: Template System (Complete)**
- Template registry with metadata
- Template composition system
- Service plugin architecture
- Legacy compatibility layer

**Phase 4: GitOps Generation (Complete)**
- Workspace management with checkpointing
- Pipeline-based generation
- Generation stage implementations
- Legacy compatibility layer

**Phase 5: Integration (Complete)**
- Feature flag integration
- Comprehensive testing
- Documentation updates
- Performance validation

**Phase 6: Production Readiness (In Progress)**
- User-facing documentation
- Performance benchmarking
- Production monitoring
- Feature flag cleanup plan

### Compatibility Layers

Each refactored component includes a compatibility layer:

**Template Engine**: `internal/template/legacy.go`
- Wraps new engine with legacy interface
- Feature flag switches between implementations
- Validates output identity

**GitOps Generator**: `internal/gitops/legacy_compat.go`
- Wraps pipeline generator with legacy interface
- Feature flag switches between implementations
- Validates repository structure identity

**Configuration Builder**: `internal/config/builder.go`
- Provides both fluent and legacy interfaces
- Feature flag switches between implementations
- Validates configuration identity

### Validation Strategy

All refactored components include validation tests:

**Output Identity Tests**: Verify new system produces identical output
```go
func TestTemplateOutputIdentity(t *testing.T) {
    legacyOutput := legacyRender(template, data)
    newOutput := newEngine.Render(template, data)
    assert.Equal(t, legacyOutput, newOutput)
}
```

**Property-Based Tests**: Verify universal properties
```go
func TestConfigBuilderIdempotency(t *testing.T) {
    properties.Property("building twice yields same result", 
        prop.ForAll(func(provider, org, cluster string) bool {
            config1 := builder.Build()
            config2 := builder.Build()
            return reflect.DeepEqual(config1, config2)
        }))
}
```

**Integration Tests**: Verify complete workflows
```go
func TestCompleteGitOpsGeneration(t *testing.T) {
    // Test full generation pipeline with rollback
}
```

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

**Status**: Planned for future release

### 2. State Management

Implement cluster state tracking and drift detection.

**Status**: Under consideration

### 3. Multi-Cluster Orchestration

Add support for managing multiple clusters as a fleet.

**Status**: Planned for future release

### 4. Observability

Integrate metrics, tracing, and logging for better visibility.

**Status**: Partially implemented (structured logging, metrics collection)

### 5. Policy Engine

Add policy-as-code for compliance and governance.

**Status**: Under consideration

### 6. Advanced Template Features

- Template inheritance and mixins
- Conditional template selection
- Dynamic template generation
- Template testing framework

**Status**: Partially implemented (composition, overlays)

### 7. Enhanced Service Management

- Service health monitoring
- Automatic service updates
- Service dependency visualization
- Service marketplace

**Status**: Foundation implemented (service registry, plugins)

## Refactored Architecture Benefits

### Improved Maintainability

**Modular Design**: Each component has clear responsibilities and interfaces
- Template engine handles all template operations
- Configuration builder manages configuration construction
- GitOps generator orchestrates repository creation
- Service registry manages service lifecycle

**Reduced Coupling**: Components interact through well-defined interfaces
- Easy to modify one component without affecting others
- Clear dependency boundaries
- Testable in isolation

### Enhanced Extensibility

**Plugin Architecture**: Add new functionality without modifying core
- Service plugins for new services
- Provider plugins for new cloud platforms
- Validator plugins for custom validation rules
- Template plugins for custom templates

**Composition System**: Build complex templates from reusable components
- Base templates provide foundation
- Overlays add provider-specific configuration
- Patches enable targeted modifications

### Better Testability

**Property-Based Testing**: Validate universal properties
- Configuration builder idempotency
- Template rendering consistency
- Service dependency resolution correctness
- Migration value preservation

**Integration Testing**: Validate complete workflows
- End-to-end GitOps generation
- Configuration migration paths
- Service dependency resolution
- Error handling and rollback

### Improved User Experience

**Error Aggregation**: Report all errors with suggestions
- No more "fix one error, run again, find next error"
- Actionable suggestions for common mistakes
- Rich error context (file, line, operation)

**Rollback Capability**: Recover from failures gracefully
- Automatic rollback on generation failure
- Checkpoint-based recovery
- No partial artifacts left behind

**Feature Flags**: Gradual adoption of new features
- Try new features without commitment
- Fallback to legacy if issues arise
- Smooth migration path

### Performance Improvements

**Template Caching**: Parsed templates cached for reuse
- Significant speedup for repeated renders
- Reduced memory allocation
- Better resource utilization

**Parallel Processing**: Independent operations run concurrently
- Parallel template rendering where possible
- Concurrent validation checks
- Faster overall execution

**Optimized Validation**: Efficient validation pipeline
- Early exit on critical errors
- Parallel validation of independent rules
- Cached validation results

## Migration Guide

For users migrating to the refactored architecture, see:
- **Migration Guide**: `docs/migration/configuration-system-refactor.md`
- **Feature Flag Timeline**: `docs/migration/feature-flag-removal-timeline.md`
- **Troubleshooting**: `docs/migration/troubleshooting-refactored-system.md`

## Conclusion

openCenter's refactored architecture represents a significant improvement in maintainability, extensibility, and user experience. The modular design with clean interfaces enables independent development and testing of components while maintaining a cohesive user experience.

**Key Achievements:**

1. **Modular Architecture**: Clean separation of concerns with well-defined interfaces
2. **Feature Flags**: Gradual adoption path without breaking changes
3. **Backward Compatibility**: Legacy systems continue to work during migration
4. **Comprehensive Testing**: Property-based and integration tests validate correctness
5. **Performance**: Caching and optimization meet or exceed legacy system
6. **Error Handling**: Rich error context with actionable suggestions
7. **Rollback Capability**: Graceful recovery from failures

The configuration-first approach with GitOps integration, combined with the new modular architecture, provides a solid foundation for cluster lifecycle management that can evolve with changing requirements while maintaining stability and reliability.

**Current Status**: The refactored architecture is production-ready and available via feature flags. Users can enable new features individually or all at once, with full backward compatibility maintained through legacy compatibility layers.
