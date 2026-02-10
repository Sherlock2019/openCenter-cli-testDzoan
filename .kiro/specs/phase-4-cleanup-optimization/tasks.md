# Implementation Plan: Phase 4 Cleanup

## Overview

This implementation plan breaks down Phase 4 cleanup into discrete, incremental tasks. The approach follows a careful migration strategy: create the base plugin foundation, migrate plugins one at a time with testing, consolidate utilities, and remove unused abstractions. Each task builds on previous work and includes validation checkpoints.

## Tasks

- [x] 1. Create Base Service Plugin Foundation
  - [x] 1.1 Implement BaseServicePlugin with composition pattern
    - Create `internal/services/base_plugin.go` with PluginMetadata struct
    - Implement BaseServicePlugin with metadata storage
    - Add accessor methods (Name, Version, Description, Type, Author, License)
    - Implement SetValidator and SetRenderer for function injection
    - Implement Validate and Render methods that delegate to injected functions
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.7, 1.8_
  
  - [x] 1.2 Write property test for base plugin metadata accessibility
    - **Property 1: Base Plugin Metadata Accessibility**
    - **Validates: Requirements 1.1, 1.2, 1.6**
    - Generate random metadata and verify all accessors return correct values
    - _Requirements: 1.1, 1.2, 1.6_
  
  - [x] 1.3 Write property test for custom logic injection
    - **Property 2: Custom Logic Injection**
    - **Validates: Requirements 1.4, 1.7, 1.8**
    - Test that injected validator and renderer functions are called correctly
    - _Requirements: 1.4, 1.7, 1.8_
  
  - [x] 1.4 Write unit test for plugin composition
    - Test that embedded plugins can access base methods
    - Test that embedded plugins can override specific methods
    - _Requirements: 1.3_

- [x] 2. Migrate Core Service Plugins
  - [x] 2.1 Migrate cert-manager plugin
    - Update `internal/services/cert_manager.go` to embed BaseServicePlugin
    - Move metadata to PluginMetadata struct
    - Convert validation logic to injected function
    - Convert rendering logic to injected function
    - Remove boilerplate accessor methods
    - _Requirements: 2.3, 2.8, 2.9_
  
  - [ ]* 2.2 Write property test for cert-manager behavioral equivalence
    - **Property 4: Migrated Plugin Behavioral Equivalence**
    - **Validates: Requirements 2.8**
    - Compare validation/rendering results before and after migration
    - _Requirements: 2.8_
  
  - [ ]* 2.3 Write property test for cert-manager backward compatibility
    - **Property 5: Configuration Backward Compatibility**
    - **Validates: Requirements 2.9**
    - Test existing configurations work with migrated plugin
    - _Requirements: 2.9_
  
  - [x] 2.4 Migrate calico plugin
    - Apply same migration pattern as cert-manager
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.3, 2.8, 2.9_
  
  - [x] 2.5 Migrate cilium plugin
    - Apply same migration pattern
    - _Requirements: 2.3, 2.8, 2.9_
  
  - [x] 2.6 Migrate kube-ovn plugin
    - Apply same migration pattern
    - _Requirements: 2.3, 2.8, 2.9_

- [x] 3. Migrate Observability Service Plugins
  - [x] 3.1 Migrate prometheus-stack plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.4, 2.8, 2.9_
  
  - [x] 3.2 Migrate loki plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.4, 2.8, 2.9_
  
  - [x] 3.3 Migrate tempo plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.4, 2.8, 2.9_
  
  - [x] 3.4 Migrate grafana plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.4, 2.8, 2.9_

- [x] 4. Migrate Application and Storage Service Plugins
  - [x] 4.1 Migrate keycloak plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.5, 2.8, 2.9_
  
  - [x] 4.2 Migrate harbor plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.5, 2.8, 2.9_
  
  - [x] 4.3 Migrate vault plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.5, 2.8, 2.9_
  
  - [x] 4.4 Migrate velero plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.6, 2.8, 2.9_
  
  - [x] 4.5 Migrate etcd-backup plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.6, 2.8, 2.9_
  
  - [x] 4.6 Migrate vsphere-csi plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.7, 2.8, 2.9_
  
  - [x] 4.7 Migrate ceph-csi plugin
    - Update to use BaseServicePlugin composition
    - _Requirements: 2.7, 2.8, 2.9_

- [x] 5. Checkpoint - Verify Plugin Migration
  - Ensure all plugin tests pass
  - Verify code reduction metrics (should see ~1,000 LOC reduction)
  - Run integration tests for plugin registration
  - Ask the user if questions arise

- [x] 6. Consolidate Path Resolution
  - [x] 6.1 Audit path resolution usage across codebase
    - Search for hardcoded path construction patterns
    - Identify all locations using filepath.Join for cluster paths
    - Document files that need updates
    - _Requirements: 3.7, 3.8_
  
  - [x] 6.2 Update PathResolver with comprehensive methods
    - Ensure ResolveConfigPath, ResolveSecretsPath, ResolveGitOpsPath exist
    - Add caching with thread-safe map
    - Add platform-specific path normalization
    - Add organization-based path handling
    - _Requirements: 3.1, 3.2, 3.3, 3.4_
  
  - [ ]* 6.3 Write property test for PathResolver caching
    - **Property 6: PathResolver Caching**
    - **Validates: Requirements 3.2**
    - Test that second call returns cached result
    - _Requirements: 3.2_
  
  - [ ]* 6.4 Write property test for platform-specific path normalization
    - **Property 7: Platform-Specific Path Normalization**
    - **Validates: Requirements 3.3**
    - Test correct path separators for current platform
    - _Requirements: 3.3_
  
  - [ ]* 6.5 Write property test for organization-based paths
    - **Property 8: Organization-Based Path Structure**
    - **Validates: Requirements 3.4**
    - Test paths include organization correctly
    - _Requirements: 3.4_
  
  - [ ]* 6.6 Write property test for PathResolver thread safety
    - **Property 10: PathResolver Thread Safety**
    - **Validates: Requirements 3.9**
    - Test concurrent access doesn't cause races or panics
    - _Requirements: 3.9_
  
  - [x] 6.7 Migrate internal/config/ to use PathResolver
    - Update all path construction in config package
    - Replace hardcoded filepath.Join calls
    - _Requirements: 3.7, 3.8_
  
  - [x] 6.8 Migrate internal/sops/ to use PathResolver
    - Update path construction in SOPS manager
    - _Requirements: 3.7, 3.8_
  
  - [x] 6.9 Migrate internal/gitops/ to use PathResolver
    - Update path construction in GitOps operations
    - _Requirements: 3.7, 3.8_
  
  - [x] 6.10 Migrate cmd/ commands to use PathResolver
    - Update CLI commands to use PathResolver
    - _Requirements: 3.7, 3.8_

- [x] 7. Complete File Operations Migration
  - [x] 7.1 Audit remaining direct file operations
    - Run: `grep -rn "os\.ReadFile\|os\.WriteFile" internal/`
    - Document all remaining instances
    - Create migration checklist
    - _Requirements: 4.1, 4.2_
  
  - [x] 7.2 Migrate file operations in internal/sops/manager.go
    - Replace os.ReadFile with fileSystem.ReadFile
    - Replace os.WriteFile with fileSystem.WriteFile
    - Update error handling to use wrapped errors
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5_
  
  - [x] 7.3 Migrate file operations in internal/template/engine.go
    - Replace direct file operations with FileSystem wrapper
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.6_
  
  - [x] 7.4 Migrate file operations in internal/gitops/copy.go
    - Replace direct file operations with FileSystem wrapper
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.7_
  
  - [x] 7.5 Migrate any remaining file operations
    - Address all instances found in audit
    - _Requirements: 4.1, 4.2, 4.3, 4.4_
  
  - [ ]* 7.6 Write property test for atomic file writes
    - **Property 12: Atomic File Writes**
    - **Validates: Requirements 4.4, 4.9**
    - Test interrupted writes don't corrupt files
    - _Requirements: 4.4, 4.9_
  
  - [ ]* 7.7 Write property test for file operation error context
    - **Property 13: File Operation Error Context**
    - **Validates: Requirements 4.8**
    - Test errors include path and operation details
    - _Requirements: 4.8_
  
  - [x] 7.8 Verify no direct os file operations remain
    - Run: `grep -rn "os\.ReadFile\|os\.WriteFile" internal/`
    - Should return zero results
    - _Requirements: 4.10_

- [x] 8. Checkpoint - Verify Utility Consolidation
  - Ensure all path resolution uses PathResolver
  - Ensure all file operations use FileSystem wrapper
  - Run all tests to verify no regressions
  - Ask the user if questions arise

- [x] 9. Remove Unused Interfaces
  - [x] 9.1 Remove ConfigLoaderInterface
    - Delete interface definition from internal/config/interfaces.go
    - Update ConfigurationManager to accept *ConfigLoader
    - Update all callers to pass concrete type
    - Update tests to use concrete type
    - _Requirements: 5.1, 5.5, 5.6, 5.7_
  
  - [x] 9.2 Remove PathResolverInterface
    - Delete interface definition
    - Update ConfigurationManager to accept *PathResolver
    - Update all callers to pass concrete type
    - Update tests to use concrete type
    - _Requirements: 5.2, 5.5, 5.6, 5.7_
  
  - [x] 9.3 Remove ConfigCacheInterface
    - Delete interface definition
    - Update ConfigurationManager to accept *ConfigCache
    - Update all callers to pass concrete type
    - Update tests to use concrete type
    - _Requirements: 5.3, 5.5, 5.6, 5.7_
  
  - [ ]* 9.4 Write property test for interface removal behavioral equivalence
    - **Property 14: Interface Removal Behavioral Equivalence**
    - **Validates: Requirements 5.8**
    - Test operations produce identical results after interface removal
    - _Requirements: 5.8_
  
  - [x] 9.5 Verify ConfigValidatorInterface is retained
    - Confirm multiple implementations exist (schema validator, business rules validator)
    - Keep interface as it provides value
    - _Requirements: 5.4_

- [x] 10. Final Verification and Metrics
  - [x] 10.1 Run complete test suite
    - Execute: `mise run test`
    - Verify all tests pass
    - _Requirements: 7.7, 7.8_
  
  - [x] 10.2 Measure code reduction
    - Run: `git diff --stat <phase-4-start-commit>`
    - Verify at least 1,000 LOC reduction
    - _Requirements: 6.1, 6.4_
  
  - [x] 10.3 Measure boilerplate reduction in plugins
    - Compare plugin LOC before and after migration
    - Verify at least 70% boilerplate reduction
    - _Requirements: 6.2, 6.5_
  
  - [x] 10.4 Measure test coverage
    - Run: `go test -cover ./...`
    - Verify at least 85% coverage
    - _Requirements: 6.3, 6.6, 7.6_
  
  - [x] 10.5 Run race detector
    - Execute: `go test -race ./...`
    - Verify no data races detected
    - _Requirements: 3.9_
  
  - [x] 10.6 Document metrics in final report
    - Create summary of all metrics
    - Compare against targets
    - Document any deviations
    - _Requirements: 6.7_

- [x] 11. Final Checkpoint
  - All tests passing
  - All metrics meet or exceed targets
  - No direct os file operations remain
  - All path resolution uses PathResolver
  - All plugins use BaseServicePlugin
  - Unused interfaces removed
  - Ask the user if questions arise

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Plugin migration follows incremental pattern: migrate one, test, move to next
- Checkpoints ensure validation at key milestones
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
