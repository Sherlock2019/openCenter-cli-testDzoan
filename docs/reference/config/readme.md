# Config Commands Reference

Configuration management commands for openCenter CLI.

## Available Commands

### [features](features.md)
Display feature flag status and manage gradual migration to refactored systems.

```bash
openCenter config features [--output table|json|env]
```

**Use Cases:**
- Check which features are currently enabled
- Generate environment variable export statements
- Automate feature flag configuration in scripts
- Debug feature flag evaluation

### ide
Generate IDE configuration files for enhanced development experience.

```bash
openCenter config ide [--vscode] [--jetbrains] [--all]
```

**Use Cases:**
- Set up YAML schema validation in VS Code
- Configure JetBrains IDEs for cluster configuration editing
- Enable autocomplete and validation in editors

## Feature Flags

openCenter uses feature flags to enable new refactored systems while maintaining backward compatibility. The `config features` command helps manage these flags.

### Available Feature Flags

- `OPENCENTER_USE_NEW_TEMPLATE_ENGINE` - Enhanced template engine with caching
- `OPENCENTER_USE_PIPELINE_GENERATOR` - Pipeline-based GitOps generation
- `OPENCENTER_USE_NEW_CONFIG_BUILDER` - Type-safe configuration builder
- `OPENCENTER_USE_SERVICE_REGISTRY` - Plugin-based service registry
- `OPENCENTER_ENABLE_ALL_NEW_FEATURES` - Enable all new features at once
- `OPENCENTER_FEATURE_FLAG_DEBUG` - Enable debug logging for feature flags

### Quick Start

```bash
# Check current feature status
openCenter config features

# Enable all new features
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true

# Enable specific feature
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true

# Verify settings
openCenter config features
```

## See Also

- [CLI Commands Reference](../cli-commands.md)
- [Configuration Reference](../configuration.md)
- [Migration Guide](../../migration/configuration-system-refactor.md)
