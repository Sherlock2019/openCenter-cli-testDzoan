# openCenter - Comprehensive Overview

## What is openCenter?

openCenter is a command-line tool that streamlines Kubernetes and OpenStack cluster bootstrapping by transforming a single, declarative YAML configuration file into a ready-to-use GitOps repository. It provides a configuration-first workflow that is both user-friendly for interactive use and automation-ready for CI/CD pipelines.

## Current Status

**Version:** 0.0.1 (Active Development)

**Maturity:** Early development phase with core functionality implemented and comprehensive test coverage.

### Implemented Features

#### Core Configuration Management
- ✅ YAML-based declarative cluster configuration
- ✅ JSON schema generation for IDE integration and validation
- ✅ Organization-based multi-tenancy support
- ✅ Configuration validation with comprehensive error reporting
- ✅ Configuration migration tools for schema updates
- ✅ Debug configuration generation for troubleshooting

#### Cluster Lifecycle Management
- ✅ Cluster initialization with sensible defaults
- ✅ Cluster validation (schema, business rules, connectivity)
- ✅ Cluster selection and active cluster tracking
- ✅ Cluster listing across organizations
- ✅ Cluster information display
- ✅ Cluster update with dot-notation overrides
- ✅ Cluster migration between schema versions

#### GitOps Integration
- ✅ Automatic GitOps repository scaffolding
- ✅ Organization-based repository structure
- ✅ Template-driven manifest generation
- ✅ FluxCD integration with reconciliation settings
- ✅ Cluster-specific and shared service configurations
- ✅ Infrastructure-as-Code (OpenTofu) integration

#### Secrets Management
- ✅ SOPS integration for encrypted secrets
- ✅ Age key generation and management
- ✅ Key rotation with automatic re-encryption
- ✅ Key backup and restore functionality
- ✅ Organization-wide SOPS configuration
- ✅ Secrets validation and status checking

#### Provider Support
- ✅ OpenStack (primary provider)
- ✅ AWS (basic support)
- ✅ Kind (local development)
- ⚠️ VMware (partial support)
- ⚠️ Bare metal (planned)

#### Network Plugins
- ✅ Calico CNI
- ✅ Cilium CNI
- ✅ Kube-OVN CNI
- ✅ Mutual exclusivity validation

#### Development Tools
- ✅ Comprehensive BDD test suite (Godog)
- ✅ Unit test coverage
- ✅ Mise-based build system
- ✅ Plugin system for extensibility
- ✅ Local Gitea setup for testing
- ✅ Kind cluster creation for development

### In Progress

- 🔄 Enhanced cloud provider validation
- 🔄 Improved error messages and user guidance
- 🔄 Additional service integrations
- 🔄 Performance optimizations
- 🔄 Documentation expansion

### Planned Features

- 📋 Interactive cluster configuration wizard
- 📋 Cluster health monitoring
- 📋 Automated cluster upgrades
- 📋 Multi-cluster management
- 📋 Disaster recovery workflows
- 📋 Cost estimation and optimization
- 📋 Compliance and security scanning

## Architecture

### High-Level Design

```
┌─────────────────────────────────────────────────────────────┐
│                     openCenter CLI                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │   Commands   │  │    Config    │  │   Plugins    │     │
│  │   (Cobra)    │  │  Management  │  │   System     │     │
│  └──────────────┘  └──────────────┘  └──────────────┘     │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
┌───────▼────────┐  ┌───────▼────────┐  ┌──────▼──────┐
│   GitOps       │  │   Secrets      │  │  Provider   │
│  Scaffolding   │  │  Management    │  │  Specific   │
│   (Templates)  │  │    (SOPS)      │  │   Logic     │
└────────────────┘  └────────────────┘  └─────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            │
                ┌───────────▼───────────┐
                │  Infrastructure       │
                │  (OpenTofu/Terraform) │
                └───────────────────────┘
```

### Directory Structure

```
openCenter/
├── cmd/                    # CLI command implementations
│   ├── root.go            # Main CLI entry point
│   ├── cluster*.go        # Cluster lifecycle commands
│   ├── sops.go            # SOPS key management
│   ├── config*.go         # Configuration commands
│   └── plugins.go         # Plugin system
├── internal/              # Private packages
│   ├── config/           # Configuration management
│   │   ├── config.go     # Core config structures
│   │   ├── schema.go     # JSON schema generation
│   │   ├── validator.go  # Validation logic
│   │   ├── manager.go    # Config manager
│   │   └── path_resolver.go  # Path resolution
│   ├── gitops/           # GitOps scaffolding
│   │   ├── copy.go       # Template copying
│   │   ├── templates/    # Embedded templates
│   │   └── gitops-base-dir/  # Base structure
│   ├── sops/             # SOPS integration
│   │   ├── keys.go       # Key management
│   │   ├── encrypt.go    # Encryption/decryption
│   │   └── manager.go    # SOPS manager
│   ├── tofu/             # OpenTofu integration
│   ├── plugins/          # Plugin system
│   └── util/             # Shared utilities
├── tests/                # BDD test suite
│   └── features/         # Gherkin feature files
├── docs/                 # Documentation
│   ├── tutorials/        # Step-by-step guides
│   ├── how-to/          # Task-oriented guides
│   ├── reference/       # Technical reference
│   └── explanation/     # Conceptual docs
├── schema/              # Generated JSON schemas
├── testdata/            # Test fixtures
├── .mise.toml           # Build system config
├── go.mod               # Go module definition
└── main.go              # Application entry point
```

## Key Concepts

### Configuration-First Workflow

openCenter uses a single YAML file as the source of truth for cluster configuration. This file contains:

- **Cluster metadata**: Name, environment, region, organization
- **Infrastructure settings**: Cloud provider, networking, compute resources
- **Kubernetes configuration**: Version, node counts, network plugins
- **GitOps settings**: Repository location, branch, FluxCD configuration
- **Service toggles**: Enable/disable managed services
- **Secrets management**: SOPS key locations and encryption settings

### Organization-Based Multi-Tenancy

Clusters are organized by organization, allowing multiple teams to manage their clusters independently:

```
~/.config/openCenter/
└── clusters/
    ├── org1/
    │   └── infrastructure/
    │       └── clusters/
    │           ├── cluster1/
    │           └── cluster2/
    └── org2/
        └── infrastructure/
            └── clusters/
                └── cluster3/
```

### GitOps-Native Design

Every cluster configuration automatically generates a complete GitOps repository structure:

```
gitops-repo/
├── infrastructure/
│   └── clusters/
│       └── <cluster-name>/
│           ├── flux-system/
│           ├── opentofu/
│           └── kustomization.yaml
└── apps/
    └── <cluster-name>/
        ├── cert-manager/
        ├── monitoring/
        └── ...
```

### Schema-Driven Validation

Configuration is validated against a comprehensive JSON schema that:
- Provides IDE autocomplete and inline documentation
- Catches errors before deployment
- Ensures consistency across clusters
- Supports versioning and migration

### Secrets Management with SOPS

All sensitive data is encrypted using SOPS with Age keys:
- Automatic key generation per cluster
- Organization-wide key sharing
- Key rotation with re-encryption
- Git-friendly encrypted files

## Technology Stack

### Core Technologies
- **Language**: Go 1.24+
- **CLI Framework**: Cobra for command structure
- **Configuration**: YAML with Go struct validation
- **Testing**: Godog (BDD) + standard Go testing
- **Build System**: Mise for task automation

### Key Dependencies
- **UI/UX**: Charm libraries (Bubble Tea, Lipgloss, Bubbles)
- **Templating**: Masterminds Sprig
- **Schema**: Automatic JSON schema generation
- **Secrets**: SOPS with Age encryption
- **Logging**: Logrus with structured logging

### Development Tools
- Go 1.24
- kubectl 1.34
- kind 0.30
- helm 3.18
- flux (latest)

## Getting Started

### Prerequisites

1. Install Mise: https://mise.jdx.dev/
2. Clone the repository
3. Run `mise install` to install all required tools

### Building

```bash
# Build the binary
mise run build

# Build for multiple platforms
mise run build-all

# Run tests
mise run test
mise run godog
```

### Basic Usage

```bash
# Initialize a new cluster
./bin/openCenter cluster init my-cluster

# Validate configuration
./bin/openCenter cluster validate my-cluster

# Setup GitOps repository
./bin/openCenter cluster setup my-cluster

# Bootstrap the cluster
./bin/openCenter cluster bootstrap my-cluster
```

## Documentation Structure

- **[Tutorials](tutorials/)**: Step-by-step learning guides
- **[How-To Guides](how-to/)**: Task-oriented recipes
- **[Reference](reference/)**: Technical specifications
- **[Explanation](explanation/)**: Conceptual overviews

## Contributing

See [CONTRIBUTING.md](../CONTRIBUTING.md) for development guidelines.

## License

Apache 2.0 License - See [LICENSE](../LICENSE) for details.
