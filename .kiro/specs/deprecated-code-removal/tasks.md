# Implementation Plan: Deprecated Code Removal

## Overview

This plan completes the removal of deprecated code from opencenter-cli. Part 1 (unused deprecated code) is complete. This plan focuses on Part 2 (removing deprecated persistence functions) and Part 3 (evaluating and removing validateServiceSecretsSimple, deferring TemplateValidator).

The approach is surgical: verify no production code uses deprecated functions, update test files, remove deprecated functions and helpers, verify all tests pass, and document Part 3 decisions.

## Tasks

- [x] 1. Pre-removal verification and analysis
  - Verify no production code calls deprecated persistence functions
  - Identify all helper functions only used by deprecated functions
  - Document current test baseline (run full test suite)
  - _Requirements: 1.1-1.10, 7.1-7.5_

- [x] 2. Update migration scanner test data
  - [x] 2.1 Update scanner_test.go example code to show modern API patterns
    - Update config.Load() examples to show manager.Load(ctx, name)
    - Update config.Save() examples to show manager.Save(ctx, cfg)
    - Update config.Validate() examples to show manager.Validate(ctx, cfg)
    - Ensure examples include context.Background() and manager initialization
    - _Requirements: 2.2, 4.1, 4.2, 4.3_
  
  - [x] 2.2 Update scanner.go migration instructions
    - Update GenerateReport() to show full modern API context
    - Include manager initialization in code examples
    - Show context.Background() usage
    - _Requirements: 4.1, 4.2, 4.3_
  
  - [x] 2.3 Run scanner tests to verify example code updates
    - Run tests in internal/config/migration/scanner_test.go
    - Verify scanner still detects deprecated patterns correctly
    - Ensure updated examples are syntactically correct
    - _Requirements: 4.5_

- [x] 3. Checkpoint - Verify scanner tests pass
  - Ensure all scanner tests pass, ask the user if questions arise.

- [x] 4. Remove deprecated persistence functions
  - [x] 4.1 Remove unused functions (zero callers)
    - Remove ListClusters() function
    - Remove SetActiveCluster() function
    - Remove GetActiveCluster() function
    - _Requirements: 1.8, 1.9, 1.10_
  
  - [x] 4.2 Remove core persistence functions
    - Remove Save(cfg Config) error function
    - Remove Load(name string) (Config, error) function
    - Remove Validate(cfg Config) []error function
    - _Requirements: 1.1, 1.2, 1.3_
  
  - [x] 4.3 Remove path and config generation functions
    - Remove ConfigPath(name string) (string, error) function
    - Remove GenerateCompleteConfig(name string) (Config, error) function
    - Remove GenerateCompleteConfigYAML(name string) ([]byte, error) function
    - Remove SaveDebugConfig(clusterName, gitDir string) error function
    - _Requirements: 1.4, 1.5, 1.6, 1.7_
  
  - [x] 4.4 Remove helper functions only used by deprecated functions
    - Remove mergeYAMLMaps(base, override map[string]any) map[string]any
    - Remove cleanEmptyValues(m map[string]any)
    - Remove isEmpty(v any) bool
    - Remove getConfigPathForSave(cfg Config) (string, error)
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5_
  
  - [x] 4.5 Remove deprecation warnings and comments
    - Remove logDeprecationWarning calls for removed functions
    - Remove deprecation comments for removed functions
    - _Requirements: 8.1, 8.2_

- [x] 5. Checkpoint - Verify build succeeds
  - Run `mise run build` to verify compilation
  - Ensure no compilation errors
  - Ask the user if questions arise.
  - _Requirements: 3.3_

- [x] 6. Part 3: Remove validateServiceSecretsSimple
  - [x] 6.1 Remove validateServiceSecretsSimple function
    - Remove function from internal/config/config.go
    - Remove deprecation comment
    - Function has zero callers, safe to remove
    - _Requirements: 5.1, 5.2, 5.3, 5.4_
  
  - [x] 6.2 Document validateServiceSecretsSimple removal decision
    - Update docs/architecture-review/DEPRECATED_CODE_REMOVAL_STATUS.md
    - Document that function was removed (zero callers)
    - Note that if similar validation needed, implement in ValidationEngine
    - _Requirements: 9.1, 9.4, 9.5_

- [x] 7. Part 3: Document TemplateValidator deferral decision
  - [x] 7.1 Analyze TemplateValidator usage and impact
    - Document extensive usage throughout template engine
    - Note that removal requires broader template engine refactoring
    - Assess migration complexity and effort
    - _Requirements: 6.1, 6.2, 6.3_
  
  - [x] 7.2 Document TemplateValidator deferral decision
    - Update docs/architecture-review/DEPRECATED_CODE_REMOVAL_STATUS.md
    - Document decision to defer to separate task
    - Explain rationale: requires template engine refactoring
    - Create recommendation for follow-up task
    - _Requirements: 6.4, 6.5, 9.2, 9.3, 9.5_

- [x] 8. Checkpoint - Run full test suite
  - Run `mise run test` to verify all unit tests pass
  - Run `mise run godog` to verify all BDD tests pass
  - Ensure zero test failures
  - Ask the user if questions arise.
  - _Requirements: 3.1, 3.2, 3.4, 3.5_

- [x] 9. Manual CLI command verification
  - [x] 9.1 Test cluster initialization command
    - Run `opencenter cluster init test-cleanup-cluster`
    - Verify command succeeds
    - Verify configuration file is created
    - _Requirements: 10.2_
  
  - [x] 9.2 Test cluster validation command
    - Run `opencenter cluster validate test-cleanup-cluster`
    - Verify command succeeds
    - Verify validation runs correctly
    - _Requirements: 10.3_
  
  - [x] 9.3 Test cluster info command
    - Run `opencenter cluster info test-cleanup-cluster`
    - Verify command succeeds
    - Verify cluster information is displayed
    - _Requirements: 10.4_
  
  - [x] 9.4 Test cluster list command
    - Run `opencenter cluster list`
    - Verify command succeeds
    - Verify clusters are listed
    - _Requirements: 10.1, 10.5_

- [x] 10. Final verification and documentation
  - [x] 10.1 Verify no deprecation warnings remain
    - Search for "Deprecated:" in internal/config/persistence.go (expect 0)
    - Search for logDeprecationWarning calls (expect 0 for removed functions)
    - Verify no comments reference removed functions
    - _Requirements: 8.1, 8.2, 8.3_
  
  - [x] 10.2 Update status document with completion
    - Mark Part 2 as COMPLETE in DEPRECATED_CODE_REMOVAL_STATUS.md
    - Mark Part 3 as COMPLETE with decisions documented
    - Update success criteria checklist
    - Document lines removed and files modified
    - _Requirements: 9.5_
  
  - [x] 10.3 Final test suite run
    - Run `mise run build` one final time
    - Run `mise run test` one final time
    - Run `mise run godog` one final time
    - Verify all pass with zero failures
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5_

- [x] 11. Final checkpoint - Confirm completion
  - Ensure all tests pass, all CLI commands work, and documentation is updated
  - Ask the user if questions arise or if ready to close the spec.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster completion
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Manual CLI testing verifies backward compatibility
- Part 3 evaluation is integrated into the implementation flow
- Status document is updated throughout to track progress
