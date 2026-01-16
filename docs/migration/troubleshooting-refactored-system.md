# Troubleshooting Guide: Refactored Configuration System

This guide helps you diagnose and resolve issues specific to the refactored configuration system, including the new template engine, pipeline generator, configuration builder, and service registry.

## Table of Contents

- [Debug Mode and Logging](#debug-mode-and-logging)
- [Feature Flag Issues](#feature-flag-issues)
- [Template Engine Issues](#template-engine-issues)
- [Configuration Builder Issues](#configuration-builder-issues)
- [GitOps Pipeline Generator Issues](#gitops-pipeline-generator-issues)
- [Service Registry Issues](#service-registry-issues)
- [Performance Issues](#performance-issues)
- [Migration Issues](#migration-issues)
- [Log Analysis](#log-analysis)
- [Common Error Messages](#common-error-messages)

## Debug Mode and Logging

### Enabling Debug Mode

The refactored system provides multiple levels of debugging:

#### 1. Feature Flag Debug Mode

Enable detailed feature flag evaluation logging:

```bash
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter cluster render my-cluster
```

This shows which systems (legacy vs new) are being used for each operation.

#### 2. Application Debug Logging

Enable debug-level logging for all operations:

```bash
# Set log level to debug
export OPENCENTER_LOG_LEVEL=debug
openCenter cluster render my-cluster

# Or use CLI flag (if implemented)
openCenter --log-level=debug cluster render my-cluster
```

#### 3. Component-Specific Debug Logging

Enable debug logging for specific components:

```bash
# Template engine debug
export OPENCENTER_TEMPLATE_DEBUG=true

# Pipeline generator debug
export OPENCENTER_PIPELINE_DEBUG=true

# Configuration builder debug
export OPENCENTER_CONFIG_BUILDER_DEBUG=true

# Service registry debug
export OPENCENTER_SERVICE_REGISTRY_DEBUG=true
```

#### 4. Structured Logging Output

Change log format for better analysis:

```bash
# JSON format for machine parsing
export OPENCENTER_LOG_FORMAT=json
openCenter cluster render my-cluster > logs.json

# YAML format for human readability
export OPENCENTER_LOG_FORMAT=yaml
openCenter cluster render my-cluster

# Text format (default)
export OPENCENTER_LOG_FORMAT=text
openCenter cluster render my-cluster
```

### Log File Configuration

Configure persistent logging to file:

```bash
# Log to file with rotation
export OPENCENTER_LOG_OUTPUT=~/.config/openCenter/logs/opencenter.log
export OPENCENTER_LOG_FILE_MAX_SIZE=100    # MB
export OPENCENTER_LOG_FILE_MAX_BACKUPS=5   # number of old files
export OPENCENTER_LOG_FILE_MAX_AGE=30      # days
export OPENCENTER_LOG_FILE_COMPRESS=true   # compress old logs

openCenter cluster render my-cluster
```

### Viewing Feature Flag Status

Check which feature flags are currently active:

```bash
# Using the config features command
openCenter config features

# Or check environment variables
env | grep OPENCENTER_
```

## Feature Flag Issues

### Issue: Not sure which system is being used

**Symptom:** Unclear whether legacy or new system is active.

**Solution:**

```bash
# Enable feature flag debug mode
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter cluster render my-cluster

# Output will show:
# [FEATURE FLAG] new template engine is enabled (OPENCENTER_USE_NEW_TEMPLATE_ENGINE, source: environment)
# [FEATURE FLAG] pipeline generator is disabled (OPENCENTER_USE_PIPELINE_GENERATOR, source: default)
```

### Issue: Feature flag not taking effect

**Symptom:** Setting feature flag doesn't change behavior.

**Cause:** Feature flag cache or incorrect value format.

**Solution:**

```bash
# Ensure correct value format (case-insensitive)
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true  # ✓ Valid
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=1     # ✓ Valid
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=yes   # ✓ Valid
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=on    # ✓ Valid
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=True  # ✓ Valid (case-insensitive)

# Invalid values (treated as false)
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false # ✗ Treated as false
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=0     # ✗ Treated as false
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=      # ✗ Empty = false

# Verify the flag is set
echo $OPENCENTER_USE_NEW_TEMPLATE_ENGINE

# Check if it's being read
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter config features
```

### Issue: Conflicting feature flags

**Symptom:** Unexpected behavior when multiple flags are set.

**Cause:** Individual flags override the global "all new features" flag.

**Solution:**

```bash
# Understand precedence:
# 1. Individual flags (highest priority)
# 2. OPENCENTER_ENABLE_ALL_NEW_FEATURES
# 3. Default (false)

# Example: Enable all except template engine
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false  # This overrides the global flag

# Verify with debug mode
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter config features
```

## Template Engine Issues

### Issue: Template rendering fails with new engine

**Symptom:** Templates that worked with legacy engine fail with new engine.

**Error Example:**
```
Error: template rendering failed: template: cluster.yaml:15:23: executing "cluster.yaml" at <.Config.InvalidField>: can't evaluate field InvalidField in type config.Config
```

**Diagnosis:**

```bash
# Enable template debug mode
export OPENCENTER_TEMPLATE_DEBUG=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster
```

**Solution:**

1. **Check template syntax:**
   ```bash
   # Validate template syntax
   openCenter cluster validate my-cluster --check-templates
   ```

2. **Compare with legacy engine:**
   ```bash
   # Render with legacy engine
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
   openCenter cluster render my-cluster --output legacy-output/
   
   # Render with new engine
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
   openCenter cluster render my-cluster --output new-output/
   
   # Compare outputs
   diff -r legacy-output/ new-output/
   ```

3. **Check for missing functions:**
   ```bash
   # The new engine may have different function availability
   # Check template for custom functions
   grep -r "{{ .*| " templates/
   ```

4. **Rollback to legacy engine:**
   ```bash
   export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
   openCenter cluster render my-cluster
   ```

### Issue: Template cache causing stale output

**Symptom:** Template changes not reflected in output.

**Cause:** Template caching enabled in new engine.

**Solution:**

```bash
# Clear template cache
export OPENCENTER_TEMPLATE_CACHE_ENABLED=false
openCenter cluster render my-cluster

# Or force cache clear
openCenter cluster render my-cluster --clear-cache

# For development, disable caching
export OPENCENTER_TEMPLATE_CACHE_ENABLED=false
```

### Issue: Template error messages unclear

**Symptom:** Error doesn't show which template or line failed.

**Solution:**

```bash
# Enable verbose error reporting
export OPENCENTER_TEMPLATE_VERBOSE_ERRORS=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster

# Output will include:
# - Template file path
# - Line and column numbers
# - Context around the error
# - Suggested fixes
```

## Configuration Builder Issues

### Issue: Type-safe builder compilation errors

**Symptom:** Code using new builder doesn't compile.

**Error Example:**
```
cannot use "invalid-value" (type string) as type int in argument to WithNodeCounts
```

**Solution:**

This is actually a feature! The type-safe builder catches errors at compile time.

```go
// ✗ Wrong: Will not compile
builder.WithNodeCounts("3", "5")

// ✓ Correct: Use proper types
builder.WithNodeCounts(3, 5)
```

### Issue: Builder validation errors

**Symptom:** Build() returns validation errors.

**Error Example:**
```
validation failed: 
  - field 'opencenter.organization' is required
  - field 'opencenter.cluster_name' must be 3-63 characters
```

**Diagnosis:**

```bash
# Enable config builder debug mode
export OPENCENTER_CONFIG_BUILDER_DEBUG=true
export OPENCENTER_USE_NEW_CONFIG_BUILDER=true
openCenter cluster init my-cluster
```

**Solution:**

```bash
# Check all validation errors at once
openCenter cluster validate my-cluster --verbose

# The new builder aggregates all errors, so fix them all together
# rather than one at a time
```

### Issue: Configuration migration fails

**Symptom:** Old configuration can't be loaded with new builder.

**Error Example:**
```
Error: configuration migration failed: unsupported schema version v1.0.0
```

**Solution:**

```bash
# Check current schema version
openCenter cluster schema --version

# Migrate configuration manually
openCenter cluster migrate my-cluster --from v1.0.0 --to v2.0.0

# Or use automatic migration
export OPENCENTER_AUTO_MIGRATE=true
openCenter cluster validate my-cluster

# Dry-run migration to preview changes
openCenter cluster migrate my-cluster --dry-run
```

## GitOps Pipeline Generator Issues

### Issue: Pipeline generation fails mid-stage

**Symptom:** Generation fails partway through, leaving partial output.

**Error Example:**
```
Error: stage 'infrastructure' failed: template rendering error
Rolling back previous stages...
Rollback complete.
```

**Diagnosis:**

```bash
# Enable pipeline debug mode
export OPENCENTER_PIPELINE_DEBUG=true
export OPENCENTER_USE_PIPELINE_GENERATOR=true
openCenter cluster render my-cluster
```

**Solution:**

The new pipeline generator automatically rolls back on failure. Check the logs to see which stage failed:

```bash
# View detailed stage execution
export OPENCENTER_PIPELINE_DEBUG=true
export OPENCENTER_LOG_LEVEL=debug
openCenter cluster render my-cluster 2>&1 | tee pipeline.log

# Look for stage failures
grep "stage.*failed" pipeline.log

# Check rollback operations
grep "rollback" pipeline.log
```

### Issue: Dry-run shows different output than actual run

**Symptom:** Dry-run preview doesn't match actual generation.

**Cause:** Dry-run mode may not execute all validation steps.

**Solution:**

```bash
# Run dry-run with full validation
openCenter cluster render my-cluster --dry-run --validate

# Compare dry-run with actual run
openCenter cluster render my-cluster --dry-run --output dry-run-preview/
openCenter cluster render my-cluster --output actual-output/
diff -r dry-run-preview/ actual-output/
```

### Issue: Workspace checkpoint/rollback not working

**Symptom:** Rollback doesn't restore previous state.

**Diagnosis:**

```bash
# Check workspace checkpoints
openCenter cluster render my-cluster --list-checkpoints

# View checkpoint details
openCenter cluster render my-cluster --checkpoint-info <checkpoint-id>
```

**Solution:**

```bash
# Manual rollback to specific checkpoint
openCenter cluster render my-cluster --rollback <checkpoint-id>

# Clear all checkpoints and start fresh
openCenter cluster render my-cluster --clear-checkpoints --force
```

### Issue: Progress reporting not showing

**Symptom:** No progress updates during long-running generation.

**Solution:**

```bash
# Enable progress reporting
export OPENCENTER_SHOW_PROGRESS=true
openCenter cluster render my-cluster

# Or use verbose mode
openCenter --verbose cluster render my-cluster
```

## Service Registry Issues

### Issue: Service plugin not loading

**Symptom:** Custom service plugin not recognized.

**Error Example:**
```
Error: service 'my-custom-service' not found in registry
```

**Diagnosis:**

```bash
# List registered services
openCenter service list

# Enable service registry debug mode
export OPENCENTER_SERVICE_REGISTRY_DEBUG=true
export OPENCENTER_USE_SERVICE_REGISTRY=true
openCenter cluster render my-cluster
```

**Solution:**

```bash
# Check plugin manifest
cat ~/.config/openCenter/plugins/my-custom-service/manifest.yaml

# Validate plugin manifest
openCenter service validate my-custom-service

# Reload service registry
openCenter service reload

# Check plugin directory permissions
ls -la ~/.config/openCenter/plugins/
```

### Issue: Circular dependency detected

**Symptom:** Service dependency resolution fails.

**Error Example:**
```
Error: circular dependency detected: service-a -> service-b -> service-c -> service-a
```

**Diagnosis:**

```bash
# Visualize service dependency graph
openCenter service dependencies --graph

# Check specific service dependencies
openCenter service dependencies my-service
```

**Solution:**

```bash
# Review service manifests and remove circular dependencies
# Edit service manifest files to break the cycle

# Validate dependencies after changes
openCenter service validate --check-dependencies
```

### Issue: Service lifecycle hooks failing

**Symptom:** Pre/post install hooks fail during service deployment.

**Error Example:**
```
Error: service 'my-service' pre-install hook failed: exit code 1
```

**Diagnosis:**

```bash
# Enable lifecycle hook debug mode
export OPENCENTER_SERVICE_LIFECYCLE_DEBUG=true
openCenter cluster render my-cluster
```

**Solution:**

```bash
# Test lifecycle hooks manually
openCenter service test-hook my-service --hook pre-install

# Skip failing hooks (not recommended for production)
openCenter cluster render my-cluster --skip-hooks

# View hook logs
cat ~/.config/openCenter/logs/service-hooks.log
```

## Performance Issues

### Issue: Slow template rendering

**Symptom:** Template rendering takes longer than expected.

**Diagnosis:**

```bash
# Enable performance metrics
export OPENCENTER_ENABLE_METRICS=true
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster

# View metrics
openCenter metrics show --component template-engine
```

**Solution:**

```bash
# Enable template caching
export OPENCENTER_TEMPLATE_CACHE_ENABLED=true

# Increase cache size
export OPENCENTER_TEMPLATE_CACHE_SIZE=1000

# Use parallel rendering (if available)
export OPENCENTER_PARALLEL_RENDERING=true
export OPENCENTER_MAX_PARALLEL_RENDERS=4

# Profile template rendering
openCenter cluster render my-cluster --profile --profile-output profile.pprof
go tool pprof profile.pprof
```

### Issue: High memory usage during generation

**Symptom:** Memory usage spikes during GitOps generation.

**Diagnosis:**

```bash
# Monitor memory usage
openCenter cluster render my-cluster --memory-profile memory.pprof

# Analyze memory profile
go tool pprof -http=:8080 memory.pprof
```

**Solution:**

```bash
# Enable streaming mode (if available)
export OPENCENTER_STREAMING_MODE=true

# Reduce parallel operations
export OPENCENTER_MAX_PARALLEL_RENDERS=2

# Clear caches before generation
openCenter cluster render my-cluster --clear-cache

# Use incremental generation
openCenter cluster render my-cluster --incremental
```

### Issue: Slow configuration validation

**Symptom:** Validation takes longer than a few seconds.

**Solution:**

```bash
# Skip connectivity checks
openCenter cluster validate my-cluster --skip-connectivity

# Use schema-only validation
openCenter cluster validate my-cluster --schema-only

# Disable expensive validations
export OPENCENTER_FAST_VALIDATION=true
openCenter cluster validate my-cluster
```

## Migration Issues

### Issue: Legacy configuration not compatible

**Symptom:** Old configuration file doesn't work with new system.

**Error Example:**
```
Error: unsupported configuration schema version: v1.0.0
```

**Solution:**

```bash
# Check configuration version
openCenter cluster info my-cluster --show-version

# Migrate to current version
openCenter cluster migrate my-cluster --auto

# Or migrate step by step
openCenter cluster migrate my-cluster --from v1.0.0 --to v1.1.0
openCenter cluster migrate my-cluster --from v1.1.0 --to v2.0.0

# Dry-run migration first
openCenter cluster migrate my-cluster --dry-run --verbose
```

### Issue: Output differs between legacy and new system

**Symptom:** New system produces different output than legacy.

**Diagnosis:**

```bash
# Generate with both systems
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
openCenter cluster render my-cluster --output legacy/

export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
openCenter cluster render my-cluster --output new/

# Compare outputs
diff -r legacy/ new/

# Or use built-in comparison
openCenter cluster compare-output legacy/ new/
```

**Solution:**

If differences are found:

1. **Check if differences are intentional** (e.g., improved formatting)
2. **Report unexpected differences** as bugs
3. **Use legacy system** until differences are resolved
4. **Update templates** if new system requires changes

### Issue: Feature flag migration path unclear

**Symptom:** Not sure when to enable new features.

**Solution:**

Follow the recommended migration path:

```bash
# Phase 1: Test individual features in development
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
# Test thoroughly

# Phase 2: Enable in staging
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
export OPENCENTER_USE_NEW_CONFIG_BUILDER=true
# Test thoroughly

# Phase 3: Enable pipeline generator
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=true
export OPENCENTER_USE_NEW_CONFIG_BUILDER=true
export OPENCENTER_USE_PIPELINE_GENERATOR=true
# Test thoroughly

# Phase 4: Enable all features
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
# Test thoroughly

# Phase 5: Make new features default (future release)
# No flags needed, new system becomes default
```

## Log Analysis

### Understanding Log Structure

The refactored system uses structured logging with consistent fields:

```json
{
  "timestamp": "2025-01-16T10:30:45.123Z",
  "level": "info",
  "component": "template_engine",
  "operation": "render",
  "template_name": "cluster.yaml",
  "duration_ms": 45,
  "cache_hit": true,
  "message": "Template rendered successfully"
}
```

### Key Log Fields

- **component**: Which system generated the log (template_engine, pipeline_generator, config_builder, service_registry, feature_flags)
- **operation**: What operation was being performed (render, validate, build, register, evaluate)
- **duration_ms**: How long the operation took
- **error**: Error message if operation failed
- **context**: Additional context (file paths, line numbers, etc.)

### Analyzing Logs

#### Find all errors:

```bash
# JSON logs
cat logs.json | jq 'select(.level == "error")'

# Text logs
grep "level=error" opencenter.log

# YAML logs
grep "level: error" opencenter.log
```

#### Find slow operations:

```bash
# JSON logs - operations taking > 1 second
cat logs.json | jq 'select(.duration_ms > 1000)'

# Find slowest operations
cat logs.json | jq -s 'sort_by(.duration_ms) | reverse | .[0:10]'
```

#### Track feature flag usage:

```bash
# JSON logs
cat logs.json | jq 'select(.component == "feature_flags")'

# See which features are enabled
cat logs.json | jq 'select(.component == "feature_flags" and .operation == "evaluation") | {feature_name, enabled, source}'
```

#### Find template rendering issues:

```bash
# JSON logs
cat logs.json | jq 'select(.component == "template_engine" and .level == "error")'

# Group by template name
cat logs.json | jq 'select(.component == "template_engine") | .template_name' | sort | uniq -c
```

#### Track pipeline stages:

```bash
# JSON logs
cat logs.json | jq 'select(.component == "pipeline_generator") | {stage_name, operation, duration_ms}'

# Find failed stages
cat logs.json | jq 'select(.component == "pipeline_generator" and .operation == "rollback")'
```

### Log Correlation

Correlate logs across components using operation IDs:

```bash
# Find all logs for a specific operation
OPERATION_ID="abc123"
cat logs.json | jq "select(.operation_id == \"$OPERATION_ID\")"

# Trace a request through the system
cat logs.json | jq "select(.request_id == \"$REQUEST_ID\") | {timestamp, component, operation, message}"
```

## Common Error Messages

### "feature flag cache inconsistency detected"

**Meaning:** Feature flag cache is out of sync with environment variables.

**Solution:**
```bash
# Clear feature flag cache
unset OPENCENTER_FEATURE_FLAG_DEBUG
export OPENCENTER_FEATURE_FLAG_DEBUG=true
openCenter config features --clear-cache
```

### "template engine not initialized"

**Meaning:** Template engine wasn't properly initialized before use.

**Solution:**
```bash
# This is likely a bug. Report it with:
export OPENCENTER_TEMPLATE_DEBUG=true
export OPENCENTER_LOG_LEVEL=debug
openCenter cluster render my-cluster 2>&1 | tee bug-report.log
```

### "configuration builder validation failed: multiple errors"

**Meaning:** Configuration has multiple validation errors.

**Solution:**
```bash
# View all errors at once
openCenter cluster validate my-cluster --verbose

# Fix all errors before retrying
# The new builder aggregates errors for efficiency
```

### "pipeline stage rollback failed"

**Meaning:** Automatic rollback after stage failure encountered an error.

**Solution:**
```bash
# Manual cleanup may be required
openCenter cluster render my-cluster --force-clean

# Check workspace state
openCenter cluster render my-cluster --workspace-status

# Report as bug if rollback consistently fails
```

### "service dependency resolution failed: cycle detected"

**Meaning:** Services have circular dependencies.

**Solution:**
```bash
# Visualize dependency graph
openCenter service dependencies --graph --output deps.dot
dot -Tpng deps.dot -o deps.png

# Review and break the cycle in service manifests
```

### "migration path not found: v1.0.0 -> v3.0.0"

**Meaning:** No direct migration path exists between versions.

**Solution:**
```bash
# Migrate step by step
openCenter cluster migrate my-cluster --from v1.0.0 --to v2.0.0
openCenter cluster migrate my-cluster --from v2.0.0 --to v3.0.0

# Or use auto-migration
openCenter cluster migrate my-cluster --auto
```

## Getting Help

### Collect Diagnostic Information

When reporting issues, collect this information:

```bash
#!/bin/bash
# diagnostic-collect.sh

echo "=== System Information ===" > diagnostic.txt
uname -a >> diagnostic.txt
echo "" >> diagnostic.txt

echo "=== openCenter Version ===" >> diagnostic.txt
openCenter --version >> diagnostic.txt
echo "" >> diagnostic.txt

echo "=== Feature Flags ===" >> diagnostic.txt
env | grep OPENCENTER_ >> diagnostic.txt
echo "" >> diagnostic.txt

echo "=== Feature Flag Status ===" >> diagnostic.txt
openCenter config features >> diagnostic.txt
echo "" >> diagnostic.txt

echo "=== Cluster Info ===" >> diagnostic.txt
openCenter cluster info my-cluster >> diagnostic.txt 2>&1
echo "" >> diagnostic.txt

echo "=== Validation Output ===" >> diagnostic.txt
openCenter --verbose cluster validate my-cluster >> diagnostic.txt 2>&1
echo "" >> diagnostic.txt

echo "=== Recent Logs ===" >> diagnostic.txt
tail -n 100 ~/.config/openCenter/logs/opencenter.log >> diagnostic.txt 2>&1

echo "Diagnostic information collected in diagnostic.txt"
```

### Enable Maximum Debug Output

```bash
# Enable all debug flags
export OPENCENTER_FEATURE_FLAG_DEBUG=true
export OPENCENTER_TEMPLATE_DEBUG=true
export OPENCENTER_PIPELINE_DEBUG=true
export OPENCENTER_CONFIG_BUILDER_DEBUG=true
export OPENCENTER_SERVICE_REGISTRY_DEBUG=true
export OPENCENTER_LOG_LEVEL=debug
export OPENCENTER_LOG_FORMAT=json

# Run command and capture output
openCenter cluster render my-cluster 2>&1 | tee debug-output.log
```

### Report Issues

Include in your bug report:

1. **Diagnostic information** (from script above)
2. **Command executed** (exact command with all flags)
3. **Expected behavior** (what should happen)
4. **Actual behavior** (what actually happened)
5. **Debug logs** (with all debug flags enabled)
6. **Configuration file** (sanitized, no secrets)
7. **Steps to reproduce** (minimal reproduction case)

Submit issues at: https://github.com/rackerlabs/openCenter-cli/issues

## Additional Resources

- [Feature Flag Migration Guide](template-engine.md#feature-flags)
- [Configuration System Design](.kiro/specs/configuration-system-refactor/design.md)
- [Implementation Tasks](.kiro/specs/configuration-system-refactor/tasks.md)
- [Performance Characteristics](../dev/performance-characteristics.md)
- [General Troubleshooting](../TROUBLESHOOTING.md)

---

**Note:** This guide covers the refactored configuration system. For general openCenter troubleshooting, see [docs/TROUBLESHOOTING.md](../TROUBLESHOOTING.md).
