# Implementation Plan: Phase 3 Configuration Unification

## Overview

This implementation consolidates three overlapping configuration management systems into a single unified ConfigurationManager. The plan follows a 4-week schedule with incremental implementation, testing, and migration. All tasks build on Phase 1 (FileSystem, PathResolver) and Phase 2 (ValidationEngine) components.

## Table of Contents

- [Overview](#overview)
- [Tasks](#tasks)
- [Notes](#notes)

## Tasks

- [x] 1. Implement ConfigCache with thread-safe operations
  - Create `internal/config/cache.go` with ConfigCache struct
  - Implement Get, Set, Invalidate, and Clear methods with RWMutex protection
  - Add cache entry expiration support
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 11.5_

- [ ]* 1.1 Write property test for cache thread-safety
  - **Property 23: Cache is thread-safe**
  - **Validates: Requirements 11.5**

- [ ]* 1.2 Write property test for cache operations
  - **Property 5: Cache checked before disk read**
  - **Property 6: Disk loads populate cache**
  - **Property 7: Save invalidates cache**
  - **Property 8: ClearCache empties cache**
  - **Validates: Requirements 3.1, 3.2, 3.3, 3.4**

- [x] 2. Implement ConfigLoader for I/O operations
  - Create `internal/config/loader.go` with ConfigLoader struct
  - Implement LoadFromFile using FileSystem.ReadFile
  - Implement LoadFromBytes for parsing YAML data
  - Implement SaveToFile using FileSystem.WriteFileAtomic
  - Implement MarshalConfig and UnmarshalConfig using gopkg.in/yaml.v3
  - _Requirements: 2.1, 10.1, 10.2, 10.3_

- [ ]* 2.1 Write property test for configuration serialization
  - **Property 22: Configuration round-trip preserves data**
  - **Validates: Requirements 10.1, 10.2, 10.3**

- [ ]* 2.2 Write unit tests for YAML parsing edge cases
  - Test special characters in configuration values
  - Test nested structure serialization
  - Test empty and nil field handling
  - _Requirements: 10.3, 10.4_

- [x] 3. Implement unified ConfigurationManager core operations
  - Create `internal/config/manager.go` with ConfigurationManager struct
  - Implement NewConfigurationManager with dependency injection
  - Implement Load method with cache-first strategy
  - Implement Save method with validation and atomic writes
  - Implement Validate method delegating to ValidationEngine
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 2.1, 3.1, 3.2, 3.3, 4.1, 4.2_

- [ ]* 3.1 Write property test for Load operation
  - **Property 5: Cache checked before disk read**
  - **Property 6: Disk loads populate cache**
  - **Property 20: File not found returns FileError**
  - **Validates: Requirements 3.1, 3.2, 9.1**

- [ ]* 3.2 Write property test for Save operation
  - **Property 1: Save failure preserves original file**
  - **Property 4: Backup created before overwrite**
  - **Property 7: Save invalidates cache**
  - **Property 10: Validation before save**
  - **Validates: Requirements 2.2, 2.5, 3.3, 4.2**

- [ ]* 3.3 Write property test for concurrent operations
  - **Property 2: Concurrent saves are atomic**
  - **Validates: Requirements 2.3**

- [ ]* 3.4 Write property test for corrupted file detection
  - **Property 3: Corrupted files are detected**
  - **Validates: Requirements 2.4**

- [ ]* 3.5 Write property test for validation integration
  - **Property 11: Invalid configs return validation errors**
  - **Validates: Requirements 4.3, 4.4, 9.2**

- [x] 4. Checkpoint - Core operations complete
  - Ensure all tests pass for ConfigCache, ConfigLoader, and ConfigurationManager
  - Verify integration with Phase 1 (FileSystem, PathResolver) and Phase 2 (ValidationEngine)
  - Ask the user if questions arise

- [x] 5. Implement List and Delete operations
  - Implement List method using PathResolver to find all configs
  - Implement List with organization filtering
  - Implement Delete method with backup creation
  - Add cache invalidation to Delete
  - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 6.1, 6.2, 6.3, 6.4, 6.5_

- [ ]* 5.1 Write property test for List operation
  - **Property 12: List returns all clusters**
  - **Property 13: Organization filter works correctly**
  - **Validates: Requirements 5.1, 5.2**

- [ ]* 5.2 Write unit tests for List edge cases
  - Test empty directory returns empty list
  - Test non-existent directory returns empty list
  - _Requirements: 5.3, 5.4_

- [ ]* 5.3 Write property test for Delete operation
  - **Property 14: Delete removes file and invalidates cache**
  - **Property 15: Delete non-existent cluster fails**
  - **Property 16: Delete creates backup**
  - **Validates: Requirements 6.1, 6.2, 6.3, 6.4**

- [x] 6. Implement ConfigBuilder for fluent API
  - Create `internal/config/builder.go` with ConfigBuilder struct
  - Implement NewBuilder and BuildFrom methods in ConfigurationManager
  - Implement WithProvider, WithRegion, WithOrganization methods
  - Implement WithDefaults method for provider-specific defaults
  - Implement Validate and Build methods
  - Implement Save method that builds and saves
  - _Requirements: 1.5, 7.1, 7.2, 7.3, 7.4, 7.5_

- [ ]* 6.1 Write property test for builder validation
  - **Property 17: Builder validates on Build**
  - **Validates: Requirements 7.3**

- [ ]* 6.2 Write unit tests for builder integration
  - Test builder uses ConfigurationManager's validation
  - Test builder uses ConfigurationManager's path resolution
  - Test method chaining works correctly
  - _Requirements: 7.4, 7.5_

- [x] 7. Implement error handling with structured errors
  - Create error types in `internal/config/errors.go`
  - Implement FileError, ValidationError, PathError, ParseError
  - Update all ConfigurationManager methods to return structured errors
  - Ensure errors include context (file paths, line numbers, field names)
  - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5_

- [ ]* 7.1 Write property tests for error handling
  - **Property 20: File not found returns FileError**
  - **Property 21: Path resolution failure returns PathError**
  - **Property 22: YAML parse failure returns ParseError**
  - **Validates: Requirements 9.1, 9.3, 9.4**

- [ ]* 7.2 Write unit tests for error context
  - Test FileError includes file path
  - Test ValidationError includes field name and details
  - Test PathError includes attempted path
  - Test ParseError includes line and column numbers
  - _Requirements: 9.1, 9.2, 9.3, 9.4_

- [x] 8. Checkpoint - Unified manager complete
  - Ensure all ConfigurationManager operations work correctly
  - Verify all property tests pass (minimum 100 iterations each)
  - Run performance benchmarks to verify 40% cache improvement
  - Ask the user if questions arise

- [x] 9. Implement migration scanner tool
  - Create `internal/config/migration/scanner.go` with MigrationScanner
  - Implement Scan method to find files using legacy config patterns
  - Search for `config.Load(`, `config.Save(`, `config.Validate(` patterns
  - Create MigrationReport struct with file lists and counts
  - Implement GenerateReport method to create markdown output
  - _Requirements: 8.5, 12.1, 12.2, 12.5_

- [ ]* 9.1 Write property test for migration scanner
  - **Property 18: Migration scanner finds legacy patterns**
  - **Property 19: Migration report accuracy**
  - **Validates: Requirements 8.5, 12.1, 12.2, 12.5**

- [ ]* 9.2 Write unit tests for scanner patterns
  - Test scanner finds config.Load calls
  - Test scanner finds config.Save calls
  - Test scanner finds config.Validate calls
  - Test scanner ignores comments and strings
  - _Requirements: 12.1_

- [ ] 10. Create migration documentation
  - Create `docs/migration/config-unification.md` with migration guide
  - Document before/after code examples for each operation
  - Create migration checklist for all 45+ files
  - Document common migration patterns and gotchas
  - Include examples of updating Load, Save, Validate, List, Delete calls
  - _Requirements: 8.1, 8.2, 12.3_

- [x] 11. Run migration scanner and generate initial report
  - Execute MigrationScanner on the codebase
  - Generate migration report showing all files to update
  - Review report to identify migration batches
  - Create tracking document for migration progress
  - _Requirements: 12.2, 12.5_

- [x] 12. Migrate command layer files (cmd/)
  - [x] 12.1 Migrate cmd/cluster_init.go
    - Replace config.Load with manager.Load
    - Replace config.Save with manager.Save
    - Add context parameter to all calls
    - _Requirements: 8.3_
  
  - [x] 12.2 Migrate cmd/cluster_validate.go
    - Replace config.Load with manager.Load
    - Replace config.Validate with manager.Validate
    - Add context parameter to all calls
    - _Requirements: 8.3_
  
  - [x] 12.3 Migrate cmd/cluster_setup.go
    - Replace config.Load with manager.Load
    - Add context parameter to all calls
    - _Requirements: 8.3_
  
  - [x] 12.4 Migrate cmd/cluster_bootstrap.go
    - Replace config.Load with manager.Load
    - Add context parameter to all calls
    - _Requirements: 8.3_
  
  - [x] 12.5 Migrate cmd/cluster_list.go
    - Replace config.List with manager.List
    - Add context parameter to all calls
    - _Requirements: 8.3_
  
  - [x] 12.6 Migrate cmd/config_*.go files (8 files)
    - Replace all config.Load/Save/Validate calls
    - Add context parameter to all calls
    - _Requirements: 8.3_

- [ ]* 12.7 Run tests after command layer migration
  - Verify all command tests pass
  - Run integration tests for CLI commands
  - _Requirements: 12.4_

- [x] 13. Checkpoint - Command layer migrated
  - Ensure all cmd/ files use new ConfigurationManager
  - Verify no legacy config calls remain in cmd/
  - Run full test suite
  - Ask the user if questions arise

- [x] 14. Migrate service layer files (internal/cluster/)
  - [x] 14.1 Migrate internal/cluster/init_service.go
    - Replace config.Load/Save with manager.Load/Save
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 14.2 Migrate internal/cluster/validate_service.go
    - Replace config.Load/Validate with manager.Load/Validate
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 14.3 Migrate internal/cluster/setup_service.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 14.4 Migrate internal/cluster/bootstrap_service.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_

- [ ]* 14.5 Run tests after service layer migration
  - Verify all service tests pass
  - Run integration tests for services
  - _Requirements: 12.4_

- [x] 15. Migrate GitOps layer files (internal/gitops/)
  - [x] 15.1 Migrate internal/gitops/generator.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 15.2 Migrate internal/gitops/workspace.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 15.3 Migrate internal/gitops/pipeline.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_

- [ ]* 15.4 Run tests after GitOps layer migration
  - Verify all GitOps tests pass
  - Run integration tests for GitOps generation
  - _Requirements: 12.4_

- [x] 16. Migrate SOPS layer files (internal/sops/)
  - [x] 16.1 Migrate internal/sops/manager.go
    - Replace config.Load/Save with manager.Load/Save
    - Add context parameter propagation
    - _Requirements: 8.3_
  
  - [x] 16.2 Migrate internal/sops/git.go
    - Replace config.Load with manager.Load
    - Add context parameter propagation
    - _Requirements: 8.3_

- [ ]* 16.3 Run tests after SOPS layer migration
  - Verify all SOPS tests pass
  - Run integration tests for secrets management
  - _Requirements: 12.4_

- [x] 17. Checkpoint - All layers migrated
  - Ensure all 45+ files have been migrated
  - Verify no legacy config calls remain in codebase
  - Run full test suite including integration tests
  - Ask the user if questions arise

- [x] 18. Remove legacy configuration code
  - Remove old Load, Save, Validate functions from internal/config/config.go
  - Remove old ConfigurationManager from internal/config/old_manager.go (if exists)
  - Remove Builder from internal/config/old_builder.go (if exists)
  - Remove abandoned internal/core/config/ directory
  - Update all imports to use new ConfigurationManager
  - _Requirements: 8.3_

- [x] 19. Run performance benchmarks
  - Benchmark Load operation with and without cache
  - Verify 40% performance improvement for cached loads
  - Benchmark Save operation with atomic writes
  - Benchmark concurrent operations
  - Document performance results
  - _Requirements: 3.5_

- [ ]* 19.1 Write benchmark for cache performance
  - **Property 9: Cached loads are faster**
  - **Validates: Requirements 3.5**

- [ ] 20. Update documentation
  - Update architecture documentation to reflect unified system
  - Update developer guide with new ConfigurationManager usage
  - Add migration guide to docs/
  - Update API reference documentation
  - Add examples for common configuration operations
  - _Requirements: 8.1, 12.3_

- [x] 21. Final checkpoint - Phase 3 complete
  - Verify all 45+ files migrated successfully
  - Confirm all tests pass (unit, property, integration)
  - Verify performance benchmarks meet 40% improvement target
  - Confirm no legacy configuration code remains
  - Ensure documentation is complete and accurate
  - Run full test suite one final time
  - Ask the user if questions arise

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation at key milestones
- Property tests validate universal correctness properties (minimum 100 iterations each)
- Unit tests validate specific examples, edge cases, and integration points
- Migration is done in batches by layer to minimize risk
- All property tests must include comment tags: `// Feature: phase-3-configuration-unification, Property N: <property text>`
