# Architectural Refactoring - Requirements

## Overview

This spec addresses systemic architectural issues identified in the opencenter-cli codebase health assessment (Health Score: 6.5/10). The refactoring will eliminate duplication, improve maintainability, and establish clear architectural patterns for future development.

## Priority Classification

Issues are prioritized based on impact and urgency:

- **P0 (Critical)**: Path resolution duplication, validation fragmentation
- **P1 (High)**: Configuration loading complexity, command layer bloat
- **P2 (Medium)**: Registry duplication, error handling inconsistency
- **P3 (Low)**: Wrapper functions, orphaned code cleanup

## User Stories

### Epic 1: Path Resolution Consolidation (P0 - Critical)

**As a** developer maintaining opencenter-cli  
**I want** a single, centralized path resolution system  
**So that** I can eliminate 40+ duplicate path construction calls and ensure consistent behavior

#### Story 1.1: Create Core PathResolver

**As a** developer  
**I want** a centralized PathResolver in `internal/core/paths/`  
**So that** all path logic exists in one place with caching and organization-awareness

**Acceptance Criteria:**
- PathResolver handles all path types: config files, secrets, GitOps directories, SSH keys, SOPS keys
- Supports organization-based structure: `clusters/<org>/infrastructure/clusters/<cluster>/`
- Implements caching with invalidation mechanism
- Thread-safe with proper mutex usage
- 100% test coverage including edge cases
- Benchmark tests show <1ms resolution time
- Documentation includes usage examples

#### Story 1.2: Migrate cmd/ Package to PathResolver

**As a** developer  
**I want** all command files to use PathResolver  
**So that** CLI commands have consistent path behavior

**Acceptance Criteria:**
- `cmd/cluster_init.go` uses PathResolver (eliminate 15+ filepath.Join calls)
- All other cluster commands migrated
- No direct `filepath.Join(..., "clusters", ...)` calls remain in cmd/
- All existing tests pass
- Integration tests verify path resolution

#### Story 1.3: Migrate internal/gitops/ to PathResolver

**As a** developer  
**I want** GitOps generation to use PathResolver  
**So that** template rendering uses consistent paths

**Acceptance Criteria:**
- All gitops stages use PathResolver
- `copy.go` and related files migrated
- No duplicate path construction logic
- GitOps generation tests pass
- Template rendering produces correct paths

#### Story 1.4: Migrate internal/operations/ to PathResolver

**As a** developer  
**I want** operations (backup, restore) to use PathResolver  
**So that** operational tasks use consistent paths

**Acceptance Criteria:**
- `backup_manager.go` uses PathResolver
- All operations migrated
- Backup/restore tests pass
- No path duplication remains

#### Story 1.5: Deprecate Old Path Functions

**As a** developer  
**I want** old path functions marked deprecated and eventually removed  
**So that** the codebase uses only PathResolver

**Acceptance Criteria:**
- Old functions in `config.go` marked deprecated
- Deprecation warnings added
- Migration guide created
- Timeline for removal documented (2 releases)

### Epic 2: Validation Engine Consolidation (P0 - Critical)

**As a** developer  
**I want** a unified validation system  
**So that** I can eliminate 50+ duplicate validation functions and ensure consistent error handling

#### Story 2.1: Create Core ValidationEngine

**As a** developer  
**I want** a ValidationEngine in `internal/core/validation/`  
**So that** all validation uses a consistent interface

**Acceptance Criteria:**
- ValidationEngine with registry pattern
- Standard ValidationResult format with errors, warnings, suggestions
- Pluggable validator architecture
- Context-aware validation (can pass metadata)
- Thread-safe validator registration
- 100% test coverage
- Documentation with examples

#### Story 2.2: Migrate Cluster Name Validation

**As a** developer  
**I want** cluster name validation consolidated  
**So that** validation occurs once at the boundary

**Acceptance Criteria:**
- Single ClusterNameValidator implementation
- Removes duplicate validation from 4 locations
- Validates format, length, special characters
- Provides helpful error messages
- Tests cover all edge cases

#### Story 2.3: Migrate All Validators to ValidationEngine

**As a** developer  
**I want** all validators using ValidationEngine interface  
**So that** validation is consistent across the codebase

**Acceptance Criteria:**
- Config validators migrated
- File validators migrated
- Template validators migrated
- Security validators migrated
- All tests pass
- Consistent error format

#### Story 2.4: Add Suggestion Engine

**As a** developer  
**I want** validation errors to include helpful suggestions  
**So that** users can quickly fix configuration issues

**Acceptance Criteria:**
- Suggestion engine integrated with ValidationEngine
- Context-aware suggestions based on error type
- Suggestions for common mistakes (typos, wrong values)
- Tests verify suggestion quality
- Documentation includes examples

### Epic 3: Configuration Management Unification (P1 - High)

**As a** developer  
**I want** a single configuration loader  
**So that** I can eliminate 3 overlapping loaders and simplify version handling

#### Story 3.1: Create Unified ConfigManager

**As a** developer  
**I want** a ConfigManager with strategy pattern  
**So that** all versions load through one interface

**Acceptance Criteria:**
- ConfigManager in `internal/core/config/`
- Strategy pattern for v1, v2, legacy loaders
- Auto-detection of version from YAML
- Single Load() method handles all versions
- Migration pipeline integrated
- 100% test coverage
- Documentation with examples

#### Story 3.2: Implement Load Strategies

**As a** developer  
**I want** separate strategies for each config version  
**So that** version-specific logic is isolated

**Acceptance Criteria:**
- V1Strategy wraps existing v1 loader
- V2Strategy wraps existing v2 loader
- LegacyStrategy handles flat files
- Each strategy implements CanLoad() and Load()
- Tests verify correct strategy selection
- Backward compatibility maintained

#### Story 3.3: Split config.go into Focused Modules

**As a** developer  
**I want** the 1984-line config.go split into focused files  
**So that** each module has a single responsibility

**Acceptance Criteria:**
- `types.go` - Config struct definitions only
- `defaults.go` - Default generation logic
- `persistence.go` - Load/Save operations
- `migration/` - Version migration logic
- `validation/` - Validation logic (uses core ValidationEngine)
- All imports updated across codebase
- All tests pass
- No functionality changes

#### Story 3.4: Migrate All Config Loading to ConfigManager

**As a** developer  
**I want** all Load() calls using ConfigManager  
**So that** configuration loading is consistent

**Acceptance Criteria:**
- All cmd/ files use ConfigManager
- All internal/ packages use ConfigManager
- Old loaders removed or deprecated
- All tests pass
- Migration guide created

### Epic 4: Command Layer Refactoring (P1 - High)

**As a** developer  
**I want** thin command layer with business logic in domain services  
**So that** CLI code is testable and reusable

#### Story 4.1: Extract InitService from cluster_init.go

**As a** developer  
**I want** business logic extracted from 1672-line cluster_init.go  
**So that** initialization logic is testable without Cobra

**Acceptance Criteria:**
- InitService in `internal/cluster/init_service.go`
- Service handles: path resolution, key generation, config creation, git init
- Command file only handles: flag parsing, service invocation
- InitService has no Cobra dependencies
- Unit tests for InitService (no CLI mocking needed)
- Integration tests verify end-to-end flow
- Cyclomatic complexity reduced from 150+ to <50

#### Story 4.2: Create Domain Services for Other Commands

**As a** developer  
**I want** domain services for validate, setup, bootstrap commands  
**So that** all commands follow the same pattern

**Acceptance Criteria:**
- ValidateService created
- SetupService created
- BootstrapService created
- All services use dependency injection
- All services testable without CLI
- Command files <200 lines each

#### Story 4.3: Implement Dependency Injection Container

**As a** developer  
**I want** a DI container for service dependencies  
**So that** services are loosely coupled and testable

**Acceptance Criteria:**
- DI container in `internal/di/`
- Container provides: PathResolver, ConfigManager, ValidationEngine, KeyGenerator, GitOpsService
- Services registered at startup
- Commands retrieve services from container
- Tests can inject mocks
- Documentation with examples

### Epic 5: Registry Consolidation (P2 - Medium)

**As a** developer  
**I want** a generic registry pattern  
**So that** ServiceRegistry and TemplateRegistry share implementation

#### Story 5.1: Create Generic Registry

**As a** developer  
**I want** a generic Registry[T] in `internal/core/registry/`  
**So that** dependency resolution logic exists once

**Acceptance Criteria:**
- Generic Registry[T DependencyAware] implementation
- Handles registration, lookup, dependency resolution
- Circular dependency detection
- Thread-safe operations
- 100% test coverage
- Documentation with examples

#### Story 5.2: Migrate ServiceRegistry

**As a** developer  
**I want** ServiceRegistry using generic Registry[T]  
**So that** service registration uses shared logic

**Acceptance Criteria:**
- ServiceRegistry = Registry[ServiceDefinition]
- All service registration tests pass
- Dependency resolution works correctly
- No behavior changes

#### Story 5.3: Migrate TemplateRegistry

**As a** developer  
**I want** TemplateRegistry using generic Registry[T]  
**So that** template registration uses shared logic

**Acceptance Criteria:**
- TemplateRegistry = Registry[TemplateDefinition]
- All template registration tests pass
- Dependency resolution works correctly
- No behavior changes

### Epic 6: Error Handling Standardization (P2 - Medium)

**As a** developer  
**I want** consistent error handling across the codebase  
**So that** errors are machine-readable and user-friendly

#### Story 6.1: Create Core Error Package

**As a** developer  
**I want** a standardized error package  
**So that** all errors have consistent format

**Acceptance Criteria:**
- Error struct with: Code, Message, Cause, Context, Suggestions, StackTrace
- Wrap() function for error wrapping
- WithContext() for adding context
- WithSuggestion() for adding suggestions
- Error codes documented
- Tests verify error formatting

#### Story 6.2: Migrate Error Handling

**As a** developer  
**I want** all error handling using core error package  
**So that** errors are consistent

**Acceptance Criteria:**
- All fmt.Errorf replaced with errors.Wrap
- Error codes assigned to common errors
- Suggestions added where helpful
- Tests verify error format
- Documentation updated

### Epic 7: Tech Debt Cleanup (P3 - Low)

**As a** developer  
**I want** deprecated code and orphaned interfaces removed  
**So that** the codebase is clean and maintainable

#### Story 7.1: Remove Deprecated Functions

**As a** developer  
**I want** deprecated wrapper functions removed  
**So that** the codebase has no dead code

**Acceptance Criteria:**
- All deprecated functions in cmd/root.go removed
- All deprecated functions in internal/config/ removed
- Callers updated to use new implementations
- All tests pass

#### Story 7.2: Remove Orphaned Interfaces

**As a** developer  
**I want** unused interface methods removed  
**So that** interfaces follow Interface Segregation Principle

**Acceptance Criteria:**
- Audit all interfaces for unused methods
- Remove unused methods
- Split large interfaces into focused ones
- All tests pass
- Documentation updated

#### Story 7.3: Separate Test Fixtures

**As a** developer  
**I want** test-only code removed from production  
**So that** production code is clean

**Acceptance Criteria:**
- Test mode checks removed from defaultConfig()
- Test fixtures in separate files
- Test builders created for complex objects
- All tests pass

#### Story 7.4: Remove Commented-Out Code

**As a** developer  
**I want** all commented-out code removed  
**So that** the codebase is clean

**Acceptance Criteria:**
- All commented-out code removed
- TODOs converted to tracked issues
- Issue references added to remaining TODOs
- Documentation updated

## Non-Functional Requirements

### Performance

- Path resolution: <1ms per call
- Config loading: <100ms (50% improvement)
- Validation: <300ms (40% improvement)
- Memory usage: <100MB peak (33% reduction)
- No performance regressions in any operation

### Reliability

- 100% backward compatibility for user-facing operations
- All existing tests pass
- Integration tests cover critical paths
- Rollback plan for each phase

### Maintainability

- All new code has >90% test coverage
- No functions >50 lines
- No files >500 lines
- Cyclomatic complexity <10
- Clear documentation for all public APIs

### Security

- No hardcoded secrets
- Input validation at boundaries
- Error messages don't leak sensitive data
- Dependencies kept up to date

## Success Metrics

### Code Quality
- 20% reduction in total lines of code
- 67% reduction in cyclomatic complexity
- 80%+ test coverage
- <5% duplicate code

### Performance
- 40% faster cluster initialization
- 50% faster configuration loading
- 33% reduction in memory usage

### Maintainability
- 30% faster feature development
- 40% faster bug fixes
- 50% faster developer onboarding
- 35% faster code reviews

## Dependencies

- Go 1.25.2
- Existing test infrastructure
- CI/CD pipeline
- Documentation system

## Constraints

- Must maintain backward compatibility
- Cannot break existing user workflows
- Must support gradual migration (old and new code coexist)
- Changes must be reviewable (PRs <500 lines)

## Risks and Mitigations

### High Risk: Path Resolution Changes
- **Mitigation**: Extensive backward compatibility tests, phased rollout, feature flags
- **Contingency**: Keep old path resolution as fallback for 2 releases

### Medium Risk: Configuration Loading Changes
- **Mitigation**: Keep old loaders as fallbacks, comprehensive migration tests
- **Contingency**: Rollback to previous loader, manual migration scripts

### Low Risk: Validation Consolidation
- **Mitigation**: Maintain existing validation behavior, regression tests
- **Contingency**: Revert to old validators, provide override flags

## Timeline

- **Phase 1 (Weeks 1-2)**: Foundation - Core abstractions
- **Phase 2 (Weeks 3-4)**: Migration - Migrate to new abstractions
- **Phase 3 (Weeks 5-6)**: Cleanup - Remove duplication and tech debt
- **Phase 4 (Week 7)**: Optimization - Performance improvements

Total: 7 weeks

## References

- [Architectural Review](../../../docs/release-1-todo.md)
- [Architecture Documentation](../../../docs/dev/architecture.md)
- [Developer Guide](../../../docs/dev/readme.md)
