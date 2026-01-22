# cluster render

**doc_type:** reference

Render cluster templates into the GitOps directory structure with safety checks and backup functionality.

## Table of Contents

- [Synopsis](#synopsis)
- [Description](#description)
- [Modes](#modes)
- [Flags](#flags)
- [Arguments](#arguments)
- [Examples](#examples)
- [Rendering Process](#rendering-process)
- [Backup Behavior](#backup-behavior)
- [Output Structure](#output-structure)
- [Organization Support](#organization-support)
- [Use Cases](#use-cases)
- [Comparison with cluster setup](#comparison-with-cluster-setup)
- [Error Handling](#error-handling)
- [See Also](#see-also)

## Synopsis

```bash
opencenter cluster render [name] [service] [flags]
```

## Description

The `cluster render` command renders cluster templates into the GitOps directory structure with safety checks to prevent accidental overwrites. It creates timestamped backups before overwriting existing files and provides granular control over what gets rendered.

Unlike `cluster setup`, this command:
- Performs safety checks before rendering
- Creates timestamped backups before overwriting
- Does not perform Git operations
- Supports selective rendering (services or infrastructure)
- Ideal for iterative development and updates

## Modes

The command operates in different modes based on arguments and flags:

1. **Default Mode** (no flags): Checks if services already rendered
   - If rendered: Shows instructions for other modes
   - If not rendered: Performs initial render of services and infrastructure

2. **All Mode** (`--all`): Renders all services and infrastructure
   - Requires `--force` if already rendered
   - Creates backups of existing files

3. **Infrastructure Mode** (`--infra`): Renders infrastructure templates only
   - Always creates backups before overwriting
   - Does not require `--force`

4. **Service Mode** (`<service>`): Renders specific service
   - Validates service exists and is enabled
   - Requires `--force` if files already exist
   - Creates backup of service files

## Flags

- `--all` - Render all services and infrastructure
- `--force` - Force overwrite existing files (creates backups)
- `--infra` - Render infrastructure templates only

## Arguments

- `name` - Cluster name (optional if active cluster is set)
- `service` - Service name for selective rendering (optional)

## Examples

### Default Mode

```bash
# Check render status for active cluster
opencenter cluster render

# Check render status for specific cluster
opencenter cluster render my-cluster
```

**Output when already rendered:**
```
Services have already been rendered for cluster 'my-cluster'.

To render all services (with backups), use:
  opencenter cluster render my-cluster --all --force

To render a specific service, use:
  opencenter cluster render my-cluster <service-name> --force

To render infrastructure only, use:
  opencenter cluster render my-cluster --infra
```

### Render All Services and Infrastructure

```bash
# Initial render (no backups needed)
opencenter cluster render my-cluster --all

# Re-render with backups
opencenter cluster render my-cluster --all --force
```

### Render Specific Service

```bash
# Render tempo service
opencenter cluster render my-cluster tempo --force

# Render loki service
opencenter cluster render my-cluster loki --force
```

### Render Infrastructure Only

```bash
# Render infrastructure templates
opencenter cluster render my-cluster --infra
```

## Rendering Process

### All Services and Infrastructure Mode

1. **Backup Existing Files** (if `--force` specified)
   - Creates timestamped backups of application files
   - Creates timestamped backups of infrastructure files

2. **Copy Base GitOps Structure**
   - Copies base templates from embedded resources
   - Creates directory structure
   - Installs base manifests

3. **Render Cluster-Specific Applications**
   - Renders application overlays for the cluster
   - Applies cluster-specific customizations
   - Generates Kustomization files

4. **Render Infrastructure Templates**
   - Renders infrastructure cluster manifests
   - Generates provider-specific configurations
   - Creates namespace and RBAC resources

5. **Provision OpenTofu**
   - Renders `main.tf` with cluster configuration
   - Generates `provider.tf` for cloud provider
   - Creates Terraform variable files

### Service-Specific Mode

1. **Validate Service**
   - Checks service exists in configuration
   - Verifies service is enabled

2. **Backup Existing Files** (if `--force` and files exist)
   - Creates timestamped backups of service directory

3. **Render Service Templates**
   - Renders service manifests
   - Generates service-specific configurations
   - Updates Kustomization files

### Infrastructure-Only Mode

1. **Backup Existing Files**
   - Creates timestamped backups of infrastructure directory

2. **Render Infrastructure Templates**
   - Renders infrastructure cluster manifests
   - Generates provider-specific configurations

3. **Provision OpenTofu**
   - Renders Terraform configuration files

## Backup Behavior

When `--force` is used or infrastructure is rendered, the command creates timestamped backups:

**Backup Format:**
```
<original-file>.bak-YYYYMMDD-HHMMSS
```

**Example:**
```
kustomization.yaml.bak-20260122-143052
main.tf.bak-20260122-143052
```

**Backup Locations:**
- Application files: `applications/overlays/<cluster>/`
- Infrastructure files: `infrastructure/clusters/<cluster>/`
- Service files: `applications/overlays/<cluster>/services/<service>/`

## Output Structure

Templates are rendered to:
```
<git_dir>/
├── infrastructure/
│   └── clusters/<cluster>/
│       ├── main.tf                    # Terraform main configuration
│       ├── provider.tf                # Provider configuration
│       ├── variables.tf               # Variable definitions
│       ├── kustomization.yaml         # Kustomization manifest
│       └── ...                        # Other infrastructure files
└── applications/
    └── overlays/<cluster>/
        ├── kustomization.yaml         # Application kustomization
        └── ...                        # Application manifests
```

## Organization Support

The command handles organization-based directory structures automatically:

```
<git_dir>/<organization>/
├── infrastructure/
│   └── clusters/<cluster>/
└── applications/
    └── overlays/<cluster>/
```

## Use Cases

### Initial Rendering

Render services and infrastructure for the first time:
```bash
opencenter cluster render my-cluster --all
```

### Iterative Development

Test configuration changes without full setup:
```bash
# Edit configuration
opencenter cluster edit my-cluster

# Re-render with backups
opencenter cluster render my-cluster --all --force

# Review changes
git diff
```

### Service Updates

Update a specific service configuration:
```bash
# Edit service configuration
opencenter cluster edit my-cluster

# Render only the updated service
opencenter cluster render my-cluster tempo --force

# Review service changes
ls -la ~/gitops/applications/overlays/my-cluster/services/tempo/
```

### Infrastructure Updates

Update infrastructure without touching services:
```bash
# Edit infrastructure configuration
opencenter cluster edit my-cluster

# Render only infrastructure
opencenter cluster render my-cluster --infra

# Review infrastructure changes
git diff infrastructure/clusters/my-cluster/
```

### Template Debugging

Debug template rendering issues:
```bash
# Render templates
opencenter cluster render my-cluster --all --force

# Check for errors in rendered files
cat ~/gitops/infrastructure/clusters/my-cluster/main.tf
```

### Configuration Testing

Test different configurations quickly:
```bash
# Update configuration
opencenter cluster update my-cluster --opencenter.cluster.kubernetes.version=1.31.4

# Re-render infrastructure
opencenter cluster render my-cluster --infra

# Verify changes
grep "kubernetes_version" ~/gitops/infrastructure/clusters/my-cluster/main.tf
```

### CI/CD Integration

Generate templates in CI/CD pipelines:
```bash
#!/bin/bash
set -e

# Render all templates
opencenter cluster render my-cluster --all --force

# Validate rendered Terraform
cd ~/gitops/infrastructure/clusters/my-cluster
terraform validate

# Validate Kubernetes manifests
kubectl apply --dry-run=client -k ~/gitops/applications/overlays/my-cluster/
```

## Comparison with cluster setup

| Feature | render | setup |
|---------|--------|-------|
| Renders templates | ✓ | ✓ |
| Git initialization | ✗ | ✓ |
| Safety checks | ✓ | ✓ |
| Creates backups | ✓ | ✗ |
| Selective rendering | ✓ | ✗ |
| Service-specific | ✓ | ✗ |
| Infrastructure-only | ✓ | ✗ |
| Use case | Updates/iteration | Initial setup |

## Error Handling

**Cluster not found:**
```
Error: failed to load cluster configuration: cluster "my-cluster" not found
```

**Service not found:**
```
Error: service 'invalid-service' not found in cluster configuration
```

**Service disabled:**
```
Error: service 'tempo' is disabled in cluster configuration
```

**Already rendered (without --force):**
```
Error: services already rendered for cluster 'my-cluster', use --force to overwrite (creates backups)
```

**Service exists (without --force):**
```
Error: service 'tempo' is enabled but files already exist, use --force to overwrite (creates backup)
```

**GitOps directory not configured:**
```
Error: git_dir is not configured. Run 'opencenter cluster setup' first or set git_dir in the configuration
```

**Template rendering failed:**
```
Error: failed to render cluster templates: template execution error
```

**Backup creation failed:**
```
Error: failed to create backups: permission denied
```

## See Also

- [cluster setup](setup.md) - Setup GitOps directory structure with Git initialization
- [cluster validate](../cli-commands.md#cluster-validate) - Validate cluster configuration
- [cluster bootstrap](bootstrap.md) - Bootstrap cluster infrastructure
