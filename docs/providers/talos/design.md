# Talos on OpenStack — Design Document

## Purpose
This document translates the Talos provider research into an implementable product strategy for the openCenter CLI. It describes **what** experience we are building for platform and security engineers who must provision Zero Trust Talos clusters on OpenStack. The document informs backlog prioritization, UX flows, and verification criteria that downstream AI agents will execute.

## Audience
- openCenter CLI contributors
- Security, platform, and network engineers partnering on the Talos provider
- AI agents that will implement the spec in code, infrastructure templates, docs, and tests

## Problem Statement
Highly regulated organizations want Kubernetes clusters that inherit Talos immutability while running on multi-tenant OpenStack clouds. Today, provisioning demands bespoke scripts, weak secrets handling, and inconsistent network policies. The Talos provider must give teams a single declarative workflow that:
1. Treats OpenStack, Talos, and the cluster lifecycle as one secure system.
2. Enforces cryptographic verification and Zero Trust controls out of the box.
3. Integrates with GitOps pipelines without exposing long-lived credentials.

## Goals
1. **Secure-by-default experience**: every generated artifact (machine config, network policy, load balancer, disk) adheres to the hardening guidance from `readme.md` without manual toggles.
2. **Declarative lifecycle**: cluster creation, upgrades, and replacement run through first-party Pulumi programs surfaced via openCenter CLI commands (no CAPI/CAPO dependency in this iteration).
3. **Auditable management**: no SSH; all mutable actions happen through Talos API guarded by WireGuard and short-lived certificates.
4. **Composable automation**: output assets and state are automation-friendly (GitOps repos, Pulumi stacks, schemas).

## Non-Goals
- Supporting non-Talos operating systems or mutable management patterns.
- Providing a generic OpenStack orchestrator unrelated to Kubernetes.
- Shipping CAPI/CAPO integrations in this iteration (future enhancement).
- Implementing enterprise UI/portal features inside openCenter CLI.

## Product Personas & Key Jobs
| Persona | Key Job | Success Signal |
| --- | --- | --- |
| Platform Engineer | Stand up production-ready Talos clusters quickly | CLI scaffolds infrastructure and configs in <30 minutes without exemptions |
| Security Engineer | Enforce Zero Trust controls | Cryptographic attestation, MFA, and secrets encryption enabled automatically |
| Network Architect | Maintain deterministic network boundaries | Three-zone Neutron topology + policy manifests generated on every run |
| Compliance Officer | Prove supply-chain integrity | Signed images verified, audit logs centralized, drift detection reports |

## Design Principles
1. **Radical Immutability** — treat every change as a replacement; expose plans that show which nodes rotate and why.
2. **Cryptographic Attestation** — surface the verification chain (Glance signatures, Barbican certs, Talos TLS bundle) inside CLI output.
3. **Non-Interactive Management** — no escape hatches. All operational tasks route through Talos API and documented runbooks.
4. **Zero Trust Networking** — default networks, security groups, and load balancers assume hostile neighbors and require explicit opt-in to relax.
5. **Defense-in-Depth Automation** — validations run before apply to short-circuit unsafe environments (e.g., Barbican unavailable).

## Solution Overview
The experience is organized into four pillars that map to CLI commands and automation hooks:

1. **Environment Assessment** (e.g., `opencenter talos validate`)
   - Checks OpenStack services (Keystone MFA, Barbican, Octavia, Glance signatures) and tenant quotas.
   - Generates a detailed readiness report consumed by CI and humans.
2. **Blueprint Authoring** (e.g., `opencenter talos init`)
   - Produces GitOps manifests containing Talos machine configs, Pulumi stack configuration, WireGuard assets, and secrets policy (SOPS + Barbican IDs).
   - Encodes network topology (management, control, data plane subnets) and security groups.
3. **Provisioning & Bootstrap** (e.g., `opencenter talos apply`)
   - Executes the embedded Go-based Pulumi program (inspired by `third-party/openstack-flex-examples/pulumi/talos-cluster-ts` and `third-party/liberty-infrastructure`) to create/update OpenStack projects, networks, load balancers, Talos nodes, and vTPM-backed disks.
   - Bootstraps Talos nodes over the WireGuard bastion and orchestrates etcd initialization + KubePrism setup.
4. **Operations & Compliance** (e.g., `opencenter talos status`, `opencenter talos rotate`)
   - Surfaces runbooks for node replacement, image rotation, secrets revocation, and drift remediation leveraging Pulumi updates/refresh.
   - Streams audit evidence (Talos + Kubernetes logs) into external sinks defined in the blueprint.

## Experience Narrative
1. Engineer runs a validation command; CLI reports Barbican availability, Glance signature enforcement, and required Neutron networks.
2. Engineer initializes a new Talos blueprint specifying tenant/project IDs and desired cluster size. CLI outputs GitOps-ready directories with encrypted credentials, Pulumi stack configuration, and documentation stubs.
3. CI pipeline applies the blueprint. openCenter invokes Pulumi via Go bindings to create the dedicated OpenStack project, enforce the network/security posture, and provision Talos nodes that expose only required ports. The CLI prints attestation fingerprints.
4. Operations team monitors status through CLI, which fetches Talos API health via WireGuard and compares deployed resources against the blueprint for drift. Rotations trigger node replacements automatically.

## Scope of Work
- Command set for Talos on OpenStack (validate, init, apply, status, rotate, destroy).
- Pulumi stack templates encapsulating Neutron networks, security groups, LBaaS, Barbican key policies, Glance image metadata, Talos machine configs, and vTPM-backed disks.
- Reference GitOps repository layout and schema definitions for pipeline integration.
- Documentation + runbooks aligned with the architecture and PRD.

## Acceptance Considerations
- CLI commands emit machine-parseable JSON alongside human summaries.
- Secrets are never written unencrypted to disk; SOPS + `.sops.yaml` patterns enforced.
- Validation fails fast if prerequisites (Barbican, Octavia, signatures, MFA) are missing, with actionable remediation hints.
- Node replacement workflows include drain + cordon automation and watchers for workload disruption.

## Risks & Open Questions
- **Tenant Constraints**: Some OpenStack environments may not allow application credentials with the required scope. Mitigation: support user-supplied tokens + document limitations.
- **Octavia Availability**: Some regions might lack Octavia; fallback to floating IP + HAProxy needs confirmation.
- **Audit Log Shipping**: Requires selecting at least one supported backend (e.g., Loki, Splunk). Need to prioritize initial target.
- **vTPM Support**: Not all hypervisors expose vTPM. Determine alternative key wrapping for disk encryption when missing.

## References
- `architecture.md` — detailed component relationships and control flows.
- `requirements.md` — prioritized stories, functional and non-functional requirements.
