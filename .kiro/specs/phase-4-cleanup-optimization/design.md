# Design Document: Phase 4 Cleanup

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Components and Interfaces](#components-and-interfaces)
- [Data Models](#data-models)
- [Correctness Properties](#correctness-properties)
- [Error Handling](#error-handling)
- [Testing Strategy](#testing-strategy)

## Overview

Phase 4 completes the architectural refactoring by eliminating boilerplate code, consolidating utilities, and removing unused abstractions. This design focuses on four main areas:

1. **Service Plugin Consolidation**: Create a base service plugin using composition to eliminate 1,230 LOC of duplicated boilerplate across 15+ plugins
2. **Path Resolution Consolidation**: Ensure all path operations use the unified PathResolver with proper caching and platform compatibility
3. **File Operations Migration**: Complete the migration to the FileSystem wrapper, eliminating all remaining direct os.ReadFile/os.WriteFile calls
4. **Interface Cleanup**: Remove unused interfaces that have only single implementations to reduce unnecessary abstraction

### Design Principles

- **Composition over Inheritance**: Use Go's embedding to provide base functionality
- **Single Responsibility**: Each component has one clear purpose
- **YAGNI (You Aren't Gonna Need It)**: Remove abstractions that don't provide value
- **Consistency**: All similar operations use the same patterns
- **Simplicity**: Prefer concrete types over interfaces when only one implementation exists

### Dependencies

This phase builds on:
- **Phase 1**: FileSystem wrapper and utility foundations
- **Phase 2**: ValidationEngine for validation patterns
- **Phase 3**: ConfigurationManager for configuration management

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     Service Plugins                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ CertManager  │  │    Loki      │  │   Velero     │      │
│  │   Plugin     │  │   Plugin     │  │   Plugin     │ ...  │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘      │
│         │                  │                  │              │
│         └──────────────────┴──────────────────┘              │
│                            │                                 │
│                   ┌────────▼────────┐                        │
│                   │ BaseServicePlugin│                       │
│                   │  (Composition)   │                       │
│                   └─────────────────┘                        │
└─────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────┐
│                    Core Utilities                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐      │
│  │ PathResolver │  │  FileSystem  │  │Configuration │      │
│  │  (Concrete)  │  │   Wrapper    │  │   Manager    │      │
│  │              │  │  (Concrete)  │  │  (Concrete)  │      │
│  └──────────────┘  └──────────────┘  └──────────────┘      │
└─────────────────────────────────────────────────────────────┘
```

### Component Relationships

1. **Service Plugins** embed BaseServicePlugin for common functionality
2. **PathResolver** is used directly (no interface) by all components needing path resolution
3. **FileSystem** wrapper is used directly (no interface) by all file operations
4. **ConfigurationManager** uses concrete PathResolver and FileSystem types

### Migration Strategy

The migration follows this sequence:

1. **Create BaseServicePlugin** with composition pattern
2. **Migrate plugins incrementally** (one at a time, test each)
3. **Consolidate path resolution** (update all callers to use PathResolver)
4. **Complete file operations migration** (eliminate all direct os calls)
5. **Remove unused interfaces** (replace with concrete types)

## Components and Interfaces

### BaseServicePlugin

The base service plugin provides common functionality through composition:

```go
// internal/services/base_plugin.go

// PluginMetadata contains all standard plugin metadata
type PluginMetadata struct {
    Name        string
    Version     string
    Description string
    Type        string  // "core", "observability", "application", "backup", "storage"
    Author      string
    License     string
}

// BaseServicePlugin provides common functionality for all service plugins
type BaseServicePlugin struct {
    metadata  PluginMetadata
    validator func(interface{}) error
    renderer  func(interface{}) ([]byte, error)
}

// NewBasePlugin creates a new base plugin with the given metadata
func NewBasePlugin(metadata PluginMetadata) *BaseServicePlugin {
    return &BaseServicePlugin{
        metadata:  metadata,
        validator: func(interface{}) error { return nil },
        renderer:  func(interface{}) ([]byte, error) { return nil, nil },
    }
}

// Metadata accessor methods (implements ServicePlugin interface)
func (p *BaseServicePlugin) Name() string        { return p.metadata.Name }
func (p *BaseServicePlugin) Version() string     { return p.metadata.Version }
func (p *BaseServicePlugin) Description() string { return p.metadata.Description }
func (p *BaseServicePlugin) Type() string        { return p.metadata.Type }
func (p *BaseServicePlugin) Author() string      { return p.metadata.Author }
func (p *BaseServicePlugin) License() string     { return p.metadata.License }

// Validate delegates to the injected validator function
func (p *BaseServicePlugin) Validate(config interface{}) error {
    return p.validator(config)
}

// Render delegates to the injected renderer function
func (p *BaseServicePlugin) Render(config interface{}) ([]byte, error) {
    return p.renderer(config)
}

// SetValidator allows plugins to inject custom validation logic
func (p *BaseServicePlugin) SetValidator(validator func(interface{}) error) {
    p.validator = validator
}

// SetRenderer allows plugins to inject custom rendering logic
func (p *BaseServicePlugin) SetRenderer(renderer func(interface{}) ([]byte, error)) {
    p.renderer = renderer
}
```

### Migrated Plugin Pattern

Each plugin embeds BaseServicePlugin and provides only service-specific logic:

```go
// internal/services/cert_manager.go

// CertManagerPlugin handles cert-manager service configuration
type CertManagerPlugin struct {
    *BaseServicePlugin
}

// NewCertManagerPlugin creates a new cert-manager plugin
func NewCertManagerPlugin() *CertManagerPlugin {
    base := NewBasePlugin(PluginMetadata{
        Name:        "cert-manager",
        Version:     "1.0.0",
        Description: "Certificate management for Kubernetes",
        Type:        "core",
        Author:      "opencenter",
        License:     "Apache-2.0",
    })
    
    plugin := &CertManagerPlugin{BaseServicePlugin: base}
    
    // Inject service-specific logic
    base.SetValidator(plugin.validate)
    base.SetRenderer(plugin.render)
    
    return plugin
}

// validate implements cert-manager specific validation
func (p *CertManagerPlugin) validate(config interface{}) error {
    cfg, ok := config.(*services.CertManagerConfig)
    if !ok {
        return fmt.Errorf("invalid config type for cert-manager")
    }
    
    // Service-specific validation logic
    if cfg.Email == "" {
        return fmt.Errorf("email is required for cert-manager")
    }
    
    if cfg.Issuer == "" {
        return fmt.Errorf("issuer is required for cert-manager")
    }
    
    return nil
}

// render implements cert-manager specific rendering
func (p *CertManagerPlugin) render(config interface{}) ([]byte, error) {
    cfg, ok := config.(*services.CertManagerConfig)
    if !ok {
        return nil, fmt.Errorf("invalid config type for cert-manager")
    }
    
    // Service-specific rendering logic
    return yaml.Marshal(cfg)
}
```

### PathResolver (Concrete Type)

The PathResolver is used directly without an interface:

```go
// internal/core/paths/resolver.go

// PathResolver resolves file system paths for cluster resources
type PathResolver struct {
    baseDir string
    cache   map[string]string
    mu      sync.RWMutex
}

// NewPathResolver creates a new path resolver
func NewPathResolver(baseDir string) *PathResolver {
    return &PathResolver{
        baseDir: baseDir,
        cache:   make(map[string]string),
    }
}

// ResolveConfigPath resolves the path to a cluster configuration file
func (pr *PathResolver) ResolveConfigPath(clusterName, organization string) (string, error) {
    if clusterName == "" {
        return "", fmt.Errorf("cluster name cannot be empty")
    }
    
    // Check cache
    cacheKey := fmt.Sprintf("config:%s:%s", organization, clusterName)
    pr.mu.RLock()
    if cached, found := pr.cache[cacheKey]; found {
        pr.mu.RUnlock()
        return cached, nil
    }
    pr.mu.RUnlock()
    
    // Use default organization if not specified
    if organization == "" {
        organization = "opencenter"
    }
    
    // Construct path: ~/.config/opencenter/clusters/<org>/.<cluster>-config.yaml
    path := filepath.Join(pr.baseDir, organization, "."+clusterName+"-config.yaml")
    
    // Normalize for platform (handles Windows vs Unix paths)
    path = filepath.Clean(path)
    
    // Cache result
    pr.mu.Lock()
    pr.cache[cacheKey] = path
    pr.mu.Unlock()
    
    return path, nil
}

// ResolveSecretsPath resolves the path to a cluster's secrets directory
func (pr *PathResolver) ResolveSecretsPath(clusterName, organization string) (string, error) {
    if clusterName == "" {
        return "", fmt.Errorf("cluster name cannot be empty")
    }
    
    cacheKey := fmt.Sprintf("secrets:%s:%s", organization, clusterName)
    pr.mu.RLock()
    if cached, found := pr.cache[cacheKey]; found {
        pr.mu.RUnlock()
        return cached, nil
    }
    pr.mu.RUnlock()
    
    if organization == "" {
        organization = "opencenter"
    }
    
    path := filepath.Join(pr.baseDir, organization, "secrets")
    path = filepath.Clean(path)
    
    pr.mu.Lock()
    pr.cache[cacheKey] = path
    pr.mu.Unlock()
    
    return path, nil
}

// ResolveGitOpsPath resolves the path to a cluster's GitOps repository
func (pr *PathResolver) ResolveGitOpsPath(clusterName, organization string) (string, error) {
    if clusterName == "" {
        return "", fmt.Errorf("cluster name cannot be empty")
    }
    
    cacheKey := fmt.Sprintf("gitops:%s:%s", organization, clusterName)
    pr.mu.RLock()
    if cached, found := pr.cache[cacheKey]; found {
        pr.mu.RUnlock()
        return cached, nil
    }
    pr.mu.RUnlock()
    
    if organization == "" {
        organization = "opencenter"
    }
    
    path := filepath.Join(pr.baseDir, organization, "gitops")
    path = filepath.Clean(path)
    
    pr.mu.Lock()
    pr.cache[cacheKey] = path
    pr.mu.Unlock()
    
    return path, nil
}

// ClearCache clears the path resolution cache
func (pr *PathResolver) ClearCache() {
    pr.mu.Lock()
    pr.cache = make(map[string]string)
    pr.mu.Unlock()
}
```

### FileSystem Wrapper (Concrete Type)

The FileSystem wrapper is used directly without an interface:

```go
// internal/core/files/filesystem.go (already exists from Phase 1)

// FileSystem provides atomic file operations with proper error handling
type FileSystem struct {
    // Implementation already exists from Phase 1
}

// ReadFile reads a file with proper error wrapping
func (fs *FileSystem) ReadFile(path string) ([]byte, error) {
    // Implementation already exists
}

// WriteFile writes a file atomically with proper error wrapping
func (fs *FileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
    // Implementation already exists
}
```

### ConfigurationManager (Updated)

The ConfigurationManager is updated to use concrete types:

```go
// internal/config/manager.go

// ConfigurationManager manages cluster configurations
type ConfigurationManager struct {
    loader       *ConfigLoader        // Was: ConfigLoaderInterface
    pathResolver *PathResolver        // Was: PathResolverInterface
    cache        *ConfigCache         // Was: ConfigCacheInterface
    validator    ConfigValidatorInterface // Keep: multiple implementations exist
    fileSystem   *FileSystem          // Concrete type
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager(
    loader *ConfigLoader,
    pathResolver *PathResolver,
    cache *ConfigCache,
    validator ConfigValidatorInterface,
    fileSystem *FileSystem,
) *ConfigurationManager {
    return &ConfigurationManager{
        loader:       loader,
        pathResolver: pathResolver,
        cache:        cache,
        validator:    validator,
        fileSystem:   fileSystem,
    }
}
```

## Data Models

### PluginMetadata

```go
type PluginMetadata struct {
    Name        string  // Plugin identifier (e.g., "cert-manager")
    Version     string  // Plugin version (e.g., "1.0.0")
    Description string  // Human-readable description
    Type        string  // Plugin category: "core", "observability", "application", "backup", "storage"
    Author      string  // Plugin author (e.g., "opencenter")
    License     string  // License identifier (e.g., "Apache-2.0")
}
```

### Plugin Registry Entry

```go
type PluginRegistryEntry struct {
    Plugin      ServicePlugin
    Metadata    PluginMetadata
    RegisteredAt time.Time
}
```

### Path Cache Entry

```go
type PathCacheEntry struct {
    Key   string  // Format: "type:org:cluster"
    Path  string  // Resolved absolute path
    CachedAt time.Time
}
```

### File Operation Context

```go
type FileOperationContext struct {
    Path      string
    Operation string  // "read", "write", "delete"
    Timestamp time.Time
    Error     error
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*


### Property 1: Base Plugin Metadata Accessibility

*For any* PluginMetadata with valid fields (name, version, description, type, author, license), when a BaseServicePlugin is created with that metadata, all accessor methods (Name(), Version(), Description(), Type(), Author(), License()) should return the exact values from the metadata.

**Validates: Requirements 1.1, 1.2, 1.6**

### Property 2: Custom Logic Injection

*For any* custom validator function and custom renderer function, when they are injected into a BaseServicePlugin using SetValidator and SetRenderer, calling Validate() and Render() should invoke the injected functions and return their results.

**Validates: Requirements 1.4, 1.7, 1.8**

### Property 3: Plugin Composition

*For any* plugin that embeds BaseServicePlugin, the embedded plugin should have access to all base methods without reimplementing them, and should be able to override specific methods while retaining access to base functionality.

**Validates: Requirements 1.3**

### Property 4: Migrated Plugin Behavioral Equivalence

*For any* service plugin configuration, when a plugin is migrated from direct implementation to BaseServicePlugin composition, the validation and rendering results should be identical to the original implementation.

**Validates: Requirements 2.8**

### Property 5: Configuration Backward Compatibility

*For any* existing cluster configuration file, when processed by a migrated plugin, the configuration should be validated and rendered successfully with the same results as before migration.

**Validates: Requirements 2.9**

### Property 6: PathResolver Caching

*For any* cluster name and organization, when ResolveConfigPath is called twice with the same arguments, the second call should return the cached result without recomputing the path.

**Validates: Requirements 3.2**

### Property 7: Platform-Specific Path Normalization

*For any* cluster name and organization, when PathResolver resolves a path, the returned path should use the correct path separator for the current platform (backslash on Windows, forward slash on Unix).

**Validates: Requirements 3.3**

### Property 8: Organization-Based Path Structure

*For any* cluster name and organization, when PathResolver resolves a config path, the path should include the organization directory in the correct position: `<baseDir>/<organization>/.<cluster>-config.yaml`.

**Validates: Requirements 3.4**

### Property 9: Path Type Handling

*For any* path resolution operation, when given a relative path, PathResolver should resolve it relative to the base directory, and when given an absolute path, PathResolver should preserve it as absolute.

**Validates: Requirements 3.6**

### Property 10: PathResolver Thread Safety

*For any* number of concurrent goroutines calling PathResolver methods, the resolver should not panic, produce data races, or return corrupted results.

**Validates: Requirements 3.9**

### Property 11: Path Resolution Error Messages

*For any* invalid input to PathResolver (empty cluster name, invalid characters), the error message should clearly describe what input was invalid and why.

**Validates: Requirements 3.10**

### Property 12: Atomic File Writes

*For any* file write operation using FileSystem.WriteFile, if the write is interrupted or fails, the original file should remain unchanged (no partial writes or corruption).

**Validates: Requirements 4.4, 4.9**

### Property 13: File Operation Error Context

*For any* file operation that fails (read or write), the error message should include the file path and the operation that was attempted.

**Validates: Requirements 4.8**

### Property 14: Interface Removal Behavioral Equivalence

*For any* operation that previously used an interface (ConfigLoaderInterface, PathResolverInterface, ConfigCacheInterface), when the interface is removed and replaced with a concrete type, the operation should produce identical results.

**Validates: Requirements 5.8**

## Error Handling

### Error Categories

1. **Plugin Errors**
   - Invalid metadata (empty name, invalid version format)
   - Type assertion failures in validators/renderers
   - Validation failures from custom logic

2. **Path Resolution Errors**
   - Empty cluster name
   - Invalid characters in paths
   - Cache corruption
   - Platform-specific path issues

3. **File Operation Errors**
   - File not found
   - Permission denied
   - Disk full
   - Atomic write failures

4. **Migration Errors**
   - Plugin registration conflicts
   - Backward compatibility breaks
   - Test failures after migration

### Error Handling Patterns

```go
// Plugin validation errors
func (p *BaseServicePlugin) Validate(config interface{}) error {
    if err := p.validator(config); err != nil {
        return fmt.Errorf("validation failed for plugin %s: %w", p.Name(), err)
    }
    return nil
}

// Path resolution errors
func (pr *PathResolver) ResolveConfigPath(clusterName, organization string) (string, error) {
    if clusterName == "" {
        return "", fmt.Errorf("cluster name cannot be empty")
    }
    // ... resolution logic
    return path, nil
}

// File operation errors
func (fs *FileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
    if err := fs.writeAtomic(path, data, perm); err != nil {
        return fmt.Errorf("failed to write file %s: %w", path, err)
    }
    return nil
}
```

### Error Recovery

1. **Plugin Errors**: Log error, skip plugin, continue with others
2. **Path Resolution Errors**: Clear cache, retry once, then fail
3. **File Operation Errors**: Rollback atomic writes, preserve original file
4. **Migration Errors**: Revert to previous implementation, document issue

## Testing Strategy

### Dual Testing Approach

This phase uses both unit tests and property-based tests to ensure comprehensive coverage:

- **Unit tests**: Verify specific examples, edge cases, and error conditions
- **Property tests**: Verify universal properties across all inputs using gopter

Both approaches are complementary and necessary for complete validation.

### Unit Testing

Unit tests focus on:

1. **Specific Examples**
   - Creating a BaseServicePlugin with known metadata
   - Migrating a specific plugin (e.g., cert-manager)
   - Resolving a specific path (e.g., "my-cluster" in "opencenter" org)
   - Writing a specific file atomically

2. **Edge Cases**
   - Empty metadata fields
   - Nil validator/renderer functions
   - Empty cluster names
   - Special characters in paths
   - Symlinks in path resolution
   - Concurrent file writes to the same path

3. **Error Conditions**
   - Invalid plugin metadata
   - Type assertion failures
   - Path resolution with invalid inputs
   - File operations with permission errors
   - Disk full scenarios

### Property-Based Testing

Property tests use gopter to verify universal properties with randomly generated inputs:

1. **Base Plugin Properties**
   - Metadata accessibility (Property 1)
   - Custom logic injection (Property 2)
   - Plugin composition (Property 3)

2. **Migration Properties**
   - Behavioral equivalence (Property 4)
   - Backward compatibility (Property 5)

3. **Path Resolution Properties**
   - Caching behavior (Property 6)
   - Platform normalization (Property 7)
   - Organization-based structure (Property 8)
   - Path type handling (Property 9)
   - Thread safety (Property 10)
   - Error messages (Property 11)

4. **File Operations Properties**
   - Atomic writes (Property 12)
   - Error context (Property 13)

5. **Interface Removal Properties**
   - Behavioral equivalence (Property 14)

### Property Test Configuration

Each property test must:
- Run minimum 100 iterations (due to randomization)
- Reference its design document property in a comment
- Use tag format: `// Feature: phase-4-cleanup-optimization, Property N: <property_text>`

Example:

```go
// Feature: phase-4-cleanup-optimization, Property 1: Base Plugin Metadata Accessibility
func TestBasePluginMetadataAccessibility(t *testing.T) {
    properties := gopter.NewProperties(nil)
    
    properties.Property("metadata fields are accessible", prop.ForAll(
        func(name, version, description, pluginType, author, license string) bool {
            metadata := PluginMetadata{
                Name:        name,
                Version:     version,
                Description: description,
                Type:        pluginType,
                Author:      author,
                License:     license,
            }
            
            plugin := NewBasePlugin(metadata)
            
            return plugin.Name() == name &&
                   plugin.Version() == version &&
                   plugin.Description() == description &&
                   plugin.Type() == pluginType &&
                   plugin.Author() == author &&
                   plugin.License() == license
        },
        gen.AnyString(),
        gen.AnyString(),
        gen.AnyString(),
        gen.AnyString(),
        gen.AnyString(),
        gen.AnyString(),
    ))
    
    properties.TestingRun(t, gopter.ConsoleReporter(false))
}
```

### Integration Testing

Integration tests verify:

1. **Plugin Registration**: All migrated plugins register correctly
2. **End-to-End Workflows**: Config load → validate → render → write
3. **Cross-Component**: PathResolver + FileSystem + ConfigurationManager
4. **Backward Compatibility**: Old configs work with new code

### Test Coverage Goals

- Overall coverage: >85%
- Critical paths: 100% (BaseServicePlugin, PathResolver, FileSystem)
- Error handling: >90%
- Edge cases: >80%

### Testing Tools

- **Unit tests**: Standard Go testing package
- **Property tests**: gopter library
- **Benchmarks**: Go benchmark framework (for measuring improvements)
- **Coverage**: `go test -cover ./...`
- **Race detection**: `go test -race ./...`
