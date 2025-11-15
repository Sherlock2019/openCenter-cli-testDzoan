# `openCenter cluster setup` - Setup GitOps Directory

## Synopsis
```bash
openCenter cluster setup [name] [OPTIONS]
```

## Description

Setup the GitOps directory by copying or rendering templates and initializing a git repository. This command creates the complete GitOps repository structure with cluster-specific configurations, SOPS encryption setup, and infrastructure templates.

The setup process creates an organization-based GitOps repository structure that can host multiple clusters within the same organization.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster (format: `cluster` or `organization/cluster`). If not provided, uses the currently active cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `--render`
- **Description**: Render templates (process `.tmpl` files) rather than copy them as-is
- **Type**: Boolean
- **Default**: `false`

### `--force`
- **Description**: Overwrite existing files and reinitialize git repository
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Setup GitOps for active cluster
```bash
openCenter cluster setup
```

### Setup specific cluster
```bash
openCenter cluster setup my-cluster
```

### Setup with template rendering
```bash
openCenter cluster setup my-cluster --render
```

### Force reinitialize
```bash
openCenter cluster setup my-cluster --force
```

### Setup cluster in organization
```bash
openCenter cluster setup production/prod-cluster
```

## Output

```
Created GitOps repo
Setup complete.
```

## GitOps Directory Structure

The command creates the following structure:

```
<organization>/gitops/
├── .git/                                    # Git repository
├── .gitignore                               # Git ignore rules
├── .opencenter                              # Marker file with cluster list
├── .sops.yaml                               # SOPS configuration
├── applications/
│   ├── base/                                # Base application manifests
│   └── overlays/
│       └── <cluster>/                       # Cluster-specific overlays
│           ├── kustomization.yaml
│           └── ...
└── infrastructure/
    ├── base/                                # Base infrastructure manifests
    └── clusters/
        └── <cluster>/                       # Cluster-specific infrastructure
            ├── kustomization.yaml
            ├── Makefile
            ├── provider.tf
            └── ...
```

## Setup Process

The setup command performs the following steps:

1. **Create Organization Structure** - Creates organization-based directory hierarchy
2. **Copy Base Templates** - Copies GitOps base structure from embedded templates
3. **Render Cluster Templates** - Renders cluster-specific application and infrastructure templates
4. **Provision OpenTofu** - Generates provider.tf and main.tf for infrastructure
5. **Setup SOPS** - Configures SOPS encryption for cluster secrets
6. **Initialize Git** - Initializes git repository with main branch
7. **Create Marker** - Creates .opencenter marker file with cluster information
8. **Commit Changes** - Commits initial setup to git repository

## Template Rendering

When `--render` flag is used:

- All `.tmpl` files are processed with cluster configuration
- Template variables are replaced with actual values
- Conditional sections are evaluated
- Output files have `.tmpl` extension removed

Without `--render`:
- Templates are copied as-is
- Manual rendering required later

## SOPS Configuration

The command creates organization-wide SOPS configuration:

```yaml
# SOPS configuration for organization
# Each cluster's key encrypts only its specific directories
creation_rules:
  - path_regex: (applications/overlays/my-cluster/.*|infrastructure/clusters/my-cluster/.*)\.ya?ml$
    age: >-
      age1xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

## Notes

- The command is idempotent - running it multiple times is safe
- Use `--force` to reinitialize an existing GitOps repository
- The GitOps directory is shared across clusters in the same organization
- Each cluster has isolated application overlays and infrastructure directories
- SOPS keys are cluster-specific within the organization
- Git repository is initialized with `main` branch
- The `.opencenter` marker tracks which clusters use this GitOps repository
- Template rendering can be done separately with `cluster render` command
- The command validates configuration before setup

## See Also

- `openCenter cluster render` - Render templates without git initialization
- `openCenter cluster bootstrap` - Bootstrap cluster after setup
- `openCenter cluster validate` - Validate configuration before setup
