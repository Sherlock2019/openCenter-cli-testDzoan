# Session-Scoped Cluster Selection

This guide shows you how to work with multiple clusters simultaneously using session-scoped cluster selection.

## Problem

When working with multiple clusters, the default `opencenter cluster select` command sets a global cluster selection that affects all terminal sessions. This can be problematic when:

- Two engineers are working on different clusters on the same server
- You need to run commands on multiple clusters simultaneously
- You want to test changes on one cluster while monitoring another

## Solution

Enable shell integration to use session-scoped cluster selection with `opencenter cluster use`.

## Quick Start

### 1. Install Shell Integration

```bash
opencenter config ide --shell-integration
```

### 2. Reload Your Shell

```bash
source ~/.bashrc  # or ~/.zshrc for zsh
```

### 3. Use Session-Scoped Selection

```bash
# Terminal 1
opencenter cluster use prod-cluster
opencenter cluster status

# Terminal 2 (independent)
opencenter cluster use dev-cluster
opencenter cluster validate
```

## Use Cases

### Multiple Engineers on Same Server

**Engineer 1** (working on production):
```bash
opencenter cluster use prod-cluster
opencenter cluster status
opencenter cluster drift
```

**Engineer 2** (working on staging):
```bash
opencenter cluster use staging-cluster
opencenter cluster validate
opencenter cluster bootstrap
```

Both engineers can work independently without interfering with each other's cluster context.

### Parallel Operations

Monitor multiple clusters simultaneously:

```bash
# Terminal 1: Monitor production
opencenter cluster use prod-cluster
watch opencenter cluster status

# Terminal 2: Monitor staging
opencenter cluster use staging-cluster
watch opencenter cluster status

# Terminal 3: Work on development
opencenter cluster use dev-cluster
opencenter cluster update
```

### Testing and Validation

Test changes on one cluster while keeping another as reference:

```bash
# Terminal 1: Test cluster
opencenter cluster use test-cluster
opencenter cluster validate
opencenter cluster bootstrap

# Terminal 2: Reference cluster
opencenter cluster use prod-cluster
opencenter cluster info  # Compare configuration
```

## Advanced Usage

### Temporary Environment Variable Override

You can temporarily override the cluster selection without using `opencenter cluster use`:

```bash
# One-off command on different cluster
OPENCENTER_CLUSTER=prod-cluster opencenter cluster status

# Multiple commands
export OPENCENTER_CLUSTER=prod-cluster
opencenter cluster status
opencenter cluster info
unset OPENCENTER_CLUSTER
```

### Scripting with Session Selection

```bash
#!/bin/bash
# Script that works with multiple clusters

# Enable shell integration in script
eval "$(opencenter shell-init)"

# Work with cluster 1
opencenter cluster use cluster1
echo "Validating cluster1..."
opencenter cluster validate

# Work with cluster 2
opencenter cluster use cluster2
echo "Validating cluster2..."
opencenter cluster validate
```

### Check Current Cluster and Source

```bash
# Show cluster with source information
opencenter cluster current
# Output: prod-cluster (session)

# Just the cluster name (for scripting)
CLUSTER=$(opencenter cluster current --quiet)
echo "Working on: $CLUSTER"
```

## Best Practices

1. **Use `cluster use` for temporary work**: When you need to switch clusters temporarily in a specific terminal
2. **Use `cluster select` for default cluster**: Set your most-used cluster as the persistent default
3. **Add prompt integration**: Always know which cluster you're working on
4. **Document your workflow**: Add comments in scripts that use session-scoped selection

## Troubleshooting

### Commands still use wrong cluster

Make sure you're using `opencenter cluster use` (not `select`) and that shell integration is active:

```bash
# Check if shell integration is active
echo $OPENCENTER_SESSION_FILE
# Should output: /home/user/.config/openCenter/.session-XXXXX

# If empty, reload shell integration
eval "$(opencenter shell-init)"
```

### Session not isolated

Verify that each terminal has a unique session ID:

```bash
echo $OPENCENTER_SESSION_ID
# Should be unique per terminal
```

## See Also

- [Shell Integration Reference](../reference/shell-integration.md)
- [CLI Commands](../reference/cli-commands.md)
- [Multi-Cluster Management](multi-cluster.md)
