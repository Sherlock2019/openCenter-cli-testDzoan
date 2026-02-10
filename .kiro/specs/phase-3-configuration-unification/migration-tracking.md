# Configuration Migration Tracking Document

## Overview

This document tracks the progress of migrating 19 files from legacy config functions to the unified ConfigurationManager. Files are organized by layer to enable systematic batch migration.

**Generated**: 2026-02-03  
**Total Files**: 19  
**Status**: Not Started

## Migration Statistics

- **Total files to migrate**: 19
- **Files using config.Load**: 15
- **Files using config.Save**: 5
- **Files using config.Validate**: 3
- **Files completed**: 0
- **Files in progress**: 0
- **Files remaining**: 19

## Migration Batches

### Batch 1: Command Layer - Core Operations (Priority: High)

**Files**: 8  
**Status**: Not Started  
**Dependencies**: None

- [ ] `cmd/cluster_config.go` (Load)
- [ ] `cmd/cluster_info.go` (Load, Validate)
- [ ] `cmd/cluster_select.go` (Load)
- [ ] `cmd/cluster_status.go` (Load)
- [ ] `cmd/cluster_destroy.go` (Save)
- [ ] `cmd/cluster_service.go` (Save)
- [ ] `cmd/cluster_update.go` (Save, Validate)
- [ ] `cmd/config_helpers.go` (Load)

**Notes**:
- These are user-facing commands
- High visibility, test thoroughly
- Update integration tests after migration

### Batch 2: Command Layer - Credentials & Secrets (Priority: High)

**Files**: 4  
**Status**: Not Started  
**Dependencies**: None

- [ ] `cmd/cluster_credentials_export.go` (Load)
- [ ] `cmd/cluster_env.go` (Load)
- [ ] `cmd/cluster_lock.go` (Load)
- [ ] `cmd/secrets.go` (Load)

**Notes**:
- Security-sensitive operations
- Verify credential handling after migration
- Test with encrypted configs

### Batch 3: Command Layer - Validation & Preflight (Priority: Medium)

**Files**: 2  
**Status**: Not Started  
**Dependencies**: None

- [ ] `cmd/cluster_validate_manifests.go` (Load)
- [ ] `cmd/cluster_preflight.go` (Load)

**Notes**:
- Validation workflows
- Ensure error messages remain clear
- Test with invalid configs

### Batch 4: Service Layer (Priority: High)

**Files**: 3  
**Status**: Not Started  
**Dependencies**: Batch 1 (command layer)

- [ ] `internal/cluster/init_service.go` (Validate)
- [ ] `internal/cluster/setup_service.go` (Load)
- [ ] `internal/cluster/bootstrap_service.go` (Load)

**Notes**:
- Core business logic
- Update service tests
- Verify dependency injection works

### Batch 5: Internal Infrastructure (Priority: Medium)

**Files**: 1  
**Status**: Not Started  
**Dependencies**: All command and service layers

- [ ] `internal/config/persistence.go` (Load, Save)

**Notes**:
- Low-level config operations
- May need refactoring to use manager directly
- Consider deprecating if redundant

### Batch 6: Test Infrastructure (Priority: Low)

**Files**: 1  
**Status**: Not Started  
**Dependencies**: All other batches

- [ ] `tests/features/steps/helpers.go` (Load, Save)

**Notes**:
- BDD test helpers
- Update after all production code migrated
- Verify all feature tests still pass

## Files by Operation Type

### config.Load (15 files)

**Command Layer (10)**:
- cmd/cluster_config.go
- cmd/cluster_credentials_export.go
- cmd/cluster_env.go
- cmd/cluster_info.go
- cmd/cluster_select.go
- cmd/cluster_status.go
- cmd/cluster_validate_manifests.go
- cmd/cluster_lock.go
- cmd/cluster_preflight.go
- cmd/config_helpers.go

**Service Layer (2)**:
- internal/cluster/bootstrap_service.go
- internal/cluster/setup_service.go

**Infrastructure (2)**:
- internal/config/persistence.go
- cmd/secrets.go

**Test Infrastructure (1)**:
- tests/features/steps/helpers.go

### config.Save (5 files)

**Command Layer (3)**:
- cmd/cluster_destroy.go
- cmd/cluster_service.go
- cmd/cluster_update.go

**Infrastructure (1)**:
- internal/config/persistence.go

**Test Infrastructure (1)**:
- tests/features/steps/helpers.go

### config.Validate (3 files)

**Command Layer (2)**:
- cmd/cluster_info.go
- cmd/cluster_update.go

**Service Layer (1)**:
- internal/cluster/init_service.go

## Migration Workflow

### Per-File Checklist

For each file being migrated:

1. **Pre-Migration**
   - [ ] Read file and identify all legacy config calls
   - [ ] Identify function signatures that need context parameter
   - [ ] Review existing tests for the file
   - [ ] Note any special error handling

2. **Migration**
   - [ ] Add context parameter to function signatures
   - [ ] Replace `config.Load()` with `manager.Load(ctx, ...)`
   - [ ] Replace `config.Save()` with `manager.Save(ctx, ...)`
   - [ ] Replace `config.Validate()` with `manager.Validate(ctx, ...)`
   - [ ] Update error handling if needed
   - [ ] Inject ConfigurationManager via DI

3. **Testing**
   - [ ] Update unit tests
   - [ ] Run unit tests: `mise run test`
   - [ ] Run integration tests if applicable
   - [ ] Manual testing of affected commands

4. **Verification**
   - [ ] Code compiles without errors
   - [ ] All tests pass
   - [ ] No legacy config calls remain
   - [ ] Error messages are clear and helpful

### Batch Migration Process

1. **Select Batch**: Choose next batch based on priority and dependencies
2. **Update Status**: Mark batch as "In Progress"
3. **Migrate Files**: Follow per-file checklist for each file in batch
4. **Batch Testing**: Run full test suite after batch completion
5. **Update Tracking**: Mark batch as "Complete" and update statistics
6. **Review**: Verify no regressions before moving to next batch

## Risk Assessment

### High Risk Files

Files that require extra care during migration:

1. **cmd/secrets.go**: Handles sensitive credential operations
2. **internal/config/persistence.go**: Core config I/O, may need refactoring
3. **internal/cluster/init_service.go**: Critical initialization logic
4. **tests/features/steps/helpers.go**: Affects all BDD tests

### Migration Risks

- **Breaking Changes**: Context parameter addition changes function signatures
- **Dependency Injection**: Need to ensure manager is available in all contexts
- **Error Handling**: Error types may change, affecting error handling code
- **Test Coverage**: Must maintain or improve test coverage during migration

### Mitigation Strategies

- Migrate in small batches with full testing between batches
- Keep legacy functions temporarily for gradual migration
- Add deprecation warnings to legacy functions
- Comprehensive integration testing after each batch
- Rollback plan: Git branches for each batch

## Progress Log

### 2026-02-03: Initial Scan
- Ran migration scanner on codebase
- Identified 19 files requiring migration
- Created migration tracking document
- Organized files into 6 migration batches
- Status: Ready to begin migration

---

## Next Steps

1. Review migration report with team
2. Confirm batch priorities and dependencies
3. Begin Batch 1: Command Layer - Core Operations
4. Update this document after each batch completion

## Notes

- This is a living document - update after each file/batch migration
- Track any issues or blockers in the Progress Log section
- Document any deviations from the planned migration approach
- Keep migration report (.kiro/specs/phase-3-configuration-unification/migration-report.md) as reference
