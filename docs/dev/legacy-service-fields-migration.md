# Legacy Service Fields Migration Guide

## Overview

This guide helps users migrate from deprecated service configuration fields to service-specific configuration in opencenter-cli v2.0.0.

## Table of Contents

- [Overview](#overview)
- [Removed Fields](#removed-fields)
  - [BaseConfig/ServiceCfg Deprecated Fields](#baseconfigservicecfg-deprecated-fields)
  - [VSphereCSIConfig Deprecated Fields](#vspherecsiconfig-deprecated-fields)
  - [LokiConfig Deprecated Fields](#lokiconfig-deprecated-fields)
- [Migration Examples](#migration-examples)
  - [Cert-Manager Email and Region](#cert-manager-email-and-region)
  - [Etcd-Backup S3 Configuration](#etcd-backup-s3-configuration)
  - [vSphere CSI Storage Classes](#vsphere-csi-storage-classes)
  - [Loki Swift Credentials](#loki-swift-credentials)
- [Validation](#validation)
- [Troubleshooting](#troubleshooting)

## Removed Fields

### BaseConfig/ServiceCfg Deprecated Fields

The following fields have been removed from the generic `ServiceCfg` struct. Services that need these fields now have service-specific configuration:

- `Email` - Use service-specific email fields (e.g., `CertManagerConfig.Email`)
- `Region` - Use service-specific region fields (e.g., `CertManagerConfig.Region`, `EtcdBackupConfig.S3Region`)
- `S3Host` - Use service-specific S3 configuration (e.g., `EtcdBackupConfig.S3Host`)
- `S3Region` - Use service-specific S3 configuration (e.g., `EtcdBackupConfig.S3Region`)
- `AlertManagerBaseUrl` - Use alert-proxy specific configuration
- `HTTPRouteFQDN` - Use service-specific hostname configuration

### VSphereCSIConfig Deprecated Fields

The following fields have been removed from `VSphereCSIConfig`. Use the `StorageClasses` array instead:

- `DataStore` - Replaced by `StorageClasses[].Name`
- `DataStoreURL` - Replaced by `StorageClasses[].DatastoreURL`
- `DeleteDataStoreUUID` - No longer needed (use `StorageClasses[].ReclaimPolicy: Delete`)
- `RetainDataStoreName` - Replaced by `StorageClasses[].Name`
- `RetainDataStoreUUID` - No longer needed (use `StorageClasses[].ReclaimPolicy: Retain`)

### LokiConfig Deprecated Fields

The following fields have been removed from `LokiConfig`. Use application credentials instead:

- `SwiftUsername` - Use `SwiftApplicationCredentialID` instead
- `SwiftProjectName` - Use `SwiftApplicationCredentialID` instead

Note: `SwiftPassword` in `LokiSecrets` is still present but deprecated. Use `SwiftApplicationCredentialSecret` instead.

## Migration Examples

### Cert-Manager Email and Region

**Before (v1.x - using deprecated fields):**
```yaml
services:
  cert-manager:
    enabled: true
    email: admin@example.com          # Deprecated generic field
    region: us-east-1                 # Deprecated generic field
    letsencrypt_server: https://acme-v02.api.letsencrypt.org/directory
```

**After (v2.0.0 - using service-specific fields):**
```yaml
services:
  cert-manager:
    enabled: true
    email: admin@example.com          # Service-specific field (CertManagerConfig.Email)
    region: us-east-1                 # Service-specific field (CertManagerConfig.Region)
    letsencrypt_server: https://acme-v02.api.letsencrypt.org/directory
```

**Note:** The field names remain the same, but they are now part of the `CertManagerConfig` struct, not the generic `ServiceCfg` struct. No configuration changes are needed for cert-manager.

### Etcd-Backup S3 Configuration

**Before (v1.x - using deprecated fields):**
```yaml
services:
  etcd-backup:
    enabled: true
    s3_host: https://s3.amazonaws.com  # Deprecated generic field
    s3_region: us-east-1               # Deprecated generic field
```

**After (v2.0.0 - using service-specific fields):**
```yaml
services:
  etcd-backup:
    enabled: true
    s3_host: https://s3.amazonaws.com  # Service-specific field (EtcdBackupConfig.S3Host)
    s3_region: us-east-1               # Service-specific field (EtcdBackupConfig.S3Region)
```

**Note:** The field names remain the same, but they are now part of the `EtcdBackupConfig` struct. No configuration changes are needed for etcd-backup.

### vSphere CSI Storage Classes

**Before (v1.x - using deprecated fields):**
```yaml
services:
  vsphere-csi:
    enabled: true
    datastore_name: san-fc-hlu1-gold
    datastoreurl: ds:///vmfs/volumes/1375553-san-fc-hlu1-Gold
    delete_datastore_uuid: 1375553-san-fc-hlu1-Gold
    retain_datastore_name: san-fc-hlu1-gold-retain
    retain_datastore_uuid: 1375553-san-fc-hlu1-Gold
```

**After (v2.0.0 - using StorageClasses array):**
```yaml
services:
  vsphere-csi:
    enabled: true
    storage_classes:
      - name: san-fc-hlu1-gold-delete
        datastore_url: ds:///vmfs/volumes/1375553-san-fc-hlu1-Gold
        reclaim_policy: Delete
        volume_binding_mode: Immediate
        allow_expansion: true
      - name: san-fc-hlu1-gold-retain
        datastore_url: ds:///vmfs/volumes/1375553-san-fc-hlu1-Gold
        reclaim_policy: Retain
        volume_binding_mode: Immediate
        allow_expansion: true
```

**Migration Steps:**
1. Create a `storage_classes` array
2. Add one entry for each storage class you need
3. Set `reclaim_policy` to `Delete` or `Retain` as appropriate
4. Remove the old deprecated fields

### Loki Swift Credentials

**Before (v1.x - using deprecated username/password):**
```yaml
services:
  loki:
    enabled: true
    loki_storage_type: swift
    swift_auth_url: https://identity.api.rackspacecloud.com/v3
    swift_region: DFW
    swift_username: my-username        # Deprecated
    swift_project_name: my-project     # Deprecated
    swift_container_name: loki-logs

secrets:
  loki:
    swift_password: my-password        # Deprecated
```

**After (v2.0.0 - using application credentials):**
```yaml
services:
  loki:
    enabled: true
    loki_storage_type: swift
    swift_auth_url: https://identity.api.rackspacecloud.com/v3
    swift_region: DFW
    swift_application_credential_id: abc123-def456-ghi789  # Recommended
    swift_container_name: loki-logs
    swift_user_domain_name: Default
    swift_domain_name: Default

secrets:
  loki:
    swift_application_credential_secret: my-app-cred-secret  # Recommended
```

**Migration Steps:**
1. Create an OpenStack application credential in your cloud provider
2. Replace `swift_username` and `swift_project_name` with `swift_application_credential_id`
3. Replace `swift_password` with `swift_application_credential_secret` in secrets
4. Add `swift_user_domain_name` and `swift_domain_name` if required by your provider

**Why Application Credentials?**
- More secure than username/password
- Can be scoped to specific projects and roles
- Can be revoked without changing user password
- Recommended by OpenStack security best practices

## Validation

After migrating your configuration, validate it:

```bash
# Build the latest version
mise run build

# Validate your cluster configuration
./bin/opencenter cluster validate my-cluster

# Check for any deprecation warnings
./bin/opencenter cluster config update my-cluster --dry-run
```

## Troubleshooting

### Error: "unknown field 'email' in ServiceCfg"

**Cause:** You're using the deprecated generic `email` field in a service that doesn't have a service-specific email field.

**Solution:** Check if the service has a service-specific configuration struct with an `Email` field. If not, remove the field from your configuration.

### Error: "unknown field 'datastore_name' in VSphereCSIConfig"

**Cause:** You're using deprecated vSphere CSI fields.

**Solution:** Migrate to the `storage_classes` array as shown in the [vSphere CSI Storage Classes](#vsphere-csi-storage-classes) example.

### Error: "unknown field 'swift_username' in LokiConfig"

**Cause:** You're using deprecated Swift username/password authentication.

**Solution:** Migrate to application credentials as shown in the [Loki Swift Credentials](#loki-swift-credentials) example.

### Configuration not rendering all fields

**Cause:** You may be using deprecated fields that are no longer recognized.

**Solution:** Run `opencenter cluster config update my-cluster --dry-run` to see the complete configuration with all available fields.

## Related Documentation

- [v2.0.0 Breaking Changes](../../BREAKING_CHANGES.md)
- [v1 to v2 Migration Guide](../migration/v1-to-v2.md)
- [Service Configuration Reference](../reference/services.md)
- [vSphere CSI Configuration](../providers/vsphere-csi.md)
- [Loki Configuration](../reference/loki.md)
