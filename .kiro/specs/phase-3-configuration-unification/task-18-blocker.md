# Task 18 Blocker Report

**Date**: 2026-02-03  
**Task**: 18. Remove legacy configuration code  
**Status**: ⛔ BLOCKED - Cannot proceed

## Summary

Task 18 requires removing legacy configuration code from `internal/config/persistence.go` and other locations. However, this task **cannot be completed** because the migration from legacy functions to the new ConfigurationManager is not finished.

## Current Situation

### Legacy Functions Still in Use

Based on codebase analysis, the following legacy functions are still actively used:

#### config.Load() - 19+ active references
- `cmd/cluster_config.go` (1 occurrence)
- `cmd/cluster_select.go` (2 occurrences)
- `cmd/cluster_credentials_export.go` (1 occurrence)
- `cmd/cluster_env.go` (1 occurrence)
- `cmd/cluster_validate_manifests.go` (1 occurrence)
- `cmd/secrets.go` (7 occurrences)
- `internal/cluster/bootstrap_service.go` (1 occurrence - fallback)
- `internal/cluster/setup_service.go` (1 occurrence - fallback)
- `tests/features/steps/helpers.go` (2 occurrences)

#### config.Save() - 10+ active references
- `cmd/cluster_service.go` (2 occurrences)
- `cmd/cluster_destroy.go` (1 occurrence)
- `cmd/cluster_update.go` (1 occurrence)
- `tests/features/steps/helpers.go` (2 occurrences)

#### config.Validate() - 3+ active references
- `internal/cluster/init_service.go` (1 occurrence)
- `cmd/cluster_update.go` (1 occurrence)

### Impact of Premature Removal

Removing these legacy functions now would:

1. **Break compilation** for 11+ command files
2. **Break service layer** bootstrap and setup operations
3. **Break test infrastructure** for BDD tests
4. **Break user-facing commands**:
   - All secrets management commands
   - Cluster selection and listing
   - Service enable/disable operations
   - Configuration export
   - Environment variable export

## Prerequisites for Task 18

Before Task 18 can be completed, the following must be done:

### 1. Complete File Migrations (Tasks 12-16)

All files must be migrated to use the new ConfigurationManager:

- [ ] Complete command layer migration (Task 12)
  - [ ] cmd/cluster_config.go
  - [ ] cmd/cluster_select.go
  - [ ] cmd/cluster_credentials_export.go
  - [ ] cmd/cluster_env.go
  - [ ] cmd/cluster_validate_manifests.go
  - [ ] cmd/secrets.go (7 calls)
  - [ ] cmd/cluster_service.go (2 calls)
  - [ ] cmd/cluster_destroy.go
  - [ ] cmd/cluster_update.go (2 calls)

- [ ] Complete service layer migration (Task 14)
  - [ ] internal/cluster/bootstrap_service.go (remove fallback)
  - [ ] internal/cluster/setup_service.go (remove fallback)
  - [ ] internal/cluster/init_service.go

- [ ] Migrate test infrastructure
  - [ ] tests/features/steps/helpers.go

### 2. Verify All Tests Pass

- [ ] All unit tests passing
- [ ] All integration tests passing
- [ ] All BDD tests passing
- [ ] No compilation errors

### 3. Create Migration Documentation (Task 10)

- [ ] Migration guide with before/after examples
- [ ] Migration checklist
- [ ] Common patterns and gotchas documented

## Recommended Action

**Do NOT proceed with Task 18 until:**

1. All prerequisite tasks (12-16) are completed
2. All files are migrated to ConfigurationManager
3. All tests pass
4. Migration documentation is complete

## Alternative Approach

If you want to proceed with the spec despite this blocker, consider:

1. **Mark Task 18 as "Blocked"** in tasks.md
2. **Complete remaining migration tasks first** (12-16)
3. **Return to Task 18** after migration is complete
4. **Update task dependencies** to reflect the correct order

## Verification Command

To verify migration status before attempting Task 18:

```bash
# Check for remaining legacy calls (excluding tests and migration code)
grep -r "config\.Load(" --include="*.go" --exclude="*_test.go" --exclude="scanner.go" cmd/ internal/

grep -r "config\.Save(" --include="*.go" --exclude="*_test.go" --exclude="scanner.go" cmd/ internal/

grep -r "config\.Validate(" --include="*.go" --exclude="*_test.go" --exclude="scanner.go" cmd/ internal/
```

If these commands return any results, Task 18 cannot proceed.

## Conclusion

Task 18 is **blocked** and cannot be completed until the migration is finished. Attempting to remove legacy code now would break the application.

**Estimated time to unblock**: 1-2 days to complete remaining migrations (Tasks 12-16).

