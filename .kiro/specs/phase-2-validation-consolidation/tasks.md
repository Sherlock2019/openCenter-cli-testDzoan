# Implementation Plan: Phase 2 Validation Consolidation

## Overview

This implementation plan consolidates validation logic scattered across 15+ packages into a unified ValidationEngine. The plan follows a direct replacement strategy to ensure clean migration without backward compatibility overhead.

The implementation is organized into three main phases:
1. **Week 3**: Build ValidationEngine core and unified validators
2. **Week 4**: Migrate config and SOPS validation
3. **Week 5**: Migrate service validation, remove old code

## Tasks

- [x] 1. Implement ValidationEngine Core
  - Create the central validation engine with thread-safe validator registry
  - Implement validator registration, lookup, and execution
  - Build result aggregation for ValidateAll operations
  - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5, 1.6, 1.7, 1.8, 1.9_

- [ ]* 1.1 Write property test for validator registration uniqueness
  - **Property 1: Validator Registration Uniqueness**
  - **Validates: Requirements 1.1, 1.2**
  - Generate random validators and verify registration/retrieval
  - Test duplicate registration returns error
  - _Requirements: 1.1, 1.2_

- [ ]* 1.2 Write property test for thread-safe concurrent operations
  - **Property 2: Thread-Safe Concurrent Operations**
  - **Validates: Requirements 1.6**
  - Test concurrent registrations and validations
  - Run with race detector enabled
  - _Requirements: 1.6_

- [ ]* 1.3 Write property test for validation result aggregation
  - **Property 3: Validation Result Aggregation**
  - **Validates: Requirements 1.7, 1.8**
  - Test ValidateAll with mixed passing/failing validators
  - Verify error aggregation correctness
  - _Requirements: 1.7, 1.8_

- [ ]* 1.4 Write property test for context propagation
  - **Property 4: Context Propagation**
  - **Validates: Requirements 1.9**
  - Verify context passed to validators unchanged
  - Test context cancellation handling
  - _Requirements: 1.9_

- [ ]* 1.5 Write unit tests for ValidationEngine edge cases
  - Test empty validator registry
  - Test validation with non-existent validator
  - Test ListValidators functionality
  - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 2. Implement ValidationResult and Error Structures
  - Create ValidationResult with Valid, Errors, Warnings, Suggestions fields
  - Implement ValidationError with Field, Message, Suggestions, Context
  - Build ToError method for converting results to StructuredError
  - Implement HasWarnings and HasErrors helper methods
  - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10_

- [ ]* 2.1 Write property test for ToError conversion
  - **Property 7: ValidationResult ToError Conversion**
  - **Validates: Requirements 3.9**
  - Test valid results return nil
  - Test invalid results return StructuredError with all messages
  - _Requirements: 3.8, 3.9_

- [ ]* 2.2 Write unit tests for ValidationResult methods
  - Test HasWarnings with various warning states
  - Test HasErrors with various error states
  - Test ValidationError.Error() formatting
  - _Requirements: 3.6, 3.10_

- [x] 3. Create ClusterNameValidator
  - Implement validator for cluster naming rules
  - Check length constraints (1-63 characters)
  - Validate character set (lowercase alphanumeric and hyphens)
  - Verify name doesn't start or end with hyphen
  - Provide actionable suggestions for common mistakes
  - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.10_

- [ ]* 3.1 Write property test for cluster name validation rules
  - **Property 5: Cluster Name Validation Rules**
  - **Validates: Requirements 2.2, 2.3, 2.4**
  - Generate valid and invalid cluster names
  - Verify all validation rules enforced
  - _Requirements: 2.2, 2.3, 2.4_

- [ ]* 3.2 Write unit tests for cluster name edge cases
  - Test empty name
  - Test exactly 63 characters
  - Test single character name
  - Test names with uppercase, underscores, special chars
  - _Requirements: 2.2, 2.3, 2.4_

- [x] 4. Create NetworkValidator
  - Implement validator for network configuration
  - Validate CIDR format for pod and service networks
  - Check for CIDR overlap between pod and service networks
  - Validate DNS server IP addresses
  - Provide network configuration suggestions
  - _Requirements: 2.5, 2.10_

- [ ]* 4.1 Write unit tests for network validation
  - Test valid CIDR formats
  - Test invalid CIDR formats
  - Test CIDR overlap detection
  - Test DNS server IP validation
  - _Requirements: 2.5_

- [x] 5. Create SOPSKeyValidator
  - Implement validator for SOPS encryption keys
  - Check key file existence and readability
  - Validate Age key format (starts with "AGE-SECRET-KEY-")
  - Verify file permissions (warn if not 0600)
  - Test decryption capability with sample data
  - _Requirements: 2.7, 5.2, 5.3, 5.4, 5.9, 2.10_

- [ ]* 5.1 Write property test for SOPS key validation
  - **Property 8: SOPS Key Validation**
  - **Validates: Requirements 5.2, 5.3, 5.4, 5.9**
  - Test with valid and invalid key files
  - Verify format, permissions, and decryption checks
  - _Requirements: 5.2, 5.3, 5.4, 5.9_

- [ ]* 5.2 Write unit tests for SOPS key edge cases
  - Test missing key file
  - Test unreadable key file
  - Test invalid key format
  - Test insecure permissions (not 0600)
  - _Requirements: 5.2, 5.3, 5.4_

- [x] 6. Create SecurityValidator
  - Implement validator for security issues
  - Detect path traversal attempts (".." in paths)
  - Detect command injection patterns (shell metacharacters)
  - Provide generic error messages without system details
  - Log security violations for audit trail
  - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.7, 7.8, 2.10_

- [ ]* 6.1 Write property test for security validation detection
  - **Property 10: Security Validation Detection**
  - **Validates: Requirements 7.2, 7.5**
  - Generate inputs with malicious patterns
  - Verify all patterns detected
  - _Requirements: 7.2, 7.5_

- [ ]* 6.2 Write property test for security error messages
  - **Property 11: Security Error Messages**
  - **Validates: Requirements 7.8**
  - Verify error messages don't leak system details
  - Test across all security validators
  - _Requirements: 7.8_

- [ ]* 6.3 Write unit tests for security validation
  - Test path traversal detection
  - Test command injection detection
  - Test security audit logging
  - Test safe inputs pass validation
  - _Requirements: 7.2, 7.5, 7.7_

- [x] 7. Create ServiceValidator Base
  - Implement base validator for service configurations
  - Validate common service fields (name, namespace, enabled)
  - Support service-specific validator extensions
  - Use naming convention "service:{service_name}"
  - Provide service-specific suggestions
  - _Requirements: 2.9, 6.2, 6.3, 6.4, 6.9, 2.10_

- [ ]* 7.1 Write property test for service validator naming
  - **Property 9: Service Validator Naming Convention**
  - **Validates: Requirements 6.2**
  - Verify correct validator name format used
  - Test with various service names
  - _Requirements: 6.2_

- [ ]* 7.2 Write unit tests for service field validation
  - Test common field validation (name, namespace, enabled)
  - Test service-specific extensions
  - Test error suggestions
  - _Requirements: 6.4, 6.9_

- [x] 8. Create GitOpsValidator and ProviderValidator
  - Implement GitOpsValidator for repository structure validation
  - Check for required files and directories
  - Validate manifest structure
  - Implement ProviderValidator for cloud provider configs
  - Validate provider-specific requirements
  - _Requirements: 2.6, 2.8, 2.10_

- [x]* 8.1 Write unit tests for GitOps and Provider validators
  - Test GitOps repository structure validation
  - Test provider configuration validation
  - Test error messages and suggestions
  - _Requirements: 2.6, 2.8_

- [x] 10. Checkpoint - Ensure all validators and engine tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. Implement Config Validation Migration
  - Add ValidationEngine to ConfigurationManager
  - Replace old validation with ValidationEngine calls
  - Update ConfigurationManager to use ValidationEngine
  - Ensure all existing config tests pass
  - _Requirements: 4.4, 4.5, 4.6, 10.3_

- [ ]* 11.1 Write integration tests for config validation migration
  - Test with real config files
  - Verify validation behavior is correct
  - Test error messages and suggestions
  - _Requirements: 4.2, 4.3, 4.9_

- [x] 12. Migrate Config Validator Logic
  - Move validation rules from config/validator.go to ClusterNameValidator
  - Move validation rules from config/enhanced_validator.go to NetworkValidator
  - Move validation rules from config/multilayer_validator.go to appropriate validators
  - Update ConfigurationManager to use ValidationEngine
  - Ensure all existing config tests pass
  - _Requirements: 4.10, 4.9_

- [ ]* 12.1 Write property test for validation errors include suggestions
  - **Property 6: Validation Errors Include Suggestions**
  - **Validates: Requirements 2.10**
  - Test across all validators
  - Verify suggestions are actionable
  - _Requirements: 2.10_

- [ ] 13. Checkpoint - Verify config validation migration
  - Ensure all tests pass, ask the user if questions arise.

- [x] 14. Implement SOPS Validation Migration
  - Add ValidationEngine to SOPSManager
  - Update ValidateEncryption to use SOPSKeyValidator
  - Migrate validation logic from internal/sops/validator.go
  - Ensure all security checks maintained
  - Update SOPS tests to use new validation
  - _Requirements: 5.1, 5.6, 5.7, 5.8, 5.10_

- [ ]* 14.1 Write integration tests for SOPS validation
  - Test with real SOPS key files
  - Test encryption/decryption validation
  - Test security checks maintained
  - _Requirements: 5.7, 5.8_

- [x] 15. Implement Service Validation Migration
  - Update BaseServicePlugin to use ValidationEngine
  - Register service validators with engine
  - Remove individual Validate methods from service plugins
  - Update service registry to register validators
  - Ensure all service tests pass
  - _Requirements: 6.1, 6.5, 6.6, 6.7, 6.8, 6.10_

- [ ]* 16.1 Write integration tests for service validation
  - Test service validation across all plugins
  - Verify validator naming convention
  - Test error messages and suggestions
  - _Requirements: 6.2, 6.4, 6.9_

- [x] 17. Implement Security Validator Always-Run Enforcement
  - Ensure security validators registered at engine creation
  - Verify security validators run in all validation operations
  - Add tests confirming security checks cannot be bypassed
  - _Requirements: 7.6, 7.9, 7.10_

- [ ]* 17.1 Write property test for security validators always run
  - **Property 12: Security Validators Always Run**
  - **Validates: Requirements 7.10**
  - Test with various validation scenarios
  - Verify security checks always execute
  - _Requirements: 7.10_

- [x] 18. Implement Validation Result Caching
  - Add caching layer to ValidationEngine
  - Cache results based on validator name and data hash
  - Implement cache invalidation on data changes
  - Add cache expiration (time-based)
  - _Requirements: 8.8_

- [ ]* 18.1 Write property test for validation result caching
  - **Property 13: Validation Result Caching**
  - **Validates: Requirements 8.8**
  - Test cache hit on repeated validation
  - Test cache invalidation on data change
  - _Requirements: 8.8_

- [x] 19. Implement Validator Prioritization
  - Add Priority field to Validator interface
  - Sort validators by priority before execution
  - Run fast validators first (cluster name, format checks)
  - Run slow validators last (network checks, file I/O)
  - _Requirements: 8.9_

- [ ]* 19.1 Write unit tests for validator prioritization
  - Test validators execute in priority order
  - Test with various priority values
  - Verify fast validators run first
  - _Requirements: 8.9_

- [ ]* 20. Performance Testing and Optimization
  - Write benchmark tests for single validation
  - Write benchmark tests for ValidateAll
  - Write benchmark tests for concurrent validation
  - Verify <1ms overhead per validation
  - Verify <10ms for full validation
  - Optimize hot paths if needed
  - _Requirements: 1.10, 1.11, 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7, 8.10_

- [ ]* 20.1 Write performance assertion tests
  - Test single validation meets <1ms target
  - Test full validation meets <10ms target
  - Test concurrent validation performance
  - _Requirements: 1.10, 1.11, 8.1, 8.2, 8.3_

- [x] 21. Checkpoint - Verify all migrations complete
  - Ensure all tests pass, ask the user if questions arise.

- [x] 22. Remove Old Validation Code
  - Remove config/validator.go
  - Remove config/enhanced_validator.go
  - Remove config/multilayer_validator.go
  - Remove internal/sops/validator.go
  - Remove Validate methods from service plugins
  - Update imports across codebase
  - _Requirements: 2.11, 4.7, 5.6, 6.5, 6.8, 10.7_

- [x] 23. Verify All Tests Pass After Cleanup
  - Run full test suite
  - Verify no broken imports
  - Verify all validation tests pass
  - Check test coverage meets targets (>85%)
  - _Requirements: 4.8, 5.8, 6.7, 9.1, 10.8_

- [ ] 25. Create Architecture Decision Record
  - Document validation consolidation decision
  - Explain rationale for ValidationEngine approach
  - Document migration strategy and lessons learned
  - Include performance metrics and improvements
  - _Requirements: 10.9_

- [ ] 25. Update Documentation
  - Update package documentation for ValidationEngine
  - Create examples of custom validator creation
  - Document validator registration process
  - Create migration guide for developers
  - Document validation error handling patterns
  - Update architecture documentation
  - Create troubleshooting guide
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8, 11.9, 11.10_

- [x] 26. Final Checkpoint - Complete validation consolidation
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Direct replacement strategy - no backward compatibility needed
- Performance testing ensures validation meets <1ms and <10ms targets
- Security validators always run and cannot be bypassed
- Documentation updates ensure developers can use and extend the system
