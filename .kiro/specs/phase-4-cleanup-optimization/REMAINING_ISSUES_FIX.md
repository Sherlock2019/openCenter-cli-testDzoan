# Remaining Issues Fix - Phase 4

## Summary

This document addresses the three remaining high-priority issues identified in Phase 4:
1. Config API Test Failures (compilation errors)
2. Config/v2 Property Test Failures (test logic issues)
3. Template Registration Issues (missing template name)

## Issue 1: Config API Test Failures

### Problem Description

Multiple tests in `internal/config/config_test.go` fail to compile with undefined function errors:
- `undefined: Save`
- `undefined: Load`
- `undefined: Validate`

### Root Cause

During Phase 4 refactoring, the public API functions `Save()`, `Load()`, and `Validate()` were removed or renamed, but the tests still reference them. The actual implementations exist as:
- `saveConfig()` (private function in persistence.go)
- No direct `Load()` function (replaced by ConfigurationManager)
- `Validate()` moved to validation package

### Solution

**Option 1: Update tests to use ConfigurationManager (Recommended)**

The tests should use the new ConfigurationManager API instead of deprecated functions:

```go
// Before (BROKEN)
cfg := NewDefault("test")
if err := Save(cfg); err != nil {
    t.Fatal(err)
}
loaded, err := Load("test")

// After (FIXED)
manager := NewConfigurationManager(pathResolver, loader, cache, validator)
if err := manager.SaveConfiguration(ctx, cfg); err != nil {
    t.Fatal(err)
}
loaded, err := manager.LoadConfiguration(ctx, "test")
```

**Option 2: Add compatibility shims (Quick fix)**

Add temporary wrapper functions for backward compatibility:

```go
// In persistence.go
func Save(cfg Config) error {
    return saveConfig(cfg, false)
}

func Load(name string) (Config, error) {
    // Implementation using ConfigurationManager
    manager := getDefaultManager()
    return manager.LoadConfiguration(context.Background(), name)
}
```

### Files to Modify

1. **internal/config/config_test.go**
   - Update all test functions to use ConfigurationManager
   - Remove references to deprecated Save/Load/Validate functions
   - Add proper setup/teardown for ConfigurationManager

2. **internal/config/persistence.go** (if using Option 2)
   - Add compatibility wrapper functions
   - Mark as deprecated with clear migration path

### Acceptance Criteria

- [ ] All tests in `internal/config/config_test.go` compile successfully
- [ ] All tests pass with `go test ./internal/config -v`
- [ ] No undefined function errors
- [ ] Tests use modern ConfigurationManager API

---

## Issue 2: Config/v2 Property Test Failures

### Problem Description

Property tests in `internal/config/v2` have logic issues:
- Some tests "gave up" after insufficient passed tests (e.g., 98/100)
- Tests are discarding too many generated inputs (492 discarded)
- This indicates generator constraints are too restrictive

### Root Cause

The property test generators are creating inputs that don't satisfy the test preconditions, causing the test framework to discard them. When too many inputs are discarded, the test gives up.

Example from output:
```
! Kamaji requires kube_vip_enabled to be false: Gave up after only 98 passed tests. 492 tests were discarded.
```

This means the generator is creating configs where:
- Kamaji is enabled
- But kube_vip_enabled is also true (invalid combination)
- The test discards these invalid combinations
- Eventually runs out of valid inputs

### Solution

**Fix Generator Constraints**

Update the property test generators to produce valid inputs more frequently:

```go
// Before (TOO RESTRICTIVE)
gen.Struct(reflect.TypeOf(Config{}), map[string]gopter.Gen{
    "Deployment": genDeployment(),
    "Infrastructure": genInfrastructure(),
})

// After (SMARTER GENERATION)
gen.Struct(reflect.TypeOf(Config{}), map[string]gopter.Gen{
    "Deployment": genDeployment(),
    "Infrastructure": genInfrastructureForDeployment(), // Conditional generation
}).SuchThat(func(cfg Config) bool {
    // Only generate valid combinations
    if cfg.Deployment.Method == "kamaji" {
        return !cfg.Infrastructure.Networking.KubeVIPEnabled
    }
    return true
})
```

**Increase Test Attempts**

Alternatively, increase the number of generation attempts:

```go
parameters := gopter.DefaultTestParameters()
parameters.MinSuccessfulTests = 100
parameters.MaxDiscardRatio = 10 // Allow more discards (default is 5)
```

### Files to Modify

1. **internal/config/v2/loader_property_test.go**
   - Update `TestProperty_KamajiDeploymentConstraints`
   - Fix generator to produce valid Kamaji configurations
   - Reduce discard ratio by improving preconditions

2. **internal/testing/generators.go** (if generators are shared)
   - Add conditional generation helpers
   - Create deployment-aware infrastructure generators

### Acceptance Criteria

- [ ] All property tests pass with 100/100 successful tests
- [ ] Discard ratio < 5:1 (fewer than 500 discarded for 100 passed)
- [ ] Tests complete in reasonable time (< 5 seconds per property)
- [ ] No "gave up" messages in test output

---

## Issue 3: Template Registration Issues

### Problem Description

Test `TestRegisterRealGitOpsTemplates` fails with:
```
template main.tf not found
expected template main.tf should be registered
```

### Root Cause

The test expects a template named `main.tf` but the actual file is named `main.tf.tmpl` (with `.tmpl` extension). The template registration code strips the `.tpl` extension but not `.tmpl`.

### Solution

**Option 1: Fix Template Name Generation (Recommended)**

Update the template name generation logic to handle both `.tpl` and `.tmpl` extensions:

```go
// In internal/template/embedded_registry.go or similar

func generateTemplateName(path string) string {
    name := filepath.Base(path)
    
    // Strip .tpl extension
    name = strings.TrimSuffix(name, ".tpl")
    
    // Strip .tmpl extension
    name = strings.TrimSuffix(name, ".tmpl")
    
    return name
}
```

**Option 2: Rename Template File**

Rename `internal/provision/templates/main.tf.tmpl` to `main.tf.tpl` for consistency:

```bash
mv internal/provision/templates/main.tf.tmpl internal/provision/templates/main.tf.tpl
```

**Option 3: Update Test Expectations**

Change the test to expect `main.tf.tmpl` instead of `main.tf`:

```go
expectedTemplates := []string{
    "main.tf.tmpl",  // Changed from "main.tf"
    "variables.tf",
    "Makefile",
}
```

### Files to Modify

1. **internal/template/embedded_registry.go** (Option 1)
   - Update `generateTemplateName()` function
   - Add test for `.tmpl` extension handling

2. **internal/provision/templates/main.tf.tmpl** (Option 2)
   - Rename to `main.tf.tpl`
   - Update any references in code

3. **internal/template/embedded_integration_test.go** (Option 3)
   - Update expected template names
   - Document why `.tmpl` extension is used

### Acceptance Criteria

- [ ] Test `TestRegisterRealGitOpsTemplates` passes
- [ ] Template `main.tf` (or `main.tf.tmpl`) is registered successfully
- [ ] All other templates register correctly
- [ ] No template name conflicts

---

## Implementation Plan

### Phase 1: Config API Test Failures (Priority: HIGH)

**Estimated Time**: 2-3 hours

1. Analyze current ConfigurationManager API
2. Update `config_test.go` to use ConfigurationManager
3. Remove deprecated function calls
4. Run tests and verify compilation
5. Fix any remaining test failures

### Phase 2: Template Registration (Priority: HIGH)

**Estimated Time**: 1-2 hours

1. Investigate template name generation logic
2. Implement fix (Option 1 recommended)
3. Update tests
4. Verify all templates register correctly
5. Run full template test suite

### Phase 3: Config/v2 Property Tests (Priority: MEDIUM)

**Estimated Time**: 3-4 hours

1. Analyze failing property tests
2. Identify generator constraint issues
3. Update generators to produce valid inputs
4. Adjust discard ratios if needed
5. Run property tests multiple times to verify stability

---

## Testing Strategy

### Unit Tests

```bash
# Test config package
go test ./internal/config -v

# Test config/v2 package
go test ./internal/config/v2/... -v

# Test template package
go test ./internal/template/... -v
```

### Property Tests

```bash
# Run property tests with verbose output
go test ./internal/config/v2/... -v -run Property

# Run with increased test count for confidence
go test ./internal/config/v2/... -v -run Property -count=5
```

### Integration Tests

```bash
# Run full test suite
mise run test

# Check for compilation errors
go build ./...
```

---

## Success Criteria

### All Issues Resolved When:

- [ ] All tests in `internal/config` compile and pass
- [ ] All property tests in `internal/config/v2` pass with 100/100 tests
- [ ] Template registration test passes
- [ ] No undefined function errors
- [ ] No "gave up" messages in property tests
- [ ] Full test suite passes: `mise run test`

### Code Quality:

- [ ] No deprecated function usage
- [ ] Modern ConfigurationManager API used throughout
- [ ] Property test generators produce valid inputs efficiently
- [ ] Template naming is consistent and documented

---

## Risk Assessment

### Low Risk
- Template registration fix (isolated change)
- Adding compatibility shims (backward compatible)

### Medium Risk
- Updating config tests to use ConfigurationManager (requires understanding new API)
- Property test generator fixes (may need multiple iterations)

### Mitigation
- Test each fix independently
- Run full test suite after each change
- Keep deprecated functions temporarily if needed
- Document migration path for other developers

---

## Related Documents

- [STATUS_UPDATE.md](./STATUS_UPDATE.md) - Overall Phase 4 status
- [BACKUP_RESTORE_FIX.md](./BACKUP_RESTORE_FIX.md) - Similar fix pattern
- [GITOPS_TEMPLATE_FIX.md](./GITOPS_TEMPLATE_FIX.md) - Template syntax fixes

---

**Created**: 2026-02-04
**Status**: Ready for Implementation
**Priority**: HIGH (blocks Phase 4 completion)
