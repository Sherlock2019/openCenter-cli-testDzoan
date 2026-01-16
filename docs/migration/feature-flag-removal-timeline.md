# Feature Flag Removal Timeline

## Overview

This document defines the timeline, success criteria, and process for removing feature flags introduced during the configuration system refactor. Feature flags enable gradual migration from legacy systems to new implementations while maintaining the ability to rollback if issues are discovered.

## Feature Flags

The following feature flags control the refactored systems:

1. **OPENCENTER_USE_NEW_TEMPLATE_ENGINE** - Enhanced template engine with caching
2. **OPENCENTER_USE_PIPELINE_GENERATOR** - Pipeline-based GitOps generation
3. **OPENCENTER_USE_NEW_CONFIG_BUILDER** - Type-safe fluent configuration builder
4. **OPENCENTER_USE_SERVICE_REGISTRY** - Plugin-based service registry
5. **OPENCENTER_ENABLE_ALL_NEW_FEATURES** - Enable all new systems at once
6. **OPENCENTER_FEATURE_FLAG_DEBUG** - Debug logging for feature flags

## Removal Timeline

### Phase 1: Validation Period (Weeks 1-4)

**Objective**: Validate new systems in production with feature flags enabled

**Activities**:
- Deploy with feature flags enabled in staging environments
- Monitor performance metrics and error rates
- Collect user feedback on new features
- Run comprehensive test suites with flags enabled
- Document any issues or regressions

**Success Criteria**:
- All integration tests pass with flags enabled
- Performance meets or exceeds legacy system benchmarks
- No critical bugs reported in new systems
- User feedback is positive or neutral
- Production deployments successful in staging

**Timeline**: 4 weeks from Phase 6 completion

### Phase 2: Default Transition (Weeks 5-8)

**Objective**: Make new systems the default while keeping legacy systems available

**Activities**:
- Change default behavior to use new systems (flags default to `true`)
- Add deprecation warnings when legacy systems are explicitly enabled
- Update documentation to recommend new systems
- Provide migration guides for users still on legacy systems
- Monitor production deployments closely

**Success Criteria**:
- New systems work as default without issues
- Deprecation warnings are clear and actionable
- Migration documentation is complete and accurate
- Users can still opt-in to legacy systems if needed
- No increase in support requests or bug reports

**Timeline**: 4 weeks after Phase 1 completion

**Implementation**:
```go
// internal/config/feature_flags.go
const (
    // Default to true after Phase 2
    defaultUseNewTemplateEngine = true
    defaultUsePipelineGenerator = true
    defaultUseNewConfigBuilder  = true
    defaultUseServiceRegistry   = true
)

func (ff *FeatureFlags) evaluateFlag(envVar string) bool {
    // Check specific flag first
    if value := os.Getenv(envVar); value != "" {
        return parseBoolEnv(envVar)
    }

    // Check global flag
    if ff.allNewFeaturesEnabled {
        return true
    }

    // Return new defaults (changed in Phase 2)
    switch envVar {
    case EnvUseNewTemplateEngine:
        return defaultUseNewTemplateEngine
    case EnvUsePipelineGenerator:
        return defaultUsePipelineGenerator
    case EnvUseNewConfigBuilder:
        return defaultUseNewConfigBuilder
    case EnvUseServiceRegistry:
        return defaultUseServiceRegistry
    default:
        return false
    }
}
```

### Phase 3: Deprecation Period (Weeks 9-16)

**Objective**: Deprecate legacy systems and prepare for removal

**Activities**:
- Add prominent deprecation warnings for legacy system usage
- Set removal date for legacy systems (8 weeks from Phase 3 start)
- Update all documentation to remove legacy system references
- Notify users of upcoming legacy system removal
- Provide support for users migrating from legacy systems

**Success Criteria**:
- Less than 5% of users still using legacy systems
- All known migration blockers resolved
- Deprecation warnings visible in logs and CLI output
- Documentation updated to reflect new systems only
- Support team prepared for legacy system removal

**Timeline**: 8 weeks after Phase 2 completion

**Deprecation Warning Implementation**:
```go
// internal/config/feature_flags.go
func (ff *FeatureFlags) logFlagEvaluation(envVar, featureName string, enabled bool) {
    // ... existing logging ...

    // Add deprecation warning if legacy system is explicitly enabled
    if !enabled && os.Getenv(envVar) != "" {
        ff.logger.WithFields(logrus.Fields{
            "component":    "feature_flags",
            "feature_name": featureName,
            "env_var":      envVar,
            "removal_date": "2025-06-01", // Update with actual date
        }).Warn("DEPRECATION: Legacy system explicitly enabled. This will be removed in a future release.")
        
        fmt.Fprintf(os.Stderr, "\n⚠️  DEPRECATION WARNING: %s is using the legacy system.\n", featureName)
        fmt.Fprintf(os.Stderr, "   The legacy system will be removed on 2025-06-01.\n")
        fmt.Fprintf(os.Stderr, "   Please test with the new system by removing %s or setting it to 'true'.\n", envVar)
        fmt.Fprintf(os.Stderr, "   See docs/migration/feature-flag-removal-timeline.md for details.\n\n")
    }
}
```

### Phase 4: Legacy System Removal (Week 17)

**Objective**: Remove legacy systems and feature flags from codebase

**Activities**:
- Remove legacy implementation code
- Remove feature flag checks and conditional logic
- Update tests to only test new systems
- Clean up compatibility layers
- Update version number to indicate breaking change

**Success Criteria**:
- All legacy code removed from codebase
- All feature flag checks removed
- All tests pass without feature flags
- Documentation updated to remove flag references
- Release notes clearly document breaking changes

**Timeline**: Week 17 (after 8-week deprecation period)

**Files to Remove**:
```
internal/template/legacy.go
internal/gitops/legacy_compat.go
internal/config/feature_flags.go (most of it)
```

**Code Changes**:
```go
// Before (with feature flags):
func RenderTemplate(ctx context.Context, template string, data interface{}) ([]byte, error) {
    if config.UseNewTemplateEngine() {
        return newEngine.Render(ctx, template, data)
    }
    return legacyRender(template, data)
}

// After (Phase 4):
func RenderTemplate(ctx context.Context, template string, data interface{}) ([]byte, error) {
    return engine.Render(ctx, template, data)
}
```

## Success Criteria by Feature Flag

### 1. OPENCENTER_USE_NEW_TEMPLATE_ENGINE

**Phase 1 Success Criteria**:
- [ ] Template rendering performance equal or better than legacy (measured via benchmarks)
- [ ] All existing templates render identically to legacy system (golden file tests)
- [ ] Template error messages are clear and actionable
- [ ] Template caching reduces rendering time by at least 20%
- [ ] No template rendering failures in production

**Phase 2 Success Criteria**:
- [ ] New template engine is default for 4 weeks without issues
- [ ] Less than 1% of users explicitly disable new engine
- [ ] Template rendering errors decrease or stay constant

**Phase 3 Success Criteria**:
- [ ] Deprecation warnings visible to users still on legacy system
- [ ] Migration guide helps users resolve any template compatibility issues
- [ ] Less than 5% of users still using legacy template engine

**Phase 4 Success Criteria**:
- [ ] Legacy template code removed without breaking changes
- [ ] All tests pass with only new template engine
- [ ] Documentation updated to remove legacy references

### 2. OPENCENTER_USE_PIPELINE_GENERATOR

**Phase 1 Success Criteria**:
- [ ] GitOps generation produces identical output to legacy system (diff tests)
- [ ] Pipeline generation completes in similar or better time than legacy
- [ ] Rollback functionality works correctly in failure scenarios
- [ ] Progress reporting provides useful feedback to users
- [ ] No GitOps generation failures in production

**Phase 2 Success Criteria**:
- [ ] New pipeline generator is default for 4 weeks without issues
- [ ] Less than 1% of users explicitly disable new generator
- [ ] GitOps generation errors decrease or stay constant

**Phase 3 Success Criteria**:
- [ ] Deprecation warnings visible to users still on legacy system
- [ ] Migration guide helps users resolve any generation issues
- [ ] Less than 5% of users still using legacy generator

**Phase 4 Success Criteria**:
- [ ] Legacy generator code removed without breaking changes
- [ ] All tests pass with only new pipeline generator
- [ ] Documentation updated to remove legacy references

### 3. OPENCENTER_USE_NEW_CONFIG_BUILDER

**Phase 1 Success Criteria**:
- [ ] Configuration building produces identical configs to legacy system
- [ ] Type-safe builder prevents invalid configurations at compile time
- [ ] Validation errors are clear and include suggestions
- [ ] Builder API is intuitive and well-documented
- [ ] No configuration building failures in production

**Phase 2 Success Criteria**:
- [ ] New config builder is default for 4 weeks without issues
- [ ] Less than 1% of users explicitly disable new builder
- [ ] Configuration validation errors decrease or stay constant

**Phase 3 Success Criteria**:
- [ ] Deprecation warnings visible to users still on legacy system
- [ ] Migration guide helps users resolve any builder issues
- [ ] Less than 5% of users still using legacy builder

**Phase 4 Success Criteria**:
- [ ] Legacy builder code removed without breaking changes
- [ ] All tests pass with only new config builder
- [ ] Documentation updated to remove legacy references

### 4. OPENCENTER_USE_SERVICE_REGISTRY

**Phase 1 Success Criteria**:
- [ ] Service registry manages all services correctly
- [ ] Dependency resolution works for all service combinations
- [ ] Plugin loading is reliable and performant
- [ ] Service status reporting is accurate
- [ ] No service management failures in production

**Phase 2 Success Criteria**:
- [ ] New service registry is default for 4 weeks without issues
- [ ] Less than 1% of users explicitly disable new registry
- [ ] Service-related errors decrease or stay constant

**Phase 3 Success Criteria**:
- [ ] Deprecation warnings visible to users still on legacy system
- [ ] Migration guide helps users resolve any service issues
- [ ] Less than 5% of users still using legacy service management

**Phase 4 Success Criteria**:
- [ ] Legacy service code removed without breaking changes
- [ ] All tests pass with only new service registry
- [ ] Documentation updated to remove legacy references

## Deprecation Warning Schedule

### Week 1-4 (Phase 1: Validation)
- No deprecation warnings
- Feature flags optional, defaults to legacy systems
- Focus on validation and testing

### Week 5-8 (Phase 2: Default Transition)
- Soft deprecation warnings in logs (INFO level)
- Feature flags optional, defaults to new systems
- Legacy systems still fully supported

### Week 9-12 (Phase 3: Deprecation Period - First Half)
- Prominent deprecation warnings in logs (WARN level)
- CLI output includes deprecation notices
- Documentation updated to show new systems only
- Legacy systems supported but discouraged

### Week 13-16 (Phase 3: Deprecation Period - Second Half)
- Loud deprecation warnings in logs (ERROR level)
- CLI output includes removal date
- Support team notifies users of upcoming removal
- Legacy systems supported but strongly discouraged

### Week 17+ (Phase 4: Removal)
- Legacy systems removed
- Feature flags removed
- Breaking change documented in release notes

## Rollback Plan

If critical issues are discovered during any phase:

### Immediate Rollback (Critical Issues)
1. Revert default behavior to legacy systems
2. Communicate issue to users via release notes
3. Investigate and fix the issue
4. Resume timeline after fix is validated

### Partial Rollback (Single Feature Issues)
1. Disable specific problematic feature flag
2. Keep other feature flags enabled
3. Investigate and fix the specific issue
4. Resume timeline for that feature after fix

### Timeline Extension
If issues require more time to resolve:
1. Extend current phase by 2-4 weeks
2. Communicate new timeline to users
3. Continue validation and testing
4. Resume timeline when ready

## Communication Plan

### Phase 1 (Validation)
- **Audience**: Internal team, early adopters
- **Channels**: Internal docs, team meetings
- **Message**: "New systems available for testing via feature flags"

### Phase 2 (Default Transition)
- **Audience**: All users
- **Channels**: Release notes, documentation, blog post
- **Message**: "New systems now default, legacy systems still available"

### Phase 3 (Deprecation)
- **Audience**: All users, especially those on legacy systems
- **Channels**: Release notes, CLI warnings, email notifications, blog post
- **Message**: "Legacy systems deprecated, will be removed on [DATE]"

### Phase 4 (Removal)
- **Audience**: All users
- **Channels**: Release notes, documentation, blog post, breaking change notice
- **Message**: "Legacy systems removed, new systems are now the only option"

## Monitoring and Metrics

### Key Metrics to Track

1. **Feature Flag Usage**
   - Percentage of users with each flag enabled
   - Percentage of users explicitly disabling flags
   - Trend over time

2. **Performance Metrics**
   - Template rendering time (new vs legacy)
   - GitOps generation time (new vs legacy)
   - Configuration building time (new vs legacy)
   - Memory usage comparison

3. **Error Rates**
   - Template rendering errors (new vs legacy)
   - GitOps generation errors (new vs legacy)
   - Configuration validation errors (new vs legacy)
   - Service management errors (new vs legacy)

4. **User Feedback**
   - Support tickets related to new systems
   - Bug reports for new systems
   - User satisfaction surveys
   - Community feedback

### Monitoring Implementation

```go
// internal/config/feature_flags.go
type FeatureFlagMetrics struct {
    FlagEvaluations map[string]int64
    FlagEnabled     map[string]int64
    FlagDisabled    map[string]int64
    LastUpdated     time.Time
}

func (ff *FeatureFlags) RecordMetrics() *FeatureFlagMetrics {
    ff.mu.RLock()
    defer ff.mu.RUnlock()
    
    metrics := &FeatureFlagMetrics{
        FlagEvaluations: make(map[string]int64),
        FlagEnabled:     make(map[string]int64),
        FlagDisabled:    make(map[string]int64),
        LastUpdated:     time.Now(),
    }
    
    // Record current flag states
    for flag, enabled := range ff.cache {
        if enabled {
            metrics.FlagEnabled[flag]++
        } else {
            metrics.FlagDisabled[flag]++
        }
    }
    
    return metrics
}
```

## Testing Requirements

### Phase 1 Testing
- [ ] All unit tests pass with flags enabled
- [ ] All integration tests pass with flags enabled
- [ ] All property-based tests pass with flags enabled
- [ ] Performance benchmarks meet or exceed legacy systems
- [ ] Golden file tests show identical output

### Phase 2 Testing
- [ ] All tests pass with new defaults
- [ ] Deprecation warnings appear correctly
- [ ] Legacy systems still work when explicitly enabled
- [ ] Migration documentation is accurate

### Phase 3 Testing
- [ ] Deprecation warnings are prominent and clear
- [ ] Legacy systems still work when explicitly enabled
- [ ] Migration guides help users successfully migrate

### Phase 4 Testing
- [ ] All tests pass without feature flags
- [ ] No references to legacy systems remain
- [ ] Documentation is accurate and complete
- [ ] Breaking changes are clearly documented

## Risk Mitigation

### High-Risk Scenarios

1. **Critical Bug in New System**
   - **Risk**: Production outage due to new system bug
   - **Mitigation**: Immediate rollback capability, comprehensive testing
   - **Response**: Disable flag, investigate, fix, re-enable

2. **Performance Regression**
   - **Risk**: New system slower than legacy
   - **Mitigation**: Continuous performance monitoring, benchmarks
   - **Response**: Optimize or rollback, extend timeline if needed

3. **User Adoption Issues**
   - **Risk**: Users unable or unwilling to migrate
   - **Mitigation**: Clear documentation, migration support
   - **Response**: Extend deprecation period, provide additional support

4. **Breaking Changes**
   - **Risk**: New system incompatible with user workflows
   - **Mitigation**: Comprehensive compatibility testing
   - **Response**: Fix compatibility issues or provide migration path

### Mitigation Strategies

1. **Comprehensive Testing**: All phases include extensive testing
2. **Gradual Rollout**: Feature flags enable gradual adoption
3. **Clear Communication**: Users informed at every phase
4. **Rollback Capability**: Can revert to legacy systems if needed
5. **Extended Timeline**: Can extend phases if issues arise
6. **User Support**: Migration guides and support available

## Target Dates

**Note**: These dates are tentative and will be updated based on Phase 6 completion date.

- **Phase 1 Start**: TBD (after Phase 6 completion)
- **Phase 1 End**: TBD + 4 weeks
- **Phase 2 Start**: TBD + 4 weeks
- **Phase 2 End**: TBD + 8 weeks
- **Phase 3 Start**: TBD + 8 weeks
- **Phase 3 End**: TBD + 16 weeks
- **Phase 4 Start**: TBD + 16 weeks
- **Phase 4 End**: TBD + 17 weeks

**Total Timeline**: 17 weeks from Phase 6 completion

## Automated Tests for Flag Removal

### Test Suite for Phase 4

```go
// internal/config/feature_flags_removal_test.go
package config_test

import (
    "testing"
    "go/ast"
    "go/parser"
    "go/token"
    "os"
    "path/filepath"
    "strings"
)

// TestNoLegacyCodeReferences ensures no legacy code remains after Phase 4
func TestNoLegacyCodeReferences(t *testing.T) {
    if os.Getenv("PHASE_4_REMOVAL_TESTS") != "true" {
        t.Skip("Phase 4 removal tests not enabled")
    }
    
    legacyPatterns := []string{
        "legacyRender",
        "legacyGenerate",
        "legacyBuild",
        "UseNewTemplateEngine",
        "UsePipelineGenerator",
        "UseNewConfigBuilder",
        "UseServiceRegistry",
    }
    
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !strings.HasSuffix(path, ".go") || strings.Contains(path, "_test.go") {
            return nil
        }
        
        content, err := os.ReadFile(path)
        if err != nil {
            return err
        }
        
        for _, pattern := range legacyPatterns {
            if strings.Contains(string(content), pattern) {
                t.Errorf("Found legacy code reference '%s' in %s", pattern, path)
            }
        }
        
        return nil
    })
    
    if err != nil {
        t.Fatalf("Error walking directory: %v", err)
    }
}

// TestNoFeatureFlagChecks ensures no feature flag checks remain after Phase 4
func TestNoFeatureFlagChecks(t *testing.T) {
    if os.Getenv("PHASE_4_REMOVAL_TESTS") != "true" {
        t.Skip("Phase 4 removal tests not enabled")
    }
    
    fset := token.NewFileSet()
    
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !strings.HasSuffix(path, ".go") || strings.Contains(path, "feature_flags") {
            return nil
        }
        
        node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
        if err != nil {
            return err
        }
        
        ast.Inspect(node, func(n ast.Node) bool {
            if call, ok := n.(*ast.CallExpr); ok {
                if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
                    if strings.HasPrefix(sel.Sel.Name, "Use") {
                        t.Errorf("Found feature flag check in %s: %s", path, sel.Sel.Name)
                    }
                }
            }
            return true
        })
        
        return nil
    })
    
    if err != nil {
        t.Fatalf("Error walking directory: %v", err)
    }
}

// TestNewSystemsWorkWithoutFlags ensures new systems work without feature flags
func TestNewSystemsWorkWithoutFlags(t *testing.T) {
    if os.Getenv("PHASE_4_REMOVAL_TESTS") != "true" {
        t.Skip("Phase 4 removal tests not enabled")
    }
    
    // Clear all feature flag environment variables
    os.Unsetenv("OPENCENTER_USE_NEW_TEMPLATE_ENGINE")
    os.Unsetenv("OPENCENTER_USE_PIPELINE_GENERATOR")
    os.Unsetenv("OPENCENTER_USE_NEW_CONFIG_BUILDER")
    os.Unsetenv("OPENCENTER_USE_SERVICE_REGISTRY")
    os.Unsetenv("OPENCENTER_ENABLE_ALL_NEW_FEATURES")
    
    // Run comprehensive integration tests
    // These should pass without any feature flags set
    t.Run("TemplateRendering", testTemplateRendering)
    t.Run("GitOpsGeneration", testGitOpsGeneration)
    t.Run("ConfigurationBuilding", testConfigurationBuilding)
    t.Run("ServiceManagement", testServiceManagement)
}
```

## Documentation Updates

### Files to Update in Phase 2
- [ ] `README.md` - Update to mention new systems as default
- [ ] `docs/architecture.md` - Update architecture diagrams
- [ ] `docs/migration/template-engine.md` - Update migration status
- [ ] `docs/reference/cli-commands.md` - Update command documentation
- [ ] All feature-specific documentation

### Files to Update in Phase 3
- [ ] All documentation - Remove legacy system references
- [ ] `CHANGELOG.md` - Document deprecation
- [ ] Migration guides - Add removal date

### Files to Update in Phase 4
- [ ] All documentation - Remove feature flag references
- [ ] `CHANGELOG.md` - Document breaking changes
- [ ] `README.md` - Update version requirements
- [ ] Release notes - Clearly document breaking changes

## Conclusion

This timeline provides a structured approach to removing feature flags after successful validation of the refactored systems. The 17-week timeline allows for thorough validation, gradual transition, and proper deprecation before final removal.

Key principles:
- **Safety First**: Multiple validation phases before removal
- **Clear Communication**: Users informed at every step
- **Rollback Capability**: Can revert if issues arise
- **Gradual Transition**: No sudden breaking changes
- **User Support**: Migration guides and support available

The timeline can be extended if issues arise, and individual feature flags can be handled independently if needed.
