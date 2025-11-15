# `openCenter cluster config-update` - Update Configuration with Current Defaults

## Synopsis
```bash
openCenter cluster config-update [name]
```

## Description

Update a cluster configuration by reloading it with current defaults and saving it back. This command loads the existing cluster configuration, merges it with current schema defaults, and saves the updated configuration.

This is useful for applying new default values from schema updates, normalizing configuration format, or migrating configurations to new structure.

If no cluster name is provided, the currently selected cluster is updated.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster to update (format: `cluster` or `organization/cluster`). If not provided, updates the currently selected cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Update currently selected cluster
```bash
openCenter cluster config-update
```

### Update specific cluster
```bash
openCenter cluster config-update my-cluster
```

### Update cluster in organization
```bash
openCenter cluster config-update production/prod-cluster
```

## Output

```
Loading configuration for cluster 'my-cluster'...
Validating configuration...
Creating backup at: /home/user/.config/openCenter/clusters/org/.my-cluster-config.yaml.20251117-103000.backup
Saving updated configuration...
Successfully updated configuration for cluster 'my-cluster'
Configuration saved to: /home/user/.config/openCenter/clusters/org/.my-cluster-config.yaml
```

With validation warnings:
```
Loading configuration for cluster 'my-cluster'...
Validating configuration...
Warning: Configuration has validation errors:
  - kubernetes version should be updated to latest stable
  - network CIDR uses deprecated format

Proceeding with update anyway...
Creating backup at: /home/user/.config/openCenter/clusters/org/.my-cluster-config.yaml.20251117-103000.backup
Saving updated configuration...
Successfully updated configuration for cluster 'my-cluster'
Configuration saved to: /home/user/.config/openCenter/clusters/org/.my-cluster-config.yaml
```

## Update Process

The command performs the following steps:

1. **Load Configuration** - Loads existing cluster configuration
2. **Validate** - Validates configuration (warnings only, does not fail)
3. **Create Backup** - Creates timestamped backup of current configuration
4. **Merge Defaults** - Merges configuration with current schema defaults
5. **Save Configuration** - Saves updated configuration to file

## Backup

A backup is automatically created before updating:

```
<config-path>.<timestamp>.backup
```

Example:
```
/home/user/.config/openCenter/clusters/org/.my-cluster-config.yaml.20251117-103000.backup
```

## Use Cases

### Apply Schema Updates
After updating openCenter to a new version with schema changes:
```bash
openCenter cluster config-update my-cluster
```

### Normalize Configuration Format
Standardize configuration format across clusters:
```bash
for cluster in $(openCenter cluster list); do
  openCenter cluster config-update "$cluster"
done
```

### Migrate to New Defaults
Apply new default values without manual editing:
```bash
openCenter cluster config-update my-cluster
```

### Fix Configuration Issues
Resolve configuration inconsistencies:
```bash
openCenter cluster config-update my-cluster
openCenter cluster validate my-cluster
```

## Notes

- A backup is always created before updating
- Validation warnings do not prevent the update
- The command preserves user-configured values
- New default values are added for missing fields
- Existing values are not overwritten
- The backup includes a timestamp for easy identification
- Use `cluster validate` after update to check for issues
- The command does not re-render GitOps templates
- Run `cluster setup --render` to apply changes to GitOps

## Troubleshooting

### Cluster not found
**Error**: `failed to load cluster configuration: configuration file not found`

**Solution**: Check available clusters:
```bash
openCenter cluster list
```

### Permission denied
**Error**: `failed to create backup: permission denied`

**Solution**: Check file permissions:
```bash
ls -la ~/.config/openCenter/clusters/org/
```

### Validation errors
**Warning**: `Configuration has validation errors`

**Solution**: The update proceeds anyway. Fix validation errors after update:
```bash
openCenter cluster config-update my-cluster
openCenter cluster validate my-cluster
openCenter cluster update my-cluster --opencenter.meta.env=prod
```

## Recovery

If the update causes issues, restore from backup:

```bash
# Find backup
ls -la ~/.config/openCenter/clusters/org/.my-cluster-config.yaml.*

# Restore backup
cp ~/.config/openCenter/clusters/org/.my-cluster-config.yaml.20251117-103000.backup \
   ~/.config/openCenter/clusters/org/.my-cluster-config.yaml

# Verify
openCenter cluster validate my-cluster
```

## See Also

- `openCenter cluster update` - Update specific configuration fields
- `openCenter cluster validate` - Validate configuration
- `openCenter cluster edit` - Edit configuration manually
- `openCenter cluster info` - Show configuration information
