# `openCenter cluster` - Manage Cluster Configurations

## Overview

The `cluster` command provides comprehensive lifecycle management for Kubernetes cluster configurations. It supports organization-based multi-tenancy and integrates seamlessly with GitOps workflows.

## Description

The cluster command group enables you to initialize, validate, update, and manage cluster configurations throughout their entire lifecycle. All cluster configurations are stored in an organization-based directory structure at `~/.config/openCenter/clusters/<organization>/<cluster>/`, providing clear separation and multi-tenancy support.

## Common Workflow

1. **Initialize** a new cluster configuration
2. **Validate** the configuration against schema and business rules
3. **Run preflight** checks to verify tool and provider requirements
4. **Setup** infrastructure and GitOps repository
5. **Bootstrap** the cluster with provider-specific actions

## Subcommands

### Cluster Lifecycle

- [`init`](init.md) - Initialize a new cluster configuration
- [`validate`](validate.md) - Validate cluster configuration
- [`preflight`](preflight.md) - Run preflight checks for tools and providers
- [`setup`](setup.md) - Setup GitOps directory and initialize git
- [`render`](render.md) - Render templates without initializing git
- [`bootstrap`](bootstrap.md) - Run provider-specific bootstrap actions
- [`destroy`](destroy.md) - Destroy a cluster and remove its configuration

### Configuration Management

- [`list`](list.md) - List all configured clusters
- [`select`](select.md) - Select the active cluster
- [`current`](current.md) - Show the current active cluster
- [`info`](info.md) - Show configuration for a cluster
- [`edit`](edit.md) - Edit a cluster configuration
- [`update`](update.md) - Update fields in an existing configuration
- [`config-update`](config-update.md) - Update configuration with current defaults

### Advanced Operations

- [`schema`](schema.md) - Export cluster JSON schema
- [`migrate`](migrate.md) - Migrate clusters to organization-based structure

## Configuration Storage

Cluster configurations are stored in an organization-based directory structure:

```
~/.config/openCenter/
└── clusters/
    └── <organization>/
        ├── .sops.yaml                    # SOPS configuration
        ├── .<cluster>-config.yaml        # Cluster configuration
        ├── secrets/
        │   ├── age/
        │   │   └── keys/
        │   │       ├── <cluster>.txt     # SOPS private key
        │   │       └── <cluster>.pub     # SOPS public key
        │   └── ssh/
        │       ├── <cluster>-<env>-<region>     # SSH private key
        │       └── <cluster>-<env>-<region>.pub # SSH public key
        └── gitops/
            ├── applications/
            │   └── overlays/<cluster>/
            └── infrastructure/
                └── clusters/<cluster>/
```

## Organization Support

The cluster command supports organization-based multi-tenancy:

- Use `--org` flag or `--opencenter.meta.organization` to specify organization
- Clusters within the same organization share a GitOps repository root
- Each cluster has isolated secrets and configuration
- SOPS keys are scoped per cluster within the organization

## Examples

### Basic Usage

```bash
# Initialize a new cluster
openCenter cluster init my-cluster

# Initialize with organization
openCenter cluster init my-cluster --org production

# List all clusters
openCenter cluster list

# Select active cluster
openCenter cluster select my-cluster

# Show current cluster
openCenter cluster current

# Validate configuration
openCenter cluster validate my-cluster
```

### Complete Workflow

```bash
# 1. Initialize cluster with custom values
openCenter cluster init prod-cluster \
  --org production \
  --opencenter.meta.env=prod \
  --opencenter.cluster.kubernetes.version=1.31.4

# 2. Validate configuration
openCenter cluster validate prod-cluster

# 3. Run preflight checks
openCenter cluster preflight prod-cluster

# 4. Setup GitOps repository
openCenter cluster setup prod-cluster

# 5. Bootstrap the cluster
openCenter cluster bootstrap prod-cluster
```

### Organization-Based Workflow

```bash
# Initialize multiple clusters in same organization
openCenter cluster init dev-cluster --org myorg
openCenter cluster init staging-cluster --org myorg
openCenter cluster init prod-cluster --org myorg

# List all clusters (shows organization/cluster format)
openCenter cluster list

# Select cluster with organization
openCenter cluster select myorg/prod-cluster
```

## Global Flags

All cluster subcommands support the following global flags:

- `--config-dir <path>` - Override default configuration directory
- `--help` - Display help information

## Environment Variables

- `OPENCENTER_CONFIG_DIR` - Override default config directory
- `OPENCENTER_DEBUG` - Enable debug logging and artifacts
- `EDITOR` or `VISUAL` - Preferred editor for `cluster edit` command

## See Also

- [Configuration Reference](../configuration.md) - Detailed configuration options
- [GitOps Integration](../../explanation/gitops.md) - GitOps workflow details
- [SOPS Integration](../../explanation/sops.md) - Secrets management
- [Provider Support](../../explanation/providers.md) - Supported cloud providers
