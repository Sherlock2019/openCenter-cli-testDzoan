# v1 to v2 Configuration Migration Guide

## Table of Contents

- [Overview](#overview)
- [Migration Timeline](#migration-timeline)
- [Key Changes in v2](#key-changes-in-v2)
- [Field Relocation Mapping](#field-relocation-mapping)
- [Migration Process](#migration-process)
- [Migration Command Usage](#migration-command-usage)
- [Post-Migration Validation](#post-migration-validation)
- [Common Migration Scenarios](#common-migration-scenarios)
- [Troubleshooting](#troubleshooting)
- [Backward Compatibility](#backward-compatibility)

## Overview

The v2 cluster configuration schema represents a significant redesign of opencenter-cli's configuration system. This guide helps you migrate existing v1 configurations to the new v2 format.

**Why Migrate?**

- **Eliminates Duplication**: Single source of truth for each setting
- **Clear Ownership**: Hierarchical domain model (Meta, Cluster, Infrastructure, Deployment, Services)
- **Provider Isolation**: Provider-specific settings cleanly separated
- **Advanced Features**: Support for Kamaji hosted control planes, reference resolution, provider-region defaults
- **Better Validation**: Multi-layered validation with clear error messages

**Migration Strategy**: The migration tool automatically relocates fields, applies hydration to make implicit defaults explicit, and validates the resulting v2 configuration.

## Migration Timeline

### Phase 1: Coexistence (Current)

- **Status**: Both v1 and v2 schemas supported simultaneously
- **Default**: v1 (when `schema_version` field is missing)
- **Action**: Begin testing v2 configurations in non-production environments
- **Timeline**: 6 months

### Phase 2: v2 Recommended (Q3 2026)

- **Status**: v2 becomes recommended default for new clusters
- **Default**: v2 for new `cluster init` commands
- **Action**: Migrate production clusters during maintenance windows
- **Timeline**: 6 months

### Phase 3: v1 Deprecated (Q1 2027)

- **Status**: v1 marked as deprecated with warnings
- **Default**: v2 required for new clusters
- **Action**: Complete migration of all remaining v1 clusters
- **Timeline**: 6 months

### Phase 4: v1 Removed (Q3 2027)

- **Status**: v1 support removed from opencenter-cli
- **Default**: v2 only
- **Action**: All clusters must use v2 schema
- **Timeline**: Ongoing

## Key Changes in v2

### 1. Schema Version Field (Required)

```yaml
# v2 configurations MUST include schema_version
schema_version: "2.0"
```

### 2. Hierarchical Domain Model

v2 organizes configuration into five clear domains:

```yaml
opencenter:
  meta:             # Cluster identity (name, org, env, region)
  cluster:          # Kubernetes configuration (provider-agnostic)
  infrastructure:   # Physical resources (provider-specific)
  services:         # Self-hosted platform workloads
  managed_services: # External/vendor-managed services
```

### 3. Deployment Configuration Promoted

Deployment configuration moved from `opencenter.deployment` to root-level `deployment`:

```yaml
# v1 (deprecated)
opencenter:
  deployment:
    method: kubespray

# v2 (correct)
deployment:
  method: kubespray
  kubespray:
    version: "v2.29.1"
```

### 4. VRRP IP Consolidation

VRRP IP now has a single location:

```yaml
# v1 (multiple locations - ambiguous)
cluster:
  networking:
    vrrp_ip: "10.2.128.5"
infrastructure:
  cloud:
    openstack:
      vrrp_ip: "10.2.128.5"  # Duplicate!

# v2 (single source of truth)
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "10.2.128.5"
```

### 5. Storage Configuration Relocated

Storage moved from `opencenter.storage` to `infrastructure.storage`:

```yaml
# v1 (deprecated)
opencenter:
  storage:
    default_storage_class: "csi-cinder-sc-delete"
    worker_volume_size: 100

# v2 (correct)
opencenter:
  infrastructure:
    storage:
      default_storage_class: "csi-cinder-sc-delete"
      worker_volume_size: 100
```

### 6. Worker Pools Relocated

Additional worker pools moved from `cluster.kubernetes` to `infrastructure.compute`:

```yaml
# v1 (deprecated)
opencenter:
  cluster:
    kubernetes:
      additional_server_pools_worker:
        - name: "high-memory"
          worker_count: 2

# v2 (correct)
opencenter:
  infrastructure:
    compute:
      additional_server_pools_worker:
        - name: "high-memory"
          worker_count: 2
```

### 7. CSI Plugin Selection

CSI plugin selection added (similar to CNI):

```yaml
opencenter:
  cluster:
    kubernetes:
      # CNI plugin selection
      network_plugin:
        calico:
          enabled: true
      
      # NEW: CSI plugin selection
      storage_plugin:
        cinder_csi:
          enabled: true
          version: "v1.28.0"
```

### 8. Reference Resolution

v2 supports explicit reference syntax:

```yaml
opencenter:
  infrastructure:
    networking:
      vrrp_ip: "10.2.128.5"
  
  services:
    calico:
      calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:6443"
```

## Field Relocation Mapping

### Complete Field Mapping Table

| v1 Location | v2 Location | Notes |
|-------------|-------------|-------|
| `cluster.networking.vrrp_ip` | `infrastructure.networking.vrrp_ip` | Single source of truth |
| `cluster.networking.subnet_nodes` | `infrastructure.networking.subnet_nodes` | Infrastructure-level networking |
| `cluster.networking.allocation_pool_*` | `infrastructure.networking.allocation_pool_*` | Infrastructure-level networking |
| `cluster.networking.dns_nameservers` | `infrastructure.networking.dns_nameservers` | Infrastructure-level networking |
| `cluster.networking.ntp_servers` | `infrastructure.networking.ntp_servers` | Infrastructure-level networking |
| `cluster.networking.use_octavia` | `infrastructure.networking.use_octavia` | Infrastructure-level networking |
| `cluster.networking.use_designate` | `infrastructure.networking.use_designate` | Infrastructure-level networking |
| `cluster.kubernetes.additional_server_pools_worker` | `infrastructure.compute.additional_server_pools_worker` | Compute resources |
| `opencenter.storage.*` | `infrastructure.storage.*` | Storage configuration |
| `opencenter.deployment.*` | `deployment.*` | Root-level deployment config |
| `cluster.ssh.*` | `infrastructure.ssh.*` | Infrastructure-level SSH config |
| `infrastructure.cloud.openstack.vrrp_ip` | `infrastructure.networking.vrrp_ip` | Consolidated to single location |

### Networking Fields

#### Infrastructure Networking (Physical Network)

**v1 Location**: `cluster.networking.*`  
**v2 Location**: `infrastructure.networking.*`

```yaml
# v1
opencenter:
  cluster:
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.128.10"
      allocation_pool_end: "10.2.131.254"
      vrrp_ip: "10.2.128.5"
      vrrp_enabled: true
      use_octavia: true
      loadbalancer_provider: "octavia"
      use_designate: true
      dns_zone_name: "acme-corp.com"
      dns_nameservers:
        - "8.8.8.8"
      ntp_servers:
        - "time.google.com"

# v2
opencenter:
  infrastructure:
    networking:
      subnet_nodes: "10.2.128.0/22"
      allocation_pool_start: "10.2.128.10"
      allocation_pool_end: "10.2.131.254"
      vrrp_ip: "10.2.128.5"
      vrrp_enabled: true
      use_octavia: true
      loadbalancer_provider: "octavia"
      use_designate: true
      dns_zone_name: "acme-corp.com"
      dns_nameservers:
        - "8.8.8.8"
      ntp_servers:
        - "time.google.com"
```

#### Kubernetes Networking (CNI-Managed)

**v1 Location**: `cluster.kubernetes.subnet_*`  
**v2 Location**: `cluster.kubernetes.subnet_*` (unchanged)

```yaml
# Both v1 and v2 (no change)
opencenter:
  cluster:
    kubernetes:
      subnet_pods: "10.233.64.0/18"
      subnet_services: "10.233.0.0/18"
      network_plugin:
        calico:
          enabled: true
```

### Compute Fields

**v1 Location**: `cluster.kubernetes.additional_server_pools_worker`  
**v2 Location**: `infrastructure.compute.additional_server_pools_worker`

```yaml
# v1
opencenter:
  cluster:
    kubernetes:
      additional_server_pools_worker:
        - name: "high-memory"
          worker_count: 2
          flavor_worker: "gp.0.8.64"

# v2
opencenter:
  infrastructure:
    compute:
      additional_server_pools_worker:
        - name: "high-memory"
          worker_count: 2
          flavor_worker: "gp.0.8.64"
```

### Storage Fields

**v1 Location**: `opencenter.storage.*`  
**v2 Location**: `infrastructure.storage.*`

```yaml
# v1
opencenter:
  storage:
    default_storage_class: "csi-cinder-sc-delete"
    worker_volume_size: 100
    worker_volume_destination_type: "volume"
    worker_volume_source_type: "image"
    worker_volume_type: "HA-Standard"
    additional_block_devices:
      - device_name: "/dev/vdb"
        volume_size: 500

# v2
opencenter:
  infrastructure:
    storage:
      default_storage_class: "csi-cinder-sc-delete"
      worker_volume_size: 100
      worker_volume_destination_type: "volume"
      worker_volume_source_type: "image"
      worker_volume_type: "HA-Standard"
      additional_block_devices:
        - device_name: "/dev/vdb"
          volume_size: 500
```

### SSH Configuration

**v1 Location**: `cluster.ssh.*`  
**v2 Location**: `infrastructure.ssh.*`

```yaml
# v1
opencenter:
  cluster:
    ssh:
      user: "ubuntu"
      key_path: "~/.ssh/cluster-key"

# v2
opencenter:
  infrastructure:
    ssh:
      user: "ubuntu"
      key_path: "~/.ssh/cluster-key"
```

### Deployment Configuration

**v1 Location**: `opencenter.deployment.*`  
**v2 Location**: `deployment.*` (root level)

```yaml
# v1
opencenter:
  deployment:
    method: "kubespray"
    kubespray:
      version: "v2.29.1"

# v2
deployment:
  auto_deploy: true
  method: "kubespray"
  kubespray:
    version: "v2.29.1"
```

## Migration Process

### Step 1: Backup Existing Configuration

```bash
# Backup your v1 configuration
cp ~/.config/opencenter/clusters/myorg/.mycluster-config.yaml \
   ~/.config/opencenter/clusters/myorg/.mycluster-config.yaml.v1.backup
```

### Step 2: Run Migration Command

```bash
# Migrate v1 to v2
opencenter cluster migrate-config \
  --input ~/.config/opencenter/clusters/myorg/.mycluster-config.yaml \
  --output ~/.config/opencenter/clusters/myorg/.mycluster-config-v2.yaml
```

### Step 3: Review Migration Report

The migration tool generates a detailed report:

```
Migration Report for: mycluster
================================

Schema Version: v1.0 → v2.0

Field Relocations:
------------------
✓ cluster.networking.vrrp_ip → infrastructure.networking.vrrp_ip
✓ cluster.networking.subnet_nodes → infrastructure.networking.subnet_nodes
✓ cluster.networking.dns_nameservers → infrastructure.networking.dns_nameservers
✓ cluster.networking.ntp_servers → infrastructure.networking.ntp_servers
✓ cluster.kubernetes.additional_server_pools_worker → infrastructure.compute.additional_server_pools_worker
✓ opencenter.storage.* → infrastructure.storage.*
✓ opencenter.deployment.* → deployment.*
✓ cluster.ssh.* → infrastructure.ssh.*

Applied Defaults (Hydration):
------------------------------
✓ infrastructure.cloud.openstack.image_id: "799dcf97-3656-4361-8187-13ab1b295e33" (provider-region: sjc3)
✓ infrastructure.cloud.openstack.availability_zones: ["az1", "az2", "az3"] (provider-region: sjc3)
✓ infrastructure.networking.ntp_servers: ["time.sjc3.rackspace.com"] (provider-region: sjc3)

Warnings:
---------
⚠ VRRP IP found in multiple locations (cluster.networking and infrastructure.cloud.openstack)
  → Consolidated to infrastructure.networking.vrrp_ip
⚠ Deprecated field: opencenter.storage (use infrastructure.storage in v2)

Validation: PASSED ✓

Next Steps:
-----------
1. Review the migrated configuration: ~/.config/opencenter/clusters/myorg/.mycluster-config-v2.yaml
2. Validate the configuration: opencenter cluster validate --config .mycluster-config-v2.yaml
3. Test in non-production environment before deploying
4. Update your GitOps repository with the new v2 configuration
```

### Step 4: Validate Migrated Configuration

```bash
# Validate the v2 configuration
opencenter cluster validate \
  --config ~/.config/opencenter/clusters/myorg/.mycluster-config-v2.yaml

# Expected output:
# ✓ Schema validation passed
# ✓ Business rules validation passed
# ✓ Provider validation passed (OpenStack)
# ✓ Deployment validation passed (Kubespray)
# ✓ Service dependencies validated
# ✓ Required secrets validated
# 
# Configuration is valid ✓
```

### Step 5: Test in Non-Production

```bash
# Create a test cluster with v2 config
opencenter cluster init test-v2-migration \
  --schema-version 2.0 \
  --provider openstack \
  --region sjc3

# Compare with your migrated config
diff ~/.config/opencenter/clusters/myorg/.test-v2-migration-config.yaml \
     ~/.config/opencenter/clusters/myorg/.mycluster-config-v2.yaml
```

### Step 6: Deploy to Production

```bash
# Replace v1 config with v2 config
mv ~/.config/opencenter/clusters/myorg/.mycluster-config-v2.yaml \
   ~/.config/opencenter/clusters/myorg/.mycluster-config.yaml

# Validate one more time
opencenter cluster validate --config .mycluster-config.yaml

# Update cluster (if already deployed)
opencenter cluster update mycluster
```

### Step 7: Update IDE Configuration

After migrating to v2, update your IDE to use the v2 schema for validation and autocomplete:

**Visual Studio Code:**

Update `.vscode/settings.json`:
```json
{
  "yaml.schemas": {
    "./schema/cluster-v2.schema.json": [
      "**/clusters/**/*-v2.yaml",
      "**/v2/**/*.yaml"
    ]
  }
}
```

**Generate v2 Schema:**
```bash
opencenter cluster schema --version 2.0 --out schema/cluster-v2.schema.json
```

**Automatic Setup:**
```bash
opencenter config ide --schema-version 2.0
```

For complete IDE integration instructions, see [JSON Schema Documentation](../reference/json-schema.md).

## Migration Command Usage

### Basic Usage

```bash
opencenter cluster migrate-config --input <v1-config> --output <v2-config>
```

### Options

```bash
opencenter cluster migrate-config [flags]

Flags:
  -i, --input string       Path to v1 configuration file (required)
  -o, --output string      Path to write v2 configuration file (required)
      --dry-run            Show migration report without writing output
      --skip-validation    Skip validation of migrated configuration
      --apply-defaults     Apply provider-region defaults during migration (default: true)
      --format string      Output format: yaml, json (default: yaml)
  -h, --help               Help for migrate-config
```

### Examples

#### Dry Run (Preview Migration)

```bash
# Preview migration without writing output
opencenter cluster migrate-config \
  --input prod-cluster-v1.yaml \
  --output prod-cluster-v2.yaml \
  --dry-run
```

#### Skip Validation (Faster Migration)

```bash
# Migrate without validation (validate separately later)
opencenter cluster migrate-config \
  --input prod-cluster-v1.yaml \
  --output prod-cluster-v2.yaml \
  --skip-validation
```

#### JSON Output

```bash
# Generate v2 config in JSON format
opencenter cluster migrate-config \
  --input prod-cluster-v1.yaml \
  --output prod-cluster-v2.json \
  --format json
```

#### Batch Migration

```bash
# Migrate multiple clusters
for config in ~/.config/opencenter/clusters/myorg/.*.yaml; do
  basename=$(basename "$config" .yaml)
  opencenter cluster migrate-config \
    --input "$config" \
    --output "${config%.yaml}-v2.yaml"
done
```

## Post-Migration Validation

### Validation Checklist

After migration, verify the following:

- [ ] Schema version is `"2.0"`
- [ ] VRRP IP appears only in `infrastructure.networking.vrrp_ip`
- [ ] Storage configuration is under `infrastructure.storage`
- [ ] Worker pools are under `infrastructure.compute`
- [ ] Deployment config is at root level `deployment`
- [ ] SSH config is under `infrastructure.ssh`
- [ ] All provider-specific settings are under `infrastructure.cloud.<provider>`
- [ ] Kubernetes networking (pods/services CIDRs) remains in `cluster.kubernetes`
- [ ] All services are properly configured
- [ ] All required secrets are present

### Validation Commands

```bash
# Full validation
opencenter cluster validate --config mycluster-v2.yaml

# Schema validation only
opencenter cluster validate --config mycluster-v2.yaml --schema-only

# Provider validation only
opencenter cluster validate --config mycluster-v2.yaml --provider-only

# Check for deprecated fields
grep -r "opencenter.storage" mycluster-v2.yaml && echo "⚠ Deprecated field found"
```

### Common Validation Errors

#### E001: Missing Required Field

```
E001: infrastructure.networking.vrrp_ip: required field is missing
```

**Solution**: Ensure VRRP IP was migrated from v1 location.

#### E002: Invalid CIDR Notation

```
E002: infrastructure.networking.subnet_nodes: invalid CIDR notation "10.2.128.0/33"
```

**Solution**: Verify CIDR notation is valid (prefix must be 0-32 for IPv4).

#### E003: Reference Not Found

```
E003: services.calico.calico_kube_api_server: reference ${infrastructure.networking.vrrp_ip} not found
```

**Solution**: Ensure referenced field exists in v2 configuration.

#### E004: Provider Configuration Missing

```
E004: infrastructure.cloud.openstack.auth_url: required field for OpenStack provider
```

**Solution**: Verify all provider-specific fields were migrated.

## Common Migration Scenarios

### Scenario 1: OpenStack + Kubespray (Standard)

**v1 Configuration**: Standard OpenStack cluster with Kubespray deployment

**Migration Steps**:
1. Run migration command
2. Verify VRRP IP consolidated to `infrastructure.networking.vrrp_ip`
3. Verify storage moved to `infrastructure.storage`
4. Verify deployment moved to root-level `deployment`
5. Validate and deploy

**Example**:

```bash
opencenter cluster migrate-config \
  --input openstack-kubespray-v1.yaml \
  --output openstack-kubespray-v2.yaml

opencenter cluster validate --config openstack-kubespray-v2.yaml
```

### Scenario 2: OpenStack + Kamaji (Advanced)

**v1 Configuration**: Not supported (Kamaji is v2-only feature)

**Migration Steps**:
1. Migrate base configuration to v2
2. Manually add Kamaji deployment configuration
3. Configure worker pools with mixed OS support
4. Validate and deploy

**Example**:

```bash
# Migrate base config
opencenter cluster migrate-config \
  --input openstack-kubespray-v1.yaml \
  --output openstack-kamaji-v2.yaml

# Edit to add Kamaji configuration
vim openstack-kamaji-v2.yaml

# Add deployment.kamaji section (see examples/v2/openstack-kamaji-mixed-os.yaml)
```

### Scenario 3: Multiple Worker Pools

**v1 Configuration**: Cluster with additional worker pools

**Migration Steps**:
1. Run migration command
2. Verify worker pools moved from `cluster.kubernetes.additional_server_pools_worker` to `infrastructure.compute.additional_server_pools_worker`
3. Verify pool-specific storage configurations migrated
4. Validate and deploy

**Example**:

```bash
opencenter cluster migrate-config \
  --input multi-pool-v1.yaml \
  --output multi-pool-v2.yaml

# Verify worker pools location
grep -A 10 "additional_server_pools_worker" multi-pool-v2.yaml
```

### Scenario 4: Custom Service Configurations

**v1 Configuration**: Cluster with many custom service configurations

**Migration Steps**:
1. Run migration command
2. Verify all service configurations preserved
3. Check service dependencies are satisfied
4. Verify required secrets are configured
5. Validate and deploy

**Example**:

```bash
opencenter cluster migrate-config \
  --input custom-services-v1.yaml \
  --output custom-services-v2.yaml

# Validate service dependencies
opencenter cluster validate --config custom-services-v2.yaml
```

## Troubleshooting

### Issue: Migration Command Not Found

**Symptom**:
```
Error: unknown command "migrate-config" for "opencenter cluster"
```

**Solution**: Ensure you're using opencenter-cli version 2.0 or later:

```bash
opencenter version
# Should show: opencenter version 2.0.0 or higher
```

### Issue: VRRP IP Conflict

**Symptom**:
```
⚠ VRRP IP found in multiple locations
```

**Solution**: The migration tool automatically consolidates to `infrastructure.networking.vrrp_ip`. Verify the correct IP was selected:

```bash
grep -r "vrrp_ip" mycluster-v2.yaml
# Should show only one location: infrastructure.networking.vrrp_ip
```

### Issue: Missing Provider Defaults

**Symptom**:
```
E004: infrastructure.cloud.openstack.image_id: required field is missing
```

**Solution**: Ensure provider-region defaults are applied during migration:

```bash
opencenter cluster migrate-config \
  --input mycluster-v1.yaml \
  --output mycluster-v2.yaml \
  --apply-defaults  # Ensure this is enabled
```

### Issue: Validation Fails After Migration

**Symptom**:
```
E011: deployment.kamaji.cluster_api.providers.infrastructure: must match infrastructure.provider
```

**Solution**: Manually review and fix validation errors. The migration tool relocates fields but may not catch all logical inconsistencies:

```bash
# Review validation errors
opencenter cluster validate --config mycluster-v2.yaml

# Fix errors manually
vim mycluster-v2.yaml

# Re-validate
opencenter cluster validate --config mycluster-v2.yaml
```

### Issue: Deprecated Fields Warning

**Symptom**:
```
⚠ Deprecated field: opencenter.storage (use infrastructure.storage in v2)
```

**Solution**: This is informational. The migration tool should have moved these fields. Verify manually:

```bash
# Check for deprecated fields
grep -r "opencenter.storage" mycluster-v2.yaml

# Should return no results (all moved to infrastructure.storage)
```

### Issue: Reference Resolution Errors

**Symptom**:
```
E003: services.calico.calico_kube_api_server: reference ${cluster.networking.vrrp_ip} not found
```

**Solution**: Update references to use v2 field locations:

```yaml
# v1 reference (broken in v2)
calico_kube_api_server: "${cluster.networking.vrrp_ip}:6443"

# v2 reference (correct)
calico_kube_api_server: "${infrastructure.networking.vrrp_ip}:6443"
```

## Backward Compatibility

### v1 Support Timeline

- **Current - Q3 2026**: Full v1 support with no warnings
- **Q3 2026 - Q1 2027**: v1 supported with deprecation warnings
- **Q1 2027 - Q3 2027**: v1 supported with strong deprecation warnings
- **Q3 2027+**: v1 support removed

### Running v1 Configurations

During the coexistence period, v1 configurations continue to work:

```bash
# v1 config (no schema_version field)
opencenter cluster validate --config mycluster-v1.yaml

# Output:
# ⚠ WARNING: Using v1 schema (deprecated)
# ⚠ Please migrate to v2 using: opencenter cluster migrate-config
# ✓ Configuration is valid (v1 schema)
```

### Detecting Schema Version

The system detects schema version from the `schema_version` field:

```yaml
# v1 (no schema_version field or schema_version: "1.0")
opencenter:
  cluster:
    cluster_name: "mycluster"

# v2 (schema_version: "2.0" required)
schema_version: "2.0"
opencenter:
  cluster:
    cluster_name: "mycluster"
```

### Mixed Environment Support

You can run both v1 and v2 clusters simultaneously:

```bash
# List all clusters (shows schema version)
opencenter cluster list

# Output:
# NAME              ORG         ENV     REGION  SCHEMA  STATUS
# prod-cluster-v1   acme-corp   prod    sjc3    v1.0    active
# prod-cluster-v2   acme-corp   prod    sjc3    v2.0    active
# test-cluster-v2   acme-corp   test    dfw3    v2.0    active
```

### Gradual Migration Strategy

Recommended approach for large deployments:

1. **Week 1-2**: Migrate dev/test clusters to v2
2. **Week 3-4**: Validate v2 in staging environments
3. **Week 5-8**: Migrate production clusters during maintenance windows
4. **Week 9-12**: Complete migration of all clusters

```bash
# Migrate by environment
for env in dev test staging prod; do
  echo "Migrating $env clusters..."
  for config in ~/.config/opencenter/clusters/*/.${env}-*-config.yaml; do
    opencenter cluster migrate-config \
      --input "$config" \
      --output "${config%.yaml}-v2.yaml"
  done
done
```

## Additional Resources

- [v2 Configuration Reference](./v2-reference.md) - Complete v2 schema documentation
- [v2 Configuration Examples](./examples/v2/) - Example v2 configurations
- [JSON Schema Documentation](../reference/json-schema.md) - Schema documentation and IDE integration
- [IDE Integration Guide](../how-to/ide-integration.md) - Complete IDE setup instructions
- [Validation Error Codes](./validation-errors.md) - Complete error code reference

## Getting Help

If you encounter issues during migration:

1. **Check Migration Report**: Review warnings and applied defaults
2. **Validate Configuration**: Run `opencenter cluster validate` for detailed errors
3. **Review Examples**: Compare with example v2 configurations
4. **Community Support**: Ask questions in GitHub Discussions
5. **Enterprise Support**: Contact Rackspace support for assistance

---

**Last Updated**: January 2026  
**Schema Version**: v2.0  
**CLI Version**: 2.0.0+
