# `openCenter cluster render` - Render Templates

## Synopsis
```bash
openCenter cluster render [name]
```

## Description

Render templates into the GitOps directory without initializing git. This command processes all `.tmpl` files in the GitOps structure, replacing template variables with actual configuration values, but does not perform any git operations.

This is useful for inspecting template output, testing configurations, or when git initialization is handled separately.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster (format: `cluster` or `organization/cluster`). If not provided, uses the currently active cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Render templates for active cluster
```bash
openCenter cluster render
```

### Render templates for specific cluster
```bash
openCenter cluster render my-cluster
```

### Render for organization cluster
```bash
openCenter cluster render production/prod-cluster
```

### Inspect rendered output
```bash
openCenter cluster render my-cluster
cat ~/config/openCenter/clusters/myorg/gitops/infrastructure/clusters/my-cluster/kustomization.yaml
```

### Using Feature Flags

Feature flags enable new refactored systems for improved performance and functionality:

```bash
# Use new template engine with caching and better error messages
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster

# Compare legacy vs new template engine output
openCenter cluster render my-cluster > /tmp/legacy-output.txt
OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true openCenter cluster render my-cluster > /tmp/new-output.txt
diff /tmp/legacy-output.txt /tmp/new-output.txt

# Enable all new features
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
openCenter cluster render my-cluster

# Enable debug logging to see feature flag evaluation
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter cluster render my-cluster

# Check which features are currently enabled
openCenter config features
```

**Feature Flag Benefits:**
- **New Template Engine**: 
  - Improved performance through template caching
  - Better error messages with line numbers and context
  - Template validation before rendering
  - Reduced memory usage for large configurations

See `openCenter config features --help` for more information about feature flags.

## Output

```
Render complete.
```

## Rendering Process

The command performs the following operations:

1. **Copy Base Templates** - Copies GitOps base structure with rendering enabled
2. **Render Cluster Apps** - Processes application overlay templates
3. **Render Infrastructure** - Processes infrastructure cluster templates
4. **Provision OpenTofu** - Generates provider.tf for infrastructure

## Template Processing

### Template Variables

Templates have access to the full cluster configuration:

```yaml
# Example template: kustomization.yaml.tmpl
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: {{ .ClusterName }}
  namespace: {{ .Namespace }}
resources:
  - namespace.yaml
  {{- if .EnableMonitoring }}
  - monitoring.yaml
  {{- end }}
```

### Template Functions

Templates can use Sprig template functions:

- String manipulation: `upper`, `lower`, `trim`, `replace`
- Lists: `join`, `split`, `append`
- Conditionals: `if`, `else`, `with`
- Loops: `range`
- And many more from [Sprig](http://masterminds.github.io/sprig/)

## Rendered Files

The command renders the following template types:

### Application Overlays
```
applications/overlays/<cluster>/
├── kustomization.yaml
├── namespace.yaml
├── flux-system.yaml
└── ...
```

### Infrastructure Cluster
```
infrastructure/clusters/<cluster>/
├── kustomization.yaml
├── Makefile
├── provider.tf
├── main.tf
└── ...
```

## Differences from Setup

| Feature | render | setup |
|---------|--------|-------|
| Template rendering | ✓ | ✓ |
| Git initialization | ✗ | ✓ |
| SOPS configuration | ✗ | ✓ |
| Git commit | ✗ | ✓ |
| .opencenter marker | ✗ | ✓ |

## Use Cases

### Testing Configuration Changes
```bash
# Make configuration changes
openCenter cluster update my-cluster --opencenter.meta.env=staging

# Render to see the effect
openCenter cluster render my-cluster

# Inspect rendered files
diff -r gitops/infrastructure/clusters/my-cluster/ backup/
```

### Debugging Templates
```bash
# Render templates
openCenter cluster render my-cluster

# Check for rendering errors
grep -r "{{" gitops/infrastructure/clusters/my-cluster/
```

### Dry Run Before Setup
```bash
# Render first to verify output
openCenter cluster render my-cluster

# Review rendered files
ls -la gitops/infrastructure/clusters/my-cluster/

# Then setup with git
openCenter cluster setup my-cluster
```

## Notes

- The command does not initialize or modify git repository
- Rendered files are written to the GitOps directory
- Existing files are overwritten
- Template errors are reported during rendering
- The command validates configuration before rendering
- Use `cluster setup` for complete GitOps initialization
- Rendering is idempotent - can be run multiple times
- Template variables come from cluster configuration
- Sprig template functions are available in all templates

## Troubleshooting

### Template rendering errors
**Error**: `failed to render cluster apps templates: template: parse error`

**Solution**: Check template syntax in the affected file:
```bash
# Find template files
find gitops/ -name "*.tmpl"

# Validate YAML syntax
yamllint gitops/infrastructure/clusters/my-cluster/
```

### Missing template variables
**Error**: `template: undefined variable ".SomeField"`

**Solution**: Ensure the configuration field exists:
```bash
openCenter cluster info my-cluster --json | jq '.metadata'
```

### OpenTofu provisioning failed
**Error**: `failed to provision opentofu`

**Solution**: Check provider configuration:
```bash
openCenter cluster validate my-cluster
```

## See Also

- `openCenter cluster setup` - Setup GitOps with git initialization
- `openCenter cluster validate` - Validate configuration before rendering
- `openCenter cluster update` - Update configuration values
