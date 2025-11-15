# `openCenter cluster init` - Initialize a New Cluster Configuration

## Synopsis
```bash
openCenter cluster init [name] [OPTIONS]
```

## Description

Initialize a new cluster configuration with default values based on the JSON schema. This command creates a complete cluster configuration file with sensible defaults that can be customized using command-line flags with dot notation.

The configuration is created in an organization-based directory structure at `~/.config/openCenter/clusters/<organization>/<cluster>/`. By default, the cluster name is used as the organization name unless specified otherwise.

SOPS Age encryption keys and SSH key pairs are automatically generated and stored in the cluster's secrets directory unless `--no-keygen` is specified.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster to initialize. If not provided, uses the currently active cluster (requires `--force` to overwrite)
- **Example**: `my-cluster`

## Options

### `--org <organization>`
- **Description**: Organization name for the cluster
- **Type**: String
- **Default**: Uses cluster name if not specified
- **Example**: `--org production`

### `--type <provider>`
- **Description**: Cluster infrastructure provider type
- **Type**: String
- **Default**: `openstack`
- **Valid Values**: `openstack`, `baremetal`, `kind`, `vmware`
- **Example**: `--type baremetal`

### `--force`
- **Description**: Overwrite existing cluster configuration if it already exists
- **Type**: Boolean
- **Default**: `false`

### `--strict`
- **Description**: Enable strict validation during initialization (fails if required values are missing)
- **Type**: Boolean
- **Default**: `false`

### `--no-keygen`
- **Description**: Do not auto-generate SOPS age keys and SSH key pairs
- **Type**: Boolean
- **Default**: `false`

### `--<config.path>=<value>`
- **Description**: Override any configuration value using dot notation
- **Type**: Dynamic (string, int, bool based on schema)
- **Example**: `--opencenter.meta.env=prod`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Configuration Override

You can override any configuration value using dot notation flags:

### Organization
```bash
--org myorg
--opencenter.meta.organization=myorg
```

### Cluster Type
```bash
--type baremetal
--opencenter.infrastructure.provider=aws
```

### Metadata
```bash
--opencenter.meta.env=prod
--opencenter.meta.region=us-east-1
--opencenter.meta.status=planned
```

### Kubernetes Configuration
```bash
--opencenter.cluster.kubernetes.version=1.31.4
--opencenter.cluster.cluster_name=my-k8s-cluster
```

### GitOps Configuration
```bash
--opencenter.gitops.git_dir=/custom/path/to/gitops
--opencenter.gitops.git_ssh_key=/path/to/ssh/key
```

## Examples

### Basic usage
```bash
openCenter cluster init my-cluster
```
Creates a cluster with default OpenStack provider, using cluster name as organization.

### Initialize bare metal cluster
```bash
openCenter cluster init my-cluster --org myorg --type baremetal
```
Creates a bare metal cluster in the "myorg" organization.

### Initialize with organization using --org flag
```bash
openCenter cluster init my-cluster --org production
```
Creates a cluster in the "production" organization.

### Initialize with organization using dot notation
```bash
openCenter cluster init my-cluster --opencenter.meta.organization=production
```
Alternative way to specify organization using configuration path.

### Initialize with custom values
```bash
openCenter cluster init my-cluster \
  --org production \
  --opencenter.meta.env=prod \
  --opencenter.cluster.kubernetes.version=1.31.4 \
  --opencenter.infrastructure.provider=aws
```
Creates a production cluster with custom Kubernetes version and AWS provider.

### Initialize without key generation
```bash
openCenter cluster init my-cluster --no-keygen
```
Creates cluster configuration without generating SOPS and SSH keys (useful for testing or when keys are managed externally).

### Force overwrite existing configuration
```bash
openCenter cluster init my-cluster --force
```
Overwrites existing cluster configuration if it already exists.

### Force overwrite active cluster configuration
```bash
openCenter cluster init --force
```
Overwrites the currently active cluster configuration.

### Initialize with strict validation
```bash
openCenter cluster init my-cluster --strict
```
Validates configuration during initialization and fails if validation errors are found.

## Output

The command creates the following structure:

```
~/.config/openCenter/clusters/<organization>/
├── .sops.yaml                           # SOPS configuration
├── .<cluster>-config.yaml               # Cluster configuration
├── secrets/
│   ├── age/
│   │   └── keys/
│   │       ├── <cluster>.txt            # SOPS private key
│   │       └── <cluster>.pub            # SOPS public key
│   └── ssh/
│       ├── <cluster>-<env>-<region>     # SSH private key
│       └── <cluster>-<env>-<region>.pub # SSH public key
└── gitops/                              # GitOps repository root
```

Success output:
```
Created cluster configuration in organization 'myorg' at '/home/user/.config/openCenter/clusters/myorg/my-cluster'
GitOps repository root: /home/user/.config/openCenter/clusters/myorg/gitops
SOPS key location: /home/user/.config/openCenter/clusters/myorg/secrets/age/keys/my-cluster.txt
Generated ed25519 SSH key pair at /home/user/.config/openCenter/clusters/myorg/secrets/ssh/my-cluster-dev-local
```

## Troubleshooting

### Cluster already exists
**Error**: `cluster configuration directory 'my-cluster' already exists in organization 'myorg', use --force to overwrite`

**Solution**: Use `--force` flag to overwrite existing configuration:
```bash
openCenter cluster init my-cluster --force
```

### No cluster name provided
**Error**: `no cluster name provided and no active cluster set`

**Solution**: Either provide a cluster name or set an active cluster:
```bash
openCenter cluster init my-cluster
# or
openCenter cluster select existing-cluster
openCenter cluster init --force
```

### Validation failures with --strict
**Error**: `validation failed`

**Solution**: Check error messages for specific validation failures and provide required values:
```bash
openCenter cluster init my-cluster --strict \
  --opencenter.meta.env=dev \
  --opencenter.meta.region=local
```

### Invalid organization name
**Error**: `invalid organization name 'my org': organization name cannot contain spaces`

**Solution**: Use valid organization names (alphanumeric, hyphens, underscores):
```bash
openCenter cluster init my-cluster --org my-org
```

## Notes

- Configuration files are created with restrictive permissions (0600) for security
- SOPS keys are automatically generated with proper Age key format
- SSH keys are generated using ed25519 cipher by default (can be customized)
- Organization directories are created with .gitignore files automatically
- The cluster name must be unique within an organization
- Use `--no-keygen` when keys are managed externally or for testing purposes
- Check `~/.config/openCenter/clusters/` for created files and structure

## See Also

- `openCenter cluster validate` - Validate cluster configuration
- `openCenter cluster edit` - Edit cluster configuration
- `openCenter cluster list` - List all clusters
- `openCenter cluster select` - Select active cluster
- `openCenter cluster setup` - Setup GitOps repository
