# Implementation Plan: TODO Completion Phase 1

## Overview

This implementation plan addresses 12 outstanding TODO items (4-15) organized into 8 logical epics. The plan prioritizes Epic 1 (Terraform State Integration) as the foundational piece that enables accurate drift detection. Implementation follows an incremental approach with testing integrated at each step.

## Tasks

- [ ] 1. Epic 1: Terraform State Integration (Foundation)
  - [ ] 1.1 Create Terraform state reader package
    - Create `internal/terraform/state_reader.go` with StateReader interface
    - Implement State, Resource, and ResourceInstance structs
    - Implement ReadState method to parse JSON state files
    - Handle both Terraform and OpenTofu state formats
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.7_
  
  - [ ]* 1.2 Write property test for Terraform state parsing
    - **Property 2: Valid Terraform state parsing**
    - **Validates: Requirements 1.3, 1.4**
  
  - [ ]* 1.3 Write property test for Terraform/OpenTofu compatibility
    - **Property 3: Terraform and OpenTofu compatibility**
    - **Validates: Requirements 1.7**
  
  - [ ] 1.4 Implement resource type mapping
    - Create resource type mapping table (Terraform type → InfrastructureState type)
    - Implement ConvertToInfrastructureState method
    - Map all 8 OpenStack resource types (compute, network, security group, load balancer, volume, floating IP)
    - Handle unknown resource types with warning and skip
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5, 2.6, 2.7, 2.8, 2.9_
  
  - [ ]* 1.5 Write property test for resource type mapping
    - **Property 4: Resource type mapping completeness**
    - **Validates: Requirements 2.1-2.8**
  
  - [ ]* 1.6 Write property test for unknown resource type handling
    - **Property 5: Unknown resource type handling**
    - **Validates: Requirements 2.9**
  
  - [ ] 1.7 Implement attribute extraction
    - Extract attributes for Server resources (id, name, flavor, image, status, networks, metadata)
    - Extract attributes for Network resources (id, name, subnets)
    - Extract attributes for SecurityGroup resources (id, name, rules)
    - Extract attributes for LoadBalancer resources (id, name, vip, members, protocol, port)
    - Extract attributes for Volume resources (id, name, size, status, attachments)
    - Extract attributes for FloatingIP resources (id, address, status, attached_to)
    - Handle missing attributes with sensible defaults
    - Handle nested attributes (e.g., instances[0].attributes.id)
    - _Requirements: 3.1, 3.2, 3.3, 3.4, 3.5, 3.6, 3.7, 3.8, 3.9, 3.10_
  
  - [ ]* 1.8 Write property test for attribute extraction
    - **Property 7: Attribute extraction completeness**
    - **Validates: Requirements 3.1-3.8**
  
  - [ ]* 1.9 Write property test for missing attribute handling
    - **Property 8: Missing attribute default handling**
    - **Validates: Requirements 3.9**
  
  - [ ]* 1.10 Write property test for nested attribute extraction
    - **Property 9: Nested attribute extraction**
    - **Validates: Requirements 3.10**
  
  - [ ] 1.11 Implement resource relationship preservation
    - Preserve volume attachment relationships (volume → instance)
    - Preserve floating IP associations (floating IP → instance)
    - Preserve security group associations (security group → instance)
    - Preserve subnet relationships (subnet → network)
    - _Requirements: 2.10_
  
  - [ ]* 1.12 Write property test for relationship preservation
    - **Property 6: Resource relationship preservation**
    - **Validates: Requirements 2.10**
  
  - [ ] 1.13 Integrate with buildDesiredState
    - Update `cmd/cluster_drift.go::buildDesiredState` to use TerraformStateReader
    - Replace manual state construction with Terraform state parsing
    - Locate state file at `<git_dir>/infrastructure/clusters/<cluster>/terraform.tfstate`
    - Handle missing state file error
    - Handle corrupted state file error
    - Cache parsed state for duration of drift detection operation
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 4.6, 4.7, 4.8_
  
  - [ ]* 1.14 Write property test for buildDesiredState integration
    - **Property 10: buildDesiredState integration**
    - **Validates: Requirements 4.1-4.5**
  
  - [ ]* 1.15 Write unit tests for error handling
    - Test missing state file error
    - Test corrupted JSON error
    - Test large state file handling
    - _Requirements: 1.5, 1.6, 1.8_

- [ ] 2. Checkpoint - Verify Terraform state integration
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 3. Epic 2: OpenStack Provider Enhancements
  - [ ] 3.1 Implement security groups retrieval
    - Add `listSecurityGroups` method to OpenStack provider
    - Use gophercloud security groups and rules packages
    - Filter by cluster tag
    - Extract security group ID, name, description
    - Extract all security rules (direction, protocol, port range, remote IP, description)
    - Populate InfrastructureState.SecurityGroups
    - Handle API errors with detailed error messages
    - _Requirements: 5.1, 5.2, 5.3, 5.4, 5.5, 5.6, 5.7_
  
  - [ ]* 3.2 Write property test for security groups retrieval
    - **Property 11: Security groups retrieval**
    - **Validates: Requirements 5.1-5.5**
  
  - [ ] 3.3 Implement load balancers retrieval
    - Add `listLoadBalancers` method to OpenStack provider
    - Use gophercloud loadbalancers, listeners, pools, members packages
    - Filter by cluster tag
    - Extract load balancer ID, name, VIP address, VIP port, provisioning status, operating status
    - Extract listeners (protocol, port, default pool, connection limit)
    - Extract pools (protocol, algorithm, session persistence, members)
    - Extract members (address, port, weight, health status)
    - Populate InfrastructureState.LoadBalancers
    - Handle API errors with detailed error messages
    - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 6.6, 6.7, 6.8_
  
  - [ ]* 3.4 Write property test for load balancers retrieval
    - **Property 12: Load balancers retrieval**
    - **Validates: Requirements 6.1-6.8**
  
  - [ ] 3.5 Implement volumes retrieval
    - Add `listVolumes` method to OpenStack provider
    - Use gophercloud volumes package
    - Filter by cluster metadata
    - Extract volume ID, name, size, status, volume type, availability zone
    - Extract volume attachments (instance ID, device path, attachment status)
    - Populate InfrastructureState.Volumes
    - Handle API errors with detailed error messages
    - _Requirements: 7.1, 7.2, 7.3, 7.4, 7.5, 7.6, 7.7_
  
  - [ ]* 3.6 Write property test for volumes retrieval
    - **Property 13: Volumes retrieval**
    - **Validates: Requirements 7.1-7.6**
  
  - [ ] 3.7 Implement floating IPs retrieval
    - Add `listFloatingIPs` method to OpenStack provider
    - Use gophercloud floatingips package
    - Filter by cluster tag
    - Extract floating IP ID, address, fixed IP, port ID, status
    - Resolve port ID to instance ID
    - Populate InfrastructureState.FloatingIPs
    - Handle API errors with detailed error messages
    - Handle unassociated floating IPs (nil instance ID)
    - _Requirements: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6, 8.7_
  
  - [ ]* 3.8 Write property test for floating IPs retrieval
    - **Property 14: Floating IPs retrieval**
    - **Validates: Requirements 8.1-8.7**
  
  - [ ] 3.9 Integrate with GetCurrentState
    - Update `internal/cloud/openstack/provider.go::GetCurrentState` to call new methods
    - Add security groups retrieval
    - Add load balancers retrieval
    - Add volumes retrieval
    - Add floating IPs retrieval
    - Ensure all resource types are populated in InfrastructureState
    - _Requirements: 5.1, 6.1, 7.1, 8.1_
  
  - [ ]* 3.10 Write integration test for comprehensive state retrieval
    - Test GetCurrentState returns all resource types
    - Use mock OpenStack API
    - _Requirements: 5.1, 6.1, 7.1, 8.1_

- [ ] 4. Checkpoint - Verify OpenStack provider enhancements
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 5. Epic 3: Secrets Management Enhancements
  - [ ] 5.1 Create manifest scanner package
    - Create `internal/secrets/scanner.go` with ManifestScanner interface
    - Implement Scanner struct
    - Implement ScanManifests method
    - Use filepath.Walk to recursively traverse overlay directory
    - Check for `kind: Secret` in YAML files
    - Check for SOPS encryption markers (`sops:`, `ENC[AES256_GCM,`)
    - Extract service name from directory path
    - Skip common ignore patterns (.git/, node_modules/, *.swp, *~)
    - Handle symlinks by following once (no recursive symlink following)
    - Apply service filter if provided
    - Return map[serviceName][]manifestPaths
    - _Requirements: 9.1, 9.2, 9.3, 9.4, 9.5, 9.6, 9.8, 9.9_
  
  - [ ]* 5.2 Write property test for manifest scanning
    - **Property 15: Manifest scanning completeness**
    - **Validates: Requirements 9.1-9.4**
  
  - [ ]* 5.3 Write property test for service name extraction
    - **Property 16: Service name extraction**
    - **Validates: Requirements 9.3**
  
  - [ ]* 5.4 Write property test for manifest filtering
    - **Property 17: Manifest scanning filtering**
    - **Validates: Requirements 9.8**
  
  - [ ] 5.5 Create organization detector package
    - Create `internal/config/org_detector.go` with OrganizationDetector interface
    - Implement OrgDetector struct
    - Implement DetermineOrganization method
    - Scan `~/.config/opencenter/clusters/*/` for directories
    - Check for `<cluster>/<cluster>-config.yaml` in each organization
    - Return organization name if found
    - Return error if multiple matches found (list all matches)
    - Return error if no matches found (list available clusters)
    - Cache results for duration of command execution
    - _Requirements: 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, 10.8_
  
  - [ ]* 5.6 Write property test for organization detection
    - **Property 18: Organization detection uniqueness**
    - **Validates: Requirements 10.1-10.4**
  
  - [ ]* 5.7 Write property test for organization ambiguity handling
    - **Property 19: Organization detection ambiguity handling**
    - **Validates: Requirements 10.5**
  
  - [ ]* 5.8 Write unit tests for organization detection edge cases
    - Test missing organization directory
    - Test explicit --organization flag override
    - _Requirements: 10.7, 10.8_

- [ ] 6. Checkpoint - Verify secrets management enhancements
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 7. Epic 4: Key Revocation User Context
  - [ ] 7.1 Create user context extractor package
    - Create `internal/secrets/user_context.go` with UserContextExtractor interface
    - Implement ContextExtractor struct
    - Implement ExtractUserContext method
    - Check --user flag (highest priority)
    - Check OPENCENTER_USER environment variable
    - Check OPENCENTER_SERVICE_ACCOUNT environment variable
    - Check Git user.email configuration
    - Default to "system" if no context available
    - Validate user identifier as email or service account name
    - Return UserContext with email, service account, and source
    - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5, 11.6, 11.7, 11.8_
  
  - [ ]* 7.2 Write property test for user context extraction
    - **Property 20: User context extraction priority**
    - **Validates: Requirements 11.1-11.5**
  
  - [ ]* 7.3 Write property test for user identifier validation
    - **Property 21: User identifier validation**
    - **Validates: Requirements 11.7**
  
  - [ ] 7.4 Create key ownership checker package
    - Create `internal/secrets/ownership.go` with KeyOwnershipChecker interface
    - Implement OwnershipChecker struct
    - Implement IsKeyOwnedByUser method
    - Read key registry from `~/.config/opencenter/clusters/<org>/secrets/key-registry.json`
    - Check if key's CreatedBy field matches user email
    - Check if user email is in key's UsedBy list
    - Fall back to checking .sops.yaml recipient list if no registry
    - Match Age public keys against user's registered keys
    - Return false if no ownership information found (deny by default)
    - Support --force flag to bypass ownership checks
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5, 12.6, 12.7, 12.8_
  
  - [ ]* 7.5 Write property test for key ownership verification
    - **Property 22: Key ownership verification**
    - **Validates: Requirements 12.1-12.4**
  
  - [ ]* 7.6 Write property test for key ownership fallback
    - **Property 23: Key ownership fallback**
    - **Validates: Requirements 12.5-12.6**
  
  - [ ] 7.7 Integrate with key revocation command
    - Update `internal/secrets/revocation.go` to use UserContextExtractor
    - Update `internal/secrets/revocation.go` to use KeyOwnershipChecker
    - Extract user context before revocation
    - Verify key ownership before revocation
    - Include user in RevokedBy field of key entry
    - Log revocation event to audit log with user context
    - _Requirements: 11.6, 12.8_
  
  - [ ]* 7.8 Write integration test for key revocation with user context
    - Test complete revocation workflow with user context
    - Test ownership verification
    - Test --force flag bypass
    - _Requirements: 11.1-11.8, 12.1-12.8_

- [ ] 8. Checkpoint - Verify key revocation user context
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 9. Epic 5: Configuration Validation
  - [ ] 9.1 Create configuration structure validator package
    - Create `internal/config/structure_validator.go` with ConfigStructureValidator interface
    - Implement StructureValidator struct
    - Implement Validate method
    - Use JSON Schema for structural validation (types, required fields, enums)
    - Implement custom validators for cross-field constraints
    - Check required fields (ClusterName, Provider, Kubernetes.Version)
    - Check cross-field constraints (WorkerCount >= MasterCount, UseOctavia requires OpenStack, etc.)
    - Check provider-specific requirements (OpenStack.NetworkID, AWS.Region, VMware.Datacenter)
    - Check service dependencies (Keycloak requires CertManager, Harbor requires CertManager)
    - Aggregate all validation errors (don't stop at first error)
    - Return detailed error messages with field paths
    - _Requirements: 13.1, 13.2, 13.3, 13.4, 13.5, 13.6, 13.7_
  
  - [ ]* 9.2 Write property test for required field validation
    - **Property 24: Required field validation**
    - **Validates: Requirements 13.1-13.3**
  
  - [ ]* 9.3 Write property test for cross-field constraint validation
    - **Property 25: Cross-field constraint validation**
    - **Validates: Requirements 13.4**
  
  - [ ]* 9.4 Write property test for provider-specific validation
    - **Property 26: Provider-specific validation**
    - **Validates: Requirements 13.5**
  
  - [ ]* 9.5 Write property test for service dependency validation
    - **Property 27: Service dependency validation**
    - **Validates: Requirements 13.6**
  
  - [ ]* 9.6 Write property test for validation error aggregation
    - **Property 28: Validation error aggregation**
    - **Validates: Requirements 13.7**
  
  - [ ] 9.7 Integrate with Config Manager
    - Update `internal/config/manager.go::Save` to use ConfigStructureValidator
    - Validate before saving
    - Support --skip-validation flag with warning
    - Log validation skip event to audit log
    - Return aggregated validation errors
    - _Requirements: 13.8, 13.9_
  
  - [ ]* 9.8 Write integration test for configuration validation
    - Test complete validation workflow
    - Test validation with multiple errors
    - Test --skip-validation flag
    - _Requirements: 13.1-13.9_

- [ ] 10. Checkpoint - Verify configuration validation
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 11. Epic 6: Audit Logging Implementation
  - [ ] 11.1 Create audit log signer package
    - Create `internal/security/audit_signer.go` with AuditSigner interface
    - Implement Ed25519Signer struct
    - Implement Sign method
    - Implement Verify method
    - Load or generate Ed25519 key pair on initialization
    - Store signing key at `~/.config/opencenter/audit/signing-key`
    - Store public key at `~/.config/opencenter/audit/signing-key.pub`
    - Set key permissions to 0600 (owner read/write only)
    - Serialize event to JSON (excluding signature field)
    - Sign with Ed25519
    - Return base64-encoded signature
    - _Requirements: 14.1, 14.2, 14.3, 14.4, 14.5, 14.6, 14.7_
  
  - [ ]* 11.2 Write property test for signature generation
    - **Property 29: Signature generation determinism**
    - **Validates: Requirements 14.1, 14.4-14.6**
  
  - [ ]* 11.3 Write property test for signature round-trip
    - **Property 30: Signature verification round-trip**
    - **Validates: Requirements 15.1-15.3**
  
  - [ ]* 11.4 Write property test for tampering detection
    - **Property 31: Signature tampering detection**
    - **Validates: Requirements 15.4**
  
  - [ ] 11.5 Create audit log verifier package
    - Create `internal/security/audit_verifier.go` with AuditVerifier interface
    - Implement Verifier struct
    - Implement VerifyLog method
    - Read audit log file line by line
    - Parse each JSON event
    - Verify signature using Ed25519Signer
    - Track verified and failed events
    - Return VerificationReport with total, verified, and failed events
    - Support --repair flag to re-sign events with missing/invalid signatures
    - _Requirements: 15.1, 15.2, 15.3, 15.4, 15.5, 15.6, 15.7, 15.8_
  
  - [ ]* 11.6 Write unit tests for audit log verification
    - Test verification of valid log
    - Test detection of tampered events
    - Test --repair flag functionality
    - _Requirements: 15.1-15.8_
  
  - [ ] 11.7 Create audit log parser package
    - Create `internal/security/audit_parser.go` with AuditParser interface
    - Implement Parser struct
    - Implement ParseLog method
    - Read audit log file line by line
    - Parse each line as JSON into AuditEvent struct
    - Apply filters (--since, --event-type, --user)
    - Handle corrupted lines (invalid JSON) by logging warning and skipping
    - Return events sorted by timestamp
    - Return empty list (not error) if log file doesn't exist
    - _Requirements: 16.1, 16.2, 16.3, 16.4, 16.5, 16.6, 16.7, 16.8, 16.9_
  
  - [ ]* 11.8 Write property test for audit log parsing
    - **Property 32: Audit log parsing completeness**
    - **Validates: Requirements 16.1-16.3**
  
  - [ ]* 11.9 Write property test for audit log filtering
    - **Property 33: Audit log filtering correctness**
    - **Validates: Requirements 16.4-16.6**
  
  - [ ] 11.10 Integrate with audit logging commands
    - Update `cmd/cluster_audit_log.go` to use AuditSigner
    - Update `cmd/cluster_audit_log.go` to use AuditVerifier
    - Update `cmd/cluster_audit_log.go` to use AuditParser
    - Sign all audit events before writing
    - Implement verify command to check log integrity
    - Implement parse command to filter and display events
    - _Requirements: 14.8, 15.1, 16.1_
  
  - [ ]* 11.11 Write integration test for audit logging workflow
    - Test complete audit workflow: log event → sign → write → verify
    - Test parsing workflow: read log → filter → return events
    - _Requirements: 14.1-14.8, 15.1-15.8, 16.1-16.9_

- [ ] 12. Checkpoint - Verify audit logging implementation
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 13. Epic 7: Drift Detection Callbacks
  - [ ] 13.1 Create HTTP callback client package
    - Create `internal/drift/callback.go` with CallbackClient interface
    - Implement HTTPCallbackClient struct
    - Implement SendDriftReport method
    - Marshal DriftReport to JSON
    - Create HTTP POST request to callback URL
    - Set Content-Type: application/json header
    - Set Authorization header if --callback-auth provided
    - Set 10-second timeout
    - Send request and check response status
    - Log success (2xx status) or failure (non-2xx or network error)
    - Don't block drift detection on callback failure
    - Implement retry logic with exponential backoff (up to 3 retries)
    - _Requirements: 17.1, 17.2, 17.3, 17.4, 17.5, 17.6, 17.7, 17.8, 17.9_
  
  - [ ]* 13.2 Write property test for callback request format
    - **Property 34: Callback request format**
    - **Validates: Requirements 17.1-17.3**
  
  - [ ]* 13.3 Write property test for callback authentication
    - **Property 35: Callback authentication**
    - **Validates: Requirements 17.5**
  
  - [ ]* 13.4 Write property test for callback timeout
    - **Property 36: Callback timeout enforcement**
    - **Validates: Requirements 17.8**
  
  - [ ]* 13.5 Write property test for callback retry
    - **Property 37: Callback retry behavior**
    - **Validates: Requirements 17.9**
  
  - [ ] 13.6 Integrate with drift detection commands
    - Update `cmd/cluster_drift.go::newClusterDriftDetectCmd` to use CallbackClient
    - Update `cmd/cluster_drift.go::newClusterDriftScheduleCmd` to use CallbackClient
    - Add --callback-url flag
    - Add --callback-auth flag
    - Add --callback-retry flag
    - Send drift report to callback URL after detection
    - Log callback success or failure
    - Continue drift detection even if callback fails
    - _Requirements: 17.1, 17.6, 17.7_
  
  - [ ]* 13.7 Write integration test for drift callback workflow
    - Test complete callback workflow: detect drift → send HTTP POST → verify received
    - Test callback failure handling
    - Test retry logic
    - _Requirements: 17.1-17.9_

- [ ] 14. Checkpoint - Verify drift detection callbacks
  - Ensure all tests pass, ask the user if questions arise.


- [ ] 15. Epic 8: Keycloak Backup Automation
  - [ ] 15.1 Create backup upload script template
    - Create `internal/gitops/templates/services/keycloak/backup-script.sh.tpl`
    - Implement storage backend auto-detection (S3 vs Swift)
    - Implement S3 upload function with AWS CLI
    - Implement Swift upload function with Swift CLI
    - Implement retry logic with exponential backoff (up to 3 retries)
    - Set object metadata (cluster name, timestamp, Keycloak version)
    - Set lifecycle policy for S3 (delete after retention days)
    - Set delete-after header for Swift (retention seconds)
    - Exit with non-zero status if all retries fail
    - _Requirements: 18.1, 18.2, 18.3, 18.4, 18.5, 18.6, 18.7, 18.8, 18.9, 19.1, 19.2, 19.3, 19.4, 19.5, 19.6, 19.7, 19.8, 19.9_
  
  - [ ]* 15.2 Write property test for backup upload retry
    - **Property 38: Backup upload retry**
    - **Validates: Requirements 18.6, 19.6**
  
  - [ ]* 15.3 Write property test for backup metadata
    - **Property 39: Backup metadata inclusion**
    - **Validates: Requirements 18.8, 19.8**
  
  - [ ]* 15.4 Write property test for storage backend detection
    - **Property 40: Storage backend auto-detection**
    - **Validates: Requirements 19.9**
  
  - [ ] 15.5 Create backup CronJob template
    - Create `internal/gitops/templates/services/keycloak/backup-cronjob.yaml.tpl`
    - Define CronJob with daily schedule (2 AM)
    - Mount backup script as ConfigMap
    - Configure environment variables for storage backend
    - Configure S3 credentials from Kubernetes secret
    - Configure Swift credentials from Kubernetes secret
    - Configure backup retention from ConfigMap
    - Set job history limits (3 successful, 3 failed)
    - Set restart policy to OnFailure
    - _Requirements: 18.1, 19.1_
  
  - [ ] 15.6 Create backup configuration templates
    - Create ConfigMap template for backup configuration
    - Create Secret template for S3 credentials
    - Create Secret template for Swift credentials
    - Include storage backend selection (s3 or swift)
    - Include retention days configuration
    - Include bucket/container name configuration
    - _Requirements: 18.3, 18.4, 18.9, 19.3, 19.4, 19.9_
  
  - [ ] 15.7 Integrate with GitOps template rendering
    - Update `internal/gitops/copy.go` to include backup templates
    - Render backup CronJob template with cluster-specific values
    - Render backup ConfigMap template with storage configuration
    - Render backup Secret templates with credentials
    - Include templates in services/keycloak directory
    - _Requirements: 18.1, 19.1_
  
  - [ ]* 15.8 Write integration test for backup workflow
    - Test complete backup workflow: export realm → upload to S3 → verify uploaded
    - Test complete backup workflow: export realm → upload to Swift → verify uploaded
    - Test retry logic with simulated failures
    - Use mock AWS SDK and Swift client
    - _Requirements: 18.1-18.9, 19.1-19.9_

- [ ] 16. Checkpoint - Verify Keycloak backup automation
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 17. Final Integration and Documentation
  - [ ] 17.1 Remove TODO comments
    - Remove TODO comment from `cmd/cluster_drift.go::buildDesiredState`
    - Remove TODO comment from `internal/cloud/openstack/provider.go::GetCurrentState`
    - Remove TODO comment from `internal/secrets/manager.go::ScanManifests`
    - Remove TODO comment from `internal/secrets/manager.go::DetermineOrganization`
    - Remove TODO comment from `internal/secrets/revocation.go::isKeyOwnedByUser`
    - Remove TODO comment from `internal/config/manager.go::Save`
    - Remove TODO comment from `cmd/cluster_audit_log.go::generateSignature`
    - Remove TODO comment from `cmd/cluster_audit_log.go::verifyAuditLogIntegrity`
    - Remove TODO comment from `cmd/cluster_audit_log.go::parseAuditLog`
    - Remove TODO comment from `cmd/cluster_drift.go::newClusterDriftScheduleCmd`
    - Remove TODO comment from `internal/gitops/templates/.../keycloak-backup-cronjob.yaml.tpl`
    - Remove TODO comment from `internal/gitops/templates/.../keycloak-backup-cronjob.yaml.tpl`
    - _Requirements: All epics_
  
  - [ ] 17.2 Update drift detection documentation
    - Update `docs/explanation/drift-detection.md` to explain Terraform state integration
    - Add section on Terraform state as source of truth
    - Add section on resource type mapping
    - Add examples of drift detection with all resource types
    - Document callback functionality
    - _Requirements: Epic 1, Epic 2, Epic 7_
  
  - [ ] 17.3 Update secrets management documentation
    - Document manifest scanning functionality
    - Document organization auto-detection
    - Document key revocation user context
    - Add examples of secrets management workflows
    - _Requirements: Epic 3, Epic 4_
  
  - [ ] 17.4 Update configuration documentation
    - Document configuration validation rules
    - Document cross-field constraints
    - Document provider-specific requirements
    - Document service dependencies
    - Add examples of validation errors and fixes
    - _Requirements: Epic 5_
  
  - [ ] 17.5 Update audit logging documentation
    - Document audit log signing and verification
    - Document audit log parsing and filtering
    - Add examples of audit log commands
    - Document key management for signing
    - _Requirements: Epic 6_
  
  - [ ] 17.6 Update Keycloak backup documentation
    - Document backup automation setup
    - Document S3 and Swift configuration
    - Add examples of backup CronJob configuration
    - Document backup retention policies
    - _Requirements: Epic 8_
  
  - [ ]* 17.7 Run full integration test suite
    - Test complete drift detection workflow with Terraform state
    - Test complete secrets management workflow
    - Test complete configuration validation workflow
    - Test complete audit logging workflow
    - Test complete backup workflow
    - Verify all 12 TODO items are resolved
    - _Requirements: All epics_

- [ ] 18. Final checkpoint - Verify all implementations
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- Integration tests validate end-to-end workflows
- Epic 1 (Terraform State Integration) is the foundational piece and must be completed first
- Epic 2 (OpenStack Provider Enhancements) depends on Epic 1
- All other epics are independent and can be implemented in parallel after Epic 1

