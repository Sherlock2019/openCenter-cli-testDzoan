# Requirements Document: TODO Completion Phase 1

## Introduction

This document specifies requirements for completing 12 outstanding TODO items in the openCenter-cli codebase (items 4-15 from the technical debt inventory). These items represent deferred implementation work across multiple subsystems: OpenStack provider enhancements, secrets management features, configuration validation, audit logging, drift detection callbacks, and Keycloak backup automation.

The implementation is organized into logical feature groups to enable incremental delivery while maintaining system coherence.

## Glossary

- **Terraform_State**: JSON file (`terraform.tfstate`) containing all resources provisioned by Terraform, serving as the authoritative source of truth for desired infrastructure state
- **OpenStack_Provider**: Cloud provider implementation for OpenStack infrastructure (`internal/cloud/openstack/provider.go`)
- **Drift_Detection**: System for identifying differences between desired state (from Terraform) and actual state (from cloud provider APIs)
- **Infrastructure_State**: Snapshot of cloud resources (instances, networks, security groups, volumes, load balancers, floating IPs)
- **Secrets_Manager**: Component responsible for secrets lifecycle management (`internal/secrets/manager.go`)
- **Manifest**: Kubernetes YAML files containing SOPS-encrypted secrets
- **Overlay_Directory**: Path containing cluster-specific Kubernetes manifests (`applications/overlays/<cluster>/`)
- **Key_Revocation**: Process of removing user access to encrypted secrets (`internal/secrets/revocation.go`)
- **Audit_Log**: Append-only log recording security events (`cmd/cluster_audit_log.go`)
- **Config_Manager**: Component managing cluster configuration validation and persistence (`internal/config/manager.go`)
- **ConfigStructureValidator**: Validator for complete Config struct validation (not yet implemented)
- **Drift_Callback**: HTTP webhook notification when infrastructure drift is detected
- **Keycloak_Backup**: Automated backup of Keycloak realm configuration to object storage
- **Object_Storage**: S3-compatible or Swift storage for backup artifacts

## Requirements

### Epic 1: Terraform State Integration

**User Story:** As a platform operator, I want drift detection to use Terraform state as the source of truth, so that I can accurately detect infrastructure drift for all resources that Terraform provisioned.

**Context:** The cluster configuration YAML is rendered into Terraform configuration (`main.tf`), which provisions infrastructure. Terraform maintains a state file (`terraform.tfstate`) containing all provisioned resources. This state file is the authoritative source of truth for what infrastructure should exist, not the YAML configuration.

#### Requirement 1.1: Terraform State File Reading

**Acceptance Criteria:**

1. WHEN `buildDesiredState()` is called, THE system SHALL locate the Terraform state file for the cluster
2. THE Terraform state file path SHALL be `<git_dir>/infrastructure/clusters/<cluster>/terraform.tfstate`
3. WHEN the Terraform state file exists, THE system SHALL read and parse the JSON content
4. WHEN parsing the state file, THE system SHALL extract the `resources` array from the JSON structure
5. WHEN the state file does not exist, THE system SHALL return an error indicating the cluster has not been provisioned
6. WHEN the state file is corrupted or invalid JSON, THE system SHALL return an error with the parse failure details
7. THE system SHALL support both Terraform and OpenTofu state file formats (they are compatible)
8. WHEN reading the state file, THE system SHALL handle large state files (>10MB) efficiently

#### Requirement 1.2: Terraform Resource Type Mapping

**Acceptance Criteria:**

1. WHEN processing Terraform state resources, THE system SHALL map OpenStack resource types to InfrastructureState types
2. THE system SHALL map `openstack_compute_instance_v2` resources to `Server` objects
3. THE system SHALL map `openstack_networking_network_v2` resources to `Network` objects
4. THE system SHALL map `openstack_networking_subnet_v2` resources to `Subnet` objects (nested in Network)
5. THE system SHALL map `openstack_networking_secgroup_v2` resources to `SecurityGroup` objects
6. THE system SHALL map `openstack_lb_loadbalancer_v2` resources to `LoadBalancer` objects
7. THE system SHALL map `openstack_blockstorage_volume_v3` resources to `Volume` objects
8. THE system SHALL map `openstack_networking_floatingip_v2` resources to `FloatingIP` objects
9. WHEN encountering an unknown resource type, THE system SHALL log a warning and skip the resource
10. THE system SHALL preserve resource relationships (e.g., volume attachments, floating IP associations)

#### Requirement 1.3: Terraform Attribute Extraction

**Acceptance Criteria:**

1. WHEN converting Terraform resources to InfrastructureState objects, THE system SHALL extract relevant attributes
2. FOR Server resources, THE system SHALL extract: id, name, flavor_id, image_id, status, network names, metadata/tags
3. FOR Network resources, THE system SHALL extract: id, name, and associated subnets
4. FOR SecurityGroup resources, THE system SHALL extract: id, name, description, and all security rules
5. FOR SecurityRule resources, THE system SHALL extract: direction, protocol, port_range_min, port_range_max, remote_ip_prefix, remote_group_id
6. FOR LoadBalancer resources, THE system SHALL extract: id, name, vip_address, vip_port_id, provisioning_status, operating_status
7. FOR Volume resources, THE system SHALL extract: id, name, size, status, volume_type, availability_zone, attachments
8. FOR FloatingIP resources, THE system SHALL extract: id, floating_ip_address, fixed_ip_address, port_id, status
9. WHEN an expected attribute is missing, THE system SHALL use a sensible default or empty value
10. THE system SHALL handle nested attributes (e.g., `instances[0].attributes.id`)

#### Requirement 1.4: Desired State Construction from Terraform

**Acceptance Criteria:**

1. WHEN `buildDesiredState()` is called, THE system SHALL replace manual state construction with Terraform state parsing
2. THE system SHALL read the Terraform state file using Requirement 1.1
3. THE system SHALL map all Terraform resources using Requirement 1.2
4. THE system SHALL extract all attributes using Requirement 1.3
5. THE system SHALL return a complete InfrastructureState with all resource types populated
6. WHEN the cluster has not been provisioned (no Terraform state), THE system SHALL return an error
7. THE system SHALL cache the parsed Terraform state for the duration of the drift detection operation
8. WHEN multiple drift detection operations run concurrently, THE system SHALL handle concurrent state file reads safely

### Epic 2: OpenStack Provider Enhancements

**User Story:** As a platform operator using OpenStack, I want comprehensive drift detection for all resource types, so that I can compare Terraform state against actual OpenStack infrastructure.

**Context:** After reading desired state from Terraform (Epic 1), we need to retrieve actual state from OpenStack APIs for all resource types to enable comprehensive drift detection.

#### Requirement 2.1: Security Groups State Retrieval

**Acceptance Criteria:**

1. WHEN `GetCurrentState()` is called on OpenStack provider, THE provider SHALL retrieve all security groups associated with the cluster
2. WHEN retrieving security groups, THE provider SHALL include security group rules (ingress/egress, protocols, ports, CIDR ranges)
3. WHEN security groups are retrieved, THE provider SHALL populate `InfrastructureState.SecurityGroups` with a list of SecurityGroup objects
4. EACH SecurityGroup object SHALL contain: ID, name, description, rules list, tags, and creation timestamp
5. EACH SecurityGroupRule SHALL contain: direction (ingress/egress), protocol, port range, remote IP prefix, and remote group ID
6. IF security group retrieval fails, THEN THE provider SHALL return an error with the OpenStack API error details
7. WHEN no security groups exist for the cluster, THE provider SHALL return an empty list without error

#### Requirement 2.2: Load Balancers State Retrieval

**Acceptance Criteria:**

1. WHEN `GetCurrentState()` is called on OpenStack provider, THE provider SHALL retrieve all load balancers associated with the cluster
2. WHEN retrieving load balancers, THE provider SHALL include listeners, pools, members, and health monitors
3. WHEN load balancers are retrieved, THE provider SHALL populate `InfrastructureState.LoadBalancers` with a list of LoadBalancer objects
4. EACH LoadBalancer object SHALL contain: ID, name, VIP address, VIP port, provisioning status, operating status, and listeners
5. EACH Listener SHALL contain: protocol, port, default pool ID, and connection limit
6. EACH Pool SHALL contain: protocol, load balancing algorithm, session persistence, and members list
7. EACH Member SHALL contain: address, port, weight, and health status
8. IF load balancer retrieval fails, THEN THE provider SHALL return an error with the OpenStack API error details

#### Requirement 2.3: Volumes State Retrieval

**Acceptance Criteria:**

1. WHEN `GetCurrentState()` is called on OpenStack provider, THE provider SHALL retrieve all volumes associated with the cluster
2. WHEN retrieving volumes, THE provider SHALL include volume attachments and metadata
3. WHEN volumes are retrieved, THE provider SHALL populate `InfrastructureState.Volumes` with a list of Volume objects
4. EACH Volume object SHALL contain: ID, name, size (GB), status, volume type, availability zone, and attachments
5. EACH VolumeAttachment SHALL contain: instance ID, device path, and attachment status
6. THE provider SHALL filter volumes by cluster tag or naming convention to avoid retrieving unrelated volumes
7. IF volume retrieval fails, THEN THE provider SHALL return an error with the OpenStack API error details

#### Requirement 2.4: Floating IPs State Retrieval

**Acceptance Criteria:**

1. WHEN `GetCurrentState()` is called on OpenStack provider, THE provider SHALL retrieve all floating IPs associated with the cluster
2. WHEN retrieving floating IPs, THE provider SHALL include port associations and fixed IP mappings
3. WHEN floating IPs are retrieved, THE provider SHALL populate `InfrastructureState.FloatingIPs` with a list of FloatingIP objects
4. EACH FloatingIP object SHALL contain: ID, floating IP address, fixed IP address, port ID, status, and associated instance ID
5. THE provider SHALL filter floating IPs by cluster tag or network association
6. IF floating IP retrieval fails, THEN THE provider SHALL return an error with the OpenStack API error details
7. WHEN a floating IP is not associated with any instance, THE provider SHALL include it with nil instance ID

### Epic 3: Secrets Management Enhancements

**User Story:** As a platform operator, I want automated secrets manifest scanning and organization detection, so that I can manage secrets across multiple clusters without manual path specification.

#### Requirement 3.1: Overlay Directory Manifest Scanning

**Acceptance Criteria:**

1. WHEN `ScanManifests()` is called on Secrets_Manager, THE manager SHALL recursively scan the Overlay_Directory for YAML files
2. WHEN scanning directories, THE manager SHALL identify files containing `kind: Secret` or SOPS encryption markers
3. WHEN a manifest file is found, THE manager SHALL extract the service name from the directory path (e.g., `services/keycloak/` → `keycloak`)
4. WHEN scanning completes, THE manager SHALL return a map of service names to manifest file paths
5. THE manager SHALL skip files in `.git/`, `node_modules/`, and other common ignore patterns
6. THE manager SHALL handle symlinks by following them once (no recursive symlink following)
7. IF the Overlay_Directory does not exist, THEN THE manager SHALL return an error with the expected path
8. WHEN the `--services` filter is provided, THE manager SHALL only scan manifests for specified services
9. WHEN scanning encounters a read permission error, THE manager SHALL log a warning and continue with remaining files

#### Requirement 3.2: Organization Detection from Cluster

**Acceptance Criteria:**

1. WHEN `DetermineOrganization()` is called with a cluster name, THE manager SHALL search the clusters directory for the cluster configuration
2. WHEN searching for clusters, THE manager SHALL scan `~/.config/opencenter/clusters/*/` for matching cluster names
3. WHEN a cluster is found, THE manager SHALL extract the organization from the parent directory name
4. WHEN the organization is found, THE manager SHALL return the organization ID and name
5. IF multiple organizations contain clusters with the same name, THEN THE manager SHALL return an error listing all matches
6. IF no organization contains the specified cluster, THEN THE manager SHALL return an error with available clusters
7. WHEN the `--organization` flag is explicitly provided, THE manager SHALL skip auto-detection and use the provided value
8. THE manager SHALL cache organization lookups for the duration of the command execution to avoid repeated filesystem scans

### Epic 4: Key Revocation User Context

**User Story:** As a security administrator, I want accurate audit trails showing which user performed key revocations, so that I can investigate security incidents and maintain compliance.

#### Requirement 4.1: User Context Extraction

**Acceptance Criteria:**

1. WHEN a key revocation operation is initiated, THE system SHALL extract the current user from the execution context
2. WHEN running in interactive mode, THE system SHALL use the Git user.email configuration as the user identifier
3. WHEN running in CI/CD mode, THE system SHALL use the `OPENCENTER_USER` environment variable if set
4. WHEN running as a service account, THE system SHALL use the `OPENCENTER_SERVICE_ACCOUNT` environment variable
5. IF no user context is available, THEN THE system SHALL use "system" as the default user identifier
6. WHEN the user context is determined, THE system SHALL include it in the `RevokedBy` field of the key entry
7. THE user identifier SHALL be validated as a valid email address or service account name (alphanumeric + hyphens)
8. WHEN the `--user` flag is provided to the revoke command, THE system SHALL use that as the actor (for admin operations)

#### Requirement 4.2: User-Key Mapping Implementation

**Acceptance Criteria:**

1. WHEN `isKeyOwnedByUser()` is called, THE system SHALL check the Key_Registry for key ownership records
2. WHEN checking ownership, THE system SHALL match the user email against the key's `CreatedBy` field
3. WHEN checking ownership, THE system SHALL match the user email against the key's `UsedBy` list
4. WHEN a key has multiple users in `UsedBy`, THE system SHALL return true if the specified user is in the list
5. WHEN the Key_Registry does not contain ownership information, THE system SHALL fall back to checking SOPS `.sops.yaml` recipient list
6. WHEN checking `.sops.yaml`, THE system SHALL match Age public keys against the user's registered keys
7. IF no ownership information is found, THEN THE system SHALL return false (deny revocation)
8. WHEN the `--force` flag is provided, THE system SHALL bypass ownership checks (admin override)

### Epic 5: Configuration Validation

**User Story:** As a platform operator, I want comprehensive configuration validation before saving, so that I can catch configuration errors early and prevent deployment failures.

#### Requirement 5.1: Full Config Struct Validation

**Acceptance Criteria:**

1. WHEN `Save()` is called on Config_Manager, THE manager SHALL validate the complete Config struct before persisting
2. WHEN validating the Config, THE manager SHALL use a ConfigStructureValidator that checks all required fields
3. WHEN validating the Config, THE manager SHALL verify cross-field constraints (e.g., node count ≥ control plane count)
4. WHEN validating the Config, THE manager SHALL check provider-specific requirements (e.g., OpenStack requires network ID)
5. WHEN validation fails, THE manager SHALL return a detailed error listing all validation failures (not just the first)
6. WHEN validation succeeds, THE manager SHALL proceed with saving the configuration
7. THE ConfigStructureValidator SHALL support validation rules defined in JSON Schema or similar declarative format
8. WHEN the `--skip-validation` flag is provided, THE manager SHALL bypass validation with a warning (for emergency fixes)
9. WHEN validation is skipped, THE manager SHALL log the skip event to the audit log

### Epic 6: Audit Logging Implementation

**User Story:** As a compliance officer, I want cryptographically signed audit logs with signature verification, so that I can prove log integrity during security audits.

#### Requirement 6.1: Signature Generation

**Acceptance Criteria:**

1. WHEN an audit event is logged, THE system SHALL generate a cryptographic signature for the event
2. WHEN generating signatures, THE system SHALL use Ed25519 signing keys stored in `~/.config/opencenter/audit/signing-key`
3. WHEN the signing key does not exist, THE system SHALL generate a new key pair on first use
4. WHEN signing an event, THE system SHALL create a signature over the JSON-serialized event data
5. WHEN the signature is created, THE system SHALL append it to the event as a `signature` field
6. THE signature SHALL be base64-encoded for storage in the audit log
7. WHEN the audit log is rotated, THE system SHALL sign the final entry with a rotation marker
8. IF signing fails, THEN THE system SHALL log the event without a signature and record a signing failure event

#### Requirement 6.2: Signature Verification

**Acceptance Criteria:**

1. WHEN `verifyAuditLogIntegrity()` is called, THE system SHALL verify signatures for all events in the audit log
2. WHEN verifying signatures, THE system SHALL use the public key corresponding to the signing key
3. WHEN verifying an event, THE system SHALL recompute the signature over the event data (excluding the signature field)
4. WHEN a signature does not match, THE system SHALL report the event index and timestamp of the tampered entry
5. WHEN verification completes, THE system SHALL return a report showing total events, verified events, and failed events
6. IF the signing key is not found, THEN THE system SHALL return an error indicating verification is not possible
7. WHEN the `--repair` flag is provided, THE system SHALL re-sign events with missing or invalid signatures (admin operation)
8. WHEN verification succeeds for all events, THE system SHALL return exit code 0

#### Requirement 6.3: Log Parsing Implementation

**Acceptance Criteria:**

1. WHEN `parseAuditLog()` is called, THE system SHALL read the audit log file and parse JSON events
2. WHEN parsing events, THE system SHALL handle multi-line JSON events (one event per line)
3. WHEN parsing events, THE system SHALL deserialize each event into an AuditEvent struct
4. WHEN the `--since` filter is provided, THE system SHALL only return events after the specified timestamp
5. WHEN the `--event-type` filter is provided, THE system SHALL only return events matching the specified type
6. WHEN the `--user` filter is provided, THE system SHALL only return events for the specified user
7. IF the audit log is corrupted (invalid JSON), THEN THE system SHALL skip the corrupted line and log a warning
8. WHEN parsing completes, THE system SHALL return a list of AuditEvent objects sorted by timestamp
9. WHEN the audit log does not exist, THE system SHALL return an empty list without error

### Epic 7: Drift Detection Callbacks

**User Story:** As a platform operator integrating with monitoring systems, I want HTTP callbacks when drift is detected, so that I can trigger automated remediation workflows or alert on-call engineers.

#### Requirement 7.1: Callback HTTP POST Implementation

**Acceptance Criteria:**

1. WHEN drift is detected and `--callback-url` is provided, THE system SHALL send an HTTP POST request to the callback URL
2. WHEN sending the callback, THE system SHALL include the drift report as JSON in the request body
3. THE callback request body SHALL contain: cluster name, drift count, drift items list, detection timestamp, and severity
4. WHEN sending the callback, THE system SHALL set `Content-Type: application/json` header
5. WHEN the `--callback-auth` flag is provided, THE system SHALL include the value as an `Authorization` header
6. WHEN the callback request succeeds (2xx status), THE system SHALL log the callback success
7. WHEN the callback request fails (non-2xx status or network error), THE system SHALL log the failure but continue execution
8. THE system SHALL set a 10-second timeout for callback requests to prevent hanging
9. WHEN the `--callback-retry` flag is provided, THE system SHALL retry failed callbacks up to 3 times with exponential backoff

### Epic 8: Keycloak Backup Automation

**User Story:** As a platform operator, I want automated Keycloak realm backups uploaded to object storage, so that I can recover from Keycloak failures without manual intervention.

#### Requirement 8.1: S3 Upload Implementation

**Acceptance Criteria:**

1. WHEN the Keycloak backup CronJob completes, THE backup script SHALL upload the realm backup to S3-compatible object storage
2. WHEN uploading to S3, THE script SHALL use AWS CLI or equivalent S3 client
3. WHEN uploading, THE script SHALL read S3 credentials from Kubernetes secrets (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
4. WHEN uploading, THE script SHALL read the S3 bucket name from a ConfigMap or environment variable
5. WHEN the upload succeeds, THE script SHALL log the S3 object key and upload timestamp
6. WHEN the upload fails, THE script SHALL retry up to 3 times with exponential backoff
7. IF all retries fail, THEN THE script SHALL exit with a non-zero status to trigger Kubernetes Job failure alerts
8. WHEN uploading, THE script SHALL set object metadata including backup timestamp and Keycloak version
9. WHEN the `BACKUP_RETENTION_DAYS` environment variable is set, THE script SHALL configure S3 lifecycle policy for automatic deletion

#### Requirement 8.2: Swift Upload Implementation

**Acceptance Criteria:**

1. WHEN the Keycloak backup CronJob completes, THE backup script SHALL support uploading to OpenStack Swift storage
2. WHEN uploading to Swift, THE script SHALL use Swift CLI or python-swiftclient
3. WHEN uploading, THE script SHALL read Swift credentials from Kubernetes secrets (OS_AUTH_URL, OS_USERNAME, OS_PASSWORD, OS_PROJECT_NAME)
4. WHEN uploading, THE script SHALL read the Swift container name from a ConfigMap or environment variable
5. WHEN the upload succeeds, THE script SHALL log the Swift object name and upload timestamp
6. WHEN the upload fails, THE script SHALL retry up to 3 times with exponential backoff
7. IF all retries fail, THEN THE script SHALL exit with a non-zero status
8. WHEN uploading, THE script SHALL set object metadata including backup timestamp and Keycloak version
9. THE script SHALL auto-detect storage backend (S3 vs Swift) based on environment variables present

## Non-Functional Requirements

### Performance

1. OpenStack resource retrieval SHALL complete within 30 seconds for clusters with up to 100 resources
2. Manifest scanning SHALL process up to 1000 files within 5 seconds
3. Audit log parsing SHALL process up to 10,000 events within 2 seconds
4. Drift detection callbacks SHALL not delay drift detection by more than 1 second

### Security

1. Audit log signing keys SHALL be protected with file permissions 0600
2. Object storage credentials SHALL never be logged or printed to stdout
3. Callback URLs SHALL support HTTPS with certificate validation
4. User context SHALL be validated to prevent injection attacks

### Reliability

1. All network operations SHALL have configurable timeouts
2. All external API calls SHALL implement retry logic with exponential backoff
3. Partial failures SHALL not prevent completion of independent operations
4. All errors SHALL include actionable remediation guidance

### Maintainability

1. All TODO comments SHALL be removed upon implementation completion
2. All new code SHALL include unit tests with ≥80% coverage
3. All new features SHALL include integration tests
4. All new features SHALL be documented in user-facing documentation

## Dependencies

### External Dependencies

1. OpenStack SDK (gophercloud) for resource retrieval
2. AWS SDK for Go (S3 operations)
3. python-swiftclient (Swift operations)
4. Go crypto/ed25519 for audit log signing

### Internal Dependencies

1. `internal/cloud/openstack/provider.go` - OpenStack provider implementation
2. `internal/secrets/manager.go` - Secrets management
3. `internal/secrets/revocation.go` - Key revocation
4. `internal/config/manager.go` - Configuration management
5. `cmd/cluster_audit_log.go` - Audit logging commands
6. `cmd/cluster_drift.go` - Drift detection commands
7. `internal/gitops/templates/.../keycloak-backup-cronjob.yaml.tpl` - Backup template

## Implementation Phases

### Phase 1: Foundation (Epics 1, 3, 4, 5)
- Terraform state integration (Epic 1)
- Secrets manifest scanning (Epic 3)
- Organization detection (Epic 3)
- User context extraction (Epic 4)
- Config validation (Epic 5)

### Phase 2: Observability (Epics 6, 7)
- Audit log signing (Epic 6)
- Audit log parsing (Epic 6)
- Drift callbacks (Epic 7)

### Phase 3: Infrastructure (Epics 2, 8)
- OpenStack resource retrieval (Epic 2)
- Keycloak backup automation (Epic 8)

## Success Criteria

1. All 12 TODO items (4-15) are resolved with production-ready implementations
2. Terraform state integration enables accurate drift detection for all resource types
3. OpenStack provider retrieves all resource types (security groups, load balancers, volumes, floating IPs)
4. All new code has ≥80% test coverage
5. All integration tests pass
6. Documentation explains Terraform state workflow and drift detection
7. No new TODO comments are introduced
8. All existing functionality continues to work (no regressions)

## References

- [OpenStack API Documentation](https://docs.openstack.org/api-ref/)
- [Terraform State Format](https://www.terraform.io/docs/language/state/index.html)
- [Terraform OpenStack Provider](https://registry.terraform.io/providers/terraform-provider-openstack/openstack/latest/docs)
- [AWS S3 API Documentation](https://docs.aws.amazon.com/s3/)
- [OpenStack Swift API Documentation](https://docs.openstack.org/swift/latest/api/)
- [Ed25519 Signature Scheme](https://ed25519.cr.yp.to/)
- [Multi-Cluster Secrets Management Spec](../multi-cluster-secrets-management/requirements.md)
- [Drift Detection Explanation](../../docs/explanation/drift-detection.md)
