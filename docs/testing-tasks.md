# Testing Tasks - Prioritized Fix Order

**Test Run Summary:** 145 scenarios (105 passed, 40 failed)  
**Date:** 2024-11-23

## Priority 1: Core Configuration System (Blocking Issues)

These failures indicate fundamental issues with the cluster initialization and configuration file structure that block most other functionality.

### 1.1 Cluster Init File Path Issues
**Impact:** HIGH - Blocks cluster creation workflow  
**Affected Tests:** 18 scenarios

- `cluster_init.feature:5` - Initialize a cluster with defaults
- `cluster_init.feature:21` - Init generates a SOPS key when not provided
- `cluster_init.feature:38` - Init creates clusters subdirectory and cluster directory structure
- `cluster_init.feature:45` - Init creates cluster-specific secrets directory structure
- `cluster_init.feature:70` - SOPS key generation uses cluster-specific directory
- `cluster_init.feature:75` - Cluster directory creation with special characters in name
- `cluster_init.feature:109` - Init cluster without organization uses opencenter organization
- `organization_init.feature:22` - Init cluster with organization creates cluster configuration in correct location
- `organization_init.feature:38` - Init cluster without organization uses cluster name as organization
- `organization_init.feature:46` - Init multiple clusters in same organization share GitOps root
- `organization_init.feature:57` - Init cluster with organization and force flag overwrites existing
- `cli_configuration_system.feature:152` - Cluster name is used as organization when none specified
- `cli_configuration_system.feature:223` - Custom configuration paths work correctly
- `cluster_commands.feature:69` - init <cluster-name> creates a YAML with defaults
- `cluster_commands_integration.feature:6` - Cluster select, info, and validate work with new directory structure
- `workflow.feature:17` - Initialize with org, select, validate VRRP requirement, render setup, and bootstrap

**Root Cause:** Configuration files are being created in wrong locations. Expected path pattern:
- Expected: `clusters/<org>/.{cluster}-config.yaml` or `clusters/<org>/infrastructure/clusters/<cluster>/.{cluster}-config.yaml`
- Actual: Likely `clusters/<cluster>.yaml` or similar legacy path

**Fix Strategy:**
1. Review cluster init logic in `cmd/cluster_init.go`
2. Verify organization-based directory structure creation
3. Ensure config file naming follows `.{cluster}-config.yaml` pattern
4. Update path resolution logic to handle organization-based structure

---

## Priority 2: GitOps Setup and Template Rendering

These failures affect the GitOps workflow setup, which is critical for cluster deployment.

### 2.1 GitOps Template Materialization
**Impact:** HIGH - Blocks cluster setup workflow  
**Affected Tests:** 5 scenarios

- `cli_behaviors.feature:136` - Setup materializes GitOps template into git_dir (missing README.md)
- `cli_behaviors.feature:145` - Running setup again is idempotent (should show "already initialized")
- `cli_behaviors.feature:155` - Forced setup overwrites existing files
- `gitops_setup.feature:17` - setup materializes embedded templates into git_dir
- `gitops_setup.feature:27` - setup is idempotent when run repeatedly
- `gitops_setup.feature:37` - setup --force overwrites existing files

**Root Cause:** Template rendering is not properly copying/generating files into the GitOps directory. The setup command completes but doesn't materialize expected files like README.md.

**Fix Strategy:**
1. Review `cmd/cluster_setup.go` template rendering logic
2. Verify embedded template resources are properly included in build
3. Check file copy/generation logic in setup command
4. Ensure idempotency checks work correctly

### 2.2 GitOps Directory Validation
**Impact:** MEDIUM - Validation should catch missing git_dir  
**Affected Tests:** 2 scenarios

- `cli_behaviors.feature:190` - opencenter.gitops.git_dir missing -> error on setup
- `gitops_setup.feature:49` - setup errors when no active cluster or git_dir is missing

**Root Cause:** Validation is not properly enforcing required git_dir configuration.

**Fix Strategy:**
1. Review validation logic in cluster setup/validate commands
2. Ensure git_dir is marked as required field
3. Add proper error handling for missing git_dir

---

## Priority 3: Cluster Selection and Listing

These failures affect cluster management and navigation.

### 3.1 Cluster List Command
**Impact:** MEDIUM - Affects cluster discovery  
**Affected Tests:** 2 scenarios

- `cli_configuration_system.feature:189` - Cluster list works with organization-based structure
- `cluster_commands_integration.feature:43` - Multiple clusters work correctly with new directory structure

**Root Cause:** List command not properly scanning organization-based directory structure.

**Fix Strategy:**
1. Review `cmd/cluster_list.go` directory scanning logic
2. Update to traverse `clusters/<org>/` structure
3. Format output to show `<org>/<cluster>` naming

### 3.2 Cluster Select Error Messages
**Impact:** LOW - Error messages need updating  
**Affected Tests:** 2 scenarios

- `cluster_commands_integration.feature:30` - Cluster commands handle non-existent clusters correctly
- `config_select_list_info.feature:70` - Selecting a non-existent cluster yields a helpful error

**Root Cause:** Error messages reference old path structure ("cluster configuration directory" vs "cluster configuration file").

**Fix Strategy:**
1. Update error messages in `cmd/cluster_select.go`
2. Ensure error messages reflect organization-based structure

---

## Priority 4: Validation Logic Issues

These failures indicate validation rules are not being properly enforced.

### 4.1 VRRP Validation
**Impact:** MEDIUM - Networking validation not working  
**Affected Tests:** 2 scenarios

- `validation.feature:422` - prosys.dev.dfw3 cluster VRRP validation fails when IP missing
- `workflow.feature:17` - Initialize with org, select, validate VRRP requirement (partial failure)

**Root Cause:** Validation should fail when `vrrp_enabled=true` but `vrrp_ip=""`, but it's passing.

**Fix Strategy:**
1. Review validation rules in cluster validate command
2. Add conditional validation: if `use_octavia=false` and `vrrp_enabled=true`, then `vrrp_ip` is required
3. Ensure validation runs before setup/bootstrap

### 4.2 Service Configuration Validation
**Impact:** MEDIUM - Service validation not enforcing requirements  
**Affected Tests:** 4 scenarios

- `config_template_rendering.feature:491` - Missing cert-manager secrets should fail validation
- `config_template_rendering.feature:507` - Missing loki secrets should fail validation
- `config_template_rendering.feature:538` - Invalid admin email should fail validation
- `config_template_rendering.feature:562` - Invalid cluster FQDN should fail validation

**Root Cause:** Service-specific validation rules not being enforced.

**Fix Strategy:**
1. Review service validation logic
2. Add validation for required secrets when services are enabled
3. Add format validation for email and FQDN fields

---

## Priority 5: Bootstrap Command Issues

These failures affect the final deployment step.

### 5.1 Bootstrap Cleanup Logic
**Impact:** MEDIUM - Bootstrap failing on cleanup  
**Affected Tests:** 1 scenario

- `cli_behaviors.feature:170` - Bootstrap pushes the local repo to a remote

**Root Cause:** Bootstrap is trying to clean up files that don't exist (cluster.rkestate, kube_config_cluster.yml, terraform.tfstate*), causing make clean to fail.

**Fix Strategy:**
1. Review `cmd/cluster_bootstrap.go` cleanup logic
2. Make cleanup more resilient to missing files
3. Use `rm -f` or check file existence before deletion

---

## Priority 6: Cluster Info Output Format

These failures indicate the info command output format has changed.

### 6.1 Info Command Output
**Impact:** LOW - Output format mismatch  
**Affected Tests:** 2 scenarios

- `cli_configuration_system.feature:340` - Configuration system integrates with complete cluster lifecycle
- `cluster_commands.feature:40` - "openCenter cluster" prints help with all subcommands

**Root Cause:** 
- Info command no longer outputs "provider: openstack" in expected format
- Help text changed (removed "openCenter cluster info" from subcommand list)

**Fix Strategy:**
1. Review `cmd/cluster_info.go` output format
2. Either update tests to match new format or restore old format
3. Verify help text generation includes all subcommands

---

## Priority 7: Cluster Destroy Issues

These failures affect cluster cleanup.

### 7.1 Destroy Command
**Impact:** LOW - Cleanup not removing config files  
**Affected Tests:** 2 scenarios

- `cluster.feature:169` - Destroy a cluster
- `destroy.feature:8` - Destroy removes config and GitOps directory

**Root Cause:** Destroy command not removing configuration files in organization-based structure.

**Fix Strategy:**
1. Review `cmd/cluster_destroy.go` file deletion logic
2. Update to handle organization-based directory structure
3. Ensure both config file and directories are removed

---

## Priority 8: Init Command Validation

These failures indicate init validation issues.

### 8.1 Init with Organization Validation
**Impact:** LOW - Validation too strict during init  
**Affected Tests:** 2 scenarios

- `cluster_init.feature:153` - Init cluster with organization validates organization name in config
- `cluster_init.feature:9` - Initialise a new cluster with default settings (contains "local.")

**Root Cause:** 
- Init is running validation that requires git_dir to be set
- Generated config contains "local." references that shouldn't be there

**Fix Strategy:**
1. Review init validation - should be minimal during init
2. Remove "local." prefix from generated config values
3. Defer full validation until setup/validate commands

---

## Recommended Fix Order

1. **Start with Priority 1** - Fix cluster init file path issues. This is the foundation that blocks everything else.
2. **Move to Priority 2** - Fix GitOps setup and template rendering. This enables the deployment workflow.
3. **Address Priority 3** - Fix cluster selection and listing to enable cluster management.
4. **Handle Priority 4** - Fix validation logic to ensure proper configuration enforcement.
5. **Fix Priority 5** - Resolve bootstrap issues to complete the deployment workflow.
6. **Clean up Priority 6-8** - Address remaining output format and cleanup issues.

## Testing Strategy

After each priority level is fixed:
1. Run the full test suite: `go test ./... -v`
2. Verify all tests in that priority pass
3. Check for regression in previously passing tests
4. Move to next priority level

## Notes

- Many failures cascade from the Priority 1 issues (wrong file paths)
- Fixing Priority 1 may automatically resolve some Priority 2-3 failures
- The organization-based directory structure appears to be partially implemented
- Some tests may need updates to match new behavior (especially output format tests)
