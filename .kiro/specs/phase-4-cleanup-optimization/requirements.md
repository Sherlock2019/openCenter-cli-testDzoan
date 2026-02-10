# Requirements Document: Phase 4 Cleanup

## Introduction

Phase 4 completes the architectural refactoring initiative by consolidating service plugins, finalizing utility migrations, and removing unused code. This phase eliminates the remaining 1,230 lines of boilerplate code across 15+ service plugins, completes the file operations migration started in Phase 1, consolidates path resolution, and removes unused interfaces.

This phase depends on the completion of Phases 1-3 and represents the final cleanup work to ensure the refactoring is complete and sustainable.

## Glossary

- **Service_Plugin**: A modular component that implements service-specific configuration, validation, and rendering logic for Kubernetes services (e.g., cert-manager, Loki, Velero)
- **Base_Service_Plugin**: A reusable foundation class that provides common boilerplate functionality for all service plugins using composition
- **Boilerplate_Code**: Repetitive code patterns duplicated across multiple plugins (metadata accessors, registration, lifecycle methods)
- **PathResolver**: A unified utility that resolves file system paths for cluster configurations, secrets, and GitOps repositories with caching and platform compatibility
- **FileSystem_Wrapper**: An abstraction layer over os.ReadFile/os.WriteFile that provides atomic operations, error wrapping, and consistent error handling
- **Configuration_Manager**: The unified configuration management system created in Phase 3
- **Validation_Engine**: The consolidated validation system created in Phase 2
- **Property_Based_Test**: A test that validates universal properties across randomly generated inputs using gopter library

## Requirements

### Requirement 1: Base Service Plugin Foundation

**User Story:** As a developer, I want a reusable base service plugin, so that I can create new service plugins without duplicating boilerplate code.

#### Acceptance Criteria

1. THE Base_Service_Plugin SHALL provide metadata storage for name, version, description, type, author, and license
2. THE Base_Service_Plugin SHALL implement all ServicePlugin interface methods using the stored metadata
3. THE Base_Service_Plugin SHALL support composition by allowing plugins to embed the base and override specific methods
4. THE Base_Service_Plugin SHALL provide registration helpers for validator and renderer functions
5. THE Base_Service_Plugin SHALL add less than 1ms overhead compared to direct implementations
6. WHEN a plugin embeds Base_Service_Plugin, THE System SHALL automatically provide all boilerplate methods
7. THE Base_Service_Plugin SHALL support custom validation logic through function injection
8. THE Base_Service_Plugin SHALL support custom rendering logic through function injection

### Requirement 2: Service Plugin Migration

**User Story:** As a developer, I want all existing service plugins migrated to use the base plugin, so that boilerplate code is eliminated and maintenance is simplified.

#### Acceptance Criteria

1. WHEN all plugins are migrated, THE System SHALL reduce total plugin code by at least 1,000 lines
2. WHEN all plugins are migrated, THE System SHALL reduce boilerplate code by at least 70%
3. THE System SHALL migrate all core service plugins (cert-manager, calico, cilium, kube-ovn)
4. THE System SHALL migrate all observability service plugins (prometheus-stack, loki, tempo, grafana)
5. THE System SHALL migrate all application service plugins (keycloak, harbor, vault)
6. THE System SHALL migrate all backup service plugins (velero, etcd-backup)
7. THE System SHALL migrate all storage service plugins (vsphere-csi, ceph-csi)
8. WHEN a plugin is migrated, THE System SHALL preserve all existing functionality
9. WHEN a plugin is migrated, THE System SHALL maintain backward compatibility with existing configurations
10. WHEN all plugins are migrated, THE System SHALL pass all existing plugin tests

### Requirement 3: Unified Path Resolution

**User Story:** As a developer, I want all path operations to use the unified PathResolver, so that path handling is consistent, cached, and platform-compatible.

#### Acceptance Criteria

1. THE PathResolver SHALL provide methods for resolving config paths, secrets paths, and GitOps paths
2. THE PathResolver SHALL cache resolved paths to improve performance
3. THE PathResolver SHALL normalize paths for the current platform (Windows, Linux, macOS)
4. THE PathResolver SHALL handle organization-based path structures correctly
5. THE PathResolver SHALL resolve symlinks consistently
6. THE PathResolver SHALL distinguish between relative and absolute paths
7. WHEN any code needs to resolve a cluster path, THE System SHALL use PathResolver
8. WHEN PathResolver is used, THE System SHALL eliminate all hardcoded path construction
9. THE PathResolver SHALL be thread-safe for concurrent access
10. THE PathResolver SHALL provide clear error messages for invalid paths

### Requirement 4: File Operations Migration

**User Story:** As a developer, I want all file operations to use the FileSystem wrapper, so that file operations are atomic, consistent, and properly error-handled.

#### Acceptance Criteria

1. THE System SHALL eliminate all direct os.ReadFile calls in the internal/ directory
2. THE System SHALL eliminate all direct os.WriteFile calls in the internal/ directory
3. WHEN reading files, THE System SHALL use FileSystem.ReadFile with proper error wrapping
4. WHEN writing files, THE System SHALL use FileSystem.WriteFile with atomic operations
5. THE System SHALL migrate file operations in internal/sops/manager.go
6. THE System SHALL migrate file operations in internal/template/engine.go
7. THE System SHALL migrate file operations in internal/gitops/copy.go
8. WHEN file operations fail, THE System SHALL provide contextual error messages
9. THE System SHALL ensure all file writes are atomic to prevent corruption
10. WHEN the migration is complete, THE System SHALL have zero direct os file operation calls

### Requirement 5: Interface Simplification

**User Story:** As a developer, I want unused interfaces removed, so that the codebase is simpler and easier to understand.

#### Acceptance Criteria

1. THE System SHALL remove ConfigLoaderInterface when only one implementation exists
2. THE System SHALL remove PathResolverInterface when only one implementation exists
3. THE System SHALL remove ConfigCacheInterface when only one implementation exists
4. THE System SHALL retain ConfigValidatorInterface because multiple implementations exist
5. WHEN an interface is removed, THE System SHALL update all consumers to use concrete types
6. WHEN an interface is removed, THE System SHALL update all tests to use concrete types
7. WHEN an interface is removed, THE System SHALL update constructor signatures to accept concrete types
8. THE System SHALL maintain the same functionality after interface removal
9. THE System SHALL pass all tests after interface removal
10. WHEN interfaces are removed, THE System SHALL reduce cognitive complexity

### Requirement 6: Code Quality Metrics

**User Story:** As a project maintainer, I want measurable code quality improvements, so that I can verify the refactoring success.

#### Acceptance Criteria

1. THE System SHALL achieve at least 1,000 lines of code reduction
2. THE System SHALL achieve at least 70% boilerplate code reduction in plugins
3. THE System SHALL achieve at least 85% test coverage
4. WHEN measuring code reduction, THE System SHALL use git diff statistics
5. WHEN measuring boilerplate reduction, THE System SHALL compare plugin code before and after
6. WHEN measuring test coverage, THE System SHALL use go test -cover
7. THE System SHALL document all metrics in the final report
8. THE System SHALL verify metrics meet or exceed targets

### Requirement 7: Testing Requirements

**User Story:** As a developer, I want comprehensive tests for all changes, so that I can be confident the refactoring is correct.

#### Acceptance Criteria

1. THE System SHALL provide unit tests for Base_Service_Plugin
2. THE System SHALL provide unit tests for each migrated service plugin
3. THE System SHALL provide unit tests for PathResolver caching
4. THE System SHALL provide unit tests for FileSystem wrapper operations
5. THE System SHALL provide integration tests for plugin registration
6. WHEN tests are run, THE System SHALL achieve at least 85% coverage
7. WHEN tests are run, THE System SHALL pass all existing tests
8. WHEN tests are run, THE System SHALL pass all new tests
9. THE System SHALL use property-based tests for path resolution edge cases
