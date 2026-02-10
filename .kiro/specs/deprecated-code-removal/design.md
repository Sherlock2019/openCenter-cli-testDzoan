# Design Document: Deprecated Code Removal

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
  - [Current State](#current-state)
  - [Target State](#target-state)
  - [Migration Strategy](#migration-strategy)
- [Components and Interfaces](#components-and-interfaces)
  - [Deprecated Functions to Remove](#deprecated-functions-to-remove)
  - [Modern API Replacements](#modern-api-replacements)
  - [Test Helper Infrastructure](#test-helper-infrastructure)
- [Data Models](#data-models)
  - [Function Dependency Graph](#function-dependency-graph)
  - [Test File Migration Patterns](#test-file-migration-patterns)
- [Part 3 Evaluation](#part-3-evaluation)
  - [validateServiceSecretsSimple Analysis](#validateservicesecretsimple-analysis)
  - [TemplateValidator Analysis](#templatevalidator-analysis)
- [Correctness Properties](#correctness-properties)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)

## Overview

This design completes the removal of deprecated code from opencenter-cli following Phase 1-4 modernization. Part 1 (unused deprecated code) is complete. This design focuses on:

1. **Part 2**: Removing ~300 lines of deprecated persistence functions from internal/config/persistence.go
2. **Part 3**: Evaluating validateServiceSecretsSimple and TemplateValidator for removal or deferral

The approach is surgical: remove deprecated functions, update test files to use modern APIs, verify all tests pass, and document decisions for Part 3 code.

## Architecture

### Current State

```
internal/config/persistence.go
├── Deprecated Functions (10 functions, ~300 lines)
│   ├── Save(cfg Config) error
│   ├── Load(name string) (Config, error)
│   ├── Validate(cfg Config) []error
│   ├── ConfigPath(name string) (string, error)
│   ├── GenerateCompleteConfig(name string) (Config, error)
│   ├── GenerateCompleteConfigYAML(name string) ([]byte, error)
│   ├── SaveDebugConfig(clusterName, gitDir string) error
│   ├── ListClusters() ([]string, error) - unused
│   ├── SetActiveCluster(name string) error - unused
│   └── GetActiveCluster() (string, error) - unused
│
├── Helper Functions (only used by deprecated functions)
│   ├── mergeYAMLMaps(base, override map[string]any) map[string]any
│   ├── cleanEmptyValues(m map[string]any)
│   ├── isEmpty(v any) bool
│   └── getConfigPathForSave(cfg Config) (string, error)
│
└── Modern Functions (keep these)
    ├── List() ([]string, error)
    ├── SetActive(name string) error
    ├── GetActive() (string, error)
    └── Other non-deprecated functions

Test Files Using Deprecated Functions:
└── internal/config/migration/scanner_test.go
    └── Contains example code showing deprecated patterns
```

### Target State

```
internal/config/persistence.go
├── Modern Functions Only
│   ├── List() ([]string, error)
│   ├── SetActive(name string) error
│   ├── GetActive() (string, error)
│   └── Other non-deprecated functions
│
└── No deprecated functions or helpers

Test Files:
├── internal/config/migration/scanner_test.go
│   └── Updated examples showing modern API patterns
│
└── All tests use ConfigurationManager via test helpers
```

### Migration Strategy

**Phase 1: Analyze Dependencies**
1. Identify all functions that call deprecated functions
2. Identify helper functions only used by deprecated functions
3. Map test files that need updates

**Phase 2: Update Test Files**
1. Update scanner_test.go example code to show modern patterns
2. Verify scanner tests still detect deprecated patterns correctly
3. Ensure no other test files use deprecated functions

**Phase 3: Remove Deprecated Functions**
1. Remove the 10 deprecated functions
2. Remove helper functions (mergeYAMLMaps, cleanEmptyValues, isEmpty, getConfigPathForSave)
3. Remove deprecation warning calls

**Phase 4: Verification**
1. Run `mise run build` - verify compilation
2. Run `mise run test` - verify unit tests pass
3. Run `mise run godog` - verify BDD tests pass
4. Test CLI commands manually

**Phase 5: Part 3 Evaluation**
1. Analyze validateServiceSecretsSimple usage and complexity
2. Analyze TemplateValidator usage and impact
3. Document decisions and rationale
4. Update status document

## Components and Interfaces

### Deprecated Functions to Remove

**Category 1: Core Persistence Functions (3 functions)**

```go
// REMOVE: Replaced by ConfigurationManager.Save()
func Save(cfg Config) error

// REMOVE: Replaced by ConfigurationManager.Load()
func Load(name string) (Config, error)

// REMOVE: Replaced by ConfigurationManager.Validate()
func Validate(cfg Config) []error
```

**Category 2: Path Resolution Function (1 function)**

```go
// REMOVE: Replaced by PathResolver.ResolveClusterPaths()
func ConfigPath(name string) (string, error)
```

**Category 3: Config Generation Functions (2 functions)**

```go
// REMOVE: Replaced by ConfigurationManager.Load() with merge options
func GenerateCompleteConfig(name string) (Config, error)

// REMOVE: Replaced by manual implementation in tests
func GenerateCompleteConfigYAML(name string) ([]byte, error)
```

**Category 4: Debug Function (1 function)**

```go
// REMOVE: Replaced by manual implementation in tests
func SaveDebugConfig(clusterName, gitDir string) error
```

**Category 5: Unused Functions (3 functions)**

```go
// REMOVE: Never called, can remove immediately
func ListClusters() ([]string, error)

// REMOVE: Never called, can remove immediately
func SetActiveCluster(name string) error

// REMOVE: Never called, can remove immediately
func GetActiveCluster() (string, error)
```

### Modern API Replacements

**ConfigurationManager Methods (already exist)**

```go
// Modern replacement for Save()
func (cm *ConfigurationManager) Save(ctx context.Context, config *Config) error

// Modern replacement for Save() without validation (for tests)
func (cm *ConfigurationManager) SaveWithoutValidation(ctx context.Context, config *Config) error

// Modern replacement for Load()
func (cm *ConfigurationManager) Load(ctx context.Context, name string) (*Config, error)

// Modern replacement for Load() without validation (for tests)
func (cm *ConfigurationManager) LoadWithoutValidation(ctx context.Context, name string) (*Config, error)

// Modern replacement for Validate()
func (cm *ConfigurationManager) Validate(ctx context.Context, config *Config) error

// Modern replacement for ListClusters()
func (cm *ConfigurationManager) List(ctx context.Context) ([]string, error)
```

**PathResolver Methods (already exist)**

```go
// Modern replacement for ConfigPath()
func (pr *PathResolver) ResolveClusterPaths(ctx context.Context, name, organization string) ClusterPaths
```

### Test Helper Infrastructure

**Existing Test Helpers (internal/testing/config_helpers.go)**

```go
// Save configuration without validation (for test fixtures)
func SaveConfig(t *testing.T, cfg config.Config)

// Save configuration with custom PathResolver (for temp directories)
func SaveConfigWithPathResolver(t *testing.T, cfg config.Config, pathResolver *paths.PathResolver)

// Save configuration with validation (for testing validation)
func SaveConfigWithValidation(t *testing.T, cfg config.Config) error

// Load configuration
func LoadConfig(t *testing.T, name string) config.Config

// Validate configuration
func ValidateConfig(t *testing.T, cfg config.Config) error
```

## Data Models

### Function Dependency Graph

```
Deprecated Functions → Helper Functions → Modern APIs

Save() ──────────────┐
                     ├──→ getConfigPathForSave() ──→ PathResolver
                     └──→ cleanEmptyValues() ──→ isEmpty()

Load() ──────────────┐
                     └──→ ConfigPath() ──→ PathResolver

GenerateCompleteConfig() ──→ mergeYAMLMaps()
                         └──→ Load()

GenerateCompleteConfigYAML() ──→ GenerateCompleteConfig()

SaveDebugConfig() ──→ Save()
                  └──→ ConfigPath()

Validate() ──────────→ ValidationEngine (via ConfigurationManager)

ListClusters() ──────→ UNUSED (can remove immediately)
SetActiveCluster() ──→ UNUSED (can remove immediately)
GetActiveCluster() ──→ UNUSED (can remove immediately)
```

### Test File Migration Patterns

**Pattern 1: scanner_test.go Example Code**

```go
// BEFORE (in test data strings)
func loadConfig() {
    cfg, err := config.Load("cluster-name")
    if err != nil {
        return
    }
}

// AFTER (in test data strings)
func loadConfig() {
    ctx := context.Background()
    manager, err := config.NewConfigurationManager()
    if err != nil {
        return
    }
    cfg, err := manager.Load(ctx, "cluster-name")
    if err != nil {
        return
    }
}
```

**Pattern 2: scanner.go Migration Instructions**

```go
// BEFORE (in generated report)
sb.WriteString("config, err := config.Load(clusterName)\n\n")
sb.WriteString("// After\n")
sb.WriteString("config, err := manager.Load(ctx, clusterName)\n")

// AFTER (updated to show full context)
sb.WriteString("// Before\n")
sb.WriteString("config, err := config.Load(clusterName)\n\n")
sb.WriteString("// After\n")
sb.WriteString("ctx := context.Background()\n")
sb.WriteString("manager, err := config.NewConfigurationManager()\n")
sb.WriteString("if err != nil { return err }\n")
sb.WriteString("config, err := manager.Load(ctx, clusterName)\n")
```

## Part 3 Evaluation

### validateServiceSecretsSimple Analysis

**Current Status:**
- Location: `internal/config/config.go`
- Lines: ~100 lines
- Marked as deprecated with comment
- **Usage: NONE** - Function is defined but never called

**Analysis:**
1. **Scope**: Internal function, not exported
2. **Callers**: Zero callers found in codebase
3. **Complexity**: Medium - validates service-specific secrets with fallback logic
4. **Migration Effort**: Would require creating ValidationEngine rules for each service

**Decision: REMOVE IMMEDIATELY**

**Rationale:**
- Function is never called anywhere in the codebase
- No migration needed since nothing uses it
- Removing dead code improves maintainability
- If similar validation is needed in future, implement in ValidationEngine

**Action:**
- Remove validateServiceSecretsSimple function
- Remove deprecation comment
- No test updates needed (function not tested)

### TemplateValidator Analysis

**Current Status:**
- Location: `internal/util/template/interfaces.go`
- Type: Interface combining BasicTemplateValidator, TemplateDataValidator, AdvancedTemplateValidator
- Marked as deprecated with comment
- **Usage: EXTENSIVE** - Used throughout template engine

**Analysis:**
1. **Scope**: Exported interface, part of public API
2. **Callers**: 
   - DefaultTemplateEngine embeds TemplateValidator
   - NewTemplateEngineWithDependencies accepts TemplateValidator parameter
   - GetValidator() returns TemplateValidator
   - Multiple test files use TemplateValidator
3. **Complexity**: High - removing requires refactoring entire template engine
4. **Migration Effort**: Large - would require:
   - Updating DefaultTemplateEngine to use specific validator interfaces
   - Changing all dependency injection to use specific interfaces
   - Updating all tests
   - Ensuring backward compatibility or coordinating breaking change

**Decision: DEFER TO SEPARATE TASK**

**Rationale:**
- Removing TemplateValidator requires broader template engine refactoring
- Current scope is focused on config persistence functions
- Template engine refactoring should be its own spec with:
  - Comprehensive design for interface decomposition
  - Migration plan for all consumers
  - Backward compatibility strategy
  - Separate testing and validation

**Recommendation:**
- Create follow-up task: "Template Engine Interface Refactoring"
- Scope: Replace TemplateValidator with specific validator interfaces
- Priority: Low (technical debt, not blocking functionality)
- Estimated effort: 2-3 days

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

This feature is primarily about code removal and structural refactoring. Most requirements are not testable as properties because they describe code structure rather than runtime behavior. However, we can test that the refactoring maintains system correctness.

### Testable Acceptance Criteria

From the prework analysis, the following criteria are testable as examples:

- 3.1: All unit tests pass after removal
- 3.2: All integration tests pass after removal
- 3.3: Build succeeds with no compilation errors
- 3.4: `mise run test` reports zero failures
- 3.5: `mise run godog` reports zero BDD failures
- 4.5: Scanner tests continue to validate deprecated pattern detection
- 8.1: No "Deprecated:" occurrences in persistence.go
- 8.2: No logDeprecationWarning calls for removed functions
- 10.1-10.5: CLI commands continue to work correctly

These are all example-based tests (specific scenarios) rather than properties (universal rules). This is appropriate for a refactoring task where we're verifying that existing behavior is preserved.

## Error Handling

### Compilation Errors

**Scenario**: Removing deprecated functions causes compilation errors

**Handling**:
1. Before removal, verify no production code calls deprecated functions
2. Use grep/ripgrep to search for function calls
3. If calls found, update to modern APIs first
4. Only remove functions after zero callers confirmed

### Test Failures

**Scenario**: Tests fail after deprecated function removal

**Handling**:
1. Identify which test failed
2. Check if test was using deprecated function indirectly
3. Update test to use ConfigurationManager or test helpers
4. Re-run tests to verify fix
5. If persistent failures, investigate root cause before proceeding

### Scanner Test Failures

**Scenario**: Migration scanner tests fail after updating example code

**Handling**:
1. Verify scanner still detects deprecated patterns in old code
2. Ensure test data strings are properly formatted
3. Check that scanner regex patterns still match
4. Update scanner logic if needed to handle modern patterns

### CLI Command Failures

**Scenario**: CLI commands fail after deprecated function removal

**Handling**:
1. This should not happen if all cmd/* files already migrated
2. If it does happen, it indicates missed migration
3. Identify which cmd file is affected
4. Update to use ConfigurationManager
5. Re-test CLI command

## Testing Strategy

### Dual Testing Approach

This feature uses both unit tests and manual verification:

- **Unit tests**: Verify that existing tests continue to pass after refactoring
- **Manual verification**: Verify that deprecated code is removed and CLI commands work
- Together: Comprehensive coverage (unit tests catch regressions, manual verification confirms cleanup)

### Unit Testing

**Test Categories:**

1. **Existing Test Suite** (primary validation)
   - Run `mise run test` to verify all unit tests pass
   - Run `mise run godog` to verify all BDD tests pass
   - These tests validate that ConfigurationManager works correctly
   - No new tests needed - existing tests provide coverage

2. **Scanner Tests** (verify scanner still works)
   - Run scanner tests to verify deprecated pattern detection
   - Ensure scanner can still identify old code patterns
   - Verify updated example code is syntactically correct

3. **Build Verification** (verify compilation)
   - Run `mise run build` to verify code compiles
   - Check for any compilation errors
   - Verify no undefined function references

### Manual Verification

**Verification Steps:**

1. **Code Inspection**
   - Search for "Deprecated:" in internal/config/persistence.go (should find 0)
   - Search for "logDeprecationWarning" related to removed functions (should find 0)
   - Verify deprecated functions are removed
   - Verify helper functions are removed

2. **CLI Command Testing**
   - Run `opencenter cluster init test-cluster` (should succeed)
   - Run `opencenter cluster validate test-cluster` (should succeed)
   - Run `opencenter cluster info test-cluster` (should succeed)
   - Run `opencenter cluster list` (should succeed)
   - Verify all commands work as expected

3. **Documentation Review**
   - Check that status document is updated
   - Verify migration guides don't reference removed functions
   - Ensure Part 3 decisions are documented

### Test Execution Order

1. **Before Removal**: Run full test suite to establish baseline
2. **After Test Updates**: Run scanner tests to verify example code
3. **After Function Removal**: Run full test suite to verify no regressions
4. **After Build**: Test CLI commands manually
5. **Final Verification**: Run all tests one more time

### Success Criteria

- ✅ `mise run build` succeeds with no errors
- ✅ `mise run test` reports 0 failures
- ✅ `mise run godog` reports 0 failures
- ✅ All CLI commands work correctly
- ✅ No "Deprecated:" comments in persistence.go
- ✅ No logDeprecationWarning calls for removed functions
- ✅ Status document updated with Part 3 decisions
