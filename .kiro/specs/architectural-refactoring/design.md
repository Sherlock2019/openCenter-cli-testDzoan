# Architectural Refactoring - Design Document

## Executive Summary

This design addresses systemic architectural issues in the opencenter-cli codebase (Health Score: 6.5/10). The refactoring will eliminate 97% of path resolution duplication, 92% of validation duplication, consolidate 3 configuration loaders into 1, and reduce command layer complexity by 91%.

## Design Principles

### 1. Single Responsibility Principle
- Each module has one clear purpose
- No mixed concerns (e.g., CLI + business logic)
- Clear boundaries between layers

### 2. Dependency Inversion
- High-level modules don't depend on low-level modules
- Both depend on abstractions (interfaces)
- Dependency injection for loose coupling

### 3. Don't Repeat Yourself (DRY)
- Eliminate duplicate code through centralization
- Shared logic in core packages
- Generic implementations where appropriate

### 4. Interface Segregation
- Small, focused interfaces
- Clients depend only on methods they use
- No "god" interfaces

### 5. Open/Closed Principle
- Open for extension (plugins, strategies)
- Closed for modification (stable core)
- Strategy pattern for variations

## Architecture Overview

### Current Architecture (Problematic)

```
┌─────────────────────────────────────────────────────────────┐
│                         cmd/ Layer                          │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ cluster_init.go (1672 lines, complexity 150+)        │  │
│  │ - Path construction (15+ calls)                      │  │
│  │ - Validation logic (duplicate)                       │  │
│  │ - Config loading (3 loaders)                         │  │
│  │ - Business logic (tightly coupled)                   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      internal/ Layer                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ config/config.go (1984 lines)                        │  │
│  │ - Mixed responsibilities                             │  │
│  │ - 3 overlapping loaders                              │  │
│  │ - 50+ validation functions                           │  │
│  │ - 40+ path construction calls                        │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘

Issues:
- Duplication: 40+ path calls, 50+ validation functions
- Tight coupling: CLI + business logic
- Mixed concerns: config.go does everything
- Hard to test: Cobra mocking required
- Hard to maintain: Large files, high complexity
```

### Target Architecture (Clean)

```
┌─────────────────────────────────────────────────────────────┐
│                    cmd/ Layer (Thin)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ cluster_init.go (~150 lines)                         │  │
│  │ - Flag parsing only                                  │  │
│  │ - Service invocation                                 │  │
│  │ - Result display                                     │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   Domain Services Layer                     │
│  ┌──────────────────────────────────────────────────────┐  │
│  │ internal/cluster/                                    │  │
│  │ - InitService (testable, no Cobra)                  │  │
│  │ - ValidateService                                    │  │
│  │ - SetupService                                       │  │
│  │ - BootstrapService                                   │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Core Layer                             │
│  ┌────────────────┬────────────────┬────────────────────┐  │
│  │ paths/         │ config/        │ validation/        │  │
│  │ PathResolver   │ ConfigManager  │ ValidationEngine   │  │
│  │ (1 impl)       │ (1 loader)     │ (1 engine)         │  │
│  └────────────────┴────────────────┴────────────────────┘  │
│  ┌────────────────────────────────────────────────────────┐│
│  │ di/ - Dependency Injection Container                   ││
│  └────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘

Benefits:
- No duplication: Single implementations
- Loose coupling: DI + interfaces
- Clear concerns: Each layer has one job
- Easy to test: No CLI mocking needed
- Easy to maintain: Small files, low complexity
```

## Core Components

### 1. PathResolver (internal/core/paths/)

**Purpose**: Single source of truth for all path resolution

**Key Features**:
- Organization-aware path resolution
- Organization search when organization not specified
- Caching with invalidation
- Thread-safe operations

**Interface**:
```go
type PathResolver interface {
    Resolve(clusterName, organization string) (*ClusterPaths, error)
    ResolveWithFallback(clusterName string) (*ClusterPaths, error)
    InvalidateCache(clusterName string)
}
```

**Impact**:
- Eliminates 40+ duplicate path construction calls
- Reduces path-related bugs to 0
- <1ms resolution time, <100μs cached

### 2. ValidationEngine (internal/core/validation/)

**Purpose**: Unified validation system with suggestions

**Key Features**:
- Pluggable validator architecture
- Standard ValidationResult format
- Suggestion engine for common mistakes
- Context-aware validation

**Interface**:
```go
type ValidationEngine interface {
    Register(validator Validator) error
    Validate(ctx context.Context, validatorName string, value interface{}) (ValidationResult, error)
    ValidateAll(ctx context.Context, validators []string, value interface{}) ValidationResult
}
```

**Impact**:
- Eliminates 50+ duplicate validation functions
- Consistent error format across codebase
- 80%+ suggestion accuracy
- <100μs per validation

### 3. ConfigManager (internal/core/config/)

**Purpose**: Unified configuration loading with version handling

**Key Features**:
- Strategy pattern for v1, v2, legacy loaders
- Auto-detection of version
- Integrated migration pipeline
- Caching with invalidation

**Interface**:
```go
type ConfigManager interface {
    Load(path string, opts LoadOptions) (*Config, error)
    Save(path string, config *Config) error
    InvalidateCache(path string)
}
```

**Impact**:
- Consolidates 3 overlapping loaders into 1
- 50% faster config loading (<100ms)
- 100% version detection accuracy
- Splits 1984-line config.go into focused modules

### 4. Domain Services (internal/cluster/)

**Purpose**: Business logic separated from CLI

**Key Services**:
- InitService: Cluster initialization
- ValidateService: Cluster validation
- SetupService: GitOps setup
- BootstrapService: Cluster bootstrap

**Interface Example**:
```go
type InitService interface {
    Initialize(ctx context.Context, opts InitOptions) (*InitResult, error)
}
```

**Impact**:
- Reduces cluster_init.go from 1672 to 150 lines (91%)
- Reduces cyclomatic complexity from 150+ to <50 (67%)
- 100% testable without CLI mocking
- Reusable outside CLI context

### 5. DI Container (internal/di/)

**Purpose**: Manage service dependencies

**Key Features**:
- Service registration and lookup
- Provider functions for common services
- Type-safe dependency resolution

**Interface**:
```go
type Container interface {
    Register(service interface{})
    Get(serviceType interface{}) (interface{}, error)
}
```

**Impact**:
- Loose coupling between services
- Easy to inject mocks for testing
- Clear dependency graph

## Data Flow

### Cluster Initialization Flow

```
User Command
    │
    ▼
┌─────────────────────┐
│ cluster_init.go     │ Parse flags, display results
│ (150 lines)         │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ InitService         │ Business logic
│ (300 lines)         │
└─────────────────────┘
    │
    ├──────────────────────────────────────┐
    │                                      │
    ▼                                      ▼
┌─────────────────────┐          ┌─────────────────────┐
│ PathResolver        │          │ ValidationEngine    │
│ Resolve paths       │          │ Validate name       │
└─────────────────────┘          └─────────────────────┘
    │                                      │
    ▼                                      ▼
┌─────────────────────┐          ┌─────────────────────┐
│ ConfigManager       │          │ KeyGenerator        │
│ Create & save       │          │ Generate keys       │
└─────────────────────┘          └─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ GitService          │
│ Initialize repo     │
└─────────────────────┘
```

### Configuration Loading Flow

```
Load Request
    │
    ▼
┌─────────────────────┐
│ ConfigManager       │
│ Load(path, opts)    │
└─────────────────────┘
    │
    ├─ Check cache ────────────────┐
    │                               │
    ▼                               ▼
┌─────────────────────┐    ┌─────────────────────┐
│ Read file           │    │ Return cached       │
└─────────────────────┘    └─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ Detect version      │
│ Select strategy     │
└─────────────────────┘
    │
    ├──────────────────────────────────────┐
    │                                      │
    ▼                                      ▼
┌─────────────────────┐          ┌─────────────────────┐
│ V2Strategy          │          │ V1Strategy          │
│ Load v2 config      │          │ Load v1 config      │
└─────────────────────┘          └─────────────────────┘
    │                                      │
    ▼                                      ▼
┌─────────────────────┐          ┌─────────────────────┐
│ Auto-migrate?       │          │ Migrator            │
│ (if requested)      │          │ Migrate to v2       │
└─────────────────────┘          └─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ Validate?           │
│ (if requested)      │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ Cache result        │
│ Return config       │
└─────────────────────┘
```

## Module Organization

### Core Packages (internal/core/)

```
internal/core/
├── paths/              # Path resolution
│   ├── resolver.go     # Main implementation
│   ├── types.go        # ClusterPaths struct
│   ├── strategies.go   # Resolution strategies
│   └── cache.go        # Caching mechanism
│
├── config/             # Configuration management
│   ├── manager.go      # Main implementation
│   ├── types.go        # Config struct (from config.go)
│   ├── defaults.go     # Default generation (from config.go)
│   ├── persistence.go  # Load/Save (from config.go)
│   ├── strategies/     # Version-specific loaders
│   │   ├── v1.go
│   │   ├── v2.go
│   │   └── legacy.go
│   └── migration/      # Version migration
│       ├── migrator.go
│       ├── v1_to_v2.go
│       └── legacy_to_v1.go
│
└── validation/         # Validation engine
    ├── engine.go       # Main implementation
    ├── types.go        # ValidationResult, Validator
    ├── registry.go     # Validator registration
    ├── suggestions.go  # Suggestion engine
    └── validators/     # Built-in validators
        ├── cluster.go
        ├── config.go
        ├── file.go
        └── security.go
```

### Domain Services (internal/cluster/)

```
internal/cluster/
├── init_service.go         # Cluster initialization
├── validate_service.go     # Cluster validation
├── setup_service.go        # GitOps setup
├── bootstrap_service.go    # Cluster bootstrap
├── destroy_service.go      # Cluster destruction
└── services_test.go        # Service tests
```

### Dependency Injection (internal/di/)

```
internal/di/
├── container.go        # DI container
├── providers.go        # Service providers
└── container_test.go   # Container tests
```

### Command Layer (cmd/)

```
cmd/
├── cluster_init.go         # Thin wrapper (~150 lines)
├── cluster_validate.go     # Thin wrapper (~150 lines)
├── cluster_setup.go        # Thin wrapper (~150 lines)
└── cluster_bootstrap.go    # Thin wrapper (~150 lines)
```

## Migration Strategy

### Phase 1: Foundation (Weeks 1-2)

**Goal**: Create core abstractions

**Tasks**:
1. Implement PathResolver in internal/core/paths/
2. Implement ValidationEngine in internal/core/validation/
3. Implement ConfigManager in internal/core/config/
4. Add comprehensive tests (100% coverage)
5. Document usage with examples

**Success Criteria**:
- All core packages implemented
- 100% test coverage
- Benchmark tests pass
- Documentation complete

### Phase 2: Migration (Weeks 3-4)

**Goal**: Migrate existing code to use core abstractions

**Tasks**:
1. Create domain services in internal/cluster/
2. Implement DI container in internal/di/
3. Migrate cmd/ files to use services
4. Split config.go into focused modules
5. Update all imports across codebase

**Success Criteria**:
- All services implemented
- All commands migrated
- config.go split into <500 line files
- All existing tests pass

### Phase 3: Cleanup (Weeks 5-6)

**Goal**: Remove duplication and tech debt

**Tasks**:
1. Remove duplicate path construction calls
2. Remove duplicate validation functions
3. Remove old loader functions
4. Remove deprecated code
5. Clean up commented-out code

**Success Criteria**:
- No duplicate path calls remain
- No duplicate validation functions remain
- Old loaders removed or deprecated
- Codebase clean

### Phase 4: Optimization (Week 7)

**Goal**: Performance improvements

**Tasks**:
1. Profile hot paths
2. Optimize caching strategies
3. Reduce memory allocations
4. Benchmark all operations
5. Document performance characteristics

**Success Criteria**:
- <1ms path resolution
- <100ms config loading
- <100μs validation
- <100MB peak memory

## Testing Strategy

### Unit Tests

**Coverage Target**: >90% for all new code

**Focus Areas**:
- Core packages (paths, config, validation)
- Domain services (init, validate, setup)
- DI container

**Test Types**:
- Happy path tests
- Error condition tests
- Edge case tests
- Thread safety tests

### Integration Tests

**Coverage Target**: All critical workflows

**Focus Areas**:
- Cluster initialization end-to-end
- Configuration loading and migration
- Validation with suggestions
- GitOps setup

**Test Types**:
- Full workflow tests
- Backward compatibility tests
- Migration tests

### Benchmark Tests

**Performance Targets**:
- PathResolver.Resolve: <1ms
- PathResolver.Resolve (cached): <100μs
- ConfigManager.Load: <100ms
- ConfigManager.Load (cached): <1ms
- ValidationEngine.Validate: <100μs

### Property-Based Tests

**Focus Areas**:
- Path resolution strategies
- Configuration migration
- Validation logic

**Properties**:
- Path resolution is deterministic
- Migration is reversible
- Validation is consistent

## Performance Characteristics

### Current Performance

- Cluster initialization: ~2s
- Config loading: ~200ms
- Validation: ~500ms
- Memory usage: ~150MB peak

### Target Performance

- Cluster initialization: ~1.2s (40% faster)
- Config loading: ~100ms (50% faster)
- Validation: ~300ms (40% faster)
- Memory usage: ~100MB peak (33% reduction)

### Optimization Strategies

1. **Caching**: Cache path resolution and config loading
2. **Lazy Loading**: Load config only when needed
3. **Parallel Validation**: Run independent validators concurrently
4. **Memory Pooling**: Reuse buffers for YAML parsing
5. **Profiling**: Continuous profiling to identify bottlenecks

## Backward Compatibility

### Guarantees

- All existing user workflows continue to work
- Old config formats supported (v1, legacy)
- Automatic migration to v2 when requested
- Fallback strategies for path resolution

### Deprecation Strategy

1. **Mark as deprecated**: Add deprecation warnings
2. **Document migration**: Provide migration guide
3. **Keep for 2 releases**: Allow time for migration
4. **Remove**: Remove in major version bump

### Migration Path

```
Current State
    │
    ▼
┌─────────────────────┐
│ Old + New coexist   │ Feature flags, parallel implementation
│ (1-2 releases)      │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ Deprecate old       │ Warnings, migration guide
│ (2 releases)        │
└─────────────────────┘
    │
    ▼
┌─────────────────────┐
│ Remove old          │ Major version bump
│ (v2.0.0)            │
└─────────────────────┘
```

## Success Metrics

### Code Quality

- **Lines of Code**: 20% reduction (eliminate duplication)
- **Cyclomatic Complexity**: 67% reduction (150+ → <50)
- **Test Coverage**: 80%+ (up from ~60%)
- **Duplicate Code**: <5% (down from ~15%)

### Performance

- **Initialization Time**: 40% faster (2s → 1.2s)
- **Config Loading**: 50% faster (200ms → 100ms)
- **Validation**: 40% faster (500ms → 300ms)
- **Memory Usage**: 33% reduction (150MB → 100MB)

### Maintainability

- **Feature Development**: 30% faster
- **Bug Fixes**: 40% faster
- **Developer Onboarding**: 50% faster
- **Code Reviews**: 35% faster

### Reliability

- **Path Resolution Bugs**: 0 (down from 3-4/release)
- **Validation Bugs**: 0 (down from 2-3/release)
- **Config Loading Bugs**: 0 (down from 1-2/release)
- **Backward Compatibility**: 100%

## Risk Management

### High Risk: Path Resolution Changes

**Impact**: Could break existing deployments

**Mitigation**:
- Comprehensive backward compatibility tests
- 4 fallback strategies
- Feature flag for gradual rollout
- Keep old functions for 2 releases

**Contingency**:
- Rollback to old path resolution
- Manual migration scripts
- Extended support for old structure

### Medium Risk: Configuration Loading Changes

**Impact**: Could fail to load existing configs

**Mitigation**:
- Keep old loaders as fallbacks
- Comprehensive migration tests
- Auto-detection of version
- Backup before migration

**Contingency**:
- Rollback to previous loader
- Manual config fixes
- Migration assistance tool

### Low Risk: Validation Consolidation

**Impact**: Could change validation behavior

**Mitigation**:
- Maintain existing validation behavior
- Regression tests for all validators
- Gradual rollout with feature flag

**Contingency**:
- Revert to old validators
- Override flags for specific validations
- Custom validator plugins

## Timeline

### Week 1-2: Foundation
- Implement PathResolver
- Implement ValidationEngine
- Implement ConfigManager
- Add tests and documentation

### Week 3-4: Migration
- Create domain services
- Implement DI container
- Migrate commands
- Split config.go

### Week 5-6: Cleanup
- Remove duplication
- Remove tech debt
- Clean up code
- Update documentation

### Week 7: Optimization
- Profile performance
- Optimize hot paths
- Benchmark all operations
- Final documentation

**Total**: 7 weeks

## Dependencies

- Go 1.25.2 (existing)
- gopkg.in/yaml.v3 (existing)
- Cobra (existing)
- All other dependencies are internal

## References

- [Requirements Document](requirements.md)
- [Path Resolver Spec](01-path-resolver.md)
- [Validation Engine Spec](02-validation-engine.md)
- [Config Manager Spec](03-config-manager.md)
- [Command Layer Spec](04-command-layer.md)
- [Architectural Review](../../../docs/release-1-todo.md)
