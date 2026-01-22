# Disaster Recovery Guide


## Table of Contents

- [Overview](#overview)
- [Backup Components](#backup-components)
- [Backup Procedures](#backup-procedures)
- [Restoration Procedures](#restoration-procedures)
- [Key Escrow Procedures](#key-escrow-procedures)
- [Common Disaster Scenarios](#common-disaster-scenarios)
- [Best Practices](#best-practices)
- [Backup Automation](#backup-automation)
- [Support and Troubleshooting](#support-and-troubleshooting)
- [Appendix: Backup File Format](#appendix-backup-file-format)
- [Revision History](#revision-history)
**doc_type: how-to**

This guide provides procedures for backing up and recovering opencenter cluster configurations in disaster scenarios.

## Overview

opencenter provides comprehensive backup and disaster recovery capabilities to protect against:
- Accidental configuration deletion
- Corrupted configuration files
- Lost SOPS encryption keys
- Lost SSH keys
- GitOps repository corruption
- Infrastructure state loss

## Backup Components

Each backup includes:

1. **Cluster Configuration** - The main YAML configuration file
2. **SOPS Age Keys** - Encryption keys for secrets management
3. **SSH Keys** - Keys for cluster access
4. **GitOps State** - Complete GitOps repository contents
5. **Terraform State** - Infrastructure state files

All backups are:
- Compressed with gzip for efficient storage
- Optionally encrypted with AES-256-GCM
- Protected with SHA-256 checksums for integrity verification
- Stored with metadata for easy identification

## Backup Procedures

### Creating a Backup

#### Basic Backup

Create an unencrypted backup:

```bash
opencenter cluster backup create my-cluster
```

This creates a compressed backup in `~/.config/opencenter/backups/` with the naming format:
```
<cluster>-<timestamp>.tar.gz
```

Example: `my-cluster-20260118-143000.tar.gz`

#### Encrypted Backup

Create an encrypted backup (recommended for production):

```bash
# Prompt for passphrase
opencenter cluster backup create my-cluster --encrypt

# Or provide passphrase directly
opencenter cluster backup create my-cluster --passphrase="your-secure-passphrase"
```

**Important**: Store the passphrase securely! Without it, the backup cannot be restored.

#### Automated Backups

Schedule periodic backups (future feature):

```bash
# Daily backups with 30-day retention
opencenter cluster backup schedule my-cluster --interval=24h --retention=30d
```

### Listing Backups

View all backups:

```bash
# List all backups
opencenter cluster backup list

# List backups for specific cluster
opencenter cluster backup list my-cluster
```

Output example:
```
BACKUP ID                        CLUSTER      CREATED              SIZE      LOCATION
my-cluster-20260118-143000      my-cluster   2026-01-18 14:30:00  1048576   ~/.config/opencenter/backups/...
my-cluster-20260117-143000      my-cluster   2026-01-17 14:30:00  1048576   ~/.config/opencenter/backups/...
```

### Verifying Backups

Backups include SHA-256 checksums for integrity verification. The checksum is automatically verified during restoration.

Manual verification:

```bash
# Calculate checksum
sha256sum ~/.config/opencenter/backups/my-cluster-20260118-143000.tar.gz

# Compare with stored checksum
cat ~/.config/opencenter/backups/my-cluster-20260118-143000.tar.gz.sha256
```

## Restoration Procedures

### Restoring from Backup

#### Basic Restoration

Restore from an unencrypted backup:

```bash
opencenter cluster backup restore my-cluster-20260118-143000
```

#### Encrypted Backup Restoration

Restore from an encrypted backup:

```bash
# Prompt for passphrase
opencenter cluster backup restore my-cluster-20260118-143000

# Or provide passphrase directly
opencenter cluster backup restore my-cluster-20260118-143000 --passphrase="your-secure-passphrase"
```

#### Post-Restoration Steps

After restoration, files are placed in a `restored` directory to prevent overwriting existing configurations:

```
~/.config/opencenter/
├── clusters/restored/
│   └── .restored-config.yaml
├── secrets/
│   ├── age/
│   │   └── restored-key.txt
│   └── ssh/
│       └── restored-keys
```

**Manual steps required:**

1. Review restored configuration:
   ```bash
   cat ~/.config/opencenter/clusters/restored/.restored-config.yaml
   ```

2. Move configuration to correct location:
   ```bash
   mv ~/.config/opencenter/clusters/restored/.restored-config.yaml \
      ~/.config/opencenter/clusters/<org>/<cluster>/.<cluster>-config.yaml
   ```

3. Restore Age keys:
   ```bash
   mv ~/.config/opencenter/secrets/age/restored-key.txt \
      ~/.config/opencenter/secrets/age/<cluster>-key.txt
   ```

4. Restore SSH keys:
   ```bash
   mv ~/.config/opencenter/secrets/ssh/restored-keys \
      ~/.config/opencenter/secrets/ssh/<cluster>-<env>-<region>
   chmod 600 ~/.config/opencenter/secrets/ssh/<cluster>-<env>-<region>
   ```

5. Validate restored configuration:
   ```bash
   opencenter cluster validate <cluster>
   ```

### Deleting Backups

Remove old or unnecessary backups:

```bash
# With confirmation prompt
opencenter cluster backup delete my-cluster-20260118-143000

# Without confirmation
opencenter cluster backup delete my-cluster-20260118-143000 --force
```

## Key Escrow Procedures

SOPS Age keys are critical for accessing encrypted secrets. Implement key escrow to prevent permanent data loss.

### Key Backup Strategy

1. **Primary Storage**: OS keyring (Keychain, Credential Manager, Secret Service)
2. **Secondary Storage**: Encrypted backup files
3. **Tertiary Storage**: Offline secure storage (recommended for production)

### Exporting Keys for Escrow

Export Age keys from OS keyring:

```bash
# macOS Keychain
security find-generic-password -s "opencenter" -a "<cluster>-age-key" -w

# Linux Secret Service
secret-tool lookup service opencenter account "<cluster>-age-key"

# Windows Credential Manager
cmdkey /list | findstr opencenter
```

Save exported keys to encrypted storage:

```bash
# Create encrypted key backup
echo "AGE-SECRET-KEY-..." | gpg --symmetric --armor > cluster-age-key.gpg

# Store in secure location (e.g., password manager, hardware security module)
```

### Multi-Key Configuration

Configure multiple Age keys for redundancy:

```yaml
# .sops.yaml
creation_rules:
  - path_regex: .*
    age: >-
      age1primary...,
      age1backup...,
      age1escrow...
```

Benefits:
- Any key can decrypt secrets
- Loss of one key doesn't cause data loss
- Different keys for different purposes (daily use, backup, escrow)

### Key Rotation

Rotate keys periodically (recommended: annually):

```bash
# Generate new key
opencenter sops keygen <cluster> --rotate

# Re-encrypt all secrets with new key
opencenter sops reencrypt <cluster>

# Archive old key
opencenter cluster backup create <cluster> --encrypt
```

## Common Disaster Scenarios

### Scenario 1: Accidental Configuration Deletion

**Problem**: Cluster configuration file was accidentally deleted.

**Solution**:
1. List available backups:
   ```bash
   opencenter cluster backup list my-cluster
   ```

2. Restore from most recent backup:
   ```bash
   opencenter cluster backup restore my-cluster-20260118-143000
   ```

3. Move restored configuration to correct location
4. Validate configuration:
   ```bash
   opencenter cluster validate my-cluster
   ```

**Recovery Time**: 5-10 minutes

### Scenario 2: Lost SOPS Age Key

**Problem**: SOPS Age key was lost or corrupted, cannot decrypt secrets.

**Solution**:

**If backup exists:**
1. Restore from backup:
   ```bash
   opencenter cluster backup restore my-cluster-20260118-143000 --passphrase="..."
   ```

2. Move restored Age key to correct location
3. Test decryption:
   ```bash
   sops -d <encrypted-file>
   ```

**If no backup exists but multi-key configured:**
1. Use alternate Age key to decrypt secrets
2. Generate new primary key:
   ```bash
   opencenter sops keygen my-cluster
   ```
3. Re-encrypt all secrets with new key

**If no backup and single key:**
- **Data loss is permanent**
- Secrets must be re-entered manually
- This is why key escrow is critical!

**Recovery Time**: 
- With backup: 10-15 minutes
- Without backup: Hours to days (manual secret re-entry)

### Scenario 3: Corrupted GitOps Repository

**Problem**: GitOps repository was corrupted or accidentally modified.

**Solution**:
1. Restore from backup:
   ```bash
   opencenter cluster backup restore my-cluster-20260118-143000
   ```

2. Extract GitOps state from backup
3. Re-initialize GitOps repository:
   ```bash
   opencenter cluster setup my-cluster --force
   ```

4. Verify repository contents:
   ```bash
   cd ~/.config/opencenter/gitops
   git log
   git status
   ```

**Recovery Time**: 15-30 minutes

### Scenario 4: Complete System Loss

**Problem**: Entire workstation was lost or destroyed.

**Solution**:

**Prerequisites:**
- Backups stored in remote location (S3, network storage, etc.)
- Backup passphrases stored securely (password manager)

**Steps:**
1. Install opencenter on new system:
   ```bash
   # Install from release
   curl -L https://github.com/rackerlabs/opencenter-cli/releases/latest/download/opencenter-linux-amd64 -o opencenter
   chmod +x opencenter
   sudo mv opencenter /usr/local/bin/
   ```

2. Retrieve backups from remote storage:
   ```bash
   mkdir -p ~/.config/opencenter/backups
   # Copy backups from remote storage
   ```

3. Restore each cluster:
   ```bash
   opencenter cluster backup restore my-cluster-20260118-143000 --passphrase="..."
   ```

4. Move restored files to correct locations
5. Validate all clusters:
   ```bash
   opencenter cluster list
   opencenter cluster validate my-cluster
   ```

**Recovery Time**: 1-2 hours (depending on number of clusters)

### Scenario 5: Terraform State Corruption

**Problem**: Terraform state file was corrupted or lost.

**Solution**:
1. Restore from backup:
   ```bash
   opencenter cluster backup restore my-cluster-20260118-143000
   ```

2. Move restored Terraform state:
   ```bash
   mv ~/.config/opencenter/clusters/restored/terraform.tfstate \
      ~/.config/opencenter/clusters/<org>/<cluster>/terraform.tfstate
   ```

3. Verify state:
   ```bash
   cd ~/.config/opencenter/clusters/<org>/<cluster>
   terraform show
   ```

4. If state is too old, refresh from actual infrastructure:
   ```bash
   terraform refresh
   ```

**Recovery Time**: 10-20 minutes

## Best Practices

### Backup Frequency

- **Development clusters**: Weekly backups
- **Staging clusters**: Daily backups
- **Production clusters**: Daily backups + before any major change

### Backup Retention

- **Development**: 7 days
- **Staging**: 30 days
- **Production**: 90 days

### Backup Storage

- **Local**: `~/.config/opencenter/backups/` (default)
- **Remote**: Copy to S3, network storage, or backup service
- **Offline**: Periodic copies to external media for disaster recovery

### Backup Testing

Test backup restoration regularly:

```bash
# Monthly: Test restore in isolated environment
opencenter cluster backup restore <backup-id> --passphrase="..."

# Quarterly: Full disaster recovery drill
# - Simulate complete system loss
# - Restore all clusters from backups
# - Verify functionality
```

### Security Considerations

1. **Always encrypt production backups**
   - Use strong passphrases (16+ characters)
   - Store passphrases in password manager

2. **Protect backup files**
   - Set restrictive permissions: `chmod 600`
   - Store in encrypted filesystem
   - Limit access to authorized personnel

3. **Implement key escrow**
   - Multiple Age keys per cluster
   - Offline key storage
   - Key rotation schedule

4. **Audit backup access**
   - Log all backup operations
   - Review access logs regularly
   - Alert on unauthorized access

## Backup Automation

### Cron-based Backups

Create a cron job for automated backups:

```bash
# Edit crontab
crontab -e

# Add daily backup at 2 AM
0 2 * * * /usr/local/bin/opencenter cluster backup create my-cluster --passphrase="$(cat ~/.backup-passphrase)" 2>&1 | logger -t opencenter-backup
```

### Backup Script

Create a backup script for multiple clusters:

```bash
#!/bin/bash
# backup-all-clusters.sh

CLUSTERS=$(opencenter cluster list --format=json | jq -r '.[].name')
PASSPHRASE=$(cat ~/.backup-passphrase)

for cluster in $CLUSTERS; do
    echo "Backing up $cluster..."
    opencenter cluster backup create "$cluster" --passphrase="$PASSPHRASE"
    
    # Copy to remote storage
    BACKUP_FILE=$(ls -t ~/.config/opencenter/backups/${cluster}-*.tar.gz.enc | head -1)
    aws s3 cp "$BACKUP_FILE" s3://my-backup-bucket/opencenter/
done

# Clean up old backups (keep last 30 days)
find ~/.config/opencenter/backups/ -name "*.tar.gz*" -mtime +30 -delete
```

### Monitoring

Monitor backup success/failure:

```bash
# Check last backup age
LAST_BACKUP=$(ls -t ~/.config/opencenter/backups/my-cluster-*.tar.gz | head -1)
BACKUP_AGE=$(( ($(date +%s) - $(stat -f %m "$LAST_BACKUP")) / 86400 ))

if [ $BACKUP_AGE -gt 1 ]; then
    echo "WARNING: Last backup is $BACKUP_AGE days old"
    # Send alert
fi
```

## Support and Troubleshooting

### Common Issues

**Issue**: Backup fails with "permission denied"
- **Solution**: Check directory permissions, ensure write access to backup directory

**Issue**: Restore fails with "checksum mismatch"
- **Solution**: Backup file is corrupted, try alternate backup

**Issue**: Restore fails with "decryption failed"
- **Solution**: Incorrect passphrase, verify passphrase is correct

**Issue**: Restored configuration doesn't work
- **Solution**: Configuration may be outdated, review and update as needed

### Getting Help

- Documentation: https://docs.opencenter.cloud/operations/disaster-recovery
- GitHub Issues: https://github.com/rackerlabs/opencenter-cli/issues
- Community Support: https://community.opencenter.cloud

## Appendix: Backup File Format

### Archive Structure

```
backup.tar.gz
├── config.yaml          # Cluster configuration
├── age-key.txt          # SOPS Age key
├── ssh-keys             # SSH keys
├── gitops.tar           # GitOps repository archive
└── terraform.tfstate    # Terraform state
```

### Encryption Format

Encrypted backups use AES-256-GCM with Argon2 key derivation:

```
backup.tar.gz.enc
├── [32 bytes] Salt for Argon2
└── [remaining] AES-256-GCM encrypted data
    ├── [12 bytes] Nonce
    └── [remaining] Ciphertext + authentication tag
```

### Checksum Format

SHA-256 checksum stored in separate file:

```
backup.tar.gz.sha256
└── [64 hex chars] SHA-256 hash of backup file
```

## Revision History

- **2026-01-18**: Initial version
- **Future**: Add automated scheduling, remote storage integration, monitoring
