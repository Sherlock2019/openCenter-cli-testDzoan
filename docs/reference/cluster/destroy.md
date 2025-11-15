# `openCenter cluster destroy` - Destroy a Cluster

## Synopsis
```bash
openCenter cluster destroy <name>
```

## Description

Destroy a cluster by removing its GitOps directory and configuration file. This command permanently deletes the cluster's infrastructure templates, configuration, and GitOps repository.

**WARNING**: This operation is destructive and cannot be undone. The actual cloud infrastructure is not destroyed - only the local configuration and GitOps repository are removed.

## Arguments

### `<name>`
- **Required/Optional**: Required
- **Description**: Name of the cluster to destroy (format: `cluster` or `organization/cluster`)
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Destroy a cluster
```bash
openCenter cluster destroy my-cluster
```

### Destroy cluster in organization
```bash
openCenter cluster destroy production/prod-cluster
```

## Output

```
Cluster "my-cluster" destroyed.
```

## What Gets Destroyed

The command removes the following:

### GitOps Directory
```
<organization>/gitops/
├── applications/overlays/<cluster>/     # Removed
└── infrastructure/clusters/<cluster>/   # Removed
```

### Configuration File
```
<organization>/.<cluster>-config.yaml    # Removed
```

### What Is NOT Destroyed

The following are preserved:

- **SOPS Keys**: Encryption keys in `secrets/age/keys/`
- **SSH Keys**: SSH key pairs in `secrets/ssh/`
- **Organization Directory**: The organization directory structure
- **Other Clusters**: Other clusters in the same organization
- **Cloud Infrastructure**: Actual cloud resources (VMs, networks, etc.)

## Important Notes

### Cloud Infrastructure

The `destroy` command does NOT destroy actual cloud infrastructure. To destroy cloud resources:

1. **For OpenStack/Cloud Providers**:
```bash
cd <git_dir>/infrastructure/clusters/<cluster>
make destroy
# or
terraform destroy
```

2. **For Kind Clusters**:
```bash
kind delete cluster --name <cluster>
```

### Data Loss Warning

This operation permanently deletes:
- Cluster configuration
- GitOps repository content for the cluster
- Infrastructure templates
- Application manifests

**There is no undo operation.**

### Backup Before Destroy

Consider backing up important data:

```bash
# Backup configuration
cp ~/.config/openCenter/clusters/org/.cluster-config.yaml /backup/

# Backup GitOps repository
tar -czf /backup/gitops-cluster.tar.gz ~/.config/openCenter/clusters/org/gitops/

# Then destroy
openCenter cluster destroy my-cluster
```

## Workflow

Typical workflow for destroying a cluster:

```bash
# 1. Destroy cloud infrastructure first
cd ~/.config/openCenter/clusters/org/gitops/infrastructure/clusters/my-cluster
make destroy

# 2. Verify infrastructure is destroyed
# Check cloud provider console

# 3. Destroy local configuration
openCenter cluster destroy my-cluster

# 4. Clean up SOPS keys if needed (optional)
rm ~/.config/openCenter/clusters/org/secrets/age/keys/my-cluster.*
```

## Exit Codes

- `0` - Cluster destroyed successfully
- `1` - Error destroying cluster (e.g., cluster not found)

## Troubleshooting

### Cluster not found
**Error**: `failed to load cluster my-cluster: configuration file not found`

**Solution**: Check available clusters:
```bash
openCenter cluster list
```

### GitOps directory not found
**Error**: `failed to remove gitops directory: no such file or directory`

**Solution**: The GitOps directory may have been manually deleted. The command will still remove the configuration file.

### Permission denied
**Error**: `failed to remove config file: permission denied`

**Solution**: Check file permissions:
```bash
ls -la ~/.config/openCenter/clusters/org/.my-cluster-config.yaml
```

## Recovery

If you accidentally destroyed a cluster:

### From Backup
```bash
# Restore configuration
cp /backup/.cluster-config.yaml ~/.config/openCenter/clusters/org/

# Restore GitOps repository
tar -xzf /backup/gitops-cluster.tar.gz -C ~/
```

### From Git Remote
If GitOps repository was pushed to a remote:
```bash
# Clone from remote
git clone <remote-url> ~/.config/openCenter/clusters/org/gitops

# Recreate configuration
openCenter cluster init my-cluster --force
```

### Reinitialize
If no backup exists:
```bash
# Reinitialize cluster
openCenter cluster init my-cluster

# Reconfigure as needed
openCenter cluster update my-cluster --opencenter.meta.env=prod
```

## See Also

- `openCenter cluster init` - Initialize a new cluster
- `openCenter cluster list` - List all clusters
- `openCenter cluster migrate` - Migrate cluster configurations
