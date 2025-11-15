# `openCenter cluster list` - List Configured Clusters

## Synopsis
```bash
openCenter cluster list [OPTIONS]
```

## Description

List all configured clusters from the configuration directory. Clusters are displayed in `organization/cluster` format when they belong to an organization, or just the cluster name for legacy clusters.

This command retrieves cluster names from all organization directories under `~/.config/openCenter/clusters/` and displays them in a simple, parseable format.

## Options

### `--json`
- **Description**: Output cluster names as a JSON array for machine-readable consumption
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Basic usage
```bash
openCenter cluster list
```
Output:
```
production/prod-cluster
production/staging-cluster
development/dev-cluster
my-cluster
```

### JSON output
```bash
openCenter cluster list --json
```
Output:
```json
["production/prod-cluster","production/staging-cluster","development/dev-cluster","my-cluster"]
```

### Using alias
```bash
openCenter cluster ls
```
The `ls` alias provides the same functionality as `list`.

### Filtering with grep
```bash
openCenter cluster list | grep production
```
Filter clusters to show only those in the production organization.

### Count clusters
```bash
openCenter cluster list | wc -l
```
Count the total number of configured clusters.

## Output

### Plain Text Format (Default)
One cluster name per line, in `organization/cluster` format:
```
org1/cluster1
org1/cluster2
org2/cluster3
standalone-cluster
```

### JSON Format (--json)
Array of cluster names as JSON:
```json
["org1/cluster1", "org1/cluster2", "org2/cluster3", "standalone-cluster"]
```

## Exit Codes

- `0` - Success
- `1` - Error listing clusters

## Notes

- Clusters are listed in the format `organization/cluster` for organization-based clusters
- Legacy clusters (without organization) are listed with just the cluster name
- The output is sorted alphabetically
- Empty output indicates no clusters are configured
- Use `--json` flag for scripting and automation
- The command reads from `~/.config/openCenter/clusters/` by default
- Override config directory with `OPENCENTER_CONFIG_DIR` environment variable

## See Also

- `openCenter cluster select` - Select the active cluster
- `openCenter cluster current` - Show current active cluster
- `openCenter cluster info` - Show detailed cluster information
- `openCenter cluster init` - Initialize a new cluster
