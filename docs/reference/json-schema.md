# JSON Schema Documentation

## Table of Contents

- [Overview](#overview)
- [Schema Versions](#schema-versions)
- [Schema Location](#schema-location)
- [IDE Integration](#ide-integration)
- [Schema Structure](#schema-structure)
- [Validation Rules](#validation-rules)
- [Using the Schema](#using-the-schema)
- [Schema Generation](#schema-generation)
- [Versioning Strategy](#versioning-strategy)
- [Migration Between Versions](#migration-between-versions)
- [Troubleshooting](#troubleshooting)
- [Related Documentation](#related-documentation)

## Overview

opencenter provides JSON schemas for cluster configuration files to enable IDE autocomplete, validation, and inline documentation. The schemas are automatically generated from Go struct definitions and include comprehensive validation rules, field descriptions, and type constraints.

**Key Benefits:**
- **IDE Autocomplete**: Intelligent suggestions for configuration keys and values
- **Real-time Validation**: Catch errors before deployment
- **Inline Documentation**: Hover tooltips with field descriptions
- **Type Safety**: Enforce correct data types and formats
- **Version Control**: Track schema changes alongside code

## Schema Versions

opencenter supports multiple schema versions to maintain backward compatibility during transitions:

### Version 1.0 (Current Production)

**Status**: Stable, production-ready  
**Schema ID**: `https://opencenter.io/schemas/cluster-config-v1.0.json`  
**Location**: `schema/cluster.schema.json`

Version 1.0 is the current production schema used by all existing clusters. It includes:
- Flat configuration structure
- OpenStack-focused provider settings
- Kubespray deployment method
- Basic service configuration

**Example Configuration:**
```yaml
opencenter:
  meta:
    name: my-cluster
    organization: myorg
  cluster:
    cluster_name: my-cluster
    kubernetes:
      version: 1.31.4
      master_count: 3
      worker_count: 2
```

### Version 2.0 (In Development)

**Status**: Development, not yet production-ready  
**Schema ID**: `https://opencenter.io/schemas/cluster-config-v2.0.json`  
**Location**: `schema/cluster-v2.schema.json`

Version 2.0 introduces a hierarchical configuration model with:
- Five-domain structure (Meta, Cluster, Infrastructure, Deployment, Services)
- Provider isolation under `infrastructure.cloud.<provider>`
- Deployment method abstraction
- Reference resolution with `${path.to.value}` syntax
- Provider-region default registry
- Kamaji hosted control plane support
- CSI plugin selection (similar to CNI)

**Example Configuration:**
```yaml
schema_version: "2.0"

opencenter:
  meta:
    name: my-cluster
    organization: myorg
    env: prod
    region: us-east-1
  
  infrastructure:
    provider: openstack
    networking:
      subnet_nodes: 10.0.0.0/24
      vrrp_ip: 10.0.0.100
    compute:
      master_count: 3
      worker_count: 2
    cloud:
      openstack:
        auth_url: https://keystone.example.com/v3/
        region: RegionOne
  
  cluster:
    cluster_name: my-cluster
    kubernetes:
      version: 1.31.4
      subnet_pods: 10.42.0.0/16
      subnet_services: 10.43.0.0/16
  
  deployment:
    method: kubespray
    kubespray:
      version: 2.24.0
```

## Schema Location

Schemas are stored in the `schema/` directory at the project root:

```
schema/
├── cluster.schema.json          # v1.0 schema (current)
└── cluster-v2.schema.json       # v2.0 schema (future)
```

**Published URLs:**
- v1.0: `https://opencenter.io/schemas/cluster-config-v1.0.json`
- v2.0: `https://opencenter.io/schemas/cluster-config-v2.0.json`

## IDE Integration

### Visual Studio Code

**Automatic Setup:**
```bash
opencenter config ide --ide=vscode
```

**Manual Setup:**

Add to `.vscode/settings.json`:
```json
{
  "yaml.schemas": {
    "./schema/cluster.schema.json": [
      "**/clusters/**/*.yaml",
      "**/clusters/**/*-config.yaml",
      "**/.opencenter.yaml"
    ],
    "./schema/cluster-v2.schema.json": [
      "**/clusters/**/*-v2.yaml",
      "**/v2/**/*.yaml"
    ]
  },
  "yaml.validate": true,
  "yaml.completion": true,
  "yaml.hover": true
}
```

**Required Extension:**
- [YAML Language Support by Red Hat](https://marketplace.visualstudio.com/items?itemName=redhat.vscode-yaml)

### JetBrains IDEs

**Setup Steps:**

1. Open **Settings** → **Languages & Frameworks** → **Schemas and DTDs** → **JSON Schema Mappings**

2. Add v1.0 schema:
   - Name: `opencenter v1.0`
   - Schema file: `schema/cluster.schema.json`
   - File patterns: `**/clusters/**/*.yaml`, `**/*-config.yaml`

3. Add v2.0 schema:
   - Name: `opencenter v2.0`
   - Schema file: `schema/cluster-v2.schema.json`
   - File patterns: `**/v2/**/*.yaml`, `**/*-v2.yaml`

### Vim/Neovim

**Using coc.nvim:**

Add to `coc-settings.json`:
```json
{
  "yaml.schemas": {
    "./schema/cluster.schema.json": [
      "**/clusters/**/*.yaml",
      "**/clusters/**/*-config.yaml"
    ],
    "./schema/cluster-v2.schema.json": [
      "**/v2/**/*.yaml"
    ]
  }
}
```

**Using nvim-lspconfig:**

Add to `init.lua`:
```lua
require'lspconfig'.yamlls.setup{
  settings = {
    yaml = {
      schemas = {
        ["./schema/cluster.schema.json"] = {
          "**/clusters/**/*.yaml",
          "**/clusters/**/*-config.yaml"
        },
        ["./schema/cluster-v2.schema.json"] = {
          "**/v2/**/*.yaml"
        }
      }
    }
  }
}
```

### Emacs

Add to `init.el`:
```elisp
(use-package lsp-mode
  :hook (yaml-mode . lsp)
  :config
  (setq lsp-yaml-schemas
        '(:v1 "./schema/cluster.schema.json"
          :v2 "./schema/cluster-v2.schema.json")))
```

## Schema Structure

### Schema Metadata

Every schema includes metadata for identification and versioning:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://opencenter.io/schemas/cluster-config-v1.0.json",
  "title": "OpenCenter Cluster Configuration Schema v1.0",
  "description": "JSON schema for OpenCenter cluster configuration files",
  "schemaVersion": "1.0",
  "type": "object"
}
```

**Fields:**
- `$schema`: JSON Schema specification version
- `$id`: Unique identifier for this schema
- `title`: Human-readable schema name
- `description`: Schema purpose and scope
- `schemaVersion`: opencenter schema version
- `type`: Root type (always "object")

### Property Definitions

Each configuration field includes:

```json
{
  "properties": {
    "cluster_name": {
      "type": "string",
      "description": "Unique cluster identifier",
      "pattern": "^[a-z0-9][a-z0-9-]*[a-z0-9]$",
      "minLength": 3,
      "maxLength": 63,
      "examples": ["my-cluster", "prod-k8s-01"]
    }
  }
}
```

**Common Attributes:**
- `type`: Data type (string, integer, boolean, object, array)
- `description`: Field purpose and usage
- `pattern`: Regular expression for validation
- `minLength`/`maxLength`: String length constraints
- `minimum`/`maximum`: Numeric range constraints
- `enum`: List of allowed values
- `examples`: Example values for documentation
- `default`: Default value if not specified

### Required Fields

Required fields are specified at each object level:

```json
{
  "type": "object",
  "properties": {
    "cluster_name": { "type": "string" },
    "kubernetes": { "type": "object" }
  },
  "required": ["cluster_name", "kubernetes"]
}
```

### Nested Objects

Complex structures use nested object definitions:

```json
{
  "properties": {
    "kubernetes": {
      "type": "object",
      "properties": {
        "version": { "type": "string" },
        "master_count": { "type": "integer" }
      },
      "required": ["version", "master_count"]
    }
  }
}
```

## Validation Rules

### Type Validation

**String Types:**
```json
{
  "type": "string",
  "minLength": 1,
  "maxLength": 255
}
```

**Integer Types:**
```json
{
  "type": "integer",
  "minimum": 1,
  "maximum": 100
}
```

**Boolean Types:**
```json
{
  "type": "boolean",
  "default": false
}
```

**Array Types:**
```json
{
  "type": "array",
  "items": { "type": "string" },
  "minItems": 1,
  "uniqueItems": true
}
```

### Pattern Validation

**CIDR Blocks:**
```json
{
  "type": "string",
  "pattern": "^([0-9]{1,3}\\.){3}[0-9]{1,3}/[0-9]{1,2}$",
  "examples": ["10.0.0.0/24", "192.168.1.0/24"]
}
```

**IPv4 Addresses:**
```json
{
  "type": "string",
  "pattern": "^([0-9]{1,3}\\.){3}[0-9]{1,3}$",
  "examples": ["10.0.0.1", "192.168.1.100"]
}
```

**DNS Names:**
```json
{
  "type": "string",
  "format": "hostname",
  "pattern": "^[a-z0-9][a-z0-9-]*[a-z0-9]$",
  "examples": ["cluster.example.com", "k8s-prod"]
}
```

**UUIDs:**
```json
{
  "type": "string",
  "pattern": "^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$",
  "examples": ["550e8400-e29b-41d4-a716-446655440000"]
}
```

**Semantic Versions:**
```json
{
  "type": "string",
  "pattern": "^[0-9]+\\.[0-9]+\\.[0-9]+$",
  "examples": ["1.31.4", "2.24.0"]
}
```

### Enum Validation

**Provider Types:**
```json
{
  "type": "string",
  "enum": ["openstack", "aws", "gcp", "azure", "baremetal", "vsphere"],
  "description": "Infrastructure provider"
}
```

**Deployment Methods:**
```json
{
  "type": "string",
  "enum": ["kubespray", "talos", "kamaji", "eks", "gke", "aks"],
  "description": "Kubernetes deployment method"
}
```

### Conditional Validation

**oneOf (Mutually Exclusive):**
```json
{
  "oneOf": [
    {
      "properties": {
        "calico": { "properties": { "enabled": { "const": true } } }
      }
    },
    {
      "properties": {
        "cilium": { "properties": { "enabled": { "const": true } } }
      }
    }
  ]
}
```

**if/then/else (Conditional Requirements):**
```json
{
  "if": {
    "properties": { "provider": { "const": "openstack" } }
  },
  "then": {
    "required": ["cloud"],
    "properties": {
      "cloud": {
        "required": ["openstack"]
      }
    }
  }
}
```

## Using the Schema

### Generating the Schema

**Generate v1.0 schema:**
```bash
opencenter cluster schema --out schema/cluster.schema.json
```

**Generate v2.0 schema:**
```bash
opencenter cluster schema --version 2.0 --out schema/cluster-v2.schema.json
```

**Generate with pretty printing:**
```bash
opencenter cluster schema --out schema/cluster.schema.json --pretty
```

### Validating Against Schema

**Using opencenter CLI:**
```bash
# Validate v1.0 configuration
opencenter cluster validate my-cluster

# Validate v2.0 configuration
opencenter cluster validate my-cluster --schema-version 2.0
```

**Using external tools:**

```bash
# Using ajv-cli
npm install -g ajv-cli
ajv validate -s schema/cluster.schema.json -d my-cluster-config.yaml

# Using check-jsonschema
pip install check-jsonschema
check-jsonschema --schemafile schema/cluster.schema.json my-cluster-config.yaml
```

### Programmatic Usage

**JavaScript/TypeScript:**
```javascript
const Ajv = require('ajv');
const schema = require('./schema/cluster.schema.json');
const config = require('./my-cluster-config.json');

const ajv = new Ajv();
const validate = ajv.compile(schema);
const valid = validate(config);

if (!valid) {
  console.error(validate.errors);
}
```

**Python:**
```python
import jsonschema
import yaml
import json

# Load schema
with open('schema/cluster.schema.json') as f:
    schema = json.load(f)

# Load config
with open('my-cluster-config.yaml') as f:
    config = yaml.safe_load(f)

# Validate
try:
    jsonschema.validate(config, schema)
    print("Configuration is valid")
except jsonschema.ValidationError as e:
    print(f"Validation error: {e.message}")
```

## Schema Generation

### Automatic Generation

Schemas are automatically generated from Go struct definitions using the `invopop/jsonschema` library:

```go
type Config struct {
    ClusterName string `yaml:"cluster_name" json:"cluster_name" jsonschema:"required,description=Unique cluster identifier,pattern=^[a-z0-9][a-z0-9-]*[a-z0-9]$"`
}
```

**Supported Tags:**
- `jsonschema:"required"`: Mark field as required
- `jsonschema:"description=..."`: Field description
- `jsonschema:"pattern=..."`: Regex pattern
- `jsonschema:"enum=..."`: Allowed values
- `jsonschema:"minimum=..."`: Minimum value
- `jsonschema:"maximum=..."`: Maximum value
- `jsonschema:"default=..."`: Default value

### Manual Schema Updates

For complex validation rules not expressible in struct tags, manually edit the generated schema:

1. Generate base schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

2. Edit `schema/cluster.schema.json` to add custom rules

3. Commit both the generator code and manual edits

4. Document manual changes in comments

### Schema Versioning

When making breaking changes:

1. Create new schema version:
   ```bash
   cp schema/cluster.schema.json schema/cluster-v2.schema.json
   ```

2. Update `$id` and `schemaVersion` in new schema

3. Update schema generator to support both versions

4. Provide migration path in documentation

## Versioning Strategy

### Semantic Versioning

Schema versions follow semantic versioning:

- **Major version** (1.0 → 2.0): Breaking changes, incompatible structure
- **Minor version** (1.0 → 1.1): Backward-compatible additions
- **Patch version** (1.0.0 → 1.0.1): Bug fixes, clarifications

### Version Detection

Configurations specify their schema version:

**v1.0 (implicit):**
```yaml
opencenter:
  meta:
    name: my-cluster
```

**v2.0 (explicit):**
```yaml
schema_version: "2.0"

opencenter:
  meta:
    name: my-cluster
```

### Deprecation Policy

1. **Announcement**: Deprecation announced in release notes
2. **Warning Period**: 2 major releases with deprecation warnings
3. **Removal**: Deprecated features removed in 3rd major release

**Example Timeline:**
- v1.0: Current schema
- v2.0: New schema introduced, v1.0 deprecated with warnings
- v3.0: v1.0 support removed

## Migration Between Versions

### v1.0 to v2.0 Migration

**Automated Migration:**
```bash
opencenter cluster migrate-config \
  --input my-cluster-v1.yaml \
  --output my-cluster-v2.yaml
```

**Manual Migration Steps:**

1. Add schema version:
   ```yaml
   schema_version: "2.0"
   ```

2. Restructure configuration domains:
   ```yaml
   # v1.0
   opencenter:
     cluster:
       kubernetes:
         flavor_master: gp.0.4.4
   
   # v2.0
   opencenter:
     infrastructure:
       compute:
         flavor_master: gp.0.4.4
   ```

3. Move VRRP IP:
   ```yaml
   # v1.0
   opencenter:
     cluster:
       networking:
         vrrp_ip: 10.0.0.100
   
   # v2.0
   opencenter:
     infrastructure:
       networking:
         vrrp_ip: 10.0.0.100
   ```

4. Isolate provider settings:
   ```yaml
   # v1.0
   opencenter:
     infrastructure:
       auth_url: https://keystone.example.com/v3/
   
   # v2.0
   opencenter:
     infrastructure:
       cloud:
         openstack:
           auth_url: https://keystone.example.com/v3/
   ```

5. Validate migrated configuration:
   ```bash
   opencenter cluster validate my-cluster --schema-version 2.0
   ```

### Migration Report

The migration tool generates a report:

```
Migration Report
================

Moved Fields:
  cluster.networking.vrrp_ip → infrastructure.networking.vrrp_ip
  cluster.kubernetes.flavor_master → infrastructure.compute.flavor_master
  opencenter.storage → infrastructure.storage

Applied Defaults:
  infrastructure.networking.dns_nameservers: [8.8.8.8, 8.8.4.4]
  infrastructure.compute.master_count: 3

Warnings:
  - VRRP IP moved from deprecated location
  - Storage configuration restructured
```

## Troubleshooting

### Schema Not Loading in IDE

**Symptoms:**
- No autocomplete suggestions
- No validation errors shown
- Hover tooltips don't appear

**Solutions:**

1. Verify schema file exists:
   ```bash
   ls -la schema/cluster.schema.json
   ```

2. Check schema is valid JSON:
   ```bash
   jq . schema/cluster.schema.json
   ```

3. Regenerate schema:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

4. Restart IDE or reload window

5. Check IDE logs for errors:
   - VS Code: View → Output → YAML Support
   - JetBrains: Help → Show Log

### Validation Errors on Valid Config

**Symptoms:**
- IDE shows errors for valid configuration
- Errors don't match `opencenter cluster validate` output

**Solutions:**

1. Check schema version matches opencenter version:
   ```bash
   opencenter cluster schema --version
   opencenter version
   ```

2. Regenerate schema after updating opencenter:
   ```bash
   opencenter cluster schema --out schema/cluster.schema.json
   ```

3. Verify file path matches schema patterns

4. Clear IDE cache and restart

### Schema Generation Fails

**Symptoms:**
- `opencenter cluster schema` command fails
- Generated schema is incomplete

**Solutions:**

1. Check Go struct tags are valid:
   ```go
   // Good
   ClusterName string `jsonschema:"required,description=Cluster name"`
   
   // Bad (missing comma)
   ClusterName string `jsonschema:"required description=Cluster name"`
   ```

2. Verify all referenced types are exported:
   ```go
   // Good
   type Config struct {
       Kubernetes KubernetesConfig
   }
   
   // Bad (unexported type)
   type Config struct {
       Kubernetes kubernetesConfig
   }
   ```

3. Check for circular type references

4. Run with debug logging:
   ```bash
   OPENCENTER_DEBUG=1 opencenter cluster schema --out schema/cluster.schema.json
   ```

### Performance Issues

**Symptoms:**
- IDE becomes slow with large configs
- Validation takes too long

**Solutions:**

1. Disable real-time validation for large files

2. Split large configurations into multiple files

3. Increase IDE memory limits:
   - VS Code: `"files.maxMemoryForLargeFilesMB": 4096`
   - JetBrains: Help → Edit Custom VM Options → `-Xmx4096m`

4. Use schema validation only on save, not on type

## Related Documentation

- [IDE Integration Guide](../how-to/ide-integration.md) - Complete IDE setup instructions
- [Configuration Reference](configuration.md) - Field-by-field configuration documentation
- [Migration Guide](../cluster-config/migration-guide.md) - v1 to v2 migration instructions
- [CLI Commands](cli-commands.md) - All opencenter commands including schema generation

## External Resources

- [JSON Schema Specification](https://json-schema.org/) - Official JSON Schema documentation
- [JSON Schema Validator](https://www.jsonschemavalidator.net/) - Online schema validation tool
- [Understanding JSON Schema](https://json-schema.org/understanding-json-schema/) - Comprehensive guide
- [YAML Language Server](https://github.com/redhat-developer/yaml-language-server) - LSP implementation for YAML
