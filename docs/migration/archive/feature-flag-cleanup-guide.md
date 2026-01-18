# Feature Flag Cleanup Guide

## Overview

This guide provides detailed instructions for removing feature flags after successful migration to the new configuration system. The cleanup process should only begin after all new systems have been validated in production for at least 30 days with zero critical issues.

## Prerequisites

Before starting the cleanup process, ensure:

1. ✅ All feature flags have been enabled in production for minimum 30 days
2. ✅ Zero critical issues reported with new systems
3. ✅ Performance metrics are within acceptable range (see `docs/dev/performance-characteristics.md`)
4. ✅ All automated tests pass with flags enabled (`mise run test && mise run godog`)
5. ✅ Stakeholder approval obtained (product owner, engineering lead)
6. ✅ Users notified of upcoming changes

## Cleanup Phases

### Phase 1: Validation (Current State)

**Status:** ✅ COMPLETE

**Objective:** Validate all new systems work correctly with feature flags enabled.

**Actions Completed:**
- All feature flags enabled via `OPENCENTER_ENABLE_ALL_NEW_FEATURES=true`
- Comprehensive test suite validates functionality
- Performance benchmarks confirm no regressions
- Production monitoring shows stable operation

**Validation:**
```bash
# Run automated tests for flag removal
mise run test -run TestFeatureFlagRemoval

# Verify all tests pass
mise run test && mise run godog

# Run performance benchmarks
go test -bench=. ./internal/benchmarks/
```

### Phase 2: Default Change

**Status:** 🔜 NOT STARTED

**Objective:** Change feature flag defaults to enable new systems by default.

**Actions Required:**

1. **Update feature flag defaults in `internal/config/feature_flags.go`:**

```go
// Change evaluateFlag to default to true instead of false
func (ff *FeatureFlags) evaluateFlag(envVar string) bool {
    // Check specific flag first
    if value := os.Getenv(envVar); value != "" {
        return parseBoolEnv(envVar)
    }

    // NEW: Default to true (new systems enabled)
    // OLD: return ff.allNewFeaturesEnabled
    return true
}
```

2. **Update documentation to reflect new defaults:**
   - Update `internal/config/feature_flags.go` comments
   - Update `docs/migration/configuration-system-refactor.md`
   - Update CLI help text in `cmd/config_features.go`

3. **Add deprecation warnings for legacy system usage:**

```go
func (ff *FeatureFlags) logFlagEvaluation(envVar, featureName string, enabled bool) {
    // ... existing logging ...
    
    // NEW: Warn when legacy system is explicitly enabled
    if !enabled && os.Getenv(envVar) == "false" {
        ff.logger.WithFields(logrus.Fields{
            "component":    "feature_flags",
            "feature_name": featureName,
            "env_var":      envVar,
        }).Warn("Legacy system explicitly enabled - this will be removed in future release")
    }
}
```

4. **Deploy and monitor:**
   - Deploy to staging environment first
   - Monitor for 1 week
   - Deploy to production
   - Monitor for 2 weeks
   - Keep rollback capability via environment variables

**Rollback Procedure:**
If issues are discovered, users can disable specific features:
```bash
export OPENCENTER_USE_NEW_TEMPLATE_ENGINE=false
export OPENCENTER_USE_PIPELINE_GENERATOR=false
```

### Phase 3: Deprecation Warnings

**Status:** 🔜 NOT STARTED

**Objective:** Notify users that legacy systems will be removed.

**Actions Required:**

1. **Add deprecation notices to CLI output:**

```go
// In cmd/config_features.go
func printFeatureFlagStatus() {
    status := config.GetFeatureFlags().GetStatus()
    
    fmt.Println("\n=== Feature Flag Status ===")
    for name, enabled := range status {
        statusStr := "enabled"
        if !enabled {
            statusStr = "disabled (DEPRECATED - will be removed)"
        }
        fmt.Printf("%s: %s\n", name, statusStr)
    }
    fmt.Println("===========================\n")
}
```

2. **Update documentation:**
   - Add deprecation notice to README.md
   - Update migration guide with removal timeline
   - Create announcement for users

3. **Monitor usage:**
   - Track how many users are still using legacy systems
   - Reach out to users who need migration assistance

**Timeline:** 30 days minimum before Phase 4

### Phase 4: Code Removal

**Status:** 🔜 NOT STARTED

**Objective:** Remove feature flags and legacy code completely.

**Actions Required:**

#### 4.1 Remove Feature Flag System

**Files to Delete:**
```bash
# Feature flag implementation
rm internal/config/feature_flags.go
rm internal/config/feature_flags_test.go
rm internal/config/feature_flags_logging_test.go
rm internal/config/feature_flags_example_test.go
rm internal/config/feature_flags_removal_test.go

# This cleanup guide (archive instead of delete)
mv docs/migration/feature-flag-cleanup-guide.md docs/migration/archive/
```

**Code to Remove from Remaining Files:**
- Remove all `if UseNewTemplateEngine()` conditionals
- Remove all `if UsePipelineGenerator()` conditionals
- Remove all `if UseNewConfigBuilder()` conditionals
- Remove all `if UseServiceRegistry()` conditionals
- Remove imports of `config.UseNew*` functions

#### 4.2 Remove Template Legacy Code

**Files to Delete:**
```bash
# Template compatibility layer
rm internal/template/legacy.go
rm internal/template/legacy_test.go
rm internal/template/migration_test.go
rm internal/template/migration_path_validation_test.go
```

**Code Changes:**
```go
// OLD: internal/template/legacy.go
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    if UseNewTemplateEngine() {
        return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
    }
    return renderTemplateLegacy(fsys, templatePath, outputPath, data)
}

// NEW: Direct call to new engine
func RenderTemplateToFile(fsys fs.FS, templatePath, outputPath string, data interface{}) error {
    return RenderWithEngine(defaultEngine, fsys, templatePath, outputPath, data)
}
```

#### 4.3 Remove GitOps Legacy Code

**Files to Delete:**
```bash
# GitOps compatibility layer
rm internal/gitops/legacy_compat.go
rm internal/gitops/legacy_compat_test.go
rm internal/gitops/backward_compatibility_test.go
rm internal/gitops/migration_test.go
```

**Code Changes:**
```go
// OLD: internal/gitops/legacy_compat.go
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    if usePipelineGenerator() {
        return generator.Generate(ctx, cfg)
    }
    return generateGitOpsRepositoryLegacy(cfg)
}

// NEW: Direct call to pipeline generator
func GenerateGitOpsRepository(ctx context.Context, cfg config.Config) error {
    generator := NewPipelineGenerator()
    return generator.Generate(ctx, cfg)
}
```

#### 4.4 Remove CLI Feature Flag Commands

**Files to Modify:**
```bash
# Remove feature flag management commands
# Edit cmd/config_features.go to remove feature flag status display
```

**Code Changes:**
- Remove `opencenter config features` command (or repurpose for other features)
- Remove feature flag status from `opencenter version` output
- Update help text to remove feature flag references

#### 4.5 Update Documentation

**Files to Update:**
```bash
# Update architecture documentation
docs/architecture.md
docs/dev/configuration-system.md

# Update migration guide
docs/migration/configuration-system-refactor.md

# Update README
README.md
```

**Changes:**
- Remove all references to feature flags
- Update architecture diagrams to show only new systems
- Archive migration guides to `docs/migration/archive/`
- Update getting started guide

#### 4.6 Update Tests

**Test Files to Update:**
```bash
# Remove feature flag test helpers
# Update tests that use clearFeatureFlagEnvVars()
# Remove tests that validate feature flag behavior
```

**Changes:**
- Remove `clearFeatureFlagEnvVars()` helper function
- Update tests to call new systems directly
- Remove tests that validate feature flag switching

## Cleanup Checklist

Use this checklist to track cleanup progress:

### Pre-Cleanup Validation
- [ ] All feature flags enabled in production for 30+ days
- [ ] Zero critical issues reported
- [ ] Performance metrics acceptable
- [ ] All tests pass with flags enabled
- [ ] Stakeholder approval obtained
- [ ] Users notified of changes
- [ ] Rollback plan documented
- [ ] Release tagged before cleanup (for easy rollback)

### Phase 2: Default Change
- [ ] Update `evaluateFlag()` to default to true
- [ ] Add deprecation warnings for legacy usage
- [ ] Update documentation
- [ ] Deploy to staging
- [ ] Monitor staging for 1 week
- [ ] Deploy to production
- [ ] Monitor production for 2 weeks

### Phase 3: Deprecation Warnings
- [ ] Add CLI deprecation notices
- [ ] Update documentation with removal timeline
- [ ] Create user announcement
- [ ] Monitor legacy system usage
- [ ] Wait 30 days minimum

### Phase 4: Code Removal

#### Constants and Types
- [ ] Remove `EnvUseNewTemplateEngine` constant
- [ ] Remove `EnvUsePipelineGenerator` constant
- [ ] Remove `EnvUseNewConfigBuilder` constant
- [ ] Remove `EnvUseServiceRegistry` constant
- [ ] Remove `EnvEnableAllNewFeatures` constant
- [ ] Remove `EnvFeatureFlagDebug` constant
- [ ] Remove `FeatureFlags` type
- [ ] Remove `globalFeatureFlags` variable
- [ ] Remove `once` sync.Once variable
- [ ] Remove `MigrationGuide` constant

#### Methods and Functions
- [ ] Remove `UseNewTemplateEngine()` method
- [ ] Remove `UsePipelineGenerator()` method
- [ ] Remove `UseNewConfigBuilder()` method
- [ ] Remove `UseServiceRegistry()` method
- [ ] Remove `isEnabled()` method
- [ ] Remove `evaluateFlag()` method
- [ ] Remove `ClearCache()` method
- [ ] Remove `GetStatus()` method
- [ ] Remove `PrintStatus()` method
- [ ] Remove `logInitialization()` method
- [ ] Remove `logFlagEvaluation()` method
- [ ] Remove `parseBoolEnv()` function
- [ ] Remove `GetFeatureFlags()` function
- [ ] Remove package-level convenience functions

#### Files to Delete
- [ ] Delete `internal/config/feature_flags.go`
- [ ] Delete `internal/config/feature_flags_test.go`
- [ ] Delete `internal/config/feature_flags_logging_test.go`
- [ ] Delete `internal/config/feature_flags_example_test.go`
- [ ] Delete `internal/config/feature_flags_removal_test.go`
- [ ] Delete `internal/template/legacy.go`
- [ ] Delete `internal/template/legacy_test.go`
- [ ] Delete `internal/template/migration_test.go`
- [ ] Delete `internal/template/migration_path_validation_test.go`
- [ ] Delete `internal/gitops/legacy_compat.go`
- [ ] Delete `internal/gitops/legacy_compat_test.go`
- [ ] Delete `internal/gitops/backward_compatibility_test.go`
- [ ] Delete `internal/gitops/migration_test.go`

#### Code References to Remove
- [ ] Remove all `if UseNewTemplateEngine()` conditionals
- [ ] Remove all `if UsePipelineGenerator()` conditionals
- [ ] Remove all `if UseNewConfigBuilder()` conditionals
- [ ] Remove all `if UseServiceRegistry()` conditionals
- [ ] Remove all `if usePipelineGenerator()` conditionals
- [ ] Remove imports of `config.UseNew*` functions
- [ ] Update function calls to use new systems directly

#### Documentation Updates
- [ ] Update `docs/architecture.md`
- [ ] Update `docs/dev/configuration-system.md`
- [ ] Update `docs/migration/configuration-system-refactor.md`
- [ ] Update `README.md`
- [ ] Archive migration guides to `docs/migration/archive/`
- [ ] Update CLI help text
- [ ] Update getting started guide

#### Test Updates
- [ ] Remove `clearFeatureFlagEnvVars()` helper
- [ ] Update tests to call new systems directly
- [ ] Remove feature flag switching tests
- [ ] Verify all tests pass after cleanup

### Post-Cleanup Validation
- [ ] All tests pass: `mise run test && mise run godog`
- [ ] No compilation errors: `mise run build`
- [ ] Cross-platform builds work: `mise run build-all`
- [ ] Benchmarks pass: `go test -bench=. ./internal/benchmarks/`
- [ ] Linter passes: `mise run lint`
- [ ] Code formatted: `mise run fmt`
- [ ] Documentation complete and accurate
- [ ] No broken links in documentation
- [ ] Architecture diagrams updated
- [ ] Release notes updated

## Automated Validation

Run the automated cleanup validation tests:

```bash
# Run feature flag removal tests
go test -v ./internal/config -run TestFeatureFlagRemoval

# Verify all systems work with flags enabled
export OPENCENTER_ENABLE_ALL_NEW_FEATURES=true
mise run test && mise run godog

# Run performance benchmarks
go test -bench=. ./internal/benchmarks/

# Search for remaining legacy code references
grep -r "UseNewTemplateEngine\|UsePipelineGenerator\|UseNewConfigBuilder\|UseServiceRegistry" internal/
grep -r "legacy\|Legacy\|compat\|Compat" internal/ | grep -v "_test.go" | grep -v "// "
```

## Rollback Procedure

If issues are discovered after cleanup:

1. **Immediate Rollback:**
   ```bash
   git revert <cleanup-commit-hash>
   mise run build
   mise run test
   ```

2. **Restore from Tagged Release:**
   ```bash
   git checkout <pre-cleanup-tag>
   mise run build
   ```

3. **Selective Rollback:**
   - Restore specific files from version control
   - Re-enable feature flags temporarily
   - Fix issues and retry cleanup

## Success Criteria

The cleanup is considered successful when:

1. ✅ All tests pass without feature flags
2. ✅ No compilation errors
3. ✅ Performance maintained or improved
4. ✅ Documentation complete and accurate
5. ✅ No legacy code references remain
6. ✅ Code quality metrics maintained
7. ✅ Production deployment successful
8. ✅ Zero critical issues for 30 days post-cleanup

## Timeline

**Estimated Total Duration:** 90-120 days

- Phase 1 (Validation): ✅ Complete
- Phase 2 (Default Change): 14-21 days
  - Staging deployment: 7 days
  - Production deployment: 7-14 days
- Phase 3 (Deprecation): 30 days minimum
- Phase 4 (Code Removal): 7-14 days
  - Code cleanup: 3-5 days
  - Testing and validation: 2-3 days
  - Documentation updates: 2-3 days
  - Production deployment: 1-3 days

## References

- **Feature Flag Implementation:** `internal/config/feature_flags.go`
- **Automated Tests:** `internal/config/feature_flags_removal_test.go`
- **Removal Timeline:** `docs/migration/feature-flag-removal-timeline.md`
- **Migration Guide:** `docs/migration/configuration-system-refactor.md`
- **Performance Baseline:** `docs/dev/performance-characteristics.md`
- **Design Document:** `.kiro/specs/configuration-system-refactor/design.md`
- **Tasks Document:** `.kiro/specs/configuration-system-refactor/tasks.md`

## Support

For questions or issues during cleanup:

1. Review this guide and automated tests
2. Check version control history for context
3. Consult with engineering lead
4. Document any unexpected issues for future reference

---

**Last Updated:** 2025-01-16
**Status:** Phase 1 Complete, Ready for Phase 2
