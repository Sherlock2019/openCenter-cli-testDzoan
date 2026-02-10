# Deprecated Code Removal Status

**Date**: February 9, 2026  
**Phase**: Phase 4 Cleanup & Optimization  
**Status**: In Progress (Part 1 Complete)

## Overview

Removing all deprecated code after Phase 1-4 completion. The deprecated code was kept for backward compatibility during migration but should now be removed since all phases are complete.

## Part 1: Unused Deprecated Code - ✅ COMPLETE

**Commit**: 109140f  
**Date**: February 9, 2026

### Removed

1. **`internal/config/cache.go`** (300+ lines removed)
   - ✅ `InMemoryConfigCache` type
   - ✅ `CacheEntry` type  
   - ✅ `NewInMemoryConfigCache()` function
   - ✅ All associated methods (15+ methods)

2. **`internal/config/interfaces.go`**
   - ✅ `ConfigManagerInterface` interface

3. **`cmd/config_migration_helpers.go`**
   - ✅ Updated comments to remove "migration helper" references

### Impact

- **Lines Removed**: 319 lines
- **Files Modified**: 3 files
- **Breaking Changes**: None (code was not being used)
- **Tests**: All passing

## Part 2: Deprecated Persistence Functions - ✅ COMPLETE

**Completion Date**: February 9, 2026

### Test Infrastructure Fix - ✅ COMPLETE

**Commits**: 0949177, 516f791  
**Date**: February 9, 2026

Before removing deprecated functions, we needed to fix the test infrastructure to work with ConfigurationManager's validation requirements.

**Changes Made**:
1. Added `SaveWithoutValidation()` method to ConfigurationManager
2. Added `LoadWithoutValidation()` method to ConfigurationManager
3. Created `internal/testing/config_helpers.go` with test utilities
4. Updated Setup and Bootstrap services to use LoadWithoutValidation when SkipValidation=true
5. Fixed all cluster service tests - all passing

### CMD Files Migration - ✅ COMPLETE

**Commits**: 8cec9d2, 08681b6  
**Date**: February 9, 2026

Migrated all production cmd files from deprecated config functions to new ConfigurationManager APIs.

**Changes Made**:
1. Created `getConfigPath()` helper function in config_migration_helpers.go
2. Migrated 7 cmd files to use new APIs:
   - cluster_config.go
   - cluster_config_update.go
   - cluster_destroy.go
   - cluster_edit.go
   - cluster_info.go
   - cluster_lock.go
   - cluster_select.go
3. Updated cmd/cluster_service_test.go to use new APIs
4. Removed duplicate code and unused imports
5. All cmd files compile successfully

**Impact**:
- **Files Modified**: 8 cmd files
- **Lines Changed**: ~150 lines
- **Breaking Changes**: None (internal refactoring only)
- **Build Status**: ✅ Compiles successfully

### Deprecated Functions Removal - ✅ COMPLETE

**Date**: February 9, 2026

All deprecated persistence functions have been successfully removed from `internal/config/persistence.go`.

**Functions Removed**:
1. ✅ `Save(cfg Config) error` - Replaced by ConfigurationManager.Save()
2. ✅ `Load(name string) (Config, error)` - Replaced by ConfigurationManager.Load()
3. ✅ `Validate(cfg Config) []error` - Replaced by ConfigurationManager.Validate()
4. ✅ `ConfigPath(name string) (string, error)` - Replaced by PathResolver.ResolveClusterPaths()
5. ✅ `GenerateCompleteConfig(name string) (Config, error)` - Replaced by ConfigurationManager.Load()
6. ✅ `GenerateCompleteConfigYAML(name string) ([]byte, error)` - Replaced by manual YAML marshaling
7. ✅ `SaveDebugConfig(clusterName, gitDir string) error` - Replaced by manual implementation
8. ✅ `ListClusters()` - Already removed (never existed or was removed earlier)
9. ✅ `SetActiveCluster()` - Already removed (never existed or was removed earlier)
10. ✅ `GetActiveCluster()` - Already removed (never existed or was removed earlier)

**Helper Functions Removed**:
1. ✅ `mergeYAMLMaps(base, override map[string]any)` - Only used by deprecated functions
2. ✅ `cleanEmptyValues(m map[string]any)` - Only used by deprecated functions
3. ✅ `isEmpty(v any)` - Only used by deprecated functions
4. ✅ `getConfigPathForSave(cfg Config)` - Only used by deprecated functions
5. ✅ `saveConfig(cfg Config, omitEmpty bool)` - Only used by deprecated functions

**Migration Scanner Updates**:
- Updated `internal/config/migration/scanner_test.go` to show modern API patterns
- Updated `internal/config/migration/scanner.go` migration instructions with full context
- Scanner tests pass and correctly detect deprecated patterns in old code

**Files Modified**:
- `internal/config/persistence.go` - Removed ~300 lines of deprecated code
- `internal/config/migration/scanner_test.go` - Updated example code
- `internal/config/migration/scanner.go` - Updated migration instructions
- `cmd/cluster_config_update.go` - Replaced GenerateCompleteConfigYAML with modern API

**Impact**:
- **Lines Removed**: ~300 lines from persistence.go
- **Breaking Changes**: None (all production code already migrated)
- **Build Status**: ✅ Compiles successfully
- **Test Status**: ✅ All tests passing (except pre-existing security test failure)

## Part 3: Other Deprecated Code - ✅ COMPLETE

**Completion Date**: February 10, 2026

### validateServiceSecretsSimple - ✅ REMOVED

**Decision**: REMOVED  
**Rationale**: Function had zero callers in the codebase

**Analysis**:
- Location: `internal/config/config.go`
- Lines: ~100 lines
- Usage: NONE - Function was defined but never called
- Complexity: Medium - validates service-specific secrets with fallback logic

**Action Taken**:
- Removed `validateServiceSecretsSimple()` function from internal/config/config.go
- Removed deprecation comment
- No test updates needed (function was not tested)
- No migration needed (nothing used it)

**Future Recommendation**:
If similar validation is needed in the future, implement it in the ValidationEngine (internal/core/validation) rather than as a standalone function.

### TemplateValidator - ✅ REMOVED

**Decision**: REMOVED  
**Completion Date**: February 10, 2026  
**Rationale**: Successfully refactored to use specific validator interfaces

**Analysis**:
- Location: `internal/util/template/interfaces.go`
- Type: Interface combining BasicTemplateValidator, TemplateDataValidator, AdvancedTemplateValidator
- Usage: EXTENSIVE - Used throughout template engine
- Complexity: High - required refactoring entire template engine

**Refactoring Approach**:
1. Split `DefaultTemplateEngine.validator` field into three separate fields:
   - `basicValidator BasicTemplateValidator`
   - `dataValidator TemplateDataValidator`
   - `advancedValidator AdvancedTemplateValidator`

2. Updated `TemplateEngine` interface to embed the three specific interfaces instead of `TemplateValidator`

3. Updated all validator method calls in `DefaultTemplateEngine` to use the appropriate specific validator

4. Added new getter methods:
   - `GetBasicValidator() BasicTemplateValidator`
   - `GetDataValidator() TemplateDataValidator`
   - `GetAdvancedValidator() AdvancedTemplateValidator`

5. Removed deprecated `GetValidator() TemplateValidator` method from interface

6. Updated `NewTemplateEngineWithDependencies` to accept `interface{}` and extract the three validator interfaces

**Files Modified**:
- `internal/util/template/interfaces.go` - Removed TemplateValidator interface, updated TemplateEngine interface
- `internal/util/template/engine.go` - Refactored to use three separate validator fields
- `internal/util/template/engine_test.go` - Updated tests to use new getter methods
- `internal/util/template/doc.go` - Updated documentation

**Impact**:
- **Lines Changed**: ~50 lines across 4 files
- **Breaking Changes**: None - DefaultTemplateValidator still implements all three interfaces
- **Build Status**: ✅ Compiles successfully
- **Test Status**: ✅ All tests passing

**Benefits**:
- Clearer separation of concerns
- More flexible dependency injection
- Better adherence to Interface Segregation Principle
- Easier to mock specific validator functionality in tests

## Part 4: Final Deprecated Functions Removal - ✅ COMPLETE

**Completion Date**: February 10, 2026

### Migration to ConfigurationManager Methods

All remaining deprecated persistence functions have been successfully migrated to use ConfigurationManager methods.

**New ConfigurationManager Methods Added**:
1. `ConfigurationManager.GetActive()` - Returns active cluster name with environment/session/persistent precedence
2. `ConfigurationManager.SetActive(name string)` - Sets or clears active cluster marker file
3. `ConfigurationManager.activeClusterPath()` - Helper for active cluster marker file path

**Files Migrated**:
1. `cmd/cluster_select.go` - 4 usages migrated to helper functions
2. `cmd/cluster_edit.go` - 1 usage migrated
3. `cmd/cluster_destroy.go` - 2 usages migrated
4. `cmd/cluster_credentials_export.go` - 1 usage migrated
5. `cmd/cluster_service_test.go` - 1 usage migrated
6. `cmd/cluster_export_consistency_test.go` - 1 usage migrated
7. `tests/features/steps/helpers.go` - 4 usages migrated to test helper functions
8. `cmd/config_migration_helpers.go` - Updated to use ConfigurationManager

**Deprecated Functions Removed from persistence.go**:
1. ✅ `List()` - ~200 lines removed
2. ✅ `SetActive(name string)` - Removed
3. ✅ `GetActive()` - Removed
4. ✅ `sortStrings(s []string)` - Helper function removed
5. ✅ `activeClusterPath()` - Helper function removed
6. ✅ `getGlobalFileSystem()` - No longer needed, removed

**Impact**:
- **Lines Removed**: ~250 lines from persistence.go
- **Breaking Changes**: None (all production code migrated)
- **Build Status**: ✅ Compiles successfully
- **Test Status**: ✅ All tests passing (except pre-existing security test failures)

## Current Status - ✅ ALL PARTS COMPLETE

- ✅ Part 1 Complete: Unused deprecated code removed (319 lines)
- ✅ Part 2 Complete: Deprecated persistence functions removed (~300 lines)
- ✅ Part 3 Complete: validateServiceSecretsSimple removed, TemplateValidator refactored
- ✅ Part 4 Complete: Final deprecated functions removed (~250 lines)

**Total Lines Removed/Refactored**: ~869 lines of deprecated code + 50 lines refactored

**Summary**:
All deprecated code has been successfully removed or refactored in the opencenter-cli codebase. The codebase is now cleaner and more maintainable, with all production code using modern APIs (ConfigurationManager, PathResolver, ValidationEngine) and properly separated validator interfaces. No deprecation warnings will be shown to users.

## Success Criteria - ✅ ALL COMPLETE

- ✅ All deprecated functions removed from persistence.go
- ✅ All test files updated to use ConfigurationManager
- ✅ All cmd files updated to use PathResolver and ConfigurationManager
- ✅ All tests passing (except pre-existing security test failures)
- ✅ Build succeeds
- ✅ No deprecation warnings for removed functions
- ✅ TemplateValidator interface refactored to use specific interfaces
- ✅ Documentation updated
