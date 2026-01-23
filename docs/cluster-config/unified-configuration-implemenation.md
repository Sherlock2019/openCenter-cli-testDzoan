# OpenCenter Unified Configuration Implementation Guide

This implementation guide provides concrete steps, code structures, and logic needed to implement the unified configuration system. It incorporates critical "Day 2" improvements (Hydration, Validation, and Determinism) and reflects the current v2.0 schema structure.

## Table of Contents

- [Configuration Structure Overview](#configuration-structure-overview)
- [Implementation Roadmap](#implementation-roadmap)
- [Phase 1: The Safety Net (Validation & Schema)](#phase-1-the-safety-net-validation--schema)
- [Phase 2: The Core Logic (Registry & Hydration)](#phase-2-the-core-logic-registry--hydration)
- [Phase 3: The Refactor (Struct Reorganization)](#phase-3-the-refactor-struct-reorganization)
- [Phase 4: The Intelligence (Reference Resolution)](#phase-4-the-intelligence-reference-resolution)
- [Phase 5: The Bridge (Migration Tooling)](#phase-5-the-bridge-migration-tooling)
- [Implementation Checklist](#implementation-checklist)

---

## Configuration Structure Overview

The unified configuration follows this hierarchy:

```
Config (Root)
├── schema_version: "2.0"
├── metadata
├── opencenter
│   ├── meta              # Identity & Ownership
│   ├── secrets           # Secrets backend (Barbican)
│   ├── infrastructure    # Provider, Compute, Storage, Networking
│   ├── cluster           # Kubernetes configuration
│   ├── gitops            # GitOps repository settings
│   ├── services          # Self-hosted platform services
│   └── managed_services  # External/vendor-managed services
├── deployment            # Deployment method & automation
├── opentofu              # IaC backend configuration
└── secrets               # Credentials & keys
```

**Key Architectural Decisions:**
- **Deployment at root level**: Consolidates method configuration and auto-deploy settings
- **Infrastructure separation**: Compute, storage, and networking under `infrastructure`
- **Service polymorphism**: Services configured via ServiceMap with type registry
- **Provider abstraction**: Cloud-specific config under `infrastructure.cloud.<provider>`

**Supported Cloud Providers:**
- **OpenStack**: Fully implemented and production-ready
- **AWS, GCP, Azure, VMware**: Included in schema for architectural completeness and future extensibility, but not currently scheduled or planned for implementation. These serve as reference implementations for the provider abstraction pattern.

## Implementation Roadmap

* **Phase 1: The Safety Net (Validation & Schema)**
* **Phase 2: The Core Logic (Registry & Hydration)**
* **Phase 3: The Refactor (Struct Reorganization)**
* **Phase 4: The Intelligence (Reference Resolution)**
* **Phase 5: The Bridge (Migration Tooling)**

---

## Phase 1: The Safety Net (Validation & Schema)

Before moving any fields, establish a rigorous validation framework. This prevents invalid configurations from ever reaching the deployment logic.

### Action 1: Add Validation Tags

Update your Go structs to use a validation library (like `go-playground/validator`) and standard JSON tags.

**`internal/config/types_infrastructure.go`**

```go
type InfrastructureNetworking struct {
    // CIDR validation ensures valid network notation
    SubnetNodes string `yaml:"subnet_nodes" validate:"required,cidrv4" json:"subnet_nodes"`
    
    // Cross-field validation (e.g., Start IP must be within SubnetNodes) 
    // is handled in custom Validate() methods, but simple constraints go here.
    AllocationPoolStart string `yaml:"allocation_pool_start" validate:"required,ipv4" json:"allocation_pool_start"`
    AllocationPoolEnd   string `yaml:"allocation_pool_end" validate:"required,ipv4" json:"allocation_pool_end"`
    
    VRRPIP      string `yaml:"vrrp_ip" validate:"required,ipv4" json:"vrrp_ip"`
    VRRPEnabled bool   `yaml:"vrrp_enabled" json:"vrrp_enabled"`
    
    // Enum validation
    LoadbalancerProvider string `yaml:"loadbalancer_provider" validate:"oneof=ovn octavia metallb" json:"loadbalancer_provider"`
}
```

### Action 2: Implement "Effective Configuration" Export

Create a method that generates the schema-compliant JSON/YAML for IDE autocompletion.

**`internal/config/schema.go`**

```go
// GenerateSchema outputs a JSON schema based on the struct tags
func GenerateSchema() {
    // Use a library like 'jsonschema' (invopop/jsonschema) to reflect on types
    schema := jsonschema.Reflect(&Config{})
    output, _ := json.MarshalIndent(schema, "", "  ")
    os.WriteFile("schema/cluster.schema.json", output, 0644)
}
```

### Action 3: Validate Deployment Configuration

Ensure deployment method configuration is validated at the root level.

**`internal/config/types_deployment.go`**

```go
type Deployment struct {
    AutoDeploy bool   `yaml:"auto_deploy" json:"auto_deploy"`
    Method     string `yaml:"method" validate:"required,oneof=kubespray talos kamaji eks gke aks cluster-api" json:"method"`
    
    Kubespray *KubesprayConfig `yaml:"kubespray,omitempty" json:"kubespray,omitempty"`
    Kamaji    *KamajiConfig    `yaml:"kamaji,omitempty" json:"kamaji,omitempty"`
    Talos     *TalosConfig     `yaml:"talos,omitempty" json:"talos,omitempty"`
}

// Validate ensures the selected method has corresponding configuration
func (d *Deployment) Validate() error {
    switch d.Method {
    case "kubespray":
        if d.Kubespray == nil {
            return errors.New("kubespray method selected but kubespray config is missing")
        }
    case "kamaji":
        if d.Kamaji == nil || !d.Kamaji.Enabled {
            return errors.New("kamaji method selected but kamaji config is missing or disabled")
        }
    case "talos":
        if d.Talos == nil || !d.Talos.Enabled {
            return errors.New("talos method selected but talos config is missing or disabled")
        }
    }
    return nil
}
```

---

## Phase 2: The Core Logic (Registry & Hydration)

This implements the "Provider-Region Default Registry" to ensure determinism.

### Action 1: Define the Registry Interface

Create the contract for what a provider *must* supply.

**`internal/config/defaults/interface.go`**

```go
package defaults

type ProviderDefaults interface {
    GetImageID(osVersion string) string
    GetAvailabilityZones() []string
    GetNTPServers() []string
    GetDNSNameservers() []string
    GetDefaultStorageClass() string
    GetFlavor(role string) string
}
```

### Action 2: Implement the Registry

Populate the hardcoded defaults in a dedicated package, separated from the logic.

**Note**: Currently only OpenStack defaults are production-ready. AWS, GCP, and Azure entries are included as reference implementations for the provider abstraction pattern but are not scheduled for implementation.

**`internal/config/defaults/registry.go`**

```go
package defaults

var Registry = map[string]map[string]ProviderDefaults{
    "openstack": {
        "sjc3": &OpenStackRegionDefaults{
            ImageIDs: map[string]string{
                "24": "799dcf97-3656-4361-8187-13ab1b295e33",
            },
            NTPServers: []string{
                "time.sjc3.rackspace.com",
                "time2.sjc3.rackspace.com",
            },
            DNSNameservers: []string{"8.8.8.8", "8.8.4.4"},
            DefaultStorageClass: "csi-cinder-sc-delete",
            Flavors: map[string]string{
                "bastion": "gp.0.2.2",
                "master":  "gp.0.4.8",
                "worker":  "gp.0.4.16",
            },
        },
    },
    // Future/reference implementations (not scheduled)
    // "aws": {
    //     "us-east-1": &AWSRegionDefaults{
    //         // AWS-specific defaults
    //     },
    // },
}

type OpenStackRegionDefaults struct {
    ImageIDs            map[string]string
    NTPServers          []string
    DNSNameservers      []string
    DefaultStorageClass string
    Flavors             map[string]string
}

func (o *OpenStackRegionDefaults) GetImageID(osVersion string) string {
    return o.ImageIDs[osVersion]
}

func (o *OpenStackRegionDefaults) GetNTPServers() []string {
    return o.NTPServers
}

func (o *OpenStackRegionDefaults) GetDNSNameservers() []string {
    return o.DNSNameservers
}

func (o *OpenStackRegionDefaults) GetDefaultStorageClass() string {
    return o.DefaultStorageClass
}

func (o *OpenStackRegionDefaults) GetFlavor(role string) string {
    return o.Flavors[role]
}

func (o *OpenStackRegionDefaults) GetAvailabilityZones() []string {
    return []string{"az1"}
}
```

### Action 3: The Hydration Engine

Implement the logic that fills in blank fields *without* overwriting user input.

**`internal/config/hydrate.go`**

```go
package config

import (
    "fmt"
    "github.com/rackerlabs/opencenter-cli/internal/config/defaults"
)

func (c *Config) Hydrate() error {
    infra := &c.OpenCenter.Infrastructure
    
    // 1. Identify Provider & Region
    region := c.OpenCenter.Meta.Region
    provider := infra.Provider
    
    regionDefaults, ok := defaults.Registry[provider][region]
    if !ok {
        return fmt.Errorf("unsupported region %q for provider %q", region, provider)
    }

    // 2. Apply Infrastructure Defaults
    if len(infra.Networking.NTPServers) == 0 {
        infra.Networking.NTPServers = regionDefaults.GetNTPServers()
    }
    
    if len(infra.Networking.DNSNameservers) == 0 {
        infra.Networking.DNSNameservers = regionDefaults.GetDNSNameservers()
    }
    
    if infra.Storage.DefaultStorageClass == "" {
        infra.Storage.DefaultStorageClass = regionDefaults.GetDefaultStorageClass()
    }

    // 3. Provider-Specific Hydration
    if provider == "openstack" {
        if infra.Cloud.OpenStack.ImageID == "" {
            infra.Cloud.OpenStack.ImageID = regionDefaults.GetImageID(infra.OSVersion)
        }
        
        if infra.Compute.FlavorBastion == "" {
            infra.Compute.FlavorBastion = regionDefaults.GetFlavor("bastion")
        }
        
        if infra.Compute.FlavorMaster == "" {
            infra.Compute.FlavorMaster = regionDefaults.GetFlavor("master")
        }
        
        if infra.Compute.FlavorWorker == "" {
            infra.Compute.FlavorWorker = regionDefaults.GetFlavor("worker")
        }
    }
    
    // 4. Deployment Method Hydration
    if c.Deployment.Method == "" {
        c.Deployment.Method = "kubespray" // Default deployment method
    }
    
    return nil
}
```

---

## Phase 3: The Refactor (Struct Reorganization)

Refactor the Go structs to match the unified configuration architecture.

### Action 1: Create the Polymorphic Containers

Use Go interfaces or pointer-based structs to handle the provider-specific sections.

**`internal/config/types_infrastructure.go`**

```go
package config

type Infrastructure struct {
    Provider             string                    `yaml:"provider" json:"provider" validate:"required,oneof=openstack aws gcp azure baremetal vsphere"`
    SSH                  SSHConfig                 `yaml:"ssh" json:"ssh"`
    OSVersion            string                    `yaml:"os_version" json:"os_version"`
    ServerGroupAffinity  []string                  `yaml:"server_group_affinity" json:"server_group_affinity"`
    K8sAPIIP             string                    `yaml:"k8s_api_ip" json:"k8s_api_ip"`
    NodeNaming           NodeNamingConfig          `yaml:"node_naming" json:"node_naming"`
    Bastion              BastionConfig             `yaml:"bastion" json:"bastion"`
    Compute              ComputeConfig             `yaml:"compute" json:"compute"`
    Storage              StorageConfig             `yaml:"storage" json:"storage"`
    Networking           InfrastructureNetworking  `yaml:"networking" json:"networking"`
    Cloud                CloudConfig               `yaml:"cloud" json:"cloud"`
}

type CloudConfig struct {
    // Pointers allow us to check if a section is present (nil check)
    // NOTE: Only OpenStack is production-ready; others are reference implementations
    OpenStack *OpenStackCloudConfig `yaml:"openstack,omitempty" json:"openstack,omitempty"`
    AWS       *AWSCloudConfig       `yaml:"aws,omitempty" json:"aws,omitempty"`       // Future/reference only
    GCP       *GCPCloudConfig       `yaml:"gcp,omitempty" json:"gcp,omitempty"`       // Future/reference only
    Azure     *AzureCloudConfig     `yaml:"azure,omitempty" json:"azure,omitempty"`   // Future/reference only
}

// ValidateProvider ensures clean separation between providers
func (c *CloudConfig) ValidateProvider(providerName string) error {
    switch providerName {
    case "openstack":
        if c.OpenStack == nil {
            return errors.New("missing 'openstack' config")
        }
        if c.AWS != nil || c.GCP != nil || c.Azure != nil {
            return errors.New("found multiple provider configs but only one should be present")
        }
    case "aws", "gcp", "azure":
        return fmt.Errorf("provider %q is not currently supported (OpenStack only)", providerName)
    // ... similar for other future providers
    }
    return nil
}
```

### Action 2: Consolidate Deployment Configuration

Ensure deployment is at root level with all method-specific configuration.

**`internal/config/types_deployment.go`**

```go
package config

type Deployment struct {
    AutoDeploy bool              `yaml:"auto_deploy" json:"auto_deploy"`
    Method     string            `yaml:"method" json:"method" validate:"required"`
    Kubespray  *KubesprayConfig  `yaml:"kubespray,omitempty" json:"kubespray,omitempty"`
    Kamaji     *KamajiConfig     `yaml:"kamaji,omitempty" json:"kamaji,omitempty"`
    Talos      *TalosConfig      `yaml:"talos,omitempty" json:"talos,omitempty"`
}

type KubesprayConfig struct {
    Version string                 `yaml:"version" json:"version"`
    Modules KubesprayModulesConfig `yaml:"modules" json:"modules"`
}

type KamajiConfig struct {
    Enabled      bool                  `yaml:"enabled" json:"enabled"`
    Version      string                `yaml:"version" json:"version"`
    ControlPlane KamajiControlPlane    `yaml:"control_plane" json:"control_plane"`
    ClusterAPI   KamajiClusterAPI      `yaml:"cluster_api" json:"cluster_api"`
    WorkerPools  []KamajiWorkerPool    `yaml:"worker_pools" json:"worker_pools"`
    Modules      KamajiModulesConfig   `yaml:"modules" json:"modules"`
}
```

### Action 3: Update Root Config Structure

Ensure the root Config struct matches the unified hierarchy.

**`internal/config/config.go`**

```go
package config

type Config struct {
    SchemaVersion string               `yaml:"schema_version" json:"schema_version"`
    Metadata      Metadata             `yaml:"metadata" json:"metadata"`
    OpenCenter    SimplifiedOpenCenter `yaml:"opencenter" json:"opencenter"`
    Deployment    Deployment           `yaml:"deployment" json:"deployment"`
    OpenTofu      OpenTofuConfig       `yaml:"opentofu" json:"opentofu"`
    Secrets       SecretsConfig        `yaml:"secrets" json:"secrets"`
}

type SimplifiedOpenCenter struct {
    Meta            ClusterMeta       `yaml:"meta" json:"meta"`
    Secrets         OpenCenterSecrets `yaml:"secrets,omitempty" json:"secrets,omitempty"`
    Infrastructure  Infrastructure    `yaml:"infrastructure" json:"infrastructure"`
    Cluster         ClusterConfig     `yaml:"cluster" json:"cluster"`
    GitOps          GitOpsConfig      `yaml:"gitops" json:"gitops"`
    Storage         StorageConfig     `yaml:"storage,omitempty" json:"storage,omitempty"` // Deprecated, moved to infrastructure.storage
    ManagedServices ServiceMap        `yaml:"managed_services" json:"managed_services"`
    Services        ServiceMap        `yaml:"services" json:"services"`
}
```

---

## Phase 4: The Intelligence (Reference Resolution)

Implement the logic to resolve `${infrastructure.networking.vrrp_ip}` and `${secrets.foo}`.

### Action 1: The Resolver Logic

Do not use simple string replacement. Use a walker that identifies fields needing resolution.

**`internal/config/resolve.go`**

```go
package config

import (
    "fmt"
    "reflect"
    "strings"
)

// ResolveReferences walks the struct using reflection and resolves ${...} references
func (c *Config) ResolveReferences(secrets *SecretsConfig) error {
    // 1. Create a data map for lookup
    data := map[string]interface{}{
        "infrastructure": c.OpenCenter.Infrastructure,
        "cluster":        c.OpenCenter.Cluster,
        "deployment":     c.Deployment,
        "secrets":        secrets,
        "meta":           c.OpenCenter.Meta,
    }

    // 2. Walk the struct and find strings matching ${...}
    return walkStruct(reflect.ValueOf(c).Elem(), func(field reflect.Value) error {
        if field.Kind() != reflect.String {
            return nil
        }
        
        str := field.String()
        if !strings.HasPrefix(str, "${") || !strings.HasSuffix(str, "}") {
            return nil
        }
        
        // Extract reference path
        path := str[2 : len(str)-1] // remove ${ and }
        
        // Resolve the reference
        val, err := lookupPath(data, path)
        if err != nil {
            return fmt.Errorf("reference not found: %s", path)
        }
        
        // Set the resolved value
        field.SetString(fmt.Sprintf("%v", val))
        return nil
    })
}

// walkStruct recursively walks a struct and calls fn on each field
func walkStruct(v reflect.Value, fn func(reflect.Value) error) error {
    if !v.IsValid() {
        return nil
    }
    
    switch v.Kind() {
    case reflect.Struct:
        for i := 0; i < v.NumField(); i++ {
            if err := walkStruct(v.Field(i), fn); err != nil {
                return err
            }
        }
    case reflect.Ptr:
        if !v.IsNil() {
            return walkStruct(v.Elem(), fn)
        }
    case reflect.Slice, reflect.Array:
        for i := 0; i < v.Len(); i++ {
            if err := walkStruct(v.Index(i), fn); err != nil {
                return err
            }
        }
    case reflect.Map:
        for _, key := range v.MapKeys() {
            if err := walkStruct(v.MapIndex(key), fn); err != nil {
                return err
            }
        }
    default:
        return fn(v)
    }
    return nil
}

// lookupPath resolves a dot-separated path in a nested map/struct
func lookupPath(data map[string]interface{}, path string) (interface{}, error) {
    parts := strings.Split(path, ".")
    var current interface{} = data
    
    for _, part := range parts {
        switch v := current.(type) {
        case map[string]interface{}:
            var ok bool
            current, ok = v[part]
            if !ok {
                return nil, fmt.Errorf("path not found: %s", path)
            }
        default:
            // Use reflection to access struct fields
            val := reflect.ValueOf(current)
            if val.Kind() == reflect.Ptr {
                val = val.Elem()
            }
            if val.Kind() != reflect.Struct {
                return nil, fmt.Errorf("cannot traverse non-struct: %s", path)
            }
            field := val.FieldByName(part)
            if !field.IsValid() {
                return nil, fmt.Errorf("field not found: %s", part)
            }
            current = field.Interface()
        }
    }
    
    return current, nil
}
```

### Action 2: Secret Reference Security

Ensure that secrets referenced this way are treated carefully (e.g., redacted in logs).

**`internal/config/secrets.go`**

```go
package config

import "strings"

// RedactSecrets replaces secret values with [REDACTED] for logging
func (c *Config) RedactSecrets() *Config {
    redacted := *c
    
    // Redact all fields in secrets section
    redacted.Secrets = SecretsConfig{
        SOPSAgeKeyFile: "[REDACTED]",
        SSHKey: SSHKeyConfig{
            Private: "[REDACTED]",
            Public:  "[REDACTED]",
            Cypher:  c.Secrets.SSHKey.Cypher,
        },
        // ... redact other secret fields
    }
    
    return &redacted
}

// IsSecretReference checks if a string is a reference to a secret
func IsSecretReference(s string) bool {
    return strings.HasPrefix(s, "${secrets.")
}
```

---

## Phase 5: The Bridge (Migration Tooling)

You cannot break existing users. You need a CLI tool that reads `v1` config and outputs `v2` config.

### Action 1: The Migration Command

**`cmd/cluster_migrate.go`**

```go
package cmd

import (
    "fmt"
    "log"
    
    "github.com/spf13/cobra"
    "github.com/rackerlabs/opencenter-cli/internal/config"
    "gopkg.in/yaml.v3"
)

func newClusterMigrateCmd() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "migrate <config-file>",
        Short: "Migrate v1 configuration to v2 schema",
        Long: `Migrate an existing v1 configuration file to the v2 unified schema.
        
This command:
- Reads the v1 configuration
- Maps fields to their new v2 locations
- Applies hydration to fill in defaults
- Outputs a v2-compliant configuration file`,
        Args: cobra.ExactArgs(1),
        RunE: runMigrate,
    }
    
    cmd.Flags().StringP("output", "o", "", "Output file path (default: <input>-v2.yaml)")
    
    return cmd
}

func runMigrate(cmd *cobra.Command, args []string) error {
    oldConfigPath := args[0]
    outputPath, _ := cmd.Flags().GetString("output")
    
    if outputPath == "" {
        outputPath = oldConfigPath + "-v2.yaml"
    }
    
    // 1. Load Old Config (into old struct types)
    oldCfg, err := loadV1Config(oldConfigPath)
    if err != nil {
        return fmt.Errorf("failed to load v1 config: %w", err)
    }
    
    // 2. Map to New Config
    newCfg := migrateV1ToV2(oldCfg)
    
    // 3. Run Hydration (Critical!)
    // This writes the implicit defaults from V1 into explicit values in V2
    if err := newCfg.Hydrate(); err != nil {
        log.Printf("Warning: hydration failed: %v", err)
    }
    
    // 4. Validate the new configuration
    if err := newCfg.Validate(); err != nil {
        return fmt.Errorf("migrated config is invalid: %w", err)
    }
    
    // 5. Write Output
    if err := writeYAML(newCfg, outputPath); err != nil {
        return fmt.Errorf("failed to write output: %w", err)
    }
    
    fmt.Printf("✓ Migrated configuration written to: %s\n", outputPath)
    fmt.Println("\nPlease review the migrated configuration before using it.")
    fmt.Println("Key changes in v2:")
    fmt.Println("  - deployment configuration moved to root level")
    fmt.Println("  - storage configuration moved to infrastructure.storage")
    fmt.Println("  - worker pools moved to infrastructure.compute")
    
    return nil
}

func migrateV1ToV2(oldCfg *ConfigV1) *config.Config {
    newCfg := &config.Config{
        SchemaVersion: "2.0",
        Metadata: config.Metadata{
            SchemaVersion: "2.0",
            // ... copy metadata
        },
    }
    
    // Map opencenter section
    newCfg.OpenCenter.Meta = config.ClusterMeta{
        Name:         oldCfg.OpenCenter.Meta.Name,
        Organization: oldCfg.OpenCenter.Meta.Organization,
        Env:          oldCfg.OpenCenter.Meta.Env,
        Region:       oldCfg.OpenCenter.Meta.Region,
    }
    
    // Map infrastructure
    newCfg.OpenCenter.Infrastructure = config.Infrastructure{
        Provider:  oldCfg.OpenCenter.Infrastructure.Provider,
        Compute:   oldCfg.OpenCenter.Infrastructure.Compute,
        Networking: oldCfg.OpenCenter.Infrastructure.Networking,
        // ... map other fields
    }
    
    // CRITICAL: Move storage from opencenter.storage to infrastructure.storage
    if oldCfg.OpenCenter.Storage != nil {
        newCfg.OpenCenter.Infrastructure.Storage = *oldCfg.OpenCenter.Storage
    }
    
    // CRITICAL: Move deployment from opencenter.deployment to root deployment
    newCfg.Deployment = config.Deployment{
        AutoDeploy: oldCfg.Deployment.AutoDeploy,
        Method:     oldCfg.OpenCenter.Deployment.Method,
        Kubespray:  oldCfg.OpenCenter.Deployment.Kubespray,
        Kamaji:     oldCfg.OpenCenter.Deployment.Kamaji,
        Talos:      oldCfg.OpenCenter.Talos,
    }
    
    // Map services
    newCfg.OpenCenter.Services = oldCfg.OpenCenter.Services
    newCfg.OpenCenter.ManagedServices = oldCfg.OpenCenter.ManagedServices
    
    // Map other root-level sections
    newCfg.OpenTofu = oldCfg.OpenTofu
    newCfg.Secrets = oldCfg.Secrets
    
    return newCfg
}

func writeYAML(cfg *config.Config, path string) error {
    data, err := yaml.Marshal(cfg)
    if err != nil {
        return err
    }
    return os.WriteFile(path, data, 0644)
}
```

### Action 2: Migration Testing

Create test cases to ensure migration preserves all data.

**`cmd/cluster_migrate_test.go`**

```go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestMigrateV1ToV2(t *testing.T) {
    tests := []struct {
        name     string
        v1Config string
        validate func(*testing.T, *config.Config)
    }{
        {
            name: "deployment moved to root",
            v1Config: `
opencenter:
  deployment:
    method: kubespray
    kubespray:
      version: v2.29.1
deployment:
  auto_deploy: false
`,
            validate: func(t *testing.T, cfg *config.Config) {
                assert.Equal(t, "kubespray", cfg.Deployment.Method)
                assert.Equal(t, "v2.29.1", cfg.Deployment.Kubespray.Version)
                assert.False(t, cfg.Deployment.AutoDeploy)
            },
        },
        {
            name: "storage moved to infrastructure",
            v1Config: `
opencenter:
  storage:
    default_storage_class: csi-cinder-sc-delete
    worker_volume_size: 100
`,
            validate: func(t *testing.T, cfg *config.Config) {
                assert.Equal(t, "csi-cinder-sc-delete", cfg.OpenCenter.Infrastructure.Storage.DefaultStorageClass)
                assert.Equal(t, 100, cfg.OpenCenter.Infrastructure.Storage.WorkerVolumeSize)
            },
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            oldCfg := loadV1ConfigFromString(tt.v1Config)
            newCfg := migrateV1ToV2(oldCfg)
            tt.validate(t, newCfg)
        })
    }
}
```

---


## Implementation Checklist

### Phase 1: Validation & Schema
- [ ] Add validation tags to all struct types in `internal/config/types_*.go`
- [ ] Implement `Validate()` methods for complex cross-field validation
- [ ] Create JSON schema generation in `internal/config/schema.go`
- [ ] Add deployment method validation in `types_deployment.go`
- [ ] Test schema generation with `mise run schema`

### Phase 2: Registry & Hydration
- [ ] Create `internal/config/defaults/` package
- [ ] Define `ProviderDefaults` interface
- [ ] Implement `Registry` with OpenStack defaults (production-ready)
- [ ] Add reference implementations for AWS, GCP, Azure (for architectural completeness)
- [ ] Create region-specific default structs for OpenStack
- [ ] Implement `Hydrate()` method in `internal/config/config.go`
- [ ] Add OpenStack-specific hydration logic
- [ ] Test hydration with minimal OpenStack configs

### Phase 3: Struct Reorganization
- [ ] Update `SimplifiedOpenCenter` struct to remove `deployment` field
- [ ] Ensure `Deployment` is at root level in `Config` struct
- [ ] Update `CloudConfig` with pointer-based provider configs
- [ ] Implement `ValidateProvider()` for cloud config
- [ ] Move storage from `opencenter.storage` to `infrastructure.storage`
- [ ] Update all references in codebase
- [ ] Run full test suite to catch breaking changes

### Phase 4: Reference Resolution
- [ ] Implement `ResolveReferences()` in `internal/config/resolve.go`
- [ ] Create `walkStruct()` helper for reflection-based traversal
- [ ] Implement `lookupPath()` for dot-notation resolution
- [ ] Add `IsSecretReference()` helper
- [ ] Implement `RedactSecrets()` for logging
- [ ] Test reference resolution with complex configs
- [ ] Add error handling for missing references

### Phase 5: Migration Tooling
- [ ] Create `cmd/cluster_migrate.go` command
- [ ] Implement `loadV1Config()` for backward compatibility
- [ ] Implement `migrateV1ToV2()` mapping function
- [ ] Add migration for deployment config (opencenter.deployment → deployment)
- [ ] Add migration for storage config (opencenter.storage → infrastructure.storage)
- [ ] Add migration for worker pools (cluster.kubernetes → infrastructure.compute)
- [ ] Create migration tests in `cmd/cluster_migrate_test.go`
- [ ] Document migration process in user-facing docs
- [ ] Test migration with real v1 configs

### Integration & Testing
- [ ] Update `cluster init` to generate v2 configs
- [ ] Update `cluster validate` to work with v2 schema
- [ ] Update `cluster render` to use v2 structure
- [ ] Run integration tests with all deployment methods
- [ ] Test with OpenStack provider (production-ready)
- [ ] Verify GitOps generation works with v2 configs
- [ ] Update documentation examples to v2 format

### Documentation
- [ ] Update `docs/cluster-config/unified-configuration.md` (✓ completed)
- [ ] Update `docs/cluster-config/cluster-config-full.yaml` (✓ completed)
- [ ] Update `docs/cluster-config/cluster-config-kamaji.yaml` (✓ completed)
- [ ] Update `docs/cluster-config/unified-configuration-implementation.md` (✓ completed)
- [ ] Create migration guide for users
- [ ] Update CLI reference documentation
- [ ] Add examples for each deployment method

## Key Migration Points

### Deployment Configuration
**v1 Structure:**
```yaml
opencenter:
  deployment:
    method: kubespray
    kubespray:
      version: v2.29.1
deployment:
  auto_deploy: false
```

**v2 Structure:**
```yaml
deployment:
  auto_deploy: false
  method: kubespray
  kubespray:
    version: v2.29.1
```

### Storage Configuration
**v1 Structure:**
```yaml
opencenter:
  storage:
    default_storage_class: csi-cinder-sc-delete
    worker_volume_size: 100
```

**v2 Structure:**
```yaml
opencenter:
  infrastructure:
    storage:
      default_storage_class: csi-cinder-sc-delete
      worker_volume_size: 100
```

### Worker Pools
**v1 Structure:**
```yaml
opencenter:
  cluster:
    kubernetes:
      additional_server_pools_worker:
        - name: high-memory
          worker_count: 2
```

**v2 Structure:**
```yaml
opencenter:
  infrastructure:
    compute:
      additional_server_pools_worker:
        - name: high-memory
          worker_count: 2
```

## Testing Strategy

### Unit Tests
- Test each phase independently
- Mock external dependencies (file I/O, network)
- Test error conditions and edge cases
- Verify validation rules work correctly

### Integration Tests
- Test full config loading and hydration
- Test reference resolution across sections
- Test migration from v1 to v2
- Test with real provider credentials (in CI)

### Property-Based Tests
- Use `gopter` for generative testing
- Test that hydration is idempotent
- Test that validation catches all invalid configs
- Test that migration preserves all data

### Regression Tests
- Keep v1 test configs and verify migration
- Test backward compatibility where needed
- Verify no data loss during migration

## Performance Considerations

### Hydration Performance
- Cache provider defaults to avoid repeated lookups
- Use lazy evaluation for expensive operations
- Profile hydration with large configs

### Validation Performance
- Validate in parallel where possible
- Cache validation results for unchanged sections
- Use early returns to fail fast

### Reference Resolution Performance
- Build reference graph once, reuse for multiple resolutions
- Detect circular references early
- Cache resolved values

## Security Considerations

### Secret Handling
- Never log secret values
- Redact secrets in error messages
- Use secure memory for secret storage
- Clear secrets from memory after use

### Reference Resolution Security
- Validate reference paths before resolution
- Prevent path traversal attacks
- Limit recursion depth
- Sanitize error messages

### Validation Security
- Validate all user input
- Prevent injection attacks in templates
- Validate file paths are within allowed directories
- Check for malicious YAML constructs

## Rollout Plan

### Phase 1: Internal Testing (Week 1-2)
- Implement core functionality
- Test with internal clusters
- Fix critical bugs

### Phase 2: Beta Release (Week 3-4)
- Release as beta feature flag
- Gather user feedback
- Document migration process

### Phase 3: General Availability (Week 5-6)
- Make v2 the default for new clusters
- Provide migration tool for existing clusters
- Deprecate v1 format (with warning)

### Phase 4: v1 Removal (3-6 months later)
- Remove v1 support code
- Clean up deprecated fields
- Update all documentation
