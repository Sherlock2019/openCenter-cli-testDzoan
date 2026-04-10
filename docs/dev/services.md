---
id: service-lifecycle
title: "Service Enable/Disable Lifecycle"
sidebar_label: Service Lifecycle
description: How services are enabled, disabled, rendered, and reconciled across cluster stages.
doc_type: explanation
audience: "developers, platform engineers"
tags: [services, lifecycle, fluxcd, gitops, rendering]
---

# Service Enable/Disable Lifecycle

**Purpose:** For developers and platform engineers, explains how service enable/disable flows through configuration, GitOps rendering, and FluxCD reconciliation, covering all cluster stages and edge cases.

## Concepts

A "service" in openCenter is a Kubernetes workload (Helm chart, Kustomization, or raw manifests) managed through the cluster configuration. Each service has an `enabled: bool` field in its config struct, inherited from `services.BaseConfig`. The enabled flag drives two things:

1. Whether the service's FluxCD manifests are rendered into the GitOps repository.
2. Whether FluxCD deploys (or prunes) the service in the target cluster.

Services live in two maps inside `SimplifiedOpenCenter`:

- `Services` (`ServiceMap`) — platform services (cert-manager, kyverno, loki, etc.)
- `ManagedService` (`ServiceMap`) — customer/application services

Both maps use polymorphic YAML unmarshaling via the service registry (`internal/config/registry`). Values are typed struct pointers (e.g., `*services.CertManagerConfig`), not raw maps.

## Enable/Disable Command Flow

```
┌──────────────────────────────────────────────────────────┐
│  opencenter cluster service enable <name> [--render]     │
│  opencenter cluster service disable <name>               │
└──────────────────────┬───────────────────────────────────┘
                       │
                       ▼
              ┌────────────────┐
              │  Load config   │
              │  (YAML file)   │
              └───────┬────────┘
                      │
                      ▼
              ┌────────────────┐
              │  Registry      │  registry.GetServiceConfigType(name)
              │  lookup        │  → typed struct (e.g. CertManagerConfig)
              └───────┬────────┘
                      │
           ┌──────────┴──────────┐
           │                     │
           ▼                     ▼
   ┌──────────────┐     ┌──────────────┐
   │  enable:     │     │  disable:    │
   │  Enabled=true│     │  Enabled=false│
   │  + params    │     │              │
   │  + secrets   │     │              │
   │  + validate  │     │              │
   └──────┬───────┘     └──────┬───────┘
          │                    │
          ▼                    ▼
   ┌─────────────────────────────────┐
   │  Save config to disk           │
   │  (~/.config/opencenter/...)    │
   └──────────────┬──────────────────┘
                  │
                  ▼
          ┌───────────────┐
          │  --render?    │──── no ──→ done (config-only change)
          └───────┬───────┘
                  │ yes
                  ▼
          ┌───────────────────────┐
          │  RenderSingleService  │  (internal/gitops/copy.go)
          │  → writes manifests   │
          │    to GitOps repo     │
          └───────────────────────┘
```

Key implementation details:

- `setEnabled()` uses reflection to set `BaseConfig.Enabled` on any service struct (`cmd/cluster_service.go:673`).
- Enable validates service-specific requirements (e.g., cert-manager requires an email parameter) before saving.
- Disable does not render or clean up manifests. It only flips the flag.

## Rendering: Config → GitOps Manifests

Rendering translates the enabled/disabled state into files on disk. Two paths trigger rendering:

| Trigger | Scope | Function |
|---|---|---|
| `cluster setup` | Full cluster | `gitops.RenderClusterApps()` |
| `cluster service enable --render` | Single service | `gitops.RenderSingleService()` |

### Full Render (`cluster setup`)

```
┌─────────────────────────────────────────────────────────────┐
│  cluster setup                                              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ▼
              ┌────────────────────┐
              │  Load config       │
              │  Validate          │
              └────────┬───────────┘
                       │
                       ▼
              ┌────────────────────┐
              │  CopyBase()        │  base GitOps structure
              └────────┬───────────┘
                       │
                       ▼
              ┌────────────────────┐
              │  RenderClusterApps │
              └────────┬───────────┘
                       │
                       ▼
         ┌─────────────────────────────┐
         │  planClusterAppActions()    │
         │                             │
         │  For each descriptor:       │
         │    isDescriptorEnabled()?   │
         │      ├─ service field set?  │
         │      │   → check Services   │
         │      │     map + Enabled    │
         │      ├─ managed_service?    │
         │      │   → check Managed    │
         │      │     Service map      │
         │      └─ condition?          │
         │          → evaluate         │
         │            enabled_when     │
         │                             │
         │  Skip disabled descriptors  │
         │  Expand enabled → actions   │
         └─────────────┬───────────────┘
                       │
                       ▼
         ┌─────────────────────────────┐
         │  cleanupRendererOwnedOverlay│
         │  Remove stale service dirs  │
         │  Write new manifests        │
         │  (atomic workspace)         │
         └─────────────┬───────────────┘
                       │
                       ▼
         ┌─────────────────────────────┐
         │  RenderInfrastructureCluster│
         │  Validate manifests         │
         │  Git commit                 │
         └─────────────────────────────┘
```

The descriptor registry (`internal/gitops/descriptorcfg`) maps each service to its template roots, output paths, and conditional rendering rules. `isDescriptorEnabled()` (`internal/gitops/descriptor_renderer.go:169`) checks the `Services` or `ManagedService` map and calls `IsServiceDisabled()` to inspect the `Enabled` field via reflection.

### Cleanup of Disabled Services

`cleanupDisabledServices()` (`internal/gitops/copy.go`) walks the overlay directories and removes any service directory whose config is missing or has `Enabled: false`. This ensures disabled services don't leave stale manifests in the GitOps repo.

```
applications/overlays/<cluster>/
├── services/
│   ├── sources/           ← GitRepository YAMLs (per-service)
│   ├── fluxcd/            ← Kustomization YAMLs (per-service)
│   ├── cert-manager/      ← enabled: present
│   ├── kyverno/           ← enabled: present
│   └── loki/              ← disabled: REMOVED by cleanup
└── managed-services/
    └── my-app/            ← disabled: REMOVED by cleanup
```

## Cluster Stage Paths

The effect of enable/disable depends on where the cluster is in its lifecycle.

### Path A: Before Setup (No GitOps Repo)

```
enable/disable
     │
     ▼
Config updated ──→ done
```

No GitOps repository exists yet. The config change is stored but has no immediate effect. The next `cluster setup` will generate manifests reflecting the current enabled state.

When to use: initial cluster configuration, before any `cluster setup` has run.

### Path B: After Setup, Before Bootstrap (GitOps Repo Exists, No Running Cluster)

```
enable/disable
     │
     ▼
Config updated
     │
     ▼
cluster setup (or enable --render)
     │
     ▼
Manifests regenerated
     │
     ▼
Git commit ──→ ready for bootstrap
```

The GitOps repo exists on disk but FluxCD is not running. Manifests need to be regenerated so the repo is consistent before bootstrap. Two options:

1. `cluster service enable <name> --render` — renders only the changed service.
2. `cluster setup --force` — full re-render of all manifests.

For disable, there is no `--render` flag on the disable command today. Run `cluster setup --force` to clean up the disabled service's manifests.

### Path C: After Bootstrap (Running Cluster with FluxCD)

```
enable/disable
     │
     ▼
Config updated
     │
     ▼
Render manifests
(setup --force or enable --render)
     │
     ▼
Git commit + push
     │
     ▼
FluxCD detects change
(source-controller polls, default 15m)
     │
     ▼
┌─────────────────────────────────────┐
│  Kustomization reconciliation       │
│                                     │
│  Enable: new manifests applied      │
│    → HelmRelease created            │
│    → pods scheduled                 │
│                                     │
│  Disable: manifests removed         │
│    → prune: true triggers deletion  │
│    → HelmRelease removed            │
│    → pods terminated                │
└─────────────────────────────────────┘
```

FluxCD Kustomizations are generated with `prune: true`, which means removed manifests cause FluxCD to delete the corresponding cluster resources. This is how disabling a service leads to actual teardown.

Verification after the reconciliation cycle:

```bash
# Check FluxCD reconciliation status
flux get kustomizations -n flux-system

# Check if service HelmRelease was created/removed
flux get helmreleases --all-namespaces

# Force immediate reconciliation (skip the 15m wait)
flux reconcile source git <source-name> -n flux-system
```

### Path D: Service with Dependencies (Enable)

Some services depend on others. For example, `weave-gitops` depends on `fluxcd`, and `headlamp` depends on `keycloak`. The dependency validator (`internal/config/services/dependency_validator.go`) checks these at validation time.

```
enable service-with-dependency
     │
     ▼
Validate dependencies
     │
     ├── dependency enabled? ──→ proceed
     │
     └── dependency missing/disabled?
              │
              ▼
         Error: "service X requires Y to be enabled"
```

Dependencies are also expressed in FluxCD Kustomizations via `dependsOn` blocks. Even if validation is skipped, FluxCD will not reconcile a Kustomization until its dependencies are healthy.

### Path E: Service with Persistent Data (Disable)

Disabling a service with persistent data (PVCs, CRDs, finalizers) requires care. FluxCD's `prune: true` removes the HelmRelease, but:

- PVCs with `Retain` reclaim policy survive pod deletion.
- CRDs are not removed by default (Helm does not delete CRDs on uninstall).
- Finalizers on custom resources can block namespace deletion.

```
disable service-with-state
     │
     ▼
Config updated + render + push
     │
     ▼
FluxCD removes HelmRelease
     │
     ▼
Helm uninstall runs
     │
     ├── PVCs: retained (reclaimPolicy: Retain)
     ├── CRDs: retained (Helm convention)
     └── Finalizers: may block if controller is gone
```

Operator action may be needed to clean up PVCs, CRDs, or stuck finalizers after disabling stateful services like loki, tempo, or velero.

### Path F: Re-enable a Previously Disabled Service

```
enable <name> --force --render
     │
     ▼
Config updated (Enabled=true, fresh defaults)
     │
     ▼
Manifests regenerated
     │
     ▼
Git commit + push
     │
     ▼
FluxCD reconciles → service redeployed
```

The `--force` flag is required because the service entry still exists in the config (with `enabled: false`). Without `--force`, the enable command errors with "service is already enabled" (a misleading message when the service is disabled but present — this is a known quirk).

If PVCs from a previous deployment still exist, the re-enabled service may reattach to them, depending on the Helm chart's `existingClaim` configuration.

### Path G: Drift Detection

`cluster drift detect` compares desired state (from config) against actual infrastructure. If a service is disabled in config but its resources still exist in the cluster (e.g., FluxCD hasn't reconciled yet, or pruning failed), drift detection flags it.

```
cluster drift detect
     │
     ▼
Build desired state from config
     │
     ▼
Query actual cluster state
     │
     ▼
Compare → report drift items with severity
```

## Implementation References

| Component | File | Key function |
|---|---|---|
| Enable/disable commands | `cmd/cluster_service.go` | `newClusterServiceEnableCmd`, `newClusterServiceDisableCmd` |
| Reflection helpers | `cmd/cluster_service.go:658` | `isEnabled`, `setEnabled`, `getStatus` |
| Service registry | `internal/config/registry/registry.go` | `GetServiceConfigType`, `GetRegisteredServices` |
| Base config | `internal/config/services/base.go` | `BaseConfig.IsEnabled`, `BaseConfig.GetStatus` |
| Descriptor rendering | `internal/gitops/descriptor_renderer.go` | `planClusterAppActions`, `isDescriptorEnabled` |
| Single service render | `internal/gitops/copy.go` | `RenderSingleService` |
| Full render | `internal/gitops/copy.go` | `RenderClusterApps`, `cleanupDisabledServices` |
| Disabled check | `internal/gitops/copy.go:482` | `IsServiceDisabled` |
| Dependency validation | `internal/config/services/dependency_validator.go` | `ValidateDependencies` |
| Setup orchestration | `internal/cluster/setup_service.go` | `generateGitOpsManifests` |
| Bootstrap (Kind) | `internal/cluster/kind_bootstrap_provider.go` | `BuildSteps` |
| Bootstrap (OpenStack) | `internal/cluster/openstack_bootstrap_provider.go` | `BuildSteps` |

## Known Gaps

1. `cluster service disable` has no `--render` flag. Disabling a service requires a separate `cluster setup --force` to update manifests. Adding `--render` to the disable command would make the workflow symmetric with enable.
2. No automated cleanup of PVCs or CRDs when a stateful service is disabled. Operators must handle this manually.
3. Drift detection does not yet cover service-level drift (e.g., "service disabled in config but HelmRelease still running"). It focuses on infrastructure resources.
