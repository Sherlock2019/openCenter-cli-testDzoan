# Cluster Lifecycle Codemap

**Last Updated:** 2026-05-11  
**Entry Point:** `internal/cluster/` service types  
**Package:** `internal/cluster`

## Lifecycle Flow

```
Init → Configure (optional) → Validate → Setup (Generate) → Bootstrap (Deploy) → [Destroy]
```

```
┌──────────┐    ┌─────────────┐    ┌──────────────┐    ┌───────────┐    ┌───────────────┐
│   Init   │───▶│  Configure  │───▶│   Validate   │───▶│   Setup   │───▶│   Bootstrap   │
│          │    │ (optional)  │    │              │    │ (generate)│    │   (deploy)    │
└──────────┘    └─────────────┘    └──────────────┘    └───────────┘    └───────────────┘
     │                                                                         │
     │                                                                         ▼
     │                                                                  ┌─────────────┐
     └──────────────────────────────────────────────────────────────────│   Destroy   │
                                                                        └─────────────┘
```

## Services

| Service | File | CLI Command | Purpose |
|---------|------|-------------|---------|
| `InitService` | `init_service.go` | `cluster init` | Create config YAML, generate keys, create dirs |
| `ConfigureService` | `configure_service.go` | `cluster configure` | Interactive guided config with provider discovery |
| `ValidateService` | `validate_service.go` | `cluster validate` | Schema + business + connectivity validation |
| `SetupService` | `setup_service.go` | `cluster generate` | Generate GitOps repo via pipeline |
| `BootstrapService` | `bootstrap_service.go` | `cluster deploy` | Provision infrastructure + deploy cluster |
| `DestroyProvider` | `destroy_provider.go` | `cluster destroy` | Provider-specific teardown |

## Init Service

**Input:** `InitOptions` (name, org, provider, key gen flags)  
**Output:** `InitResult` (config path, keys generated)

Steps:
1. Resolve paths via `PathResolver`
2. Create directory structure (org dir, secrets dir, state dir)
3. Generate SSH key pair (`util/crypto`)
4. Generate SOPS Age key pair (`sops.KeyManager`)
5. Create config YAML via `FluentConfigBuilder`
6. Initialize git repo with pre-commit hooks
7. Write `.sops.yaml` configuration

## Configure Service

**Input:** `ConfigureOptions` (identifier, org, provider)  
**Output:** `ConfigureResult` (created, config, paths)

Steps:
1. Discover provider resources (OpenStack: catalog of images, flavors, networks, AZs)
2. Run guided prompts via `orchestration.PromptRunner`
3. Accumulate config patches in `changeReview`
4. Apply patches to config YAML
5. Write managed files (SSH keys, credentials)

## Validate Service

**Input:** `ValidateOptions` (mode: online/offline, provider checks)  
**Output:** `ValidationResult` (valid, errors, warnings, issues)

Steps:
1. Load config via `v2.ConfigLoader` (validates schema on load)
2. Run `ValidateReadiness` business rules
3. If online mode: connectivity checks (OpenStack auth URL, API endpoints)
4. If provider checks: validate catalog (images exist, flavors available, networks reachable)
5. Aggregate results with severity levels

## Setup Service

**Input:** `SetupOptions` (dry run, skip validation)  
**Output:** `SetupResult` (gitops path, manifest count)

Steps:
1. Load and validate config
2. Invoke `gitops.PipelineGenerator.Generate()` (see [GitOps Engine](gitops-engine.md))
3. Run `ManifestValidator` on generated output
4. Run `ScanGitOpsSecrets` security scan
5. Encrypt overlay files via `sops.SOPSManager`

## Bootstrap Service

**Input:** `BootstrapOptions` (timeout, dry run, restart, step filter)  
**Output:** `BootstrapResult` (infra provisioned, cluster deployed, endpoint, duration)

Steps:
1. Build provider-specific step list:
   - **OpenStack**: terraform init/plan/apply → kubespray → wait ready
   - **Kind**: kind create cluster → wait ready
   - **VMware**: terraform → kubespray → wait ready
2. Load/create bootstrap state (JSON file for resume)
3. Execute steps sequentially with state persistence
4. Wait for cluster readiness via kubectl
5. Install FluxCD if configured

**Resume support:** Each step's state is persisted to `bootstrap-state.json`. On `--restart`, resumes from last incomplete step.

## Destroy

**Interface:** `lifecycleDestroyProvider`

Provider-specific implementations:
- **OpenStack**: `openstack_destroy_provider.go` — terraform destroy
- **Kind**: delegates to `cloud/kind.DeleteCluster()`

## Supporting Packages

| Package | Role in Lifecycle |
|---------|------------------|
| `internal/ansible` | Generates Kubespray inventory from config |
| `internal/tofu` | Executes OpenTofu/Terraform commands |
| `internal/provision` | Embedded provisioning templates |
| `internal/cloud/kind` | Kind cluster create/delete/wait |
| `internal/resilience` | Distributed locks for deploy/destroy |
| `internal/importer` | Scans live clusters for import |
| `internal/cluster/orchestration` | Provider registry, capability registry, prompt runner |

## Key Types

```go
type InitOptions struct {
    Name, Org, Provider string
    NoKeygen, NoSopsKeygen, RegenerateKeys, Force bool
    ServerPools []string
}

type BootstrapOptions struct {
    Timeout time.Duration
    DryRun, Restart, ConfirmCommit, BreakLock bool
    Step, FromStep string  // step filtering for partial runs
}

type ValidationResult struct {
    Valid    bool
    Errors   []ValidationIssue
    Warnings []ValidationIssue
    Issues   []ValidationIssue
}
```

## Dependencies

- `internal/config` — ConfigurationManager, ConfigIOHandler
- `internal/config/v2` — Config, ConfigLoader, ValidateReadiness
- `internal/config/defaults` — Registry for default values
- `internal/config/services` — Provider registry
- `internal/core/paths` — PathResolver, ClusterPaths
- `internal/core/validation` — ValidationEngine
- `internal/gitops` — PipelineGenerator, ManifestValidator
- `internal/sops` — KeyManager, SOPSManager
- `internal/security` — CommandRunner
- `internal/util/crypto` — SSH/Age key generation
- `internal/cloud/openstack` — DiscoveryClient for configure flow

## Related Areas

- [Config System](config-system.md) — config loading and validation
- [GitOps Engine](gitops-engine.md) — invoked by SetupService
- [Secrets](secrets-management.md) — key generation during init
- [Providers](providers.md) — provider-specific bootstrap/destroy logic
