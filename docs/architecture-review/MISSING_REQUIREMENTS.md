# Missing Requirements - Prioritized List

**Project**: opencenter-cli  
**Generated**: February 6, 2026  
**Source**: IMPLEMENTATION_STATUS.md  
**Overall Completion**: 84% (31/37 requirements fully implemented)

## Table of Contents

- [Executive Summary](#executive-summary)
- [Priority 1: Critical Issues](#priority-1-critical-issues)
- [Priority 2: High Impact](#priority-2-high-impact)
- [Priority 3: Medium Impact](#priority-3-medium-impact)
- [Priority 4: Low Impact / Nice to Have](#priority-4-low-impact--nice-to-have)
- [Metrics Summary](#metrics-summary)
- [Recommendations](#recommendations)

## Executive Summary

This document provides a prioritized list of missing requirements across all four phases of the architectural refactoring roadmap. Requirements are prioritized based on:

1. **Impact on functionality**: Does it block core features?
2. **Impact on quality**: Does it affect reliability, security, or maintainability?
3. **Dependencies**: Does it block other work?
4. **Effort required**: How much work is needed to complete?

**Key Findings**:
- 6 requirements have gaps or are incomplete
- Most critical: Test failures in Phase 1 error handling
- Most impactful: File operations migration (66 direct os calls remain)
- Most documentation gaps: Phase 3 migration guides

**Overall Status**:
- Phase 1: 71% complete (5/7 fully implemented)
- Phase 2: 100% complete (11/11 fully implemented) ✅
- Phase 3: 92% complete (11/12 fully implemented)
- Phase 4: 86% complete (6/7 fully implemented)

---

## Priority 1: Critical Issues

These issues affect core functionality, block other work, or represent significant quality/security concerns.

### P1.1: Fix Error Handling Test Failures (Phase 1)

**Requirement**: Phase 1, Requirement 2 - Structured Error Handling  
**Status**: ⚠️ Completed with test failures  
**Impact**: HIGH - Blocks verification of error handling correctness

**Issue**:
- Test compilation errors in `internal/util/errors/errors_test.go`
- `NewDefaultErrorHandler` signature changed, now requires `CredentialMasker` parameter
- Cannot measure test coverage due to compilation errors
- Error handling is functional but unverified

**Evidence**:
```
File: internal/util/errors/errors_test.go
Error: NewDefaultErrorHandler signature mismatch
Expected: NewDefaultErrorHandler(masker CredentialMasker)
Found: NewDefaultErrorHandler() (in tests)
```

**Action Items**:
1. Update all test calls to `NewDefaultErrorHandler` to pass `CredentialMasker`
2. Run tests to verify error handling behavior
3. Measure test coverage (target: >80%)
4. Fix any additional test failures

**Estimated Effort**: 2-4 hours  
**Dependencies**: None  
**Blocking**: Phase 1 completion verification

---

### P1.2: Increase FileSystem Test Coverage (Phase 1)

**Requirement**: Phase 1, Requirement 6 - Code Quality and Testing  
**Status**: ⚠️ Partially completed  
**Impact**: HIGH - Core infrastructure needs high reliability

**Issue**:
- Current coverage: 77.4%
- Target coverage: >95%
- Missing tests for error paths and edge cases
- No performance benchmarks to verify <5% overhead target

**Evidence**:
```
Component: internal/util/fs/wrapper.go
Current Coverage: 77.4%
Target Coverage: >95%
Gap: 17.6 percentage points
```

**Action Items**:
1. Identify untested code paths in FileSystem wrapper
2. Add tests for error conditions:
   - File not found
   - Permission denied
   - Disk full scenarios
   - Concurrent access patterns
3. Add edge case tests:
   - Empty files
   - Large files
   - Special characters in paths
4. Add benchmark tests to verify <5% performance overhead
5. Run coverage analysis to confirm >95% target met

**Estimated Effort**: 4-6 hours  
**Dependencies**: None  
**Blocking**: Phase 1 completion

---

## Priority 2: High Impact

These issues significantly affect code quality, maintainability, or developer experience.

### P2.1: Complete File Operations Migration (Phase 4)

**Requirement**: Phase 4, Requirement 4 - File Operations Migration  
**Status**: ⚠️ Partially completed  
**Impact**: HIGH - Affects consistency, error handling, and atomic operations

**Issue**:
- 66 direct `os.ReadFile`/`os.WriteFile` calls remain in `internal/`
- Inconsistent error handling across codebase
- Missing atomic write guarantees in some locations
- Bypasses FileSystem wrapper benefits

**Evidence**:
```
Direct os calls remaining: 66
Packages affected:
- internal/talos/generator/gitops_structure.go (4 calls)
- internal/cluster/init_service.go (4 calls)
- internal/cluster/bootstrap_service.go (2 calls)
- internal/cluster/validate_service.go (1 call)
- internal/core/validation/validators/gitops.go (1 call)
- internal/util/crypto/key_manager.go (2 calls)
- internal/util/security/credential_validator.go (2 calls)
- internal/util/files/file_operator.go (2 calls)
- And 50+ more...
```

**Action Items**:
1. Scan codebase for all remaining `os.ReadFile` and `os.WriteFile` calls
2. Prioritize migration by package:
   - Cluster services (init, bootstrap, validate) - 7 calls
   - Talos generator - 4 calls
   - Validation validators - 1 call
   - Utility packages - remaining calls
3. Update each call to use `FileSystem` interface
4. Add proper error wrapping with context
5. Verify atomic write behavior where needed
6. Update tests to verify FileSystem usage

**Estimated Effort**: 12-16 hours (1-2 days)  
**Dependencies**: None  
**Blocking**: Phase 4 completion

---

### P2.2: Create Phase 3 Migration Guide (Phase 3)

**Requirement**: Phase 3, Requirement 8 - Direct Migration Strategy  
**Status**: ⚠️ Partially completed  
**Impact**: HIGH - Blocks developer adoption of ConfigurationManager

**Issue**:
- No migration guide with before/after code examples
- No migration checklist for 45+ files using legacy patterns
- No deprecation warnings in legacy code
- Developers don't know how to migrate to new API

**Evidence**:
```
Missing Documentation:
- docs/migration/config-manager-migration.md (not found)
- Migration checklist (not found)
- Deprecation warnings in legacy functions (not implemented)

Migration tooling exists:
- cmd/migration-scanner/main.go ✅
- internal/config/migration/scanner.go ✅
```

**Action Items**:
1. Create migration guide: `docs/migration/config-manager-migration.md`
   - Before/after code examples
   - Common migration patterns
   - Error handling changes
   - Testing strategies
2. Create migration checklist
   - List of 45+ files to migrate
   - Priority order
   - Verification steps
3. Add deprecation warnings to legacy functions
   - `config.Load()` → suggest `ConfigurationManager.Load()`
   - `config.Save()` → suggest `ConfigurationManager.Save()`
   - Include migration guide link
4. Update architecture documentation

**Estimated Effort**: 6-8 hours  
**Dependencies**: None  
**Blocking**: Phase 3 completion, developer adoption

---

### P2.3: Improve Services Package Test Coverage (Phase 4)

**Requirement**: Phase 4, Requirement 7 - Testing Requirements  
**Status**: ⚠️ Partially completed  
**Impact**: HIGH - Service plugins are core functionality

**Issue**:
- Current coverage: 7.9%
- Target coverage: >85%
- Gap: 77.1 percentage points
- Individual plugin test coverage varies widely

**Evidence**:
```
Package: internal/services
Current Coverage: 7.9%
Target Coverage: >85%
Gap: 77.1 percentage points

Components:
- BaseServicePlugin: ✅ Comprehensive tests
- ServiceRegistry: ✅ Integration tests
- Individual plugins: ⚠️ Varies (some have tests, many don't)
```

**Action Items**:
1. Audit existing plugin tests
2. Create test template for service plugins
3. Add tests for each of 14 migrated plugins:
   - Metadata validation
   - Validate() method behavior
   - Render() method behavior
   - Status() method behavior
   - Error handling
4. Add integration tests for plugin lifecycle
5. Run coverage analysis to verify >85% target

**Estimated Effort**: 10-12 hours  
**Dependencies**: None  
**Blocking**: Phase 4 completion

---

## Priority 3: Medium Impact

These issues affect documentation, metrics, or optional features.

### P3.1: Create ADR for Orphaned Code Removal (Phase 1)

**Requirement**: Phase 1, Requirement 3 - Orphaned Code Removal  
**Status**: ✅ Completed (missing ADR)  
**Impact**: MEDIUM - Documentation gap, doesn't affect functionality

**Issue**:
- `internal/core/config/` directory successfully removed
- No Architecture Decision Record (ADR) documenting the decision
- No architecture documentation explaining the rationale
- Future developers may not understand why code was removed

**Evidence**:
```
Removed: internal/core/config/ directory ✅
ADR: Not found ⚠️
Architecture docs: Not updated ⚠️
```

**Action Items**:
1. Create ADR: `docs/architecture/decisions/001-remove-orphaned-config-code.md`
   - Context: Why the code existed
   - Decision: Why it was removed
   - Consequences: Impact on codebase
   - Alternatives considered
2. Update architecture documentation
   - Reference ADR
   - Explain new structure
3. Add to CHANGELOG

**Estimated Effort**: 2-3 hours  
**Dependencies**: None  
**Blocking**: Phase 1 documentation completion

---

### P3.2: Calculate and Document Comprehensive Metrics (Phase 4)

**Requirement**: Phase 4, Requirement 6 - Code Quality Metrics  
**Status**: ⚠️ Partially completed  
**Impact**: MEDIUM - Metrics help track progress and justify refactoring

**Issue**:
- Plugin boilerplate reduction calculated: ~700-1,120 lines ✅
- Total LOC reduction not calculated
- Code duplication percentage not measured
- Build time comparison not performed
- Comprehensive metrics report not generated

**Evidence**:
```
Calculated Metrics:
- Plugin boilerplate: ~700-1,120 lines removed ✅

Missing Metrics:
- Total LOC before/after ⚠️
- Code duplication % ⚠️
- Build time comparison ⚠️
- Test coverage trends ⚠️
```

**Action Items**:
1. Calculate total LOC reduction
   - Current LOC: 126,593 lines of Go code (measured February 6, 2026)
   - Need baseline from before Phase 1 started
   - Use `cloc internal/` on historical commit
   - Compare baseline to current
   - Break down by phase if possible
2. Measure code duplication
   - Use `jscpd` or similar tool
   - Target: <5% duplication
3. Measure build time
   - Baseline vs current
   - Target: <45 seconds
4. Generate comprehensive metrics report
   - Add to IMPLEMENTATION_STATUS.md
   - Create visualizations if helpful
5. Document methodology for future measurements

**Estimated Effort**: 4-6 hours  
**Dependencies**: None  
**Blocking**: Phase 4 metrics verification

---

### P3.3: Add FileSystem Performance Benchmarks (Phase 1)

**Requirement**: Phase 1, Requirement 6 - Code Quality and Testing  
**Status**: ⚠️ Partially completed  
**Impact**: MEDIUM - Verifies performance target, doesn't affect functionality

**Issue**:
- Acceptance criteria specifies <5% performance overhead
- No benchmark tests exist to verify this target
- Cannot prove FileSystem wrapper meets performance requirements

**Evidence**:
```
Target: <5% performance overhead
Benchmarks: Not found ⚠️
Verification: Cannot measure ⚠️
```

**Action Items**:
1. Create benchmark test file: `internal/util/fs/wrapper_benchmark_test.go`
2. Add benchmarks:
   - `BenchmarkFileSystem_ReadFile` vs `os.ReadFile`
   - `BenchmarkFileSystem_WriteFile` vs `os.WriteFile`
   - `BenchmarkFileSystem_WriteFileAtomic` vs `os.WriteFile`
   - `BenchmarkFileSystem_Exists` vs `os.Stat`
3. Run benchmarks and calculate overhead percentage
4. Verify <5% overhead target is met
5. Document results in test file

**Estimated Effort**: 3-4 hours  
**Dependencies**: None  
**Blocking**: Phase 1 performance verification

---

## Priority 4: Low Impact / Nice to Have

These issues are optional improvements or have minimal impact on functionality.

### P4.1: Complete Test Helper Migration (Phase 1)

**Requirement**: Phase 1, Requirement 4 - Consolidated Test Helpers  
**Status**: ✅ Completed (migration incomplete)  
**Impact**: LOW - Helpers exist and work, migration is optional

**Issue**:
- Consolidated test helpers fully implemented ✅
- 368 instances of raw `t.TempDir()` calls remain in test files
- Tests work correctly but don't use consolidated helpers
- Inconsistent test setup patterns across codebase

**Evidence**:
```
Consolidated helpers: ✅ Implemented
- CreateTempConfig ✅
- CreateTempDir ✅
- AssertNoError ✅
- AssertError ✅
- AssertEqual ✅

Raw t.TempDir() calls: 368 instances
Migration status: Optional, not blocking
```

**Action Items**:
1. Scan codebase for `t.TempDir()` usage
2. Prioritize migration by package
3. Update tests to use `CreateTempDir` helper
4. Update tests to use assertion helpers
5. Verify tests still pass
6. Document migration progress

**Estimated Effort**: 8-12 hours (optional)  
**Dependencies**: None  
**Blocking**: Nothing (optional improvement)

---

### P4.2: Enhance Migration Tooling with Automated Refactoring (Phase 3)

**Requirement**: Phase 3, Requirement 12 - Migration Tooling  
**Status**: ⚠️ Partially completed  
**Impact**: LOW - Scanner works, automation would be nice to have

**Issue**:
- Migration scanner successfully identifies legacy patterns ✅
- Scanner generates reports ✅
- No automated refactoring (only identifies, doesn't fix)
- No automated compilation validation
- No CI/CD integration

**Evidence**:
```
Scanner features:
- Identifies legacy patterns ✅
- Generates reports ✅
- Tracks progress ✅

Missing features:
- Automated refactoring ⚠️
- Compilation validation ⚠️
- CI/CD integration ⚠️
```

**Action Items**:
1. Add automated refactoring to scanner
   - Parse Go AST
   - Identify legacy function calls
   - Generate replacement code
   - Write updated files
2. Add compilation validation
   - Run `go build` after refactoring
   - Report compilation errors
   - Rollback on failure
3. Add CI/CD integration
   - Run scanner in CI pipeline
   - Track migration progress over time
   - Fail build if new legacy patterns introduced

**Estimated Effort**: 12-16 hours (optional)  
**Dependencies**: None  
**Blocking**: Nothing (nice to have)

---

## Metrics Summary

### Requirements Status

| Priority | Count | Estimated Effort | Impact |
|----------|-------|------------------|--------|
| P1 - Critical | 2 | 6-10 hours | HIGH - Blocks completion |
| P2 - High | 3 | 28-36 hours | HIGH - Quality/adoption |
| P3 - Medium | 3 | 9-13 hours | MEDIUM - Documentation |
| P4 - Low | 2 | 20-28 hours | LOW - Optional |
| **Total** | **10** | **63-87 hours** | - |

### Phase Breakdown

| Phase | Missing Requirements | Critical | High | Medium | Low |
|-------|---------------------|----------|------|--------|-----|
| Phase 1 | 3 | 2 | 0 | 2 | 1 |
| Phase 2 | 0 | 0 | 0 | 0 | 0 |
| Phase 3 | 2 | 0 | 1 | 0 | 1 |
| Phase 4 | 3 | 0 | 2 | 1 | 0 |
| **Total** | **8** | **2** | **3** | **3** | **2** |

### Completion Estimates

**Critical Issues Only** (P1):
- Effort: 6-10 hours
- Impact: Unblocks Phase 1 completion
- Recommended: Complete immediately

**Critical + High Priority** (P1 + P2):
- Effort: 34-46 hours (~1 week)
- Impact: Completes all phases to 95%+
- Recommended: Complete within 2 weeks

**All Priorities** (P1 + P2 + P3 + P4):
- Effort: 63-87 hours (~2 weeks)
- Impact: 100% completion with all optional improvements
- Recommended: Complete within 1 month

---

## Recommendations

### Immediate Actions (This Week)

1. **Fix Error Handling Tests** (P1.1)
   - Highest priority - blocks verification
   - Quick fix: 2-4 hours
   - Unblocks Phase 1 completion

2. **Increase FileSystem Coverage** (P1.2)
   - Critical infrastructure component
   - Effort: 4-6 hours
   - Completes Phase 1 quality targets

### Short-Term Actions (Next 2 Weeks)

3. **Complete File Operations Migration** (P2.1)
   - Highest impact on code consistency
   - Effort: 12-16 hours
   - Completes Phase 4 core requirement

4. **Create Migration Guide** (P2.2)
   - Unblocks developer adoption
   - Effort: 6-8 hours
   - Completes Phase 3 documentation

5. **Improve Services Test Coverage** (P2.3)
   - Core functionality needs verification
   - Effort: 10-12 hours
   - Completes Phase 4 quality targets

### Medium-Term Actions (Next Month)

6. **Create ADR** (P3.1)
   - Documentation best practice
   - Effort: 2-3 hours
   - Low effort, high value

7. **Calculate Metrics** (P3.2)
   - Demonstrates refactoring value
   - Effort: 4-6 hours
   - Useful for reporting progress

8. **Add Performance Benchmarks** (P3.3)
   - Verifies performance targets
   - Effort: 3-4 hours
   - Completes Phase 1 verification

### Optional Improvements (As Time Permits)

9. **Test Helper Migration** (P4.1)
   - Nice to have, not critical
   - Effort: 8-12 hours
   - Improves consistency

10. **Enhanced Migration Tooling** (P4.2)
    - Automation would be helpful
    - Effort: 12-16 hours
    - Reduces manual migration work

### Success Criteria

**Phase 1 Complete** when:
- ✅ Error handling tests pass (P1.1)
- ✅ FileSystem coverage >95% (P1.2)
- ✅ ADR created (P3.1)
- ✅ Performance benchmarks added (P3.3)

**Phase 3 Complete** when:
- ✅ Migration guide created (P2.2)

**Phase 4 Complete** when:
- ✅ File operations migrated (P2.1)
- ✅ Services coverage >85% (P2.3)
- ✅ Metrics calculated (P3.2)

**All Phases 100% Complete** when:
- ✅ All P1, P2, P3 items completed
- ✅ Optional P4 items completed (or explicitly deferred)

---

**Document Status**: Initial version  
**Next Update**: After completing P1 critical issues  
**Maintained By**: Project maintainers  
**Review Frequency**: Weekly during active development

**Generated**: February 6, 2026  
**Source**: IMPLEMENTATION_STATUS.md (verified February 4-5, 2026)  
**Verified By**: Automated analysis of implementation status
