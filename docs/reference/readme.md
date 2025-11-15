# `openCenter` - Kubernetes Cluster Configuration and GitOps Management

## Synopsis
```bash
openCenter [command] [OPTIONS]
```

## Description

openCenter is a command-line tool for managing Kubernetes cluster configurations and GitOps repositories. It provides a declarative approach to cluster lifecycle management with built-in validation, secrets management, and multi-provider support.

The tool streamlines the process of creating, validating, and deploying Kubernetes clusters by turning a single YAML configuration file into a complete, ready-to-use GitOps repository with proper secrets management and infrastructure templates.

## Key Features

- **Declarative YAML-based cluster configuration** - Define your entire cluster in a single configuration file
- **Automatic GitOps repository scaffolding** - Generate complete GitOps structures with templates
- **SOPS integration for secrets management** - Built-in encryption for sensitive data
- **Multi-cloud provider support** - OpenStack, AWS, VMware, Kind, and more
- **Comprehensive validation and preflight checks** - Catch errors before deployment
- **Organization-based multi-tenancy support** - Manage multiple clusters across organizations

## Commands

### Cluster Management
- [`cluster`](cluster/readme.md) - Manage cluster configurations throughout their lifecycle
  - `init` - Initialize new cluster configuration
  - `list` - List all configured clusters
  - `select` - Select active cluster
  - `validate` - Validate cluster configuration
  - `setup` - Setup GitOps repository
  - `bootstrap` - Bootstrap cluster infrastructure
  - And more...

### Configuration Management
- [`config`](config.md) - Manage CLI configuration settings
  - `view` - Display current configuration
  - `set` - Set configuration values
  - `get` - Get configuration values
  - `reset` - Reset to defaults
  - `path` - Show configuration file location
  - `ide` - Setup IDE integration

### Secrets Management
- [`sops`](sops.md) - SOPS key management and automation
  - `generate-key` - Generate new Age key pairs
  - `rotate-key` - Rotate keys and re-encrypt secrets
  - `backup-key` - Backup Age keys
  - `validate` - Validate SOPS configuration
  - `secrets-encrypt` - Encrypt secrets
  - `secrets-decrypt` - Decrypt secrets

### Plugin System
- [`plugins`](plugins.md) - Manage openCenter plugins
  - `list` - List discovered plugins

### Utility Commands
- [`version`](version.md) - Display version and build information

## Global Options

### `--config <path>`
- **Description**: Alternative cluster configuration file path
- **Type**: String
- **Example**: `--config /path/to/cluster-config.yaml`

### `--config-dir <path>`
- **Description**: Configuration directory (defaults to `~/.config/openCenter` on Linux/macOS)
- **Type**: String
- **Example**: `--config-dir ~/my-opencenter-config`

### `--dry-run`
- **Description**: Enable dry-run mode to print planned actions without executing them
- **Type**: Boolean
- **Default**: `false`

### `--log-level <level>`
- **Description**: Set log level explicitly
- **Type**: String
- **Valid Values**: `debug`, `info`, `warn`, `error`
- **Default**: `warn`

### `--set <key>=<value>`
- **Description**: Override configuration values using dot notation
- **Type**: String (repeatable)
- **Example**: `--set logging.level=debug --set behavior.autoConfirm=true`

### `--verbose`
- **Description**: Enable verbose logging (sets log level to debug)
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information

## Examples

### Basic Cluster Workflow

```bash
# Initialize a new cluster configuration
openCenter cluster init my-cluster

# Validate the configuration
openCenter cluster validate my-cluster

# Setup GitOps repository
openCenter cluster setup my-cluster

# Bootstrap the cluster
openCenter cluster bootstrap my-cluster
```

### Organization-Based Workflow

```bash
# Initialize cluster in organization
openCenter cluster init prod-cluster --org production

# List all clusters
openCenter cluster list

# Select active cluster
openCenter cluster select production/prod-cluster

# Show current cluster
openCenter cluster current
```

### Configuration Management

```bash
# View current CLI configuration
openCenter config view

# Set log level
openCenter config set logging.level debug

# Get configuration value
openCenter config get paths.clustersDir

# Reset to defaults
openCenter config reset
```

### SOPS Key Management

```bash
# Generate new Age key pair
openCenter sops generate-key

# Validate SOPS configuration
openCenter sops validate

# Encrypt secrets
openCenter sops secrets-encrypt

# Rotate keys
openCenter sops rotate-key --search-path ./gitops
```

### Using Global Flags

```bash
# Run with verbose logging
openCenter --verbose cluster validate my-cluster

# Dry run mode
openCenter --dry-run cluster bootstrap my-cluster

# Override configuration
openCenter --set logging.level=debug cluster init test-cluster

# Use custom config directory
openCenter --config-dir ~/custom-config cluster list
```

### Plugin Management

```bash
# List available plugins
openCenter plugins list

# Use a plugin command (example)
openCenter my-plugin <args>
```

### Version Information

```bash
# Show full version information
openCenter version

# Show short version only
openCenter version --short
```

## Configuration

### Configuration File Location

The CLI configuration is stored at:
- Linux/macOS: `~/.config/openCenter/config.yaml`
- Override with: `OPENCENTER_CONFIG_DIR` environment variable

### Cluster Configuration Location

Cluster configurations are stored in an organization-based structure:
```
~/.config/openCenter/clusters/
└── <organization>/
    ├── .sops.yaml
    ├── .<cluster>-config.yaml
    ├── secrets/
    └── gitops/
```

## Environment Variables

### `OPENCENTER_CONFIG_DIR`
Override the default configuration directory.
```bash
export OPENCENTER_CONFIG_DIR=~/my-opencenter-config
```

### `OPENCENTER_DEBUG`
Enable debug logging and artifacts.
```bash
export OPENCENTER_DEBUG=1
```

### `OPENCENTER_PLUGINS_DIR`
Directory for external plugins.
```bash
export OPENCENTER_PLUGINS_DIR=~/opencenter-plugins
```

### `EDITOR` or `VISUAL`
Preferred editor for `cluster edit` command.
```bash
export EDITOR=vim
```

### `SOPS_AGE_KEY_FILE`
Path to SOPS Age key file.
```bash
export SOPS_AGE_KEY_FILE=~/.config/sops/age/keys.txt
```

## Exit Codes

- `0` - Success
- `1` - Error occurred

## Configuration Sections

### Logging
- `logging.level` - Log level (debug, info, warn, error)
- `logging.format` - Log format (text, json, yaml)
- `logging.output` - Log output (stdout, stderr, or file path)

### Paths
- `paths.configDir` - Configuration directory
- `paths.clustersDir` - Clusters directory

### Behavior
- `behavior.autoConfirm` - Auto-confirm prompts
- `behavior.dryRun` - Enable dry-run mode by default
- `behavior.verbose` - Enable verbose logging by default

### Defaults
- `defaults.provider` - Default infrastructure provider
- `defaults.region` - Default region
- `defaults.environment` - Default environment

## Plugin System

openCenter supports external plugins that extend functionality. Plugins are discovered in:

1. `OPENCENTER_PLUGINS_DIR` environment variable
2. `<config-dir>/plugins` directory
3. System `PATH`

Plugin binaries must be named `openCenter-<plugin-name>` and be executable.

## Documentation

- **Tutorials**: Step-by-step learning guides
- **How-To Guides**: Problem-solving guides for specific tasks
- **Reference**: Technical reference material (this document)
- **Explanation**: Conceptual explanations and architecture

## Support

- **Documentation**: https://docs.opencenter.cloud
- **Issues**: https://github.com/rackerlabs/openCenter-cli/issues
- **Discussions**: https://github.com/rackerlabs/openCenter-cli/discussions

## See Also

- [Cluster Management](cluster/readme.md) - Complete cluster lifecycle management
- [Configuration Reference](configuration.md) - Detailed configuration options
- [GitOps Integration](../explanation/gitops.md) - GitOps workflow details
- [SOPS Integration](../explanation/sops.md) - Secrets management
- [Provider Support](../explanation/providers.md) - Supported cloud providers
