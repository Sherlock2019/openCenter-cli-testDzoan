# openCenter CLI

**openCenter** is a command-line tool that transforms a single declarative YAML configuration into a production-ready Kubernetes cluster with GitOps management.

It standardizes cluster deployment across OpenStack, VMware, Baremetal, and Kind, providing configuration validation, secrets management, and automated GitOps repository generation.

## What openCenter Does

- **Configuration-First Workflow:** Single YAML file defines your entire cluster (infrastructure, Kubernetes, services, secrets)
- **Multi-Provider Support:** Deploy to OpenStack, VMware, Baremetal, or Kind with the same configuration structure
- **Built-in Validation:** Schema validation, business rules, and provider-specific checks catch errors before deployment
- **GitOps Native:** Generates complete FluxCD-ready repository with Kustomize overlays for cluster-specific customization
- **Secrets Management:** SOPS Age encryption for safe version control of sensitive data
- **Platform Services:** 20+ pre-configured services (monitoring, logging, ingress, auth, storage, backup)

## Quick Start

```bash
# Install tools
mise install

# Build CLI
mise run build

# Initialize cluster
./bin/opencenter cluster init my-cluster --org my-org

# Edit configuration
$EDITOR ~/.config/opencenter/clusters/my-org/.my-cluster-config.yaml

# Validate
./bin/opencenter cluster validate my-cluster

# Generate GitOps repository
./bin/opencenter cluster generate my-cluster

# Deploy
./bin/opencenter cluster deploy my-cluster
```

**Time to first cluster:** 10 minutes configuration + 30-50 minutes deployment

See [Getting Started Tutorial](docs/tutorials/getting-started.md) for complete walkthrough.

## Key Capabilities

- **Cluster Lifecycle:** Initialize, configure, validate, generate, deploy, destroy
- **Configuration Management:** Schema-driven with defaults, validation, and override capabilities
- **Secrets Operations:** Generate keys, encrypt/decrypt, rotate, check expiration, sync, validate drift
- **GitOps Repository:** Automated generation with infrastructure (Terraform/Kubespray) and applications (FluxCD/Kustomize)
- **Provider Abstraction:** Unified interface across OpenStack, VMware, Baremetal, and Kind
- **Service Management:** Enable/disable platform services, customize configurations, view options
- **Operational Tools:** Drift detection, backup/restore, audit logging, cluster doctor, import

## Configuration Example

```yaml
opencenter:
  cluster:
    cluster_name: production
    organization: acme-corp
  
  infrastructure:
    provider: openstack
    cloud:
      openstack:
        auth_url: https://identity.api.rackspacecloud.com/v3
        region: sjc3
        application_credential_id: ${OPENSTACK_APP_CRED_ID}
        application_credential_secret: ${OPENSTACK_APP_CRED_SECRET}
  
  kubernetes:
    version: 1.33.5
    control_plane_count: 3
    worker_count: 2
    cni: calico
  
  services:
    keycloak:
      enabled: true
    kube-prometheus-stack:
      enabled: true
    loki:
      enabled: true
    velero:
      enabled: true

secrets:
  sops:
    age_keys:
      - age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p
```

See [Configuration Schema Reference](docs/reference/configuration-schema.md) for complete structure.

## CLI Commands Quick Reference

```bash
# Cluster Lifecycle
opencenter cluster init <name> --org <org>     # Initialize new cluster
opencenter cluster configure <name> --guided   # Guided provider configuration
opencenter cluster validate <name>             # Validate configuration
opencenter cluster generate <name>             # Generate GitOps repository
opencenter cluster deploy <name>               # Deploy cluster
opencenter cluster destroy <name>              # Destroy cluster

# Cluster Management
opencenter cluster list                        # List all clusters
opencenter cluster use <name>                  # Set active cluster
opencenter cluster active                      # Show active cluster
opencenter cluster status <name>               # Show cluster status
opencenter cluster describe <name>             # Detailed cluster description
opencenter cluster doctor <name>               # Check tools and readiness

# Configuration
opencenter cluster set <name> <path=value>     # Update configuration value
opencenter cluster edit <name>                 # Edit in $EDITOR
opencenter cluster normalize <name>            # Add missing defaults
opencenter cluster export <name>               # Export effective config

# Service Management
opencenter cluster service enable <svc>        # Enable a platform service
opencenter cluster service disable <svc>       # Disable a platform service
opencenter cluster service status              # Show all service states
opencenter cluster service options <svc>       # Show service config options

# Secrets Management
opencenter secrets keys generate               # Generate Age key pair
opencenter secrets keys rotate --type sops     # Rotate encryption keys
opencenter secrets keys check                  # Check key expiration
opencenter secrets keys backup                 # Backup Age keys
opencenter secrets sync <name>                 # Sync secrets to manifests
opencenter secrets validate <name>             # Validate secrets for drift
opencenter secrets encrypt                     # Encrypt secrets in YAML
opencenter secrets decrypt                     # Decrypt secrets in YAML
opencenter secrets status                      # Show encryption status
opencenter secrets login                       # Refresh Keystone token
opencenter secrets list                        # List secrets
opencenter secrets get <name>                  # Download and decrypt
opencenter secrets set <name>                  # Create or update

# Operations
opencenter cluster drift detect <name>         # Detect infrastructure drift
opencenter cluster drift reconcile <name>      # Reconcile drift
opencenter cluster backup create <name>        # Create backup
opencenter cluster backup restore <id>         # Restore from backup
opencenter cluster lock <name>                 # Lock cluster
opencenter cluster import scan                 # Scan repo for import
opencenter cluster migrate-layout --org <org>  # Migrate to secure layout

# CLI Settings
opencenter settings view                       # Display current settings
opencenter settings set <key> <value>          # Set a value (dot notation)
opencenter settings get <key>                  # Get a value
opencenter settings path                       # Show settings file path
opencenter settings edit                       # Edit settings in editor
opencenter settings ide                        # Generate schema + editor setup
opencenter settings explain                    # Explain config effects

# Plugins
opencenter plugins list                        # List external plugins

# Utilities
opencenter version                             # Show version information
opencenter shell-init                          # Output shell integration script
opencenter --help                              # Show help
```

See [CLI Commands Reference](docs/reference/cli-commands.md) for complete documentation with all flags.

## Documentation

Comprehensive documentation organized by the [Diátaxis framework](https://diataxis.fr/):

### 📚 [Tutorials](docs/tutorials/) (Learning-Oriented)
- [Getting Started](docs/tutorials/getting-started.md) - Your first cluster in 10 minutes
- [OpenStack First Cluster](docs/tutorials/openstack-first-cluster.md) - Deploy on OpenStack
- [Kind Local Development](docs/tutorials/kind-local-development.md) - Local development setup
- [VMware Deployment](docs/tutorials/vmware-deployment.md) - Deploy on pre-provisioned VMs
- [Multi-Cluster Management](docs/tutorials/multi-cluster-setup.md) - Manage multiple clusters

### 🔧 [How-To Guides](docs/how-to/) (Task-Oriented)
- [Validate Configuration](docs/how-to/validate-configuration.md) - Pre-deployment validation
- [Manage Secrets](docs/how-to/manage-secrets.md) - SOPS encryption and key rotation
- [Customize Services](docs/how-to/customize-services.md) - Configure platform services
- [Add Worker Pools](docs/how-to/add-worker-pools.md) - Scale your cluster
- [Backup and Restore](docs/how-to/backup-and-restore.md) - Disaster recovery
- [Upgrade Kubernetes](docs/how-to/upgrade-kubernetes.md) - Safe version upgrades
- [Troubleshoot Deployment](docs/how-to/troubleshoot-deployment.md) - Common issues and solutions
- [Migrate Clusters](docs/how-to/migrate-clusters.md) - Provider/region migration
- [Integrate CI/CD](docs/how-to/integrate-ci-cd.md) - Pipeline integration
- [Create and Install a CLI Plugin](docs/how-to/create-install-cli-plugin.md) - External plugins

### 📖 [Reference](docs/reference/) (Information-Oriented)
- [CLI Commands](docs/reference/cli-commands.md) - Complete command reference
- [Configuration Schema](docs/reference/configuration-schema.md) - Configuration file structure
- [Default Values](docs/reference/default-values.md) - Default configuration values
- [Platform Services](docs/reference/platform-services.md) - Available services and versions
- [Providers](docs/reference/providers.md) - Infrastructure provider details
- [Validation Rules](docs/reference/validation-rules.md) - Configuration constraints
- [Environment Variables](docs/reference/environment-variables.md) - Environment configuration
- [Exit Codes](docs/reference/exit-codes.md) - CLI exit codes and meanings
- [File Locations](docs/reference/file-locations.md) - Configuration file paths
- [Configuration Precedence](docs/reference/configuration-precedence.md) - Resolution order
- [Mise Tasks](docs/reference/mise-tasks.md) - Development and build tasks

### 💡 [Explanation](docs/explanation/) (Understanding-Oriented)
- [Architecture](docs/explanation/architecture.md) - System design and decisions
- [GitOps Workflow](docs/explanation/gitops-workflow.md) - Repository structure and reconciliation
- [Security Model](docs/explanation/security-model.md) - Security architecture and controls
- [Configuration Lifecycle](docs/explanation/configuration-lifecycle.md) - Configuration management
- [Provider Comparison](docs/explanation/provider-comparison.md) - Choosing infrastructure providers
- [Drift Detection](docs/explanation/drift-detection.md) - Configuration drift and reconciliation
- [Services & Templates](docs/explanation/services-templates.md) - Service rendering system
- [Plugin Internal Services](docs/explanation/plugin-internal-services.md) - Service plugin architecture
- [Plugin External CLI](docs/explanation/plugin-external-cli.md) - Command plugin model

### 🛠️ [Developer Documentation](docs/dev/)
- [Contributing Guide](docs/dev/contributing.md) - Contribution guidelines
- [Development Setup](docs/dev/development-setup.md) - Local development environment
- [Testing Guide](docs/dev/testing-guide.md) - Writing and running tests
- [Code Structure](docs/dev/code-structure.md) - Codebase organization
- [Adding Providers](docs/dev/adding-providers.md) - Implement new infrastructure providers
- [Adding Services](docs/dev/adding-services.md) - Add new platform services
- [Build System](docs/dev/build-system.md) - Mise-based build and release
- [Release Process](docs/dev/release-process.md) - Creating releases

### 🗺️ [Codemaps](docs/CODEMAPS/) (Architecture Maps)
- [Index](docs/CODEMAPS/INDEX.md) - Package map and cross-references
- [CLI Commands](docs/CODEMAPS/cli-commands.md) - Command tree and registration
- [Config System](docs/CODEMAPS/config-system.md) - Loader pipeline and types
- [GitOps Engine](docs/CODEMAPS/gitops-engine.md) - Generation pipeline
- [Cluster Lifecycle](docs/CODEMAPS/cluster-lifecycle.md) - Lifecycle services
- [Secrets Management](docs/CODEMAPS/secrets-management.md) - Secrets architecture
- [Providers](docs/CODEMAPS/providers.md) - Cloud provider implementations
- [DI Container](docs/CODEMAPS/di-container.md) - Dependency injection

**Start here:** [Documentation Home](docs/index.md) | [Navigation Guide](docs/_nav.md)

## Development Workflow

### Prerequisites

- [Mise](https://mise.jdx.dev/) - Tool version manager
- [Git](https://git-scm.com/) - Version control
- Go, kubectl, kind, helm (managed by Mise)

### Build and Test

```bash
# Install tools
mise install

# Build binary
mise run build

# Run unit tests
mise run test

# Run BDD tests
mise run godog

# Run property-based tests
mise run test-properties

# Lint code
mise run lint

# Format code
mise run fmt
```

### Development Tasks

```bash
# Build for multiple platforms
mise run build-all

# Create release
mise run release v1.0.0

# Generate JSON schema
mise run schema

# Validate templates
mise run validate-templates

# Run a Kind cluster with openCenter-managed CNI
opencenter cluster init dev-cluster --type kind --kind-disable-default-cni
opencenter cluster validate dev-cluster
opencenter cluster generate dev-cluster
opencenter cluster deploy dev-cluster

# Setup local Gitea for testing
mise run gitea-up
```

See [Mise Tasks Reference](docs/reference/mise-tasks.md) for complete list.

Tagged releases are published by GitHub Actions. Use `mise run release` for local preflight builds, then push a `v*` tag to create the signed release artifacts.

## Project Structure

```
openCenter-cli/
├── cmd/                    # CLI commands (Cobra)
│   ├── root.go            # Root command and global flags
│   ├── cluster*.go        # Cluster lifecycle commands
│   ├── secrets*.go        # Secrets management commands
│   ├── config*.go         # Settings commands (Cobra Use: "settings")
│   └── plugins.go         # Plugin management
├── internal/              # Internal packages
│   ├── config/           # Configuration management (loader, builder, types, v2)
│   ├── cluster/          # Cluster lifecycle services (init, validate, setup, bootstrap)
│   ├── gitops/           # GitOps repository generation (pipeline, templates, rendering)
│   ├── secrets/          # Multi-cluster secrets management (rotation, registry, hooks)
│   ├── sops/             # SOPS encryption (Age keys, file encrypt/decrypt)
│   ├── cloud/            # Provider adapters (OpenStack, VMware, Kind)
│   ├── security/         # Audit logging, input validation, command sanitization
│   ├── di/               # Dependency injection container
│   ├── services/         # Platform service plugin registry
│   ├── operations/       # Drift detection, backup, disaster recovery
│   ├── resilience/       # Retry, circuit breaker, distributed locks
│   ├── provision/        # Embedded provisioning templates
│   ├── template/         # Template engine with caching and sandboxing
│   ├── plugins/          # External CLI plugin discovery
│   ├── importer/         # Live cluster import/scan
│   ├── credentials/      # Cloud credential extraction
│   ├── barbican/         # OpenStack Key Manager client
│   ├── localdev/         # Local dev environment (Kind, Gitea, Flux)
│   ├── observability/    # Structured logging, credential masking
│   ├── ansible/          # Kubespray inventory generation
│   ├── tofu/             # OpenTofu/Terraform execution
│   ├── ui/               # Prompts, error formatting, guided flows
│   ├── core/             # Shared: path resolution, validation engine
│   └── util/             # Files, errors, crypto, security, metrics
├── docs/                  # Documentation (Diátaxis)
│   ├── CODEMAPS/         # Architecture maps (generated from code)
│   ├── tutorials/        # Learning-oriented guides
│   ├── how-to/           # Task-oriented guides
│   ├── reference/        # Information-oriented specs
│   ├── explanation/      # Understanding-oriented concepts
│   └── dev/              # Developer documentation
├── tests/                 # BDD tests (Godog)
│   └── features/         # Gherkin feature files
├── schema/                # JSON schema definitions
├── hack/                  # Development scripts and local Gitea setup
├── .mise.toml            # Mise configuration and tasks
├── go.mod                # Go module definition
└── main.go               # CLI entrypoint
```

See [Code Structure](docs/dev/code-structure.md) and [Codemaps](docs/CODEMAPS/INDEX.md) for detailed explanation.

## Configuration File Locations

- **Cluster configurations:** `~/.config/opencenter/clusters/<org>/.<cluster>-config.yaml`
- **CLI settings:** `~/.config/opencenter/config.yaml`
- **Active cluster:** `~/.config/opencenter/active`
- **SOPS Age keys:** `~/.config/opencenter/clusters/<org>/secrets/age/`
- **SSH keys:** `~/.config/opencenter/clusters/<org>/secrets/ssh/`

Override CLI configuration storage with `OPENCENTER_CONFIG_DIR` and cluster storage with `OPENCENTER_CLUSTERS_DIR`.

See [File Locations Reference](docs/reference/file-locations.md) for complete paths.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `OPENCENTER_CONFIG_DIR` | Configuration directory | `~/.config/opencenter` |
| `OPENCENTER_CLUSTERS_DIR` | Cluster storage directory | `${OPENCENTER_CONFIG_DIR}/clusters` |
| `OPENCENTER_PLUGINS_DIR` | Plugins directory | `${OPENCENTER_CONFIG_DIR}/plugins` |
| `OPENCENTER_LOG_LEVEL` | Log level (debug, info, warn, error) | `warn` |
| `SOPS_AGE_KEY_FILE` | Path to Age key file | |
| `SOPS_AGE_RECIPIENTS` | Age public keys for encryption | |
| `KUBECONFIG` | Kubernetes config file | `~/.kube/config` |

See [Environment Variables Reference](docs/reference/environment-variables.md) for complete list.

## Contributing

We welcome contributions! Please see our [Contributing Guide](docs/dev/contributing.md) to get started.

### Quick Contribution Workflow

1. Fork and clone the repository
2. Create a feature branch
3. Make your changes
4. Run tests: `mise run test && mise run godog`
5. Submit a pull request

### Extension Points

- **Custom Providers:** Add new infrastructure providers in `internal/cloud/<provider>/`
- **Custom Services:** Add platform services in `internal/config/services/<service>.go`
- **Custom Validators:** Add validation rules in `internal/core/validation/`
- **Plugins:** Create external plugins as `opencenter-<plugin>` executables

See [Developer Documentation](docs/dev/) for detailed guides.

## License

This project is licensed under the Apache 2.0 License. See [LICENSE](LICENSE) for details.

## Support

- **Documentation:** [docs/](docs/)
- **Security Policy:** [SECURITY.md](SECURITY.md)
- **Issues:** [GitHub Issues](https://github.com/opencenter-cloud/openCenter-cli/issues)
- **Discussions:** [GitHub Discussions](https://github.com/opencenter-cloud/openCenter-cli/discussions)

## Related Projects

openCenter CLI is part of the openCenter ecosystem:

- **[openCenter-gitops-base](https://github.com/opencenter-cloud/openCenter-gitops-base)** - Platform services library with security-hardened Helm values
- **[openCenter-customer-app-example](https://github.com/opencenter-cloud/openCenter-customer-app-example)** - Reference application deployment patterns
- **[openCenter-AirGap](https://github.com/opencenter-cloud/openCenter-AirGap)** - Air-gapped deployment packaging
- **[opencenter-windows](https://github.com/opencenter-cloud/opencenter-windows)** - Windows worker node support
