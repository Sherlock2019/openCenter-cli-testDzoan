# `openCenter config features` - Display Feature Flag Status

## Synopsis
```bash
openCenter config features [OPTIONS]
```

## Description

Display the current status of feature flags that control system behavior. Feature flags allow gradual migration from legacy systems to new implementations with the ability to rollback if issues are discovered.

This command shows which features are currently enabled and provides guidance on how to use them. It supports multiple output formats for different use cases: human-readable tables, JSON for automation, and shell export statements for easy configuration.

## Feature Flags Overview

openCenter uses feature flags to enable new refactored systems while maintaining backward compatibility. This allows users to:

- **Gradually adopt new features**: Enable features one at a time to validate behavior
- **Rollback if needed**: Disable features if issues are discovered
- **Test in isolation**: Enable features in development/staging before production
- **Maintain stability**: Keep legacy systems as default until new systems are proven

## Available Feature Flags

### Individual Feature Flags

#### `OPENCENTER_USE_NEW_TEMPLATE_ENGINE`
- **Description**: Enhanced template engine with caching and better error messages
- **Benefits**: 
  - Improved template rendering performance through caching
  - Better error messages with line numbers and context
  - Support for template validation before rendering
- **Default**: `false` (uses legacy template engine)

#### `OPENCENTER_USE_PIPELINE_GENERATOR`
- **Description**: Pipeline-based GitOps generation with rollback and progress reporting
- **Benefits**:
  - Staged generation with automatic rollback on failure
  - Progress reporting for long-running operations
  - Dry-run mode for previewing changes
  - Atomic operations prevent partial writes
- **Default**: `false` (uses legacy GitOps generation)

#### `OPENCENTER_USE_NEW_CONFIG_BUILDER`
- **Description**: Type-safe fluent configuration builder
- **Benefits**:
  - Compile-time type safety for configuration paths
  - Fluent API for readable configuration construction
  - Better validation error aggregation and reporting
- **Default**: `false` (uses legacy configuration building)

#### `OPENCENTER_USE_SERVICE_REGISTRY`
- **Description**: Plugin-based service registry with dependency resolution
- **Benefits**:
  - Dynamic service loading from manifests
  - Automatic dependency resolution
  - Circular dependency detection
  - Service lifecycle management
- **Default**: `false` (uses legacy service management)

### Global Feature Flags

#### `OPENCENTER_ENABLE_ALL_NEW_FEATURES`
- **Description**: Enable all new features at once
- **Benefits**: Convenient way to enable all refactored systems
- **Default**: `false`
- **Note**: When enabled, overrides individual feature flags

#### `OPENCENTER_FEATURE_FLAG_DEBUG`
- **Description**: Enable debug logging for feature flag evaluation
- **Benefits**: 
  - See which features are enabled/disabled
  - Understand feature flag evaluation logic
  - Troubleshoot feature flag issues
- **Default**: `false`

## Options

### `--output, -o <format>`
- **Description**: Output format for feature flag status
- **Type**: String
- **Default**: `table`
- **Valid Values**: `table`, `json`, `env`
- **Example**: `--output json`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Output Formats

### Table Format (Default)

Human-readable table showing feature status:

```
FEATURE              STATUS     ENVIRONMENT VARIABLE                      DESCRIPTION
-------              ------     --------------------                      -----------
Template Engine      disabled   OPENCENTER_USE_NEW_TEMPLATE_ENGINE        Enhanced template engine with caching
Pipeline Generator   disabled   OPENCENTER_USE_PIPELINE_GENERATOR         Pipeline-based GitOps generation
Config Builder       disabled   OPENCENTER_USE_NEW_CONFIG_BUILDER         Type-safe configuration builder
Service Registry     disabled   OPENCENTER_USE_SERVICE_REGISTRY           Plugin-based service registry

All New Features     disabled   OPENCENTER_ENABLE_ALL_NEW_FEATURES        Enable all new features at once
Debug Logging        disabled   OPENCENTER_FEATURE_FLAG_DEBUG             Feature flag debug logging
```

### JSON Format

Structured JSON output for automation and scripting:

```json
{
  "features": {
    "new_template_engine": {
      "enabled": false,
      "env_var": "OPENCENTER_USE_NEW_TEMPLATE_ENGINE",
      "description": "Enhanced template engine with caching and better error messages"
    },
    "pipeline_generator": {
      "enabled": false,
      "env_var": "OPENCENTER_USE_PIPELINE_GENERATOR",
      "description": "Pipeline-based GitOps generation with rollback and progress reporting"
    },
    "new_config_builder": {
      "enabled": false,
      "env_var": "OPENCENTER_USE_NEW_CONFIG_BUILDER",
      "description": "Type-safe fluent configuration builder"
    },
    "service_registry": {
      "enabled": false,
      "env_var": "OPENCENTER_USE_SERVICE_REGISTRY",
      "description": "Plugin-based service registry with dependency resolution"
    }
  },
  "global": {
    "all_new_features": {
      "enabled": false,
      "env_var": "OPENCENTER_ENABLE_ALL_NEW_FEATURES",
      "description": "Enable all new features at once"
    },
    "debug_enabled": {
      "enabled": false,
      "env_var": "OPENCENTER_FEATURE_FLAG_DEBUG",
      "description": "Enable debug logging for feature flag evaluation"
    }
  }
}
```

### Environment Format

Shell export statements for easy configuration:

```bash
# Feature Flag Environment Variables
# Copy and paste these commands to enable/disable features

# Enhanced template engine with caching
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false

# Pipeline-based GitOps generation
export OPENCENTER_USE_PIPELINE_GENERATOR=false

# Type-safe configuration builder
export OPENCENTER_USE_NEW_CONFIG_BUILDER=false

# Plugin-based service registry
export OPENCENTER_USE_SERVICE_REGISTRY=false


# Global flags

# Enable all new features at once
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=false

# Feature flag debug logging
export OPENCENTER_FEATURE_FLAG_DEBUG=false
```

## Examples

### Display feature status (default table format)
```bash
openCenter config features
```

### Display feature status as JSON
```bash
openCenter config features --output json
```

### Display feature status as environment variables
```bash
openCenter config features --output env
```

### Save environment variables to file
```bash
openCenter config features --output env > feature-flags.sh
```

### Enable all features using environment format
```bash
# Generate export statements
openCenter config features --output env > /tmp/features.sh

# Edit the file to set desired features to "true"
vim /tmp/features.sh

# Source the file to apply settings
source /tmp/features.sh

# Verify settings
openCenter config features
```

### Check feature status in scripts
```bash
# Get JSON output and parse with jq
openCenter config features --output json | jq '.features.new_template_engine.enabled'

# Use in conditional logic
if openCenter config features --output json | jq -e '.features.pipeline_generator.enabled' > /dev/null; then
  echo "Pipeline generator is enabled"
fi
```

## Setting Feature Flags

Feature flags are controlled through environment variables. Set them before running openCenter commands:

### Enable a Single Feature

```bash
# Enable new template engine
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster setup my-cluster

# Enable pipeline generator
export OPENCENTER_USE_PIPELINE_GENERATOR=true
openCenter cluster setup my-cluster
```

### Enable All Features

```bash
# Enable all new features at once
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
openCenter cluster setup my-cluster
```

### Enable Features for a Single Command

```bash
# Use new template engine for one command
OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true openCenter cluster setup my-cluster

# Enable all features for one command
OPENCENTER_ENABLE_ALL_NEW_FEATURES=true openCenter cluster setup my-cluster
```

### Persist Feature Flags

Add to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

```bash
# Enable new template engine permanently
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true

# Or enable all features
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
```

### Enable Debug Logging

```bash
# See feature flag evaluation
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter config features
```

## Valid Values

Feature flags accept the following values (case-insensitive):

**Enabled**: `true`, `1`, `yes`, `on`
**Disabled**: `false`, `0`, `no`, `off`, or unset

Examples:
```bash
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=1
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=yes
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=on
```

## Migration Strategy

### Recommended Approach

1. **Start with Debug Logging**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   openCenter config features
   ```

2. **Enable Features Individually**
   ```bash
   # Test template engine first
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   openCenter cluster setup test-cluster
   
   # If successful, enable pipeline generator
   export OPENCENTER_USE_PIPELINE_GENERATOR=true
   openCenter cluster setup test-cluster
   ```

3. **Validate in Non-Production**
   - Test in development/staging environments first
   - Verify generated GitOps repositories are correct
   - Compare output with legacy system if needed

4. **Enable in Production**
   ```bash
   # After validation, enable all features
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   ```

5. **Monitor and Rollback if Needed**
   ```bash
   # If issues occur, disable features
   unset OPENCENTER_ENABLE_ALL_NEW_FEATURES
   # Or disable specific features
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
   ```

### Testing Feature Flags

```bash
# Test with dry-run mode
export OPENCENTER_USE_PIPELINE_GENERATOR=true
openCenter --dry-run cluster setup my-cluster

# Compare outputs
openCenter cluster render my-cluster > /tmp/legacy-output.txt
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster > /tmp/new-output.txt
diff /tmp/legacy-output.txt /tmp/new-output.txt
```

## Troubleshooting

### Feature flag not taking effect

**Problem**: Feature flag is set but not being used

**Solution**: 
1. Verify the environment variable is set:
   ```bash
   echo $OPENCENTER_USE_NEW_TEMPLATE_ENGINE
   ```

2. Check feature status:
   ```bash
   openCenter config features
   ```

3. Enable debug logging:
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   openCenter config features
   ```

### Unexpected behavior with new features

**Problem**: New features cause errors or unexpected output

**Solution**:
1. Disable the problematic feature:
   ```bash
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
   ```

2. Report the issue with debug logs:
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   openCenter cluster setup my-cluster 2>&1 | tee debug.log
   ```

3. Rollback to legacy system:
   ```bash
   unset OPENCENTER_ENABLE_ALL_NEW_FEATURES
   unset OPENCENTER_USE_NEW_TEMPLATE_ENGINE
   unset OPENCENTER_USE_PIPELINE_GENERATOR
   unset OPENCENTER_USE_NEW_CONFIG_BUILDER
   unset OPENCENTER_USE_SERVICE_REGISTRY
   ```

### Global flag overrides individual flags

**Problem**: Individual feature flags are ignored when `OPENCENTER_ENABLE_ALL_NEW_FEATURES` is set

**Solution**: This is expected behavior. The global flag enables all features. To use individual flags, unset the global flag:
```bash
unset OPENCENTER_ENABLE_ALL_NEW_FEATURES
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
```

## Notes

- Feature flags are evaluated at runtime, no restart required
- Global flag (`OPENCENTER_ENABLE_ALL_NEW_FEATURES`) overrides individual flags
- Debug logging shows feature flag evaluation in real-time
- Feature flags are designed for gradual migration, not permanent configuration
- Legacy systems remain the default for backward compatibility
- See [Migration Guide](../../migration/configuration-system-refactor.md) for detailed migration instructions

## See Also

- [Configuration System Refactor Migration Guide](../../migration/configuration-system-refactor.md)
- [Feature Flag Removal Timeline](../../migration/feature-flag-removal-timeline.md)
- [Architecture Documentation](../../architecture.md)
- `openCenter cluster setup` - Setup GitOps repository (supports feature flags)
- `openCenter cluster render` - Render templates (supports feature flags)
