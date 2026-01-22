# cluster backup


## Table of Contents

- [Synopsis](#synopsis)
- [Description](#description)
- [Subcommands](#subcommands)
- [Backup Contents](#backup-contents)
- [Backup Format](#backup-format)
- [Storage Location](#storage-location)
- [See Also](#see-also)
**doc_type:** reference

Manage cluster configuration backups for disaster recovery.

## Synopsis

```bash
opencenter cluster backup <subcommand> [flags]
```

## Description

The `cluster backup` command manages backups of cluster configurations and related files. Backups include cluster configuration files, SOPS Age encryption keys, SSH keys, GitOps repository state, and Terraform state files.

Backups are compressed with gzip, include SHA-256 checksums for integrity verification, and can be encrypted with AES-256-GCM using a passphrase.

## Subcommands

### create

Create a backup of cluster configuration and related files.

```bash
opencenter cluster backup create <cluster> [flags]
```

**Flags:**
- `--encrypt` - Encrypt backup (prompts for passphrase)
- `--passphrase string` - Passphrase for backup encryption

**Examples:**

```bash
# Create a backup
opencenter cluster backup create my-cluster

# Create an encrypted backup (prompts for passphrase)
opencenter cluster backup create my-cluster --encrypt

# Create an encrypted backup with passphrase
opencenter cluster backup create my-cluster --passphrase="secret123"
```

**Output:**
- Backup ID (format: `<cluster>-YYYYMMDD-HHMMSS`)
- Backup size in bytes
- SHA-256 checksum
- Storage location path
- Retention expiration date

### restore

Restore cluster configuration from a backup.

```bash
opencenter cluster backup restore <backup-id> [flags]
```

**Flags:**
- `--passphrase string` - Passphrase for backup decryption

**Examples:**

```bash
# Restore from backup
opencenter cluster backup restore my-cluster-20260118-143000

# Restore from encrypted backup
opencenter cluster backup restore my-cluster-20260118-143000 --passphrase="secret123"
```

**Behavior:**
- Restored files are placed in a `restored` directory to avoid overwriting existing configurations
- Manual review and file movement required after restore
- Prompts for passphrase if backup is encrypted and passphrase not provided

**Restored File Locations:**
- Config: `~/.config/opencenter/clusters/restored/.restored-config.yaml`
- Age key: `~/.config/opencenter/secrets/age/restored-key.txt`
- SSH keys: `~/.config/opencenter/secrets/ssh/restored-keys`

### list

List all backups for a cluster or all clusters.

```bash
opencenter cluster backup list [cluster]
```

**Examples:**

```bash
# List all backups
opencenter cluster backup list

# List backups for a specific cluster
opencenter cluster backup list my-cluster
```

**Output Format:**

```
BACKUP ID                      CLUSTER      CREATED              SIZE    LOCATION
my-cluster-20260118-143000    my-cluster   2026-01-18 14:30:00  1024    /path/to/backup.tar.gz
```

### delete

Delete a backup by its ID.

```bash
opencenter cluster backup delete <backup-id> [flags]
```

**Flags:**
- `--force` - Delete without confirmation prompt

**Examples:**

```bash
# Delete a backup (with confirmation)
opencenter cluster backup delete my-cluster-20260118-143000

# Delete without confirmation
opencenter cluster backup delete my-cluster-20260118-143000 --force
```

**Warning:** This operation is irreversible.

### schedule

Schedule periodic backups for a cluster (not yet implemented).

```bash
opencenter cluster backup schedule <cluster> [flags]
```

**Flags:**
- `--interval string` - Backup interval (default: "24h")
- `--retention string` - Backup retention period (default: "30d")

**Examples:**

```bash
# Schedule daily backups
opencenter cluster backup schedule my-cluster --interval=24h

# Schedule with retention policy
opencenter cluster backup schedule my-cluster --interval=24h --retention=30d
```

**Status:** This feature is not yet implemented and will be available in a future release.

## Backup Contents

Each backup includes:
- Cluster configuration YAML file
- SOPS Age encryption keys
- SSH key pairs (private and public)
- GitOps repository state
- Terraform/OpenTofu state files

## Backup Format

- Compression: gzip
- Encryption: AES-256-GCM (optional, passphrase-protected)
- Integrity: SHA-256 checksum
- Naming: `<cluster>-YYYYMMDD-HHMMSS.tar.gz`

## Storage Location

Backups are stored in: `~/.config/opencenter/backups/`

## See Also

- [cluster destroy](destroy.md) - Destroy cluster infrastructure
- [cluster init](init.md) - Initialize new cluster configuration
