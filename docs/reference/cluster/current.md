# `opencenter cluster current` - Show the Current Active Cluster


## Table of Contents

- [Synopsis](#synopsis)
- [Description](#description)
- [Options](#options)
- [Examples](#examples)
- [Output](#output)
- [Exit Codes](#exit-codes)
- [Notes](#notes)
- [See Also](#see-also)
## Synopsis
```bash
opencenter cluster current [OPTIONS]
```

## Description

Display the name of the currently active cluster. The active cluster is used as the default for commands that accept an optional cluster name argument.

The active cluster is stored in `~/.config/opencenter/.active` and can be set using the `opencenter cluster select` command.

## Options

### `-q, --quiet`
- **Description**: Quiet output (just the cluster name without newline)
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Basic usage
```bash
opencenter cluster current
```
Output:
```
production/prod-cluster
```

### Quiet mode
```bash
opencenter cluster current --quiet
```
Output (no newline):
```
production/prod-cluster
```

### Using in scripts
```bash
CLUSTER=$(opencenter cluster current --quiet)
echo "Current cluster is: $CLUSTER"
```
Captures the cluster name in a variable for use in scripts.

### Check if cluster is set
```bash
if opencenter cluster current > /dev/null 2>&1; then
  echo "Active cluster: $(opencenter cluster current)"
else
  echo "No active cluster set"
fi
```
Checks if an active cluster is configured.

### Use with other commands
```bash
opencenter cluster validate $(opencenter cluster current --quiet)
```
Validates the currently active cluster.

## Output

### Normal Mode (Default)
Cluster name with newline:
```
my-cluster
```

or for organization-based clusters:
```
production/prod-cluster
```

### Quiet Mode (-q, --quiet)
Cluster name without newline (useful for scripting):
```
my-cluster
```

### No Active Cluster
If no cluster is set as active, the command produces no output and exits successfully.

## Exit Codes

- `0` - Success (even if no active cluster is set)
- `1` - Error reading active cluster configuration

## Notes

- The active cluster is stored in `~/.config/opencenter/.active`
- If no cluster is active, the command produces no output
- Use `--quiet` flag for scripting to avoid trailing newlines
- The active cluster can be set with `opencenter cluster select`
- Organization-based clusters are displayed in `organization/cluster` format
- The command reads from the configuration directory (default: `~/.config/opencenter/`)
- Override config directory with `OPENCENTER_CONFIG_DIR` environment variable

## See Also

- `opencenter cluster select` - Select the active cluster
- `opencenter cluster list` - List all clusters
- `opencenter cluster info` - Show detailed cluster information
