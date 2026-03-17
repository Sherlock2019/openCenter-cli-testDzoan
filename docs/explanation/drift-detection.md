---
id: drift-detection
title: "Infrastructure Drift Detection"
sidebar_label: Drift Detection
description: How openCenter detects and remediates differences between desired and actual infrastructure state.
doc_type: explanation
audience: "platform operators, architects"
tags: [drift, terraform, infrastructure, reconciliation]
---

# Infrastructure Drift Detection

**Purpose:** For platform operators, explains how drift detection works, covering the complete workflow from Terraform state to drift remediation.

Understanding drift detection helps you maintain infrastructure consistency and catch unauthorized changes. This explanation covers how openCenter detects differences between desired and actual infrastructure state.

## What is Infrastructure Drift?

**Definition:** Infrastructure drift occurs when the actual state of cloud resources differs from the desired state defined in your configuration.

**Common Causes:**
- Manual changes in cloud provider console (OpenStack Horizon, AWS Console)
- External automation tools modifying resources
- Infrastructure failures causing resource deletion
- Quota exhaustion preventing resource creation
- Network issues causing partial deployments

**Why Drift Matters:**
- Security risks (firewall rules changed, ports opened)
- Availability issues (load balancers deleted, volumes detached)
- Configuration inconsistency (production differs from staging)
- Compliance violations (resources not matching approved configuration)

**Evidence:** `cmd/cluster_drift.go`, `internal/cloud/factory.go`

## The Source of Truth: Terraform State

### Why Terraform State?

When you deploy a cluster with `opencenter cluster bootstrap`, the system:

1. Reads your YAML configuration
2. Renders Terraform configuration (`main.tf`)
3. Runs `terraform apply` to provision infrastructure
4. Terraform saves what it created to `terraform.tfstate`

**The Terraform state file is the authoritative source of truth** for what infrastructure should exist.

**Why not use YAML config directly?**
- YAML is high-level (node counts, flavors)
- Terraform translates to low-level resources (security groups, load balancers, volumes)
- Terraform state contains actual resource IDs, not just names
- Terraform state includes computed values (IP addresses, resource relationships)

**Evidence:** `docs/tutorials/openstack-first-cluster.md:Step 8`, `docs/explanation/configuration-lifecycle.md:Stage 5`

## Drift Detection Workflow

### Complete Flow

```
1. User Configuration (YAML)
   ↓
2. Terraform Rendering (opencenter cluster setup)
   ↓
3. Infrastructure Provisioning (terraform apply)
   ↓
4. Terraform State File (terraform.tfstate)
   ↓
5. Drift Detection (opencenter cluster drift detect)
   ├─ Read Terraform State (desired state)
   ├─ Query Cloud Provider API (actual state)
   └─ Compare and Report Differences
```

### Step-by-Step Example

**Step 1: User Configuration**

```yaml
# .k8s-prod-cluster-config.yaml
opencenter:
  cluster:
    master_count: 3
    worker_count: 3
  networking:
    use_octavia: true
```

**Step 2: Terraform Rendering**

```bash
opencenter cluster setup prod-cluster --render
```

Generates `main.tf`:

```hcl
resource "openstack_networking_secgroup_v2" "control_plane" {
  name = "prod-cluster-control-plane-sg"
}

resource "openstack_networking_secgroup_rule_v2" "api_server" {
  security_group_id = openstack_networking_secgroup_v2.control_plane.id
  direction         = "ingress"
  protocol          = "tcp"
  port_range_min    = 6443
  port_range_max    = 6443
  remote_ip_prefix  = "10.0.0.0/8"  # Only internal network
}
```


**Step 3: Infrastructure Provisioning**

```bash
opencenter cluster bootstrap prod-cluster
```

Terraform creates:
- Security group `prod-cluster-control-plane-sg`
- Security rule allowing port 6443 from 10.0.0.0/8
- 3 control plane VMs
- 3 worker VMs
- Load balancer
- Volumes
- Floating IPs

**Step 4: Terraform State**

Terraform saves to `terraform.tfstate`:

```json
{
  "resources": [
    {
      "type": "openstack_networking_secgroup_v2",
      "name": "control_plane",
      "instances": [{
        "attributes": {
          "id": "sg-abc123",
          "name": "prod-cluster-control-plane-sg"
        }
      }]
    },
    {
      "type": "openstack_networking_secgroup_rule_v2",
      "name": "api_server",
      "instances": [{
        "attributes": {
          "id": "sgr-def456",
          "security_group_id": "sg-abc123",
          "direction": "ingress",
          "protocol": "tcp",
          "port_range_min": 6443,
          "port_range_max": 6443,
          "remote_ip_prefix": "10.0.0.0/8"
        }
      }]
    }
  ]
}
```

**Step 5: Manual Change (Drift Introduced)**

Operator opens port 6443 to the internet in OpenStack Horizon:
- Changes `remote_ip_prefix` from `10.0.0.0/8` to `0.0.0.0/0`

**Step 6: Drift Detection**

```bash
opencenter cluster drift detect prod-cluster
```

System:
1. Reads `terraform.tfstate` (desired: port 6443 from 10.0.0.0/8)
2. Queries OpenStack API (actual: port 6443 from 0.0.0.0/0)
3. Compares and detects drift

**Output:**

```
Drift Report for Cluster: prod-cluster
Detected At: 2026-02-17T10:30:00Z
Overall Severity: critical
Reconcilable: true

Drifts:
  1. security_group prod-cluster-control-plane-sg (sg-abc123)
     Field: rules[0].remote_ip_prefix
     Expected: 10.0.0.0/8
     Actual: 0.0.0.0/0
     Severity: critical
     Reconcilable: true
     Message: Security group rule allows access from internet (expected: internal only)
```

**Evidence:** `cmd/cluster_drift.go:buildDesiredState()`, `internal/cloud/openstack/provider.go:GetCurrentState()`


## Terraform State Structure

### State File Location

```
<git_dir>/infrastructure/clusters/<cluster>/terraform.tfstate
```

Example:
```
~/prod-cluster-gitops/infrastructure/clusters/prod-cluster/terraform.tfstate
```

### State File Format

Terraform state is JSON with this structure:

```json
{
  "version": 4,
  "terraform_version": "1.7.0",
  "resources": [
    {
      "type": "openstack_compute_instance_v2",
      "name": "master_1",
      "provider": "provider[\"registry.terraform.io/terraform-provider-openstack/openstack\"]",
      "instances": [
        {
          "schema_version": 0,
          "attributes": {
            "id": "abc-123-def-456",
            "name": "prod-cluster-master-1",
            "flavor_id": "gp.0.4.8",
            "image_id": "799dcf97-3656-4361-8187-13ab1b295e33",
            "metadata": {
              "cluster": "prod-cluster",
              "role": "control-plane"
            },
            "network": [
              {
                "name": "prod-cluster-network",
                "fixed_ip_v4": "10.2.128.10"
              }
            ]
          }
        }
      ]
    }
  ]
}
```

### Resource Types in State

**Compute Resources:**
- `openstack_compute_instance_v2` - Virtual machines
- `openstack_compute_volume_attach_v2` - Volume attachments

**Networking Resources:**
- `openstack_networking_network_v2` - Networks
- `openstack_networking_subnet_v2` - Subnets
- `openstack_networking_router_v2` - Routers
- `openstack_networking_secgroup_v2` - Security groups
- `openstack_networking_secgroup_rule_v2` - Security rules
- `openstack_networking_floatingip_v2` - Floating IPs
- `openstack_networking_floatingip_associate_v2` - Floating IP associations

**Load Balancing Resources:**
- `openstack_lb_loadbalancer_v2` - Load balancers
- `openstack_lb_listener_v2` - Load balancer listeners
- `openstack_lb_pool_v2` - Load balancer pools
- `openstack_lb_member_v2` - Load balancer members

**Storage Resources:**
- `openstack_blockstorage_volume_v3` - Block storage volumes

**Evidence:** Terraform OpenStack provider documentation, `internal/cloud/factory.go:InfrastructureState`


## Drift Detection Components

### Component 1: Terraform State Reader

**Purpose:** Parse Terraform state file and extract resource information.

**Process:**
1. Locate state file at `<git_dir>/infrastructure/clusters/<cluster>/terraform.tfstate`
2. Read and parse JSON
3. Extract `resources` array
4. Map each resource to InfrastructureState type

**Example Mapping:**

```go
// Terraform resource
{
  "type": "openstack_networking_secgroup_v2",
  "instances": [{
    "attributes": {
      "id": "sg-abc123",
      "name": "prod-cluster-control-plane-sg",
      "description": "Control plane security group"
    }
  }]
}

// Mapped to InfrastructureState
SecurityGroup{
  ID:   "sg-abc123",
  Name: "prod-cluster-control-plane-sg",
  Rules: []SecurityRule{...}
}
```

**Evidence:** `.kiro/specs/todo-completion-phase-1/requirements.md:Epic 1`

### Component 2: Cloud Provider API Client

**Purpose:** Query cloud provider APIs to get actual resource state.

**Process:**
1. Authenticate with cloud provider (OpenStack, AWS, etc.)
2. Query APIs for each resource type:
   - Compute: List instances
   - Networking: List networks, security groups, floating IPs
   - Load Balancing: List load balancers
   - Storage: List volumes
3. Filter resources by cluster tag/name
4. Convert API responses to InfrastructureState type

**Example OpenStack Query:**

```go
// Query security groups
client := openstack.NewNetworkV2(provider, opts)
allPages, _ := secgroups.List(client, secgroups.ListOpts{
  Tags: "prod-cluster",
}).AllPages()
sgList, _ := secgroups.ExtractGroups(allPages)

// Convert to InfrastructureState
for _, sg := range sgList {
  state.SecurityGroups = append(state.SecurityGroups, SecurityGroup{
    ID:   sg.ID,
    Name: sg.Name,
    Rules: convertRules(sg.Rules),
  })
}
```

**Evidence:** `internal/cloud/openstack/provider.go:GetCurrentState()`, `.kiro/specs/todo-completion-phase-1/requirements.md:Epic 2`

### Component 3: Drift Comparator

**Purpose:** Compare desired state (Terraform) with actual state (cloud provider) and identify differences.

**Process:**
1. Create maps of resources by name for efficient lookup
2. Check for missing resources (in desired but not actual)
3. Check for extra resources (in actual but not desired)
4. Check for configuration drift (resource exists but attributes differ)
5. Classify drift by severity (critical, warning, info)
6. Determine if drift is reconcilable (can be auto-fixed)

**Example Comparison:**

```go
// Desired (from Terraform state)
desiredSG := SecurityGroup{
  Name: "prod-cluster-control-plane-sg",
  Rules: []SecurityRule{
    {Protocol: "tcp", PortMin: 6443, PortMax: 6443, RemoteIP: "10.0.0.0/8"},
  },
}

// Actual (from OpenStack API)
actualSG := SecurityGroup{
  Name: "prod-cluster-control-plane-sg",
  Rules: []SecurityRule{
    {Protocol: "tcp", PortMin: 6443, PortMax: 6443, RemoteIP: "0.0.0.0/0"},
  },
}

// Drift detected
DriftItem{
  ResourceType: "security_group",
  ResourceName: "prod-cluster-control-plane-sg",
  Field: "rules[0].remote_ip",
  Expected: "10.0.0.0/8",
  Actual: "0.0.0.0/0",
  Severity: SeverityCritical,
  Reconcilable: true,
}
```

**Evidence:** `internal/cloud/openstack/provider.go:DetectDrift()`, `internal/cloud/factory.go:DriftReport`


## Drift Severity Levels

### Critical Severity

**Definition:** Drift that impacts cluster availability, security, or data integrity.

**Examples:**
- Control plane nodes deleted
- API server security group allows public access
- etcd volumes detached
- Load balancer deleted
- Network routing broken

**Response:** Immediate remediation required.

### Warning Severity

**Definition:** Drift that may impact cluster functionality but not immediately critical.

**Examples:**
- Worker nodes deleted (but cluster still functional)
- Monitoring volumes detached
- Non-critical security group rules changed
- Floating IPs unassigned

**Response:** Remediate during next maintenance window.

### Info Severity

**Definition:** Drift that doesn't impact functionality but indicates configuration inconsistency.

**Examples:**
- Resource tags changed
- Resource descriptions modified
- Metadata updated
- Non-functional labels changed

**Response:** Optional remediation, document if intentional.

**Evidence:** `internal/cloud/factory.go:Severity`

## Drift Reconciliation

### Reconcilable vs Non-Reconcilable Drift

**Reconcilable Drift:**
- Can be automatically fixed by re-applying Terraform
- Examples: Tags updated, security rules changed, metadata modified

**Non-Reconcilable Drift:**
- Requires manual intervention or resource recreation
- Examples: Resources deleted, flavor changed (requires VM rebuild), image changed

### Reconciliation Process

```bash
# Detect drift
opencenter cluster drift detect prod-cluster

# Review drift report
# Decide: Accept drift or remediate

# Option 1: Remediate (revert to desired state)
opencenter cluster drift reconcile prod-cluster

# Option 2: Accept drift (update Terraform to match actual)
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform refresh  # Update state to match actual
git commit -m "Accept drift: security group rule updated"
```

### Reconciliation Methods

**Method 1: Terraform Apply**

For reconcilable drift, re-run Terraform:

```bash
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform plan  # Review changes
terraform apply # Apply changes
```

Terraform will:
- Create missing resources
- Update changed attributes
- Leave extra resources (manual deletion required)

**Method 2: Manual Remediation**

For non-reconcilable drift:

```bash
# Example: Recreate deleted VM
terraform taint openstack_compute_instance_v2.master_1
terraform apply
```

**Method 3: Accept Drift**

If drift is intentional:

```bash
# Update Terraform state to match actual
terraform refresh
git add terraform.tfstate
git commit -m "Accept drift: intentional change"
```

**Evidence:** `cmd/cluster_drift.go:newClusterDriftReconcileCmd()`, `internal/cloud/openstack/provider.go:ReconcileDrift()`


## Drift Prevention Strategies

### Strategy 1: Immutable Infrastructure

**Practice:** Never modify infrastructure manually. Always use Terraform.

**Enforcement:**
- Restrict OpenStack Horizon access (read-only for most users)
- Use RBAC to limit who can modify resources
- Audit all manual changes via OpenStack audit logs

**Benefits:**
- Prevents drift at the source
- Maintains configuration as code discipline
- Simplifies troubleshooting (all changes in Git)

### Strategy 2: Automated Drift Detection

**Practice:** Run drift detection on a schedule.

**Implementation:**

```bash
# Schedule drift detection every 24 hours
opencenter cluster drift schedule prod-cluster --interval=24h
```

Or use cron:

```bash
# Add to crontab
0 */6 * * * opencenter cluster drift detect prod-cluster --output=json > /var/log/drift-$(date +\%Y\%m\%d-\%H\%M).json
```

**Benefits:**
- Early detection of unauthorized changes
- Automated monitoring
- Historical drift tracking

### Strategy 3: Drift Alerts

**Practice:** Send alerts when drift is detected.

**Implementation:**

```bash
# Send drift reports to webhook
opencenter cluster drift schedule prod-cluster \
  --interval=6h \
  --callback=https://alerts.example.com/drift
```

Webhook receives:

```json
{
  "cluster_name": "prod-cluster",
  "detected_at": "2026-02-17T10:30:00Z",
  "overall_severity": "critical",
  "drifts": [
    {
      "resource_type": "security_group",
      "resource_name": "prod-cluster-control-plane-sg",
      "field": "rules[0].remote_ip_prefix",
      "expected": "10.0.0.0/8",
      "actual": "0.0.0.0/0",
      "severity": "critical"
    }
  ]
}
```

**Benefits:**
- Real-time notifications
- Integration with incident management (PagerDuty, Opsgenie)
- Automated remediation workflows

**Evidence:** `cmd/cluster_drift.go:newClusterDriftScheduleCmd()`, `.kiro/specs/todo-completion-phase-1/requirements.md:Epic 7`

## Common Drift Scenarios

### Scenario 1: Security Group Rule Changed

**Drift:**
- Expected: Port 6443 from 10.0.0.0/8
- Actual: Port 6443 from 0.0.0.0/0

**Impact:** Critical - API server exposed to internet

**Remediation:**
```bash
opencenter cluster drift reconcile prod-cluster
# Or manually:
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform apply
```

### Scenario 2: Worker Node Deleted

**Drift:**
- Expected: 3 worker nodes
- Actual: 2 worker nodes

**Impact:** Warning - Reduced capacity

**Remediation:**
```bash
# Recreate missing node
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform apply
```

### Scenario 3: Volume Detached

**Drift:**
- Expected: Volume attached to worker-1
- Actual: Volume detached

**Impact:** Critical - Data loss risk

**Remediation:**
```bash
# Reattach volume
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform apply
```

### Scenario 4: Load Balancer Deleted

**Drift:**
- Expected: Load balancer for API server
- Actual: Load balancer missing

**Impact:** Critical - API server unreachable

**Remediation:**
```bash
# Recreate load balancer
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform apply
```

### Scenario 5: Floating IP Unassigned

**Drift:**
- Expected: Floating IP assigned to load balancer
- Actual: Floating IP unassigned

**Impact:** Critical - External access broken

**Remediation:**
```bash
# Reassign floating IP
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform apply
```


## Best Practices

### 1. Run Drift Detection Regularly

**Practice:** Schedule drift detection to run automatically.

**Frequency:**
- Production: Every 6 hours
- Staging: Every 12 hours
- Development: Every 24 hours

**Rationale:** Early detection prevents small drift from becoming large problems.

### 2. Investigate All Critical Drift Immediately

**Practice:** Treat critical drift as a security incident.

**Workflow:**
1. Receive drift alert
2. Review drift report
3. Investigate cause (who, when, why)
4. Remediate or escalate
5. Document in incident log

**Rationale:** Critical drift indicates security or availability risk.

### 3. Document Intentional Drift

**Practice:** If drift is intentional, update Terraform and document why.

**Example:**

```bash
# Drift detected: worker_count changed from 3 to 5
# Decision: Accept drift (capacity increase approved)

cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform refresh
git add terraform.tfstate
git commit -m "Accept drift: increased worker count to 5 for capacity (approved by ops team)"
git push
```

**Rationale:** Maintains audit trail and prevents false alarms.

### 4. Use Terraform for All Changes

**Practice:** Never modify infrastructure manually. Always use Terraform.

**Workflow:**
1. Edit Terraform configuration
2. Run `terraform plan` to review changes
3. Run `terraform apply` to apply changes
4. Commit Terraform state to Git

**Rationale:** Prevents drift at the source.

### 5. Restrict Manual Access

**Practice:** Limit who can modify infrastructure directly.

**Implementation:**
- OpenStack: Read-only access for most users
- AWS: IAM policies restricting resource modification
- Audit logs: Monitor all manual changes

**Rationale:** Reduces drift risk.

## Troubleshooting

### Drift Detection Fails: State File Not Found

**Error:**
```
Error: failed to read terraform state: no such file or directory
```

**Cause:** Cluster has not been provisioned yet.

**Solution:**
```bash
# Bootstrap cluster first
opencenter cluster bootstrap prod-cluster

# Then run drift detection
opencenter cluster drift detect prod-cluster
```

### Drift Detection Fails: OpenStack API Error

**Error:**
```
Error: failed to get current infrastructure state: authentication failed
```

**Cause:** OpenStack credentials invalid or expired.

**Solution:**
```bash
# Verify credentials
openstack server list

# Update credentials in config
opencenter cluster edit prod-cluster

# Retry drift detection
opencenter cluster drift detect prod-cluster
```

### False Positive: Resource Reported as Drift

**Symptom:** Drift detected for resource that hasn't changed.

**Cause:** Terraform state out of sync with actual infrastructure.

**Solution:**
```bash
# Refresh Terraform state
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform refresh

# Retry drift detection
opencenter cluster drift detect prod-cluster
```

### Reconciliation Fails: Resource Already Exists

**Error:**
```
Error: resource already exists
```

**Cause:** Terraform state doesn't reflect actual infrastructure.

**Solution:**
```bash
# Import existing resource into Terraform state
cd ~/prod-cluster-gitops/infrastructure/clusters/prod-cluster
terraform import openstack_compute_instance_v2.master_1 <instance-id>

# Retry reconciliation
opencenter cluster drift reconcile prod-cluster
```

## Further Reading

- [Configuration Lifecycle](configuration-lifecycle.md) - How configuration flows through the system
- [GitOps Workflow](gitops-workflow.md) - Repository structure and reconciliation
- [OpenStack First Cluster](../tutorials/openstack-first-cluster.md) - Deploy on OpenStack
- [Troubleshoot Deployment](../how-to/troubleshoot-deployment.md) - Fix deployment issues

---

## Evidence

This explanation is based on:

- Drift detection implementation: `cmd/cluster_drift.go`, `internal/cloud/factory.go`
- OpenStack provider: `internal/cloud/openstack/provider.go`
- Terraform state integration: `.kiro/specs/todo-completion-phase-1/requirements.md:Epic 1`
- OpenStack provider enhancements: `.kiro/specs/todo-completion-phase-1/requirements.md:Epic 2`
- Configuration lifecycle: `docs/explanation/configuration-lifecycle.md`
- OpenStack tutorial: `docs/tutorials/openstack-first-cluster.md`

