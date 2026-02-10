# Requirements Document

## Introduction

This feature completes the removal of all deprecated code from the opencenter-cli codebase following the completion of Phase 1-4 modernization. Part 1 (unused deprecated code) is complete. This spec focuses on Part 2 (deprecated persistence functions) and Part 3 (evaluation of remaining deprecated code).

## Glossary

- **ConfigurationManager**: The modern configuration management system that replaced deprecated persistence functions
- **PathResolver**: The modern path resolution system that replaced deprecated ConfigPath function
- **ValidationEngine**: The modern validation system in internal/core/validation
- **Deprecated_Function**: A function marked with deprecation warnings that should be removed
- **Migration_Scanner**: Tool in internal/config/migration that detects deprecated function usage

## Requirements

### Requirement 1: Remove Deprecated Persistence Functions

**User Story:** As a developer, I want all deprecated persistence functions removed from the codebase, so that the code is clean and maintainable without legacy cruft.

#### Acceptance Criteria

1. THE System SHALL remove the Save(cfg Config) error function from internal/config/persistence.go
2. THE System SHALL remove the Load(name string) (Config, error) function from internal/config/persistence.go
3. THE System SHALL remove the Validate(cfg Config) []error function from internal/config/persistence.go
4. THE System SHALL remove the ConfigPath(name string) (string, error) function from internal/config/persistence.go
5. THE System SHALL remove the GenerateCompleteConfig(name string) (Config, error) function from internal/config/persistence.go
6. THE System SHALL remove the GenerateCompleteConfigYAML(name string) ([]byte, error) function from internal/config/persistence.go
7. THE System SHALL remove the SaveDebugConfig(clusterName, gitDir string) error function from internal/config/persistence.go
8. THE System SHALL remove the ListClusters() ([]string, error) function from internal/config/persistence.go
9. THE System SHALL remove the SetActiveCluster(name string) error function from internal/config/persistence.go
10. THE System SHALL remove the GetActiveCluster() (string, error) function from internal/config/persistence.go

### Requirement 2: Update Test Files to Use Modern APIs

**User Story:** As a developer, I want all test files updated to use ConfigurationManager, so that tests validate the modern API surface.

#### Acceptance Criteria

1. WHEN test files reference deprecated functions, THE System SHALL update them to use ConfigurationManager methods
2. WHEN internal/config/migration/scanner_test.go contains deprecated function examples, THE System SHALL update the test data to show modern API usage
3. WHEN tests need to save configurations, THE System SHALL use manager.Save(ctx, cfg) or SaveWithoutValidation for test fixtures
4. WHEN tests need to load configurations, THE System SHALL use manager.Load(ctx, name) or LoadWithoutValidation for incomplete configs
5. WHEN tests need to validate configurations, THE System SHALL use manager.Validate(ctx, cfg)

### Requirement 3: Maintain Test Coverage and Passing Tests

**User Story:** As a developer, I want all tests to continue passing after deprecated code removal, so that I can be confident the refactoring is correct.

#### Acceptance Criteria

1. WHEN deprecated functions are removed, THE System SHALL ensure all unit tests pass
2. WHEN deprecated functions are removed, THE System SHALL ensure all integration tests pass
3. WHEN deprecated functions are removed, THE System SHALL ensure the build succeeds with no compilation errors
4. WHEN running mise run test, THE System SHALL report zero test failures
5. WHEN running mise run godog, THE System SHALL report zero BDD test failures

### Requirement 4: Update Migration Scanner Test Data

**User Story:** As a developer, I want the migration scanner test data updated to reflect modern APIs, so that the scanner correctly identifies deprecated usage patterns.

#### Acceptance Criteria

1. WHEN scanner_test.go contains example code with config.Save(), THE System SHALL update examples to show manager.Save(ctx, cfg)
2. WHEN scanner_test.go contains example code with config.Load(), THE System SHALL update examples to show manager.Load(ctx, name)
3. WHEN scanner_test.go contains example code with config.Validate(), THE System SHALL update examples to show manager.Validate(ctx, cfg)
4. WHEN scanner_test.go contains example code with config.ConfigPath(), THE System SHALL update examples to show pathResolver.ResolveClusterPaths()
5. THE System SHALL ensure scanner tests continue to validate detection of deprecated patterns in old code

### Requirement 5: Evaluate validateServiceSecretsSimple Migration

**User Story:** As a developer, I want to evaluate whether validateServiceSecretsSimple should be migrated to ValidationEngine, so that I can determine if this is part of the current scope or a future task.

#### Acceptance Criteria

1. THE System SHALL analyze the usage of validateServiceSecretsSimple in internal/config/config.go
2. THE System SHALL determine if validateServiceSecretsSimple is only used internally or exposed to external callers
3. THE System SHALL assess the complexity of migrating validateServiceSecretsSimple to ValidationEngine
4. THE System SHALL document whether migration should be included in this spec or deferred to a separate task
5. IF migration is deferred, THEN THE System SHALL document the rationale and create a follow-up task recommendation

### Requirement 6: Evaluate TemplateValidator Interface Migration

**User Story:** As a developer, I want to evaluate whether the TemplateValidator interface should be removed or refactored, so that I can determine if this is part of the current scope or requires a separate template engine refactoring task.

#### Acceptance Criteria

1. THE System SHALL analyze the usage of TemplateValidator interface in internal/util/template/interfaces.go
2. THE System SHALL determine how extensively TemplateValidator is used throughout the template engine
3. THE System SHALL assess whether removing TemplateValidator requires broader template engine refactoring
4. THE System SHALL document whether removal should be included in this spec or deferred to a separate task
5. IF removal is deferred, THEN THE System SHALL document the rationale and create a follow-up task recommendation

### Requirement 7: Remove Helper Functions Only Used by Deprecated Functions

**User Story:** As a developer, I want helper functions that only support deprecated functions to be removed, so that the codebase doesn't contain orphaned utility code.

#### Acceptance Criteria

1. WHEN a helper function is only called by deprecated functions, THE System SHALL remove the helper function
2. WHEN mergeYAMLMaps is only used by GenerateCompleteConfig, THE System SHALL remove mergeYAMLMaps
3. WHEN cleanEmptyValues is only used by deprecated save functions, THE System SHALL remove cleanEmptyValues
4. WHEN isEmpty is only used by cleanEmptyValues, THE System SHALL remove isEmpty
5. WHEN getConfigPathForSave is only used by deprecated save functions, THE System SHALL remove getConfigPathForSave

### Requirement 8: Verify No Deprecation Warnings Remain

**User Story:** As a developer, I want to verify that no deprecation warnings remain in the codebase, so that I can confirm the cleanup is complete.

#### Acceptance Criteria

1. WHEN searching the codebase for "Deprecated:", THE System SHALL find zero occurrences in internal/config/persistence.go
2. WHEN searching the codebase for logDeprecationWarning calls, THE System SHALL find zero occurrences related to removed functions
3. THE System SHALL ensure no comments reference deprecated functions that have been removed
4. THE System SHALL ensure no documentation references deprecated functions that have been removed
5. THE System SHALL update any migration guides to remove references to deprecated functions

### Requirement 9: Document Removal Decisions for Part 3

**User Story:** As a developer, I want clear documentation of decisions made about Part 3 deprecated code, so that future developers understand why certain code was or wasn't removed.

#### Acceptance Criteria

1. THE System SHALL document the decision for validateServiceSecretsSimple in the status document
2. THE System SHALL document the decision for TemplateValidator in the status document
3. IF code is deferred, THEN THE System SHALL document the complexity and scope reasons
4. IF code is removed, THEN THE System SHALL document the migration approach used
5. THE System SHALL update docs/architecture-review/DEPRECATED_CODE_REMOVAL_STATUS.md with final status

### Requirement 10: Maintain Backward Compatibility in Public APIs

**User Story:** As a user of opencenter-cli, I want public CLI commands to continue working after deprecated code removal, so that my workflows are not disrupted.

#### Acceptance Criteria

1. WHEN deprecated internal functions are removed, THE System SHALL ensure no public CLI commands are affected
2. WHEN running opencenter cluster init, THE System SHALL successfully create a cluster configuration
3. WHEN running opencenter cluster validate, THE System SHALL successfully validate configurations
4. WHEN running opencenter cluster info, THE System SHALL successfully display cluster information
5. THE System SHALL ensure all cmd/* files compile and function correctly after changes
