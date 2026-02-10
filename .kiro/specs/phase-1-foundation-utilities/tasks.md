# Implementation Plan: Phase 1 Foundation Utilities

## Overview

This implementation plan breaks down the Phase 1 Foundation utilities into discrete, incremental coding tasks. Each task builds on previous work and includes testing to validate functionality early. The plan follows a two-week structure: Week 1 focuses on utility consolidation (file operations, error handling, code cleanup), and Week 2 focuses on test infrastructure and DI cleanup.

## Tasks

- [x] 1. Implement FileSystem wrapper with atomic operations
  - [x] 1.1 Create FileSystem interface and DefaultFileSystem implementation
    - Create `internal/util/fs/wrapper.go` with FileSystem interface
    - Implement DefaultFileSystem struct with ErrorHandler dependency
    - Implement ReadFile, WriteFile, Exists, MkdirAll, Remove, Stat methods
    - Add generateRandomString helper for temporary file names
    - _Requirements: 1.1, 1.4, 1.5_
  
  - [x] 1.2 Implement WriteFileAtomic with cleanup on failure
    - Implement WriteFileAtomic method with temp file creation
    - Add atomic rename operation
    - Add cleanup logic for temp files on failure
    - _Requirements: 1.2, 1.3_
  
  - [ ]* 1.3 Write property test for atomic write operations
    - **Property 1: Atomic Write Operations**
    - **Validates: Requirements 1.2, 1.3**
    - Create `internal/util/fs/wrapper_property_test.go`
    - Use gopter to generate random paths and data
    - Verify atomicity: complete write OR unchanged file with no temp files
    - Test concurrent writes to same path
    - _Requirements: 1.2, 1.3_
  
  - [ ]* 1.4 Write property test for error context consistency
    - **Property 2: Error Context Consistency**
    - **Validates: Requirements 1.4, 1.5**
    - Generate random file operations that fail
    - Verify all errors are StructuredError with operation, path, and cause
    - _Requirements: 1.4, 1.5_
  
  - [ ]* 1.5 Write property test for thread-safe concurrent operations
    - **Property 3: Thread-Safe Concurrent Operations**
    - **Validates: Requirements 1.7**
    - Generate random sets of file operations on different files
    - Execute concurrently with goroutines
    - Run with Go race detector enabled
    - Verify no race conditions and correct data
    - _Requirements: 1.7_
  
  - [ ]* 1.6 Write unit tests for FileSystem wrapper
    - Test ReadFile with valid and invalid paths
    - Test WriteFile with various permissions
    - Test Exists for existing and non-existing paths
    - Test MkdirAll with nested directories
    - Test error handling for permission denied, disk full scenarios
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_
  
  - [ ]* 1.7 Write benchmark tests for FileSystem performance
    - Benchmark WriteFileAtomic vs os.WriteFile
    - Verify <5% performance overhead
    - _Requirements: 1.6_

- [x] 2. Implement structured error handling system
  - [x] 2.1 Create StructuredError type and ErrorHandler interface
    - Create `internal/util/errors/structured.go`
    - Define ErrorType constants (ValidationError, FileError, ConfigError, OperationalError)
    - Implement StructuredError struct with all required fields
    - Implement Error(), Unwrap(), and Is() methods
    - _Requirements: 2.1, 2.2, 2.6, 2.7_
  
  - [x] 2.2 Implement ErrorHandler with error creation functions
    - Define ErrorHandler interface
    - Implement DefaultErrorHandler struct
    - Implement CreateValidationError function
    - Implement CreateFileError function with retryability detection
    - Implement CreateConfigError function
    - Implement Wrap function for general errors
    - Add isRetryableFileError helper function
    - _Requirements: 2.3, 2.4, 2.5_
  
  - [ ]* 2.3 Write property test for error retryability determination
    - **Property 4: Error Retryability Determination**
    - **Validates: Requirements 2.5**
    - Generate errors with retryable patterns
    - Generate errors with non-retryable patterns
    - Verify Retryable field is set correctly
    - _Requirements: 2.5_
  
  - [ ]* 2.4 Write property test for error formatting completeness
    - **Property 5: Error Formatting Completeness**
    - **Validates: Requirements 2.6**
    - Generate random StructuredErrors with various field combinations
    - Verify formatted string includes all non-empty fields
    - Test with nested error causes
    - _Requirements: 2.6_
  
  - [ ]* 2.5 Write property test for error wrapping preservation
    - **Property 6: Error Wrapping Preservation**
    - **Validates: Requirements 2.7**
    - Generate random error chains of varying depths
    - Unwrap repeatedly and verify each level
    - Verify root error is reached
    - _Requirements: 2.7_
  
  - [ ]* 2.6 Write unit tests for StructuredError
    - Test Error() formatting with all field combinations
    - Test Unwrap() with single and chained errors
    - Test Is() for error type matching
    - Test CreateValidationError with suggestions
    - Test CreateFileError with various causes
    - Test CreateConfigError
    - Test Wrap function
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7_

- [x] 3. Remove orphaned code and document decision
  - [x] 3.1 Remove internal/core/config/ directory
    - Verify zero references to internal/core/config/
    - Delete internal/core/config/ directory
    - Clean up any import statements
    - _Requirements: 3.1, 3.3_
  
  - [x] 3.3 Update architecture documentation
    - Update architecture diagrams to remove orphaned components
    - Update package documentation
    - _Requirements: 3.5, 7.6_
  
  - [x] 3.4 Verify all tests pass after cleanup
    - Run full test suite
    - Verify no broken imports
    - Verify build succeeds
    - _Requirements: 3.3, 3.4_

- [x] 4. Checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 5. Consolidate test helpers into single package
  - [x] 5.1 Create unified test helper package
    - Create `internal/testing/helpers.go`
    - Implement CreateTempConfig function
    - Implement CreateTempDir function
    - Implement AssertNoError function
    - Implement AssertError function
    - Implement AssertEqual function
    - Implement AssertFileExists function
    - Implement AssertFileNotExists function
    - _Requirements: 4.1, 4.2, 4.3_
  
  - [ ]* 5.2 Write property test for test helper file creation
    - **Property 7: Test Helper File Creation**
    - **Validates: Requirements 4.4**
    - Generate random file structures (flat and nested)
    - Call CreateTempDir with file map
    - Verify all files exist with correct content
    - Verify parent directories created
    - Test edge cases (empty map, deeply nested paths, special characters)
    - _Requirements: 4.4_
  
  - [ ]* 5.3 Write unit tests for test helpers
    - Test CreateTempConfig with valid YAML content
    - Test CreateTempDir with various file structures
    - Test cleanup behavior (verify t.TempDir() usage)
    - Test assertion functions
    - _Requirements: 4.3, 4.5, 4.6_
  
  - [x] 5.4 Migrate existing test code to use consolidated helpers
    - Identify all duplicate test helper implementations
    - Replace with calls to consolidated helpers
    - Remove duplicate implementations
    - Verify all tests still pass
    - _Requirements: 4.7, 4.8_

- [x] 6. Implement unified DI container initialization
  - [x] 6.1 Create DI container implementation
    - Create `internal/di/container.go`
    - Define Container interface (Singleton, Get, Initialize methods)
    - Implement defaultContainer struct with service registry
    - Implement Singleton method for service registration
    - Implement Get method for service retrieval
    - Implement Initialize method for dependency resolution
    - Add invokeProvider helper using reflection
    - _Requirements: 5.1_
  
  - [x] 6.2 Create SetupContainer function with service registrations
    - Create `internal/di/setup.go`
    - Implement SetupContainer function accepting baseDir parameter
    - Register ErrorHandler as singleton
    - Register FileSystem as singleton (depends on ErrorHandler)
    - Register PathResolver as singleton (depends on FileSystem)
    - Call Initialize() before returning
    - _Requirements: 5.2, 5.3, 5.4, 5.5_
  
  - [ ]* 6.3 Write property test for DI container error handling
    - **Property 8: DI Container Error Handling**
    - **Validates: Requirements 5.7**
    - Generate invalid service registrations (duplicate names, nil providers)
    - Verify errors are returned (not panics)
    - Verify error messages are descriptive
    - Test with provider functions that return errors
    - _Requirements: 5.7_
  
  - [ ]* 6.4 Write unit tests for DI container
    - Test service registration and retrieval
    - Test dependency resolution order
    - Test error handling for duplicate registrations
    - Test error handling for missing dependencies
    - Test error handling for initialization failures
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.7, 5.8_
  
  - [x] 6.5 Remove duplicate DI initialization from cmd/root.go
    - Update cmd/root.go to use SetupContainer
    - Remove duplicate service registration code
    - Verify CLI commands still work
    - _Requirements: 5.6_
  
  - [ ]* 6.6 Write integration tests for DI container setup
    - Test full container initialization with all services
    - Test service retrieval and usage
    - Test error propagation through dependency chain
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5_

- [x] 7. Final checkpoint and documentation
  - [x] 7.1 Run full test suite with coverage
    - Run `mise run test -cover`
    - Verify >95% coverage for FileSystem
    - Verify >80% coverage for StructuredError
    - Verify >80% coverage for Test Helpers
    - Verify >80% coverage for DI Container
    - _Requirements: 6.1, 6.2, 6.3, 6.4_
 
  - [x] 7.3 Verify code reduction and build time
    - Run `git diff --stat` to measure LOC reduction
    - Verify ≥500 LOC reduction
    - Run CI build and verify <45s build time
    - _Requirements: 6.6, 6.7_
  
  - [x] 7.4 Update package documentation
    - Add package documentation to internal/util/fs
    - Add usage examples for FileSystem methods
    - Add package documentation to internal/util/errors
    - Add documentation for when to use each error type
    - Add package documentation to internal/testing
    - Add examples for common test setup patterns
    - Add package documentation to internal/di
    - Add documentation for service registration pattern
    - _Requirements: 7.1, 7.2, 7.3, 7.4_
  
  - [x] 7.6 Final verification
    - Ensure all existing tests pass (100%)
    - Verify no broken imports
    - Verify no orphaned code remains
    - Run `mise run test -race` to check for race conditions
    - _Requirements: 6.8_

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Week 1 focuses on core utilities (FileSystem, ErrorHandler, cleanup)
- Week 2 focuses on test infrastructure and DI consolidation
- All code uses Go 1.25.2 with standard library and gopter for property testing
- Follow mise workflow: always use `mise run` commands for build, test, and validation
