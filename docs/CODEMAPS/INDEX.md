# openCenter CLI — Codemaps Index

**Last Updated:** 2026-05-11  
**Module:** `github.com/opencenter-cloud/opencenter-cli`  
**Language:** Go 1.23+  
**Entry Point:** `main.go` → `cmd.ExecuteWithContext()`

## Architecture Overview

```
main.go
  │
  ├─ config.ResolveClustersDir()     → base directory
  ├─ di.SetupContainer(baseDir)      → DI container
  └─ cmd.ExecuteWithContext(ctx)     → Cobra root command
       │
       ├─ cluster   → internal/cluster (lifecycle services)
       ├─ secrets   → internal/secrets + internal/sops
       ├─ settings  → internal/config (CLI config)
       ├─ plugins   → internal/plugins (external CLI plugins)
       ├─ version   → build info (ldflags)
       └─ shell-init → shell integration scripts
```

## Codemaps

| Codemap | Scope | Key Packages |
|---------|-------|--------------|
| [CLI Commands](cli-commands.md) | Command tree, flags, registration | `cmd/` |
| [Config System](config-system.md) | Loading, validation, schema, types | `internal/config/`, `internal/config/v2/` |
| [GitOps Engine](gitops-engine.md) | Generation pipeline, templates, rendering | `internal/gitops/` |
| [Cluster Lifecycle](cluster-lifecycle.md) | Init, validate, setup, bootstrap, destroy | `internal/cluster/` |
| [Secrets Management](secrets-management.md) | Rotation, registry, sync, hooks, SOPS | `internal/secrets/`, `internal/sops/` |
| [Providers](providers.md) | Cloud provider abstraction, drift detection | `internal/cloud/` |
| [DI Container](di-container.md) | Dependency injection, service wiring | `internal/di/` |

## Package Map (all `internal/`)

| Package | Purpose | Codemap |
|---------|---------|---------|
| `ansible` | Kubespray inventory generation | [Cluster Lifecycle](cluster-lifecycle.md) |
| `barbican` | OpenStack Key Manager client | [Secrets](secrets-management.md) |
| `benchmarks` | Performance benchmarks | — (internal tooling) |
| `cloud` | Provider abstraction + drift detection | [Providers](providers.md) |
| `cluster` | Lifecycle domain services | [Cluster Lifecycle](cluster-lifecycle.md) |
| `config` | Configuration management | [Config System](config-system.md) |
| `core` | Shared: path resolution, validation engine | [Config System](config-system.md) |
| `credentials` | Cloud credential extraction | [Providers](providers.md) |
| `di` | Dependency injection | [DI Container](di-container.md) |
| `gitops` | GitOps repo generation | [GitOps Engine](gitops-engine.md) |
| `importer` | Live cluster import/scan | [Cluster Lifecycle](cluster-lifecycle.md) |
| `localdev` | Local dev environment (Kind, Gitea, Flux) | [Providers](providers.md) |
| `observability` | Structured logging, credential masking | [DI Container](di-container.md) |
| `operations` | Drift detection, backup, disaster recovery | [Providers](providers.md) |
| `plugins` | External CLI plugin discovery | [CLI Commands](cli-commands.md) |
| `provision` | Embedded provisioning templates | [Cluster Lifecycle](cluster-lifecycle.md) |
| `resilience` | Retry, circuit breaker, distributed locks | [Cluster Lifecycle](cluster-lifecycle.md) |
| `secrets` | Multi-cluster secrets lifecycle | [Secrets](secrets-management.md) |
| `security` | Audit logging, input validation, command sanitization | [DI Container](di-container.md) |
| `services` | Platform service plugin registry | [Config System](config-system.md) |
| `sops` | SOPS encryption/decryption, Age key management | [Secrets](secrets-management.md) |
| `template` | Template engine with caching and sandboxing | [GitOps Engine](gitops-engine.md) |
| `testenv` | Test environment helpers | — (internal tooling) |
| `testing` | Unified test utilities | — (internal tooling) |
| `tofu` | OpenTofu/Terraform execution | [Cluster Lifecycle](cluster-lifecycle.md) |
| `ui` | Prompts, error formatting, guided flows | [CLI Commands](cli-commands.md) |
| `util` | Files, errors, crypto, security, metrics | — (shared utilities) |

## Cross-Cutting Concerns

- **Security**: `internal/security` provides audit logging, credential masking, input validation, and command sanitization used across all packages.
- **File I/O**: `internal/util/fs.FileSystem` interface abstracts all disk operations for testability.
- **Path Resolution**: `internal/core/paths.PathResolver` provides consistent cluster path resolution.
- **Validation**: `internal/core/validation.ValidationEngine` is the shared validation framework with pluggable validators.
