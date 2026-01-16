# Configuration System Refactor Migration Guide

## Overview

This guide helps you migrate from the legacy openCenter configuration system to the refactored architecture. The refactor introduces modular, extensible components with feature flags that enable gradual adoption and safe rollback if issues arise.

**Key Benefits of the Refactored System:**
- **Modular Architecture**: Clean separation of concerns with well-defined interfaces
- **Better Error Messages**: Detailed validation errors with actionable suggestions
- **Improved Performance**: Template caching and optimized processing
- **Type Safety**: Compile-time validation for configuration building
- **Plugin System**: Extensible service architecture
- **Rollback Capability**: Feature flags allow instant rollback to legacy systems

## Feature Flags

The refactored system uses feature flags to control which implementation is active. This allows you to test new features in production while maintaining the ability to quickly disable them if issues occur.

### Available Feature Flags

| Feature Flag | Controls | Default | Description |
|-------------|----------|---------|-------------|
| `OPENCENTER_USE_NEW_TEMPLATE_ENGINE` | Template rendering | `false` | Enhanced template engine with caching and better error messages |
| `OPENCENTER_USE_PIPELINE_GENERATOR` | GitOps generation | `false` | Pipeline-based generation with rollback and progress reporting |
| `OPENCENTER_USE_NEW_CONFIG_BUILDER` | Configuration building | `false` | Type-safe fluent builder with compile-time validation |
| `OPENCENTER_USE_SERVICE_REGISTRY` | Service management | `false` | Plugin-based service registry with dependency resolution |
| `OPENCENTER_ENABLE_ALL_NEW_FEATURES` | All systems | `false` | Enable all new features at once (individual flags override) |
| `OPENCENTER_FEATURE_FLAG_DEBUG` | Debug logging | `false` | Print feature flag evaluation to stderr |

### Setting Feature Flags

Feature flags are controlled via environment variables. Valid values for enabling a flag:
- `true`, `1`, `yes`, `on` (case-insensitive)

Any other value or unset means the flag is disabled.

**Examples:**

```bash
# Enable a single feature
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render

# Enable all new features at once
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
openCenter cluster init my-cluster

# Enable with debug logging to see which systems are active
export OPENCENTER_FEATURE_FLAG_DEBUG=true
export OPENCENTER_USE_PIPELINE_GENERATOR=true
openCenter cluster render

# Disable a specific feature when all are enabled
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false  # Override: use legacy template engine
openCenter cluster render
```

### Checking Feature Flag Status

Use the `opencenter config features` command to see which features are currently enabled:

```bash
# Display as table (default)
openCenter config features

# Display as JSON
openCenter config features --output json

# Display as environment variable exports
openCenter config features --output env
```

## Migration Strategy

We recommend a **gradual migration** approach to minimize risk:

### Phase 1: Testing (Recommended First Step)

Start by enabling features in a non-production environment:

1. **Enable debug logging** to see which systems are active:
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

2. **Enable one feature at a time** and test thoroughly:
   ```bash
   # Test new template engine
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   openCenter cluster render
   
   # Verify output is identical to legacy system
   # Run your test suite
   ```

3. **Gradually enable more features** as confidence grows:
   ```bash
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   export OPENCENTER_USE_PIPELINE_GENERATOR=true
   openCenter cluster init test-cluster
   ```

4. **Enable all features** once individual features are validated:
   ```bash
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   openCenter cluster init production-cluster
   ```

### Phase 2: Production Rollout

Once testing is complete, enable features in production:

1. **Start with low-risk operations** (read-only commands):
   ```bash
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   openCenter cluster validate
   openCenter cluster render --dry-run
   ```

2. **Monitor closely** for any issues:
   - Check logs for errors or warnings
   - Compare output with legacy system
   - Monitor performance metrics

3. **Gradually expand** to more critical operations:
   ```bash
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   openCenter cluster init new-cluster
   ```

4. **Keep feature flags enabled** for several weeks to ensure stability

### Phase 3: Permanent Migration

After successful validation (typically 4-8 weeks):

1. **Make feature flags permanent** in your environment:
   ```bash
   # Add to ~/.bashrc, ~/.zshrc, or CI/CD configuration
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   ```

2. **Update documentation** to reflect new systems as standard

3. **Plan for feature flag removal** (see timeline below)

## Feature-Specific Migration Guides

### 1. Template Engine Migration

**What Changed:**
- New template engine with caching for improved performance
- Better error messages with line numbers and context
- Enhanced validation before rendering
- Support for custom template functions

**Migration Steps:**

1. **Enable the feature flag:**
   ```bash
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   ```

2. **Test template rendering:**
   ```bash
   openCenter cluster render
   ```

3. **Verify output is identical:**
   ```bash
   # With legacy system
   openCenter cluster render > legacy-output.yaml
   
   # With new system
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   openCenter cluster render > new-output.yaml
   
   # Compare
   diff legacy-output.yaml new-output.yaml
   ```

4. **Check for improved error messages:**
   - Intentionally create a template error
   - Verify the new error message is more helpful

**Rollback:**
```bash
unset OPENCENTER_USE_NEW_TEMPLATE_ENGINE
# or
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
```

**Known Issues:**
- None currently reported

### 2. Pipeline Generator Migration

**What Changed:**
- Pipeline-based GitOps generation with discrete stages
- Automatic rollback on failure
- Progress reporting for long operations
- Dry-run mode for previewing changes

**Migration Steps:**

1. **Enable the feature flag:**
   ```bash
   export OPENCENTER_USE_PIPELINE_GENERATOR=true
   ```

2. **Test with dry-run first:**
   ```bash
   openCenter cluster render --dry-run
   ```

3. **Generate a test repository:**
   ```bash
   openCenter cluster init test-cluster
   ```

4. **Verify repository structure:**
   ```bash
   # Check that all expected files are present
   # Compare with legacy-generated repository
   ```

5. **Test rollback capability:**
   - Intentionally cause a generation failure
   - Verify automatic rollback occurs

**Rollback:**
```bash
unset OPENCENTER_USE_PIPELINE_GENERATOR
# or
export OPENCENTER_USE_PIPELINE_GENERATOR=false
```

**Known Issues:**
- None currently reported

### 3. Configuration Builder Migration

**What Changed:**
- Type-safe fluent API for building configurations
- Compile-time validation of configuration paths
- Better error aggregation and reporting
- Support for conditional configuration

**Migration Steps:**

1. **Enable the feature flag:**
   ```bash
   export OPENCENTER_USE_NEW_CONFIG_BUILDER=true
   ```

2. **Test configuration building:**
   ```bash
   openCenter cluster init test-cluster
   ```

3. **Verify configuration validation:**
   ```bash
   # Test with invalid configuration
   openCenter cluster validate invalid-config.yaml
   # Verify error messages are clear and helpful
   ```

4. **Check error aggregation:**
   - Create a configuration with multiple errors
   - Verify all errors are reported together

**Rollback:**
```bash
unset OPENCENTER_USE_NEW_CONFIG_BUILDER
# or
export OPENCENTER_USE_NEW_CONFIG_BUILDER=false
```

**Known Issues:**
- None currently reported

### 4. Service Registry Migration

**What Changed:**
- Plugin-based service architecture
- Automatic dependency resolution
- Circular dependency detection
- Dynamic service loading

**Migration Steps:**

1. **Enable the feature flag:**
   ```bash
   export OPENCENTER_USE_SERVICE_REGISTRY=true
   ```

2. **Test service management:**
   ```bash
   openCenter cluster init test-cluster
   ```

3. **Verify service dependencies:**
   - Enable a service with dependencies
   - Verify dependencies are automatically enabled

4. **Test circular dependency detection:**
   - Intentionally create circular dependencies
   - Verify detection and clear error message

**Rollback:**
```bash
unset OPENCENTER_USE_SERVICE_REGISTRY
# or
export OPENCENTER_USE_SERVICE_REGISTRY=false
```

**Known Issues:**
- None currently reported

## Troubleshooting

### Common Issues and Solutions

#### Issue: Feature flag not taking effect

**Symptoms:**
- Setting feature flag but legacy system still active
- No change in behavior after enabling flag

**Solutions:**
1. **Check environment variable is set:**
   ```bash
   echo $OPENCENTER_USE_NEW_TEMPLATE_ENGINE
   ```

2. **Verify correct value:**
   ```bash
   # Valid values: true, 1, yes, on (case-insensitive)
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   ```

3. **Enable debug logging:**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   openCenter config features
   ```

4. **Check for conflicting flags:**
   ```bash
   # Individual flags override OPENCENTER_ENABLE_ALL_NEW_FEATURES
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false  # This takes precedence
   ```

#### Issue: Different output between legacy and new systems

**Symptoms:**
- Generated files differ between systems
- Unexpected changes in output

**Solutions:**
1. **Enable debug logging:**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

2. **Compare outputs carefully:**
   ```bash
   diff -u legacy-output.yaml new-output.yaml
   ```

3. **Check for whitespace differences:**
   ```bash
   diff -w legacy-output.yaml new-output.yaml  # Ignore whitespace
   ```

4. **Report the issue:**
   - Include debug logs
   - Provide configuration file
   - Include diff output

#### Issue: Performance regression with new system

**Symptoms:**
- Operations slower with new system
- Increased memory usage

**Solutions:**
1. **Check if caching is enabled:**
   ```bash
   # Template caching should improve performance
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   ```

2. **Monitor resource usage:**
   ```bash
   time openCenter cluster render
   ```

3. **Report performance issues:**
   - Include timing comparisons
   - Provide configuration size
   - Include system specifications

#### Issue: Error messages unclear or unhelpful

**Symptoms:**
- Error messages don't explain the problem
- No suggestions for fixing errors

**Solutions:**
1. **Enable debug logging:**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

2. **Check validation errors:**
   ```bash
   openCenter cluster validate config.yaml
   ```

3. **Report unclear errors:**
   - Include full error message
   - Provide configuration that caused error
   - Describe expected vs actual behavior

### Getting Help

If you encounter issues not covered in this guide:

1. **Check existing documentation:**
   - [Architecture Documentation](../architecture.md)
   - [Feature Flag Removal Timeline](./feature-flag-removal-timeline.md)
   - [Developer Documentation](../dev/configuration-system.md)

2. **Enable debug logging:**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

3. **Collect diagnostic information:**
   ```bash
   # Feature flag status
   openCenter config features --output json > feature-flags.json
   
   # Configuration
   openCenter cluster validate config.yaml > validation.log 2>&1
   
   # Debug logs
   openCenter cluster render > render.log 2>&1
   ```

4. **Report the issue:**
   - Include all diagnostic information
   - Describe steps to reproduce
   - Specify which feature flags are enabled
   - Include openCenter version: `openCenter version`

## Rollback Procedures

If you encounter critical issues with the new systems, you can quickly rollback to the legacy implementation.

### Immediate Rollback (Critical Issues)

If you encounter a critical issue that blocks operations:

1. **Disable all new features immediately:**
   ```bash
   unset OPENCENTER_ENABLE_ALL_NEW_FEATURES
   unset OPENCENTER_USE_NEW_TEMPLATE_ENGINE
   unset OPENCENTER_USE_PIPELINE_GENERATOR
   unset OPENCENTER_USE_NEW_CONFIG_BUILDER
   unset OPENCENTER_USE_SERVICE_REGISTRY
   ```

2. **Verify legacy system is active:**
   ```bash
   openCenter config features
   # All features should show "disabled"
   ```

3. **Resume operations:**
   ```bash
   openCenter cluster render
   # Should now use legacy systems
   ```

4. **Report the issue** with debug logs

### Partial Rollback (Single Feature Issues)

If only one feature is problematic:

1. **Disable the specific feature:**
   ```bash
   # Keep other features enabled
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   export OPENCENTER_USE_PIPELINE_GENERATOR=false  # Disable only this one
   ```

2. **Verify the change:**
   ```bash
   openCenter config features
   ```

3. **Continue using other new features:**
   ```bash
   openCenter cluster render
   # Uses new template engine but legacy generator
   ```

### Rollback in CI/CD

If you've enabled features in CI/CD pipelines:

1. **Update environment variables:**
   ```yaml
   # GitHub Actions example
   env:
     OPENCENTER_ENABLE_ALL_NEW_FEATURES: false
     # or remove the variable entirely
   ```

2. **Redeploy pipeline:**
   ```bash
   git commit -m "Rollback to legacy openCenter systems"
   git push
   ```

3. **Verify rollback:**
   ```bash
   # Check pipeline logs for feature flag status
   ```

### Rollback Verification

After rollback, verify the legacy system is working correctly:

```bash
# 1. Check feature flag status
openCenter config features

# 2. Test basic operations
openCenter cluster validate config.yaml

# 3. Test generation
openCenter cluster render --dry-run

# 4. Verify output matches previous behavior
openCenter cluster render > output.yaml
diff output.yaml previous-output.yaml
```

## Performance Considerations

The refactored system includes several performance improvements:

### Template Caching

The new template engine caches parsed templates for reuse:

**Benefits:**
- 20-50% faster template rendering on subsequent renders
- Reduced memory allocations
- Lower CPU usage for repeated operations

**Monitoring:**
```bash
# Enable debug logging to see cache hits
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter cluster render
```

### Pipeline Generation

The new pipeline generator optimizes GitOps generation:

**Benefits:**
- Parallel template rendering where possible
- Progress reporting for long operations
- Efficient memory usage for large repositories

**Monitoring:**
```bash
# Time the operation
time openCenter cluster init large-cluster
```

### Configuration Building

The new configuration builder reduces validation overhead:

**Benefits:**
- Compile-time validation catches errors early
- Reduced runtime validation overhead
- Better error aggregation (single pass)

**Monitoring:**
```bash
# Time configuration validation
time openCenter cluster validate config.yaml
```

## Migration Timeline

The feature flag removal timeline defines when legacy systems will be deprecated and removed:

| Phase | Duration | Description | Action Required |
|-------|----------|-------------|-----------------|
| **Phase 1: Validation** | 4 weeks | Test new systems in production | Enable feature flags, monitor closely |
| **Phase 2: Default Transition** | 4 weeks | New systems become default | Update documentation, add deprecation warnings |
| **Phase 3: Deprecation** | 8 weeks | Legacy systems deprecated | Migrate all users to new systems |
| **Phase 4: Removal** | 1 week | Legacy systems removed | No action if already migrated |

**Total Timeline:** 17 weeks from Phase 1 start

See [Feature Flag Removal Timeline](./feature-flag-removal-timeline.md) for detailed information.

### Current Phase

**Status:** Phase 1 (Validation)

**Recommended Actions:**
1. Enable feature flags in non-production environments
2. Test thoroughly with your configurations
3. Report any issues or feedback
4. Monitor performance and error rates

**Timeline:**
- Phase 1 ends: TBD (4 weeks after start)
- Legacy system removal: TBD (17 weeks after start)

## Best Practices

### For Development

1. **Always use feature flags in development:**
   ```bash
   # Add to ~/.bashrc or ~/.zshrc
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

2. **Test both systems during development:**
   ```bash
   # Test with new systems
   export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
   mise run test
   
   # Test with legacy systems
   unset OPENCENTER_ENABLE_ALL_NEW_FEATURES
   mise run test
   ```

3. **Use debug logging for troubleshooting:**
   ```bash
   export OPENCENTER_FEATURE_FLAG_DEBUG=true
   ```

### For Production

1. **Enable features gradually:**
   - Start with one feature
   - Monitor for 1-2 weeks
   - Enable next feature

2. **Monitor closely after enabling:**
   - Check logs for errors
   - Monitor performance metrics
   - Compare output with legacy system

3. **Keep rollback plan ready:**
   - Document how to disable features
   - Test rollback procedure
   - Communicate plan to team

4. **Document your configuration:**
   ```bash
   # Save current feature flag status
   openCenter config features --output env > feature-flags.sh
   ```

### For CI/CD

1. **Set feature flags in pipeline configuration:**
   ```yaml
   # GitHub Actions example
   env:
     OPENCENTER_ENABLE_ALL_NEW_FEATURES: true
     OPENCENTER_FEATURE_FLAG_DEBUG: true
   ```

2. **Test both systems in CI:**
   ```yaml
   strategy:
     matrix:
       feature_flags: [true, false]
   env:
     OPENCENTER_ENABLE_ALL_NEW_FEATURES: ${{ matrix.feature_flags }}
   ```

3. **Monitor pipeline performance:**
   - Track execution time
   - Compare with baseline
   - Alert on regressions

## FAQ

### Q: Do I need to migrate immediately?

**A:** No. The legacy systems will remain available for several months (see timeline above). You can migrate at your own pace.

### Q: Can I use some new features but not others?

**A:** Yes. Each feature flag is independent. You can enable only the features you want to use.

### Q: What happens if I don't migrate before legacy removal?

**A:** After Phase 4 (legacy removal), the feature flags will be removed and only the new systems will be available. If you haven't migrated by then, you'll need to update to the new systems at that time.

### Q: Will my existing configurations work with the new systems?

**A:** Yes. The new systems maintain full backward compatibility with existing configurations. Output should be identical to the legacy systems.

### Q: How do I know which system is currently active?

**A:** Use `openCenter config features` to see which features are enabled, or enable debug logging with `OPENCENTER_FEATURE_FLAG_DEBUG=true`.

### Q: Can I rollback after enabling a feature?

**A:** Yes. Simply unset the feature flag or set it to `false`. The legacy system will be used immediately.

### Q: Are there any breaking changes?

**A:** No. The refactored systems maintain full backward compatibility. Breaking changes will only occur in Phase 4 when legacy systems are removed.

### Q: How do I report issues with the new systems?

**A:** Enable debug logging (`OPENCENTER_FEATURE_FLAG_DEBUG=true`), collect diagnostic information, and report the issue with full details including which feature flags are enabled.

### Q: Will performance improve with the new systems?

**A:** Yes. The new systems include several performance optimizations, particularly template caching and optimized GitOps generation. You should see 20-50% improvement in template rendering.

### Q: Do I need to change my code or configurations?

**A:** No. The new systems are designed to be drop-in replacements. No code or configuration changes are required.

## Additional Resources

- **Architecture Documentation:** [docs/architecture.md](../architecture.md)
- **Feature Flag Removal Timeline:** [docs/migration/feature-flag-removal-timeline.md](./feature-flag-removal-timeline.md)
- **Developer Guide:** [docs/dev/configuration-system.md](../dev/configuration-system.md)
- **Requirements Document:** [.kiro/specs/configuration-system-refactor/requirements.md](../../.kiro/specs/configuration-system-refactor/requirements.md)
- **Design Document:** [.kiro/specs/configuration-system-refactor/design.md](../../.kiro/specs/configuration-system-refactor/design.md)
- **Tasks Document:** [.kiro/specs/configuration-system-refactor/tasks.md](../../.kiro/specs/configuration-system-refactor/tasks.md)

## Conclusion

The configuration system refactor provides a more maintainable, extensible, and performant architecture while maintaining full backward compatibility. Feature flags enable safe, gradual migration with the ability to quickly rollback if issues arise.

**Key Takeaways:**
- ✅ Feature flags enable gradual, safe migration
- ✅ Legacy systems remain available during transition
- ✅ Rollback is instant and simple
- ✅ No breaking changes until Phase 4 (17 weeks)
- ✅ Performance improvements with new systems
- ✅ Better error messages and debugging

**Next Steps:**
1. Review this migration guide
2. Enable feature flags in development environment
3. Test thoroughly with your configurations
4. Enable in production when confident
5. Monitor and report any issues

For questions or issues, refer to the troubleshooting section or contact the development team.
