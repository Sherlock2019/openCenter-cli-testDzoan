# `openCenter cluster info` - Show Configuration for a Cluster

## Synopsis
```bash
openCenter cluster info [name] [OPTIONS]
```

## Description

Display configuration information for a cluster, including metadata, cluster name, and configuration file path. The command can output in human-readable YAML format or machine-readable JSON format.

If no cluster name is provided, displays information for the currently active cluster.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster (format: `cluster` or `organization/cluster`). If not provided, uses the currently active cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `--validate`
- **Description**: Validate cluster configuration invariants and display validation results
- **Type**: Boolean
- **Default**: `false`

### `--json`
- **Description**: Output information in JSON format instead of YAML
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Basic usage
```bash
openCenter cluster info my-cluster
```
Displays cluster information in human-readable format.

### Show info for active cluster
```bash
openCenter cluster info
```
Displays information for the currently active cluster.

### JSON output
```bash
openCenter cluster info my-cluster --json
```
Outputs cluster information as JSON for machine processing.

### Validate configuration
```bash
openCenter cluster info my-cluster --validate
```
Validates the cluster configuration and displays validation results.

### Organization-based cluster
```bash
openCenter cluster info production/prod-cluster
```
Shows information for a cluster in a specific organization.

### Pipe to jq for processing
```bash
openCenter cluster info my-cluster --json | jq '.metadata.env'
```
Extracts specific fields using jq.

## Output

### Human-Readable Format (Default)

```
Cluster: my-cluster
Config Path: /home/user/.config/openCenter/clusters/myorg/.my-cluster-config.yaml

Metadata:
name: my-cluster
cluster_name: my-k8s-cluster
env: dev
region: local
status: planned
organization: myorg
```

### JSON Format (--json)

```json
{
  "config_path": "/home/user/.config/openCenter/clusters/myorg/.my-cluster-config.yaml",
  "metadata": {
    "name": "my-cluster",
    "cluster_name": "my-k8s-cluster",
    "env": "dev",
    "region": "local",
    "status": "planned",
    "organization": "myorg"
  }
}
```

### Validation Output (--validate)

Success:
```
Validation successful.
```

Failure:
```
validation error: kubernetes version must be specified
validation error: infrastructure provider must be one of: openstack, baremetal, kind, vmware
validation failed
```

## Metadata Fields

The command displays the following metadata fields:

### name
The cluster identifier used in openCenter commands.

### cluster_name
The actual Kubernetes cluster name used in cluster resources.

### env
Environment designation (e.g., dev, staging, prod).

### region
Geographic or logical region for the cluster.

### status
Current cluster status (e.g., planned, deployed, destroyed).

### organization
Organization that owns the cluster.

## Exit Codes

- `0` - Success
- `1` - Error loading cluster configuration or validation failure

## Notes

- If no cluster name is provided, uses the currently active cluster
- The `--validate` flag performs comprehensive validation checks
- JSON output is useful for scripting and automation
- Configuration path shows the actual location of the cluster config file
- Metadata includes both `name` (openCenter identifier) and `cluster_name` (Kubernetes name)
- Organization information is included in the metadata
- The command reads from `~/.config/openCenter/clusters/` by default
- Override config directory with `OPENCENTER_CONFIG_DIR` environment variable

## See Also

- `openCenter cluster validate` - Comprehensive cluster validation
- `openCenter cluster edit` - Edit cluster configuration
- `openCenter cluster current` - Show current active cluster
- `openCenter cluster list` - List all clusters
