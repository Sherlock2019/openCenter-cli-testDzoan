# Requirements Document: Phase 2 Validation Consolidation

## Introduction

This document specifies the requirements for consolidating validation logic scattered across 15+ packages into a unified ValidationEngine. This consolidation eliminates 1,800 lines of duplicate validation code, provides consistent error messages with actionable suggestions, and establishes a single source of truth for all validation rules.

Phase 2 builds on the foundation utilities from Phase 1 (StructuredError and FileSystem) and provides the validation infrastructure that Phase 3 (Configuration) and Phase 4 (Services) will depend on. Without this consolidation, validation inconsistencies will continue to cause data corruption, security vulnerabilities, and user confusion.

## Glossary

- **ValidationEngine**: Central service that manages validator registration and executes validation operations
- **Validator**: Component that implements validation logic for a specific domain (cluster names, networking, SOPS keys, etc.)
- **ValidationResult**: Data structure containing validation outcome, errors, warnings, and suggestions
- **ValidationError**: Structured error specific to validation failures, including field context and suggestions
- **Validator_Registry**: Internal map of registered validators accessible by name
- **Validation_Context**: Contextual information passed through validation chain (user, operation, environment)
- **Feature_Flag**: Configuration toggle enabling gradual rollout of new validation system
- **Parallel_Validation**: Running both old and new validation systems simultaneously for comparison
- **Validation_Rule**: Specific constraint that data must satisfy (e.g., "cluster name must be lowercase")
- **Security_Validator**: Validator that checks for security issues (path traversal, command injection, etc.)

## Requirements

### Requirement 1: ValidationEngine Core Implementation

**User Story:** As a developer, I want a central validation engine, so that all validation logic is registered in one place and executed consistently across the application.

#### Acceptance Criteria

1. THE ValidationEngine SHALL provide a Register method that accepts a Validator and stores it in the validator registry
2. WHEN Register is called with a validator name that already exists, THE ValidationEngine SHALL return an error indicating duplicate registration
3. THE ValidationEngine SHALL provide a Validate method that accepts a context, validator name, and data to validate
4. WHEN Validate is called with a non-existent validator name, THE ValidationEngine SHALL return an error indicating validator not found
5. THE ValidationEngine SHALL provide a ValidateAll method that executes all registered validators against provided data
6. THE ValidationEngine SHALL use sync.RWMutex to ensure thread-safe access to the validator registry
7. THE ValidationEngine SHALL aggregate validation results from multiple validators into a single ValidationResult
8. WHEN any validator returns errors, THE ValidationEngine SHALL set the aggregated result Valid field to false
9. THE ValidationEngine SHALL pass context through the entire validation chain to enable contextual validation
10. THE ValidationEngine SHALL achieve less than 1ms overhead per individual validation operation
11. THE ValidationEngine SHALL complete full validation (all validators) in less than 10ms

### Requirement 2: Unified Validator Implementations

**User Story:** As a developer, I want unified validators for all validation domains, so that validation rules are consistent and maintainable in a single location.

#### Acceptance Criteria

1. THE System SHALL provide a ClusterNameValidator that validates cluster naming rules
2. WHEN ClusterNameValidator validates a name, THE System SHALL check that length is between 1 and 63 characters
3. WHEN ClusterNameValidator validates a name, THE System SHALL check that it contains only lowercase alphanumeric characters and hyphens
4. WHEN ClusterNameValidator validates a name, THE System SHALL check that it does not start or end with a hyphen
5. THE System SHALL provide a NetworkValidator that validates network configuration including CIDR ranges and IP addresses
6. THE System SHALL provide a ProviderValidator that validates cloud provider-specific configuration
7. THE System SHALL provide a SOPSKeyValidator that validates SOPS encryption key format and accessibility
8. THE System SHALL provide a GitOpsValidator that validates GitOps repository structure and required files
9. THE System SHALL provide a ServiceValidator that validates service configuration against service-specific schemas
10. WHEN any validator detects an error, THE System SHALL include actionable suggestions in the ValidationError
11. THE System SHALL eliminate all duplicate validation logic from config, SOPS, gitops, and services packages

### Requirement 3: Validation Result Structure

**User Story:** As a developer, I want rich validation results, so that I can provide users with clear error messages and actionable guidance.

#### Acceptance Criteria

1. THE ValidationResult SHALL include a Valid boolean field indicating overall validation success
2. THE ValidationResult SHALL include an Errors slice containing all ValidationError instances
3. THE ValidationResult SHALL include a Warnings slice containing non-fatal validation issues
4. THE ValidationResult SHALL include a Suggestions slice containing actionable recommendations
5. THE ValidationError SHALL include Field, Message, Suggestions, and Context fields
6. THE ValidationError SHALL implement the error interface for compatibility with existing error handling
7. THE ValidationResult SHALL provide a ToError method that converts the result to a StructuredError
8. WHEN ToError is called on a valid result, THE System SHALL return nil
9. WHEN ToError is called on an invalid result, THE System SHALL return a StructuredError containing all validation errors
10. THE ValidationResult SHALL provide a HasWarnings method that returns true if warnings exist

### Requirement 4: Config Validation Migration

**User Story:** As a developer, I want config validation migrated to ValidationEngine, so that configuration validation is consistent across all config operations.

#### Acceptance Criteria

1. THE System SHALL implement a feature flag OPENCENTER_NEW_VALIDATION to control validation system selection
2. WHEN the feature flag is disabled, THE System SHALL use the existing config validator
3. WHEN the feature flag is enabled, THE System SHALL use the ValidationEngine for config validation
4. THE System SHALL support parallel validation mode where both old and new validators run simultaneously
5. WHEN parallel validation detects result mismatches, THE System SHALL log a warning with details of the difference
6. THE ConfigurationManager SHALL use ValidationEngine for all config validation operations
7. THE System SHALL remove the old config validator code after migration is complete and verified
8. THE System SHALL maintain or improve validation performance compared to the old validator
9. THE System SHALL ensure all existing config validation tests pass with the new ValidationEngine
10. THE System SHALL migrate validation logic from config/validator.go, config/enhanced_validator.go, and config/multilayer_validator.go

### Requirement 5: SOPS Validation Migration

**User Story:** As a developer, I want SOPS validation migrated to ValidationEngine, so that encryption key validation is consistent and security-focused.

#### Acceptance Criteria

1. THE SOPSManager SHALL use ValidationEngine for all SOPS-related validation
2. THE SOPSKeyValidator SHALL validate Age key format and structure
3. THE SOPSKeyValidator SHALL verify Age key files are readable and have correct permissions
4. THE SOPSKeyValidator SHALL check that SOPS configuration files reference valid Age keys
5. WHEN SOPSKeyValidator detects missing keys, THE System SHALL provide suggestions for key generation
6. THE System SHALL remove the old SOPS validator code after migration is complete
7. THE System SHALL maintain all existing security checks during migration
8. THE System SHALL ensure SOPS validation tests pass with the new ValidationEngine
9. THE SOPSKeyValidator SHALL validate that encrypted files can be decrypted with available keys
10. THE System SHALL migrate validation logic from internal/sops/validator.go

### Requirement 6: Service Validation Migration

**User Story:** As a developer, I want service validation migrated to ValidationEngine, so that all 15+ service plugins use consistent validation.

#### Acceptance Criteria

1. THE BaseServicePlugin SHALL use ValidationEngine for service configuration validation
2. WHEN a service plugin validates configuration, THE System SHALL use a validator named "service:{service_name}"
3. THE System SHALL provide a ServiceValidator base implementation that service-specific validators can extend
4. THE ServiceValidator SHALL validate common service fields (name, namespace, enabled status)
5. THE System SHALL remove individual Validate methods from service plugin implementations
6. THE System SHALL update the service registry to register validators with the ValidationEngine
7. THE System SHALL ensure all service validation tests pass with the new ValidationEngine
8. THE System SHALL eliminate duplicate validation code across all service plugins
9. WHEN service validation fails, THE System SHALL provide service-specific suggestions
10. THE System SHALL migrate validation logic from internal/services/plugins/*.go

### Requirement 7: Security Validation

**User Story:** As a security-conscious developer, I want security validations centralized, so that security checks are consistently applied across all inputs.

#### Acceptance Criteria

1. THE System SHALL provide a PathTraversalValidator that detects path traversal attempts
2. WHEN PathTraversalValidator detects ".." in paths, THE System SHALL return a validation error
3. THE System SHALL provide a CommandInjectionValidator that detects shell command injection attempts
4. THE System SHALL provide an InputSanitizationValidator that validates user input against injection patterns
5. THE SecurityValidator SHALL check for common security patterns (SQL injection, XSS, command injection)
6. THE ValidationEngine SHALL automatically apply security validators to all user-provided input
7. WHEN security validation fails, THE System SHALL log the attempt for security auditing
8. THE System SHALL provide clear error messages for security violations without revealing system internals
9. THE SecurityValidator SHALL be extensible to add new security checks without modifying existing code
10. THE System SHALL ensure security validations cannot be bypassed through feature flags

### Requirement 8: Performance and Scalability

**User Story:** As a developer, I want validation to be performant, so that it does not slow down application operations.

#### Acceptance Criteria

1. THE ValidationEngine SHALL achieve less than 1ms overhead per individual validation operation
2. THE ValidationEngine SHALL complete full validation (all validators) in less than 10ms
3. THE ValidationEngine SHALL support concurrent validation operations without performance degradation
4. THE System SHALL include benchmark tests measuring validation performance
5. WHEN benchmark tests run, THE System SHALL verify performance targets are met
6. THE ValidationEngine SHALL use efficient data structures (maps for O(1) validator lookup)
7. THE ValidationEngine SHALL avoid unnecessary allocations during validation
8. THE System SHALL cache validation results when appropriate to avoid redundant validation
9. THE ValidationEngine SHALL support validator prioritization to run fast validators first
10. THE System SHALL ensure validation performance is equal to or better than existing validators

### Requirement 9: Testing and Quality

**User Story:** As a project maintainer, I want comprehensive test coverage for validation, so that validation logic is reliable and correct.

#### Acceptance Criteria

1. THE ValidationEngine implementation SHALL achieve greater than 85% test coverage
2. THE System SHALL include property-based tests for all validators
3. THE System SHALL include unit tests for specific validation rules and edge cases
4. THE System SHALL test validator registration including duplicate registration errors
5. THE System SHALL test validation with non-existent validators
6. THE System SHALL test concurrent validation operations with race detection enabled
7. THE System SHALL test validation result aggregation with multiple validators
8. THE System SHALL test feature flag behavior for gradual rollout
9. THE System SHALL test parallel validation mode and result comparison
10. THE System SHALL include integration tests validating end-to-end validation flows

### Requirement 10: Migration Strategy and Rollback

**User Story:** As a project maintainer, I want a safe migration strategy, so that we can roll back if issues are discovered.

#### Acceptance Criteria

1. THE System SHALL implement feature flags enabling gradual rollout of new validation
2. THE System SHALL support parallel validation mode for comparing old and new validators
3. WHEN parallel validation detects mismatches, THE System SHALL log detailed comparison information
4. THE System SHALL provide a rollback mechanism to disable new validation if issues arise
5. THE System SHALL document the migration process including rollback procedures
6. THE System SHALL maintain backward compatibility during the migration period
7. THE System SHALL remove feature flags and old validation code only after successful migration
8. THE System SHALL verify all existing tests pass before removing old validation code
9. THE System SHALL create an Architecture Decision Record documenting the validation consolidation
10. THE System SHALL update all documentation to reflect the new validation architecture

### Requirement 11: Documentation and Developer Experience

**User Story:** As a developer using the validation system, I want clear documentation, so that I can easily add new validators and understand validation behavior.

#### Acceptance Criteria

1. THE ValidationEngine SHALL include package documentation with usage examples
2. THE System SHALL provide examples of creating custom validators
3. THE System SHALL document the validator registration process
4. THE System SHALL document how to use ValidationEngine in different contexts (config, SOPS, services)
5. THE System SHALL provide migration guides for updating code to use ValidationEngine
6. THE System SHALL document validation error handling patterns
7. THE System SHALL include examples of validation result interpretation
8. THE System SHALL document performance characteristics and optimization techniques
9. THE System SHALL provide troubleshooting guides for common validation issues
10. THE System SHALL update architecture documentation to reflect validation consolidation
