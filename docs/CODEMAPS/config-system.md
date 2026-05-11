# Config System Codemap

**Last Updated:** 2026-05-11  
**Entry Point:** `internal/config/manager.go` → `ConfigurationManager`  
**Package:** `internal/config`

## Architecture

```
                         YAML File on Disk
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│  ConfigurationManager.Load(ctx, name)                       │
│    ├─ PathResolver.Resolve(name, org) → file path           │
│    ├─ ConfigCache (check hit)                               │
│    └─ ConfigIOHandler.LoadFromFile(ctx, path)               │
│         └─ v2.ConfigLoader.LoadFromFile(path)               │
│              ├─ Stage 1: parseYAML (yaml.v3)                │
│              ├─ Stage 2: normalize (canonicalize providers)  │
│              ├─ Stage 3: resolveReferences (${ref:},${env:})│
│              ├─ Stage 4: applyDefaults (Hydrator)           │
│              └─ Stage 5: validate                           │
│                    ├─ ValidateCleanBreakRules                │
│                    ├─ ValidateSchema (struct tags)           │
│                    ├─ ValidateBusinessRules (cross-field)    │
│                    ├─ ValidateProvider (provider-specific)   │
│                    ├─ ValidateDeployment (method compat)     │
│                    └─ ValidateServices (service deps)        │
│                              │                              │
│                              ▼                              │
│                     *v2.Config (validated)                   │
└─────────────────────────────────────────────────────────────┘
```

## Key Modules

| File/Package | Purpose | Key Exports |
|-------------|---------|-------------|
| `manager.go` | Top-level orchestrator | `ConfigurationManager` — Load, Save, Validate, List, Delete, GetActive, SetActive |
| `loader.go` | Adapter wrapping v2 loader | `ConfigIOHandler` — LoadFromFile, SaveToFile, MarshalConfig |
| `builder.go` | Fluent builder for programmatic construction | `FluentConfigBuilder` — WithProvider, WithOrganization, WithDefaults, Build |
| `resolver.go` | Cross-reference resolution via reflection | `referenceResolver` — resolves `${path.to.field}` with cycle detection |
| `defaults.go` | Default config factory | `defaultConfig(name)`, `ApplyDefaults()`, `applyCLIDefaults()` |
| `paths.go` | Type-safe config path constants | `TypedConfigPath[T]` — compile-time safe path references |
| `interfaces.go` | Core interfaces | `ConfigLoader`, `ConfigValidator`, `ConfigValidationResult` |
| `config.go` | v1 Config methods | Credential accessors with fallback logic |
| `errors.go` | Typed errors | `ConfigNotFoundError`, `ValidationError`, `SchemaError` |
| `cli_config.go` | CLI settings management | `CLIConfig` — user preferences, cluster defaults, plugin checksums |

## Subpackages

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `v2/` | Authoritative v2 config pipeline | `Config`, `ConfigLoader`, `Validator`, `InfrastructureConfig` |
| `v2schema/` | JSON Schema generator for IDE support | `Generate(opts) ([]byte, error)` |
| `services/` | Typed service configs + validation | `BaseConfig`, per-service configs, `DependencyValidator` |
| `defaults/` | Provider-region defaults hydration | `Registry`, `Hydrator`, `ProviderDefaults` interface |
| `validation/` | Shared validation utilities | `IsValidUUID`, `IsValidURL`, `SubnetsOverlap` |
| `overlay/` | GitOps overlay customization types | `UnitsConfig`, `SOPSGenerationConfig`, `CustomerManagedConfig` |
| `flags/` | CLI flag parsing + path-based mutation | Reflection engine for `cluster set` |
| `cache/` | In-memory config caching | Named cache utilities |
| `persistence/` | YAML serialization helpers | Path resolution for on-disk storage |
| `registry/` | Config type registry | Type lookups |

## Type Hierarchy

```
Config (v1 — internal/config/types.go)
├── SchemaVersion string
├── Metadata      ConfigMetadata
├── OpenCenter    OpenCenter
│   ├── Cluster        ClusterMeta
│   ├── Infrastructure Infrastructure
│   │   ├── Provider   string
│   │   └── Cloud      CloudConfig (OpenStack/AWS/VMware/Kind)
│   ├── Kubernetes     KubernetesConfig
│   ├── GitOps         GitOpsConfig
│   ├── Services       ServiceMap
│   ├── Storage        StorageConfig
│   └── Identity       IdentityConfig
├── OpenTofu      SimplifiedOpenTofu
├── Secrets       Secrets
├── Deployment    Deployment
└── Overrides     map[string]any

Config (v2 — internal/config/v2/config.go)
├── Meta           MetaConfig
├── OpenCenter     OpenCenterConfig
├── Infrastructure InfrastructureConfig
├── Kubernetes     KubernetesConfig
├── GitOps         GitOpsConfig
├── Services       ServiceMap
├── Secrets        SecretsConfig
├── OpenTofu       OpenTofuConfig
└── Deployment     DeploymentConfig
```

## Services Registry (`services/`)

Platform services with typed configs:

| Service | Config Type | Key Fields |
|---------|------------|------------|
| Calico | `CalicoConfig` | Mode, VXLAN, BGP, eBPF |
| cert-manager | `CertManagerConfig` | Issuers, DNS providers |
| Keycloak | `KeycloakConfig` | Realm, clients, OIDC |
| Loki | `LokiConfig` | Storage backend, retention |
| Prometheus | `PrometheusStackConfig` | Retention, alerting |
| Harbor | `HarborConfig` | Storage, TLS |
| Velero | `VeleroConfig` | Backup provider, schedule |
| MetalLB | `MetalLBConfig` | Address pools |
| Tempo | `TempoConfig` | Storage, sampling |
| OpenTelemetry | `OpenTelemetryConfig` | Collectors, exporters |
| vSphere CSI | `VSphereCSIConfig` | Datastore, storage policy |
| Longhorn | `LonghornConfig` | Replicas, storage class |
| Gateway | `GatewayConfig` | Routes, TLS |
| Headlamp | `HeadlampConfig` | OIDC, plugins |
| etcd-backup | `EtcdBackupConfig` | Schedule, retention |
| AlertProxy | `AlertProxyConfig` | Endpoints, routing |

All embed `BaseConfig` (Enabled, Namespace, Source, Image, AdoptionMode).

## Data Flow

1. **CLI** calls `ConfigurationManager.Load(ctx, "org/cluster")`
2. **PathResolver** maps name → `~/.config/opencenter/clusters/org/.cluster-config.yaml`
3. **ConfigCache** checks for cached version (invalidated on file mtime change)
4. **ConfigIOHandler** delegates to `v2.ConfigLoader` 5-stage pipeline
5. **Hydrator** fills empty fields from provider-region defaults
6. **Validator** runs schema + business + provider + deployment + services checks
7. **Validated Config** returned and cached

## Related Areas

- [CLI Commands](cli-commands.md) — `settings` command manages CLI config
- [Cluster Lifecycle](cluster-lifecycle.md) — `InitService` uses builder to create configs
- [GitOps Engine](gitops-engine.md) — reads validated config for template rendering
