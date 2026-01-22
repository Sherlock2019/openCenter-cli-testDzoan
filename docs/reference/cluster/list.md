# cluster list


## Table of Contents

- [Synopsis](#synopsis)
- [Aliases](#aliases)
- [Description](#description)
- [Flags](#flags)
- [Examples](#examples)
- [Output Format](#output-format)
- [Directory Structure Support](#directory-structure-support)
- [Active Cluster Indicator](#active-cluster-indicator)
- [Sorting](#sorting)
- [Use Cases](#use-cases)
- [Error Handling](#error-handling)
- [See Also](#see-also)
**doc_type:** reference

List all configured clusters across all organizations.

## Synopsis

```bash
opencenter cluster list [flags]
```

## Aliases

- `ls`

## Description

The `cluster list` command retrieves and displays the names of all configured clusters from the configuration directory. It supports both organization-based and legacy directory structures.

The active cluster is indicated with an asterisk (`*`) prefix.

## Flags

- `--json` - Output cluster names as JSON array

## Examples

```bash
# List all clusters
opencenter cluster list

# List clusters using alias
opencenter cluster ls

# Output as JSON for scripting
opencenter cluster list --json
```

## Output Format

### Default (Plain Text)

```
* my-cluster
  prod-cluster
  staging-cluster
  test-cluster
```

The asterisk (`*`) indicates the currently active cluster.

### JSON Format

```json
[
  "my-cluster",
  "prod-cluster",
  "staging-cluster",
  "test-cluster"
]
```

## Directory Structure Support

The command discovers clusters from multiple directory structures:

### Organization-Based Structure
```
~/.config/opencenter/clusters/<organization>/
└── .<cluster>-config.yaml
```

### Legacy Structure
```
~/.config/opencenter/clusters/<cluster>/
└── .<cluster>-config.yaml
```

### Flat Structure
```
~/.config/opencenter/
└── .<cluster>-config.yaml
```

## Active Cluster Indicator

The active cluster is marked with an asterisk (`*`) in plain text output:

```
* my-cluster      ← Active cluster
  other-cluster
```

To set the active cluster:
```bash
opencenter cluster select my-cluster
```

## Sorting

Cluster names are sorted alphabetically for consistent output.

## Use Cases

### Quick Cluster Discovery
```bash
opencenter cluster list
```

### Scripting Integration
```bash
# Get all cluster names as JSON array
CLUSTERS=$(opencenter cluster list --json)

# Iterate over clusters
for cluster in $(opencenter cluster list --json | jq -r '.[]'); do
  echo "Processing $cluster"
  opencenter cluster validate "$cluster"
done
```

### Check Active Cluster
```bash
# Find active cluster
ACTIVE=$(opencenter cluster list | grep '^\*' | sed 's/^\* //')
echo "Active cluster: $ACTIVE"
```

### Count Clusters
```bash
# Count total clusters
opencenter cluster list | wc -l

# Count using JSON
opencenter cluster list --json | jq 'length'
```

## Error Handling

**No clusters found:**
```
# No output (empty list)
```

**Configuration directory not found:**
```
Error: failed to list clusters: configuration directory not found
```

## See Also

- [cluster select](../cli-commands.md#cluster-select) - Set active cluster
- [cluster current](../cli-commands.md#cluster-current) - Display active cluster
- [cluster info](info.md) - Display cluster information
- [cluster init](init.md) - Initialize new cluster
