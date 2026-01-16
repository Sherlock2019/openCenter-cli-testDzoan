# Legacy Code Removal Process

## Overview

This document provides a comprehensive guide for identifying and removing legacy code after the configuration system refactor is complete and all feature flags have been successfully removed. This process should only begin after Phase 4 of the feature flag cleanup (see `feature-flag-cleanup-guide.md`).

## Prerequisites

Before starting legacy code removal:

1. ✅ Feature flags have been removed (Phase 4 complete)
2. ✅ All tests pass without feature flags
3. ✅ Production has been running new systems for 60+ days
4. ✅ Zero critical issues reported
5. ✅ Stakeholder approval obtained
6. ✅ Release tagged before removal (for rollback capability)

## Legacy Code Inventory

### 1. Template System Legacy Code

#### Files to Remove

```bash
# Legacy compatibility layer
internal/template/legacy.go
internal/template/legacy_test.go

# Migration validation tests (archive, don't delete)
internal/template/migration_test.go
internal/template/migration_path_validation_test.go
```

#### Code Patterns to Remove

**Pattern 1: Legacy Template Rendering Functions**

```go
// REMOVE: Legacy template rendering function
func renderLegacyTemplate(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    // ... legacy implementation ...
}

// REMOVE: Legacy compatibility wrapper
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    if UseNewTemplateEngine() {
        return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
    }
    return renderLegacyTemplate(fsys, templatePath, outputPath, data)
}

// KEEP: Direct call to new engine
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
}
```

**Pattern 2: Legacy Template String Rendering**

```go
// REMOVE: Legacy string rendering
func RenderTemplateString(name, templateStr string, data interface{}) (string, error) {
    if UseNewTemplateEngine() {
        return renderWithNewEngine(name, templateStr, data)
    }
    return renderLegacyString(name, templateStr, data)
}

// KEEP: Direct call to new engine
func RenderTemplateString(name, templateStr string, data interface{}) (string, error) {
    return renderWithNewEngine(name, templateStr, data)
}
```

#### Functions to Remove

- `renderLegacyTemplate()` - Legacy template file rendering
- `renderLegacyString()` - Legacy template string rendering
- `legacyTemplateCache` - Legacy template cache (if separate from new cache)
- Any `*Legacy()` or `*Compat()` suffixed functions

#### Impact Analysis

**Files that import template legacy code:**
```bash
# Find all files that import legacy template functions
grep -r "renderLegacyTemplate\|RenderTemplateString" internal/ cmd/
```

**Expected imports to update:**
- `internal/gitops/copy.go` - May use template rendering
- `cmd/cluster_render.go` - CLI command for template rendering
- Any service plugins that render templates

### 2. GitOps Generation Legacy Code

#### Files to Remove

```bash
# Legacy compatibility layer
internal/gitops/legacy_compat.go
internal/gitops/legacy_compat_test.go

# Backward compatibility tests (archive, don't delete)
internal/gitops/backward_compatibility_test.go
internal/gitops/migration_test.go
```

#### Code Patterns to Remove

**Pattern 1: Legacy Generation Wrapper**

```go
// REMOVE: Legacy generation wrapper
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    if usePipelineGenerator() {
        generator := NewPipelineGenerator()
        return generator.Generate(ctx, cfg)
    }
    return generateGitOpsRepositoryLegacy(cfg)
}

// KEEP: Direct call to pipeline generator
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    generator := NewPipelineGenerator()
    return generator.Generate(ctx, cfg)
}
```

**Pattern 2: Legacy Generation Functions**

```go
// REMOVE: Legacy generation implementation
func generateGitOpsRepositoryLegacy(cfg config.Config) error {
    if err := CopyBase(cfg, true); err != nil {
        return err
    }
    if err := RenderClusterApps(cfg); err != nil {
        return err
    }
    return RenderInfrastructureCluster(cfg)
}
```

**Pattern 3: Legacy Generation Wrapper Type**

```go
// REMOVE: Entire LegacyGenerationWrapper type and methods
type LegacyGenerationWrapper struct {
    config config.Config
}

func NewLegacyGenerationWrapper(cfg config.Config) *LegacyGenerationWrapper { ... }
func (w *LegacyGenerationWrapper) Generate() error { ... }
func (w *LegacyGenerationWrapper) CopyBase(render bool) error { ... }
func (w *LegacyGenerationWrapper) RenderClusterApps() error { ... }
func (w *LegacyGenerationWrapper) RenderInfrastructureCluster() error { ... }
```

#### Functions to Remove

- `generateGitOpsRepositoryLegacy()` - Legacy generation flow
- `GenerateWithOptions()` - Legacy options wrapper (if not used by new system)
- `NewLegacyGenerationWrapper()` - Legacy wrapper constructor
- All `LegacyGenerationWrapper` methods
- `RenderService()` - Legacy service rendering wrapper (if separate from new system)

#### Functions to Keep (But Simplify)

These functions are still used by the new pipeline system, but can be simplified:

```go
// KEEP: CopyBase is used by pipeline stages
func CopyBase(cfg config.Config, render bool) error {
    // Keep implementation, but remove any legacy-specific logic
}

// KEEP: RenderClusterApps is used by pipeline stages
func RenderClusterApps(cfg config.Config) error {
    // Keep implementation, but remove any legacy-specific logic
}

// KEEP: RenderInfrastructureCluster is used by pipeline stages
func RenderInfrastructureCluster(cfg config.Config) error {
    // Keep implementation, but remove any legacy-specific logic
}
```

#### Impact Analysis

**Files that import gitops legacy code:**
```bash
# Find all files that import legacy gitops functions
grep -r "GenerateGitOpsRepository\|LegacyGenerationWrapper" internal/ cmd/
```

**Expected imports to update:**
- `cmd/cluster_render.go` - CLI command for GitOps generation
- `cmd/cluster_bootstrap.go` - Bootstrap command
- Any integration tests that generate GitOps repositories

### 3. Configuration Builder Legacy Code

#### Files to Review

```bash
# Configuration builder files (may contain legacy compatibility code)
internal/config/builder.go
internal/config/builder_test.go
```

#### Code Patterns to Remove

**Pattern 1: Legacy Builder Compatibility**

```go
// REMOVE: Legacy builder wrapper (if exists)
func NewConfigBuilder() ConfigBuilder {
    if UseNewConfigBuilder() {
        return NewFluentConfigBuilder()
    }
    return newLegacyConfigBuilder()
}

// KEEP: Direct call to new builder
func NewConfigBuilder() ConfigBuilder {
    return NewFluentConfigBuilder()
}
```

**Pattern 2: Legacy Builder Implementation**

```go
// REMOVE: Legacy builder type and methods (if separate from new builder)
type legacyConfigBuilder struct { ... }
func newLegacyConfigBuilder() *legacyConfigBuilder { ... }
```

#### Impact Analysis

**Files that use configuration builder:**
```bash
# Find all files that create configuration builders
grep -r "NewConfigBuilder\|NewFluentConfigBuilder" internal/ cmd/
```

**Expected imports to update:**
- `cmd/cluster_init.go` - Cluster initialization
- `cmd/cluster_validate.go` - Configuration validation
- Any tests that build configurations

### 4. Service Registry Legacy Code

#### Files to Review

```bash
# Service registry files (may contain legacy compatibility code)
internal/services/registry.go
internal/services/registry_test.go
```

#### Code Patterns to Remove

**Pattern 1: Legacy Service Registry**

```go
// REMOVE: Legacy registry wrapper (if exists)
func GetServiceRegistry() ServiceRegistry {
    if UseServiceRegistry() {
        return getNewServiceRegistry()
    }
    return getLegacyServiceRegistry()
}

// KEEP: Direct call to new registry
func GetServiceRegistry() ServiceRegistry {
    return getNewServiceRegistry()
}
```

**Pattern 2: Legacy Service Management**

```go
// REMOVE: Legacy service management functions (if separate from new system)
func registerServiceLegacy(service ServiceDefinition) error { ... }
func getEnabledServicesLegacy(config Config) []ServiceDefinition { ... }
```

#### Impact Analysis

**Files that use service registry:**
```bash
# Find all files that use service registry
grep -r "GetServiceRegistry\|ServiceRegistry" internal/ cmd/
```

**Expected imports to update:**
- `internal/gitops/generator.go` - GitOps generation uses service registry
- `cmd/cluster_service.go` - Service management commands
- Any service plugins

### 5. Feature Flag System (Already Removed in Phase 4)

These files should already be removed in Phase 4 of feature flag cleanup:

```bash
# Already removed in Phase 4
internal/config/feature_flags.go
internal/config/feature_flags_test.go
internal/config/feature_flags_logging_test.go
internal/config/feature_flags_example_test.go
internal/config/feature_flags_removal_test.go
```

If these files still exist, remove them as part of this process.

## Removal Process

### Step 1: Identify All Legacy Code References

Run comprehensive searches to find all legacy code:

```bash
# Search for legacy function calls
grep -r "legacy\|Legacy" internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "// "

# Search for compatibility wrappers
grep -r "compat\|Compat" internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "// "

# Search for backward compatibility code
grep -r "backward\|Backward" internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "// "

# Search for migration code
grep -r "migration\|Migration" internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "// "

# Search for feature flag checks (should be none after Phase 4)
grep -r "UseNew\|usePipeline" internal/ cmd/ --include="*.go"
```

### Step 2: Create Removal Branch

```bash
# Create a dedicated branch for legacy code removal
git checkout -b refactor/remove-legacy-code

# Tag current state for easy rollback
git tag pre-legacy-removal
```

### Step 3: Remove Template Legacy Code

```bash
# Remove legacy template files
git rm internal/template/legacy.go
git rm internal/template/legacy_test.go

# Archive migration tests (don't delete, move to archive)
mkdir -p internal/template/archive
git mv internal/template/migration_test.go internal/template/archive/
git mv internal/template/migration_path_validation_test.go internal/template/archive/

# Commit template legacy removal
git commit -m "refactor: remove template legacy compatibility layer

- Remove legacy.go and legacy_test.go
- Archive migration validation tests for historical reference
- All template rendering now uses new engine directly

Refs: #<issue-number>"
```

### Step 4: Update Template Function Calls

For each file that uses template rendering:

```go
// Before: Conditional rendering
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    if UseNewTemplateEngine() {
        return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
    }
    return renderLegacyTemplate(fsys, templatePath, outputPath, data)
}

// After: Direct rendering
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
}
```

**Files to update:**
- `internal/template/engine.go` - Remove conditional logic
- `internal/gitops/copy.go` - Update template rendering calls
- Any other files identified in Step 1

```bash
# Commit template function updates
git commit -m "refactor: simplify template rendering to use new engine directly

- Remove conditional logic from template functions
- Update all callers to use new engine directly
- No functional changes, just code simplification

Refs: #<issue-number>"
```

### Step 5: Remove GitOps Legacy Code

```bash
# Remove legacy gitops files
git rm internal/gitops/legacy_compat.go
git rm internal/gitops/legacy_compat_test.go

# Archive backward compatibility tests
mkdir -p internal/gitops/archive
git mv internal/gitops/backward_compatibility_test.go internal/gitops/archive/
git mv internal/gitops/migration_test.go internal/gitops/archive/

# Commit gitops legacy removal
git commit -m "refactor: remove gitops legacy compatibility layer

- Remove legacy_compat.go and legacy_compat_test.go
- Archive backward compatibility tests for historical reference
- All GitOps generation now uses pipeline system directly

Refs: #<issue-number>"
```

### Step 6: Update GitOps Function Calls

For each file that uses GitOps generation:

```go
// Before: Conditional generation
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    if usePipelineGenerator() {
        generator := NewPipelineGenerator()
        return generator.Generate(ctx, cfg)
    }
    return generateGitOpsRepositoryLegacy(cfg)
}

// After: Direct generation
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    generator := NewPipelineGenerator()
    return generator.Generate(ctx, cfg)
}
```

**Files to update:**
- `internal/gitops/generator.go` - Remove conditional logic
- `cmd/cluster_render.go` - Update generation calls
- `cmd/cluster_bootstrap.go` - Update bootstrap flow
- Any other files identified in Step 1

```bash
# Commit gitops function updates
git commit -m "refactor: simplify gitops generation to use pipeline system directly

- Remove conditional logic from generation functions
- Update all callers to use pipeline system directly
- No functional changes, just code simplification

Refs: #<issue-number>"
```

### Step 7: Remove Configuration Builder Legacy Code

If separate legacy builder exists:

```bash
# Remove legacy builder code (if exists as separate file)
# git rm internal/config/legacy_builder.go

# Update builder.go to remove conditional logic
# Edit internal/config/builder.go
```

```go
// Before: Conditional builder
func NewConfigBuilder() ConfigBuilder {
    if UseNewConfigBuilder() {
        return NewFluentConfigBuilder()
    }
    return newLegacyConfigBuilder()
}

// After: Direct builder
func NewConfigBuilder() ConfigBuilder {
    return NewFluentConfigBuilder()
}
```

```bash
# Commit builder updates
git commit -m "refactor: simplify config builder to use fluent builder directly

- Remove conditional logic from builder creation
- Update all callers to use fluent builder directly
- No functional changes, just code simplification

Refs: #<issue-number>"
```

### Step 8: Remove Service Registry Legacy Code

If separate legacy registry exists:

```bash
# Remove legacy registry code (if exists as separate file)
# git rm internal/services/legacy_registry.go

# Update registry.go to remove conditional logic
# Edit internal/services/registry.go
```

```go
// Before: Conditional registry
func GetServiceRegistry() ServiceRegistry {
    if UseServiceRegistry() {
        return getNewServiceRegistry()
    }
    return getLegacyServiceRegistry()
}

// After: Direct registry
func GetServiceRegistry() ServiceRegistry {
    return getNewServiceRegistry()
}
```

```bash
# Commit registry updates
git commit -m "refactor: simplify service registry to use new registry directly

- Remove conditional logic from registry access
- Update all callers to use new registry directly
- No functional changes, just code simplification

Refs: #<issue-number>"
```

### Step 9: Remove Test Helpers

Remove test helpers that were only needed for feature flag testing:

```go
// REMOVE: Feature flag test helper
func clearFeatureFlagEnvVars(t *testing.T) {
    t.Setenv("OPENCENTER_USE_NEW_TEMPLATE_ENGINE", "")
    t.Setenv("OPENCENTER_USE_PIPELINE_GENERATOR", "")
    t.Setenv("OPENCENTER_USE_NEW_CONFIG_BUILDER", "")
    t.Setenv("OPENCENTER_USE_SERVICE_REGISTRY", "")
    t.Setenv("OPENCENTER_ENABLE_ALL_NEW_FEATURES", "")
}
```

**Files to update:**
- Any test files that use `clearFeatureFlagEnvVars()`
- Remove the helper function definition
- Remove calls to the helper function

```bash
# Commit test helper removal
git commit -m "refactor: remove feature flag test helpers

- Remove clearFeatureFlagEnvVars() helper
- Update tests to not use feature flag helpers
- Tests now test new systems directly

Refs: #<issue-number>"
```

### Step 10: Update Documentation

Update all documentation to remove legacy references:

```bash
# Update architecture documentation
# Edit docs/architecture.md - Remove legacy system diagrams

# Update developer documentation
# Edit docs/dev/configuration-system.md - Remove legacy system references

# Update migration guide
# Edit docs/migration/configuration-system-refactor.md - Mark as complete

# Update README
# Edit README.md - Remove feature flag references

# Archive migration guides
mkdir -p docs/migration/archive
git mv docs/migration/feature-flag-cleanup-guide.md docs/migration/archive/
git mv docs/migration/feature-flag-removal-timeline.md docs/migration/archive/
git mv docs/migration/legacy-code-removal-process.md docs/migration/archive/

# Commit documentation updates
git commit -m "docs: update documentation to reflect legacy code removal

- Remove legacy system references from architecture docs
- Update developer guides to show only new systems
- Archive migration guides for historical reference
- Update README to remove feature flag references

Refs: #<issue-number>"
```

### Step 11: Run Comprehensive Tests

```bash
# Run all tests
mise run test

# Run BDD tests
mise run godog

# Run benchmarks
go test -bench=. ./internal/benchmarks/

# Run linter
mise run lint

# Format code
mise run fmt

# Build for all platforms
mise run build-all
```

If any tests fail, fix them before proceeding.

### Step 12: Verify No Legacy References Remain

```bash
# Search for any remaining legacy references
grep -r "legacy\|Legacy" internal/ cmd/ --include="*.go" | grep -v "archive/" | grep -v "_test.go" | grep -v "// "

# Search for any remaining compatibility references
grep -r "compat\|Compat" internal/ cmd/ --include="*.go" | grep -v "archive/" | grep -v "_test.go" | grep -v "// "

# Search for any remaining feature flag references
grep -r "UseNew\|usePipeline\|FeatureFlag" internal/ cmd/ --include="*.go"

# If any references found, review and remove them
```

### Step 13: Create Pull Request

```bash
# Push branch
git push origin refactor/remove-legacy-code

# Create pull request with detailed description
```

**Pull Request Template:**

```markdown
## Legacy Code Removal

This PR removes all legacy compatibility code after successful migration to the new configuration system.

### Changes

- ✅ Removed template legacy compatibility layer
- ✅ Removed GitOps legacy compatibility layer
- ✅ Removed configuration builder legacy code
- ✅ Removed service registry legacy code
- ✅ Removed test helpers for feature flags
- ✅ Updated documentation to remove legacy references
- ✅ Archived migration guides for historical reference

### Testing

- ✅ All unit tests pass: `mise run test`
- ✅ All BDD tests pass: `mise run godog`
- ✅ All benchmarks pass: `go test -bench=. ./internal/benchmarks/`
- ✅ Linter passes: `mise run lint`
- ✅ Code formatted: `mise run fmt`
- ✅ Cross-platform builds work: `mise run build-all`

### Verification

- ✅ No legacy code references remain
- ✅ No feature flag references remain
- ✅ Documentation updated
- ✅ Migration guides archived

### Rollback Plan

If issues are discovered:
1. Revert this PR
2. Or checkout tag `pre-legacy-removal`
3. Investigate and fix issues
4. Retry removal

### Breaking Changes

None - all legacy code was already deprecated and unused.

Refs: #<issue-number>
```

## Validation Checklist

Use this checklist to ensure complete legacy code removal:

### Code Removal
- [ ] Template legacy files removed (`legacy.go`, `legacy_test.go`)
- [ ] GitOps legacy files removed (`legacy_compat.go`, `legacy_compat_test.go`)
- [ ] Configuration builder legacy code removed (if separate)
- [ ] Service registry legacy code removed (if separate)
- [ ] Feature flag files removed (should be done in Phase 4)
- [ ] Test helpers for feature flags removed

### Code Updates
- [ ] Template rendering functions simplified (no conditionals)
- [ ] GitOps generation functions simplified (no conditionals)
- [ ] Configuration builder functions simplified (no conditionals)
- [ ] Service registry functions simplified (no conditionals)
- [ ] All imports updated to remove legacy references

### Test Updates
- [ ] Migration tests archived (not deleted)
- [ ] Backward compatibility tests archived (not deleted)
- [ ] Feature flag tests removed
- [ ] All remaining tests pass
- [ ] No tests reference legacy code

### Documentation Updates
- [ ] Architecture documentation updated
- [ ] Developer documentation updated
- [ ] Migration guides archived
- [ ] README updated
- [ ] CLI help text updated (if needed)
- [ ] No documentation references legacy systems

### Verification
- [ ] No `legacy` or `Legacy` references in code (except archives)
- [ ] No `compat` or `Compat` references in code (except archives)
- [ ] No `UseNew*` or `usePipeline*` references in code
- [ ] No feature flag environment variable references
- [ ] All tests pass: `mise run test && mise run godog`
- [ ] All benchmarks pass: `go test -bench=. ./internal/benchmarks/`
- [ ] Linter passes: `mise run lint`
- [ ] Code formatted: `mise run fmt`
- [ ] Cross-platform builds work: `mise run build-all`

### Release Preparation
- [ ] Release notes updated with breaking changes (if any)
- [ ] Version number incremented appropriately
- [ ] Changelog updated
- [ ] Migration guide for users (if needed)
- [ ] Rollback plan documented

## Post-Removal Monitoring

After legacy code removal is deployed to production:

### Week 1: Intensive Monitoring
- Monitor error rates daily
- Check performance metrics daily
- Review user feedback daily
- Be ready for immediate rollback

### Week 2-4: Regular Monitoring
- Monitor error rates weekly
- Check performance metrics weekly
- Review user feedback weekly
- Document any issues

### Week 5+: Normal Monitoring
- Monitor error rates as normal
- Check performance metrics as normal
- Review user feedback as normal
- Consider removal successful if no issues

## Rollback Procedure

If critical issues are discovered after legacy code removal:

### Immediate Rollback (Critical Issues)

```bash
# Option 1: Revert the PR
git revert <pr-merge-commit>
mise run build
mise run test

# Option 2: Checkout pre-removal tag
git checkout pre-legacy-removal
mise run build
mise run test

# Deploy reverted version
mise run build-all
# Deploy to production
```

### Partial Rollback (Specific Component Issues)

If only one component has issues (e.g., template system):

```bash
# Restore specific files from pre-removal tag
git checkout pre-legacy-removal -- internal/template/legacy.go
git checkout pre-legacy-removal -- internal/template/legacy_test.go

# Restore conditional logic
# Edit internal/template/engine.go to add back conditionals

# Test and deploy
mise run test
mise run build
```

### Investigation and Fix

After rollback:

1. Investigate the root cause
2. Fix the issue in the new system
3. Test thoroughly
4. Retry legacy code removal

## Success Criteria

Legacy code removal is considered successful when:

1. ✅ All legacy files removed or archived
2. ✅ All conditional logic removed
3. ✅ All tests pass without legacy code
4. ✅ No compilation errors
5. ✅ Performance maintained or improved
6. ✅ Documentation complete and accurate
7. ✅ Production deployment successful
8. ✅ Zero critical issues for 30 days post-removal
9. ✅ Code quality metrics maintained or improved
10. ✅ No user complaints about missing functionality

## Timeline

**Estimated Duration:** 5-7 days

- Day 1: Identify all legacy code (Steps 1-2)
- Day 2: Remove template legacy code (Steps 3-4)
- Day 3: Remove GitOps legacy code (Steps 5-6)
- Day 4: Remove builder/registry legacy code (Steps 7-8)
- Day 5: Remove test helpers and update docs (Steps 9-10)
- Day 6: Testing and verification (Steps 11-12)
- Day 7: PR review and merge (Step 13)

## References

- **Feature Flag Cleanup Guide:** `docs/migration/feature-flag-cleanup-guide.md`
- **Feature Flag Removal Timeline:** `docs/migration/feature-flag-removal-timeline.md`
- **Design Document:** `.kiro/specs/configuration-system-refactor/design.md`
- **Tasks Document:** `.kiro/specs/configuration-system-refactor/tasks.md`
- **Requirements:** `.kiro/specs/configuration-system-refactor/requirements.md` (Requirement 10.3)

## Support

For questions or issues during legacy code removal:

1. Review this guide and related documentation
2. Check version control history for context
3. Consult with engineering lead
4. Document any unexpected issues for future reference

---

**Last Updated:** 2025-01-16
**Status:** Ready for execution after Phase 4 completion
**Validates:** Requirements 10.3 (Ensure no breaking changes during legacy code removal)
