# `openCenter cluster schema` - Export Cluster JSON Schema

## Synopsis
```bash
openCenter cluster schema [OPTIONS]
```

## Description

Export the JSON schema for openCenter cluster configuration. The schema includes comprehensive validation rules, constraints, and documentation for all configuration sections. It can be used for IDE integration, validation, and documentation purposes.

The schema is automatically generated from Go struct definitions and includes all configuration options, data types, validation rules, and descriptions.

## Options

### `--out <path>`
- **Description**: Output file path (if not specified, prints to stdout)
- **Type**: String
- **Default**: stdout

### `--pretty`
- **Description**: Pretty print JSON schema with indentation
- **Type**: Boolean
- **Default**: `true`

### `--version`
- **Description**: Show schema version and exit
- **Type**: Boolean
- **Default**: `false`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Examples

### Print schema to stdout
```bash
openCenter cluster schema
```

### Save schema to file with pretty formatting
```bash
openCenter cluster schema --out schema/cluster.schema.json --pretty
```

### Save schema without formatting
```bash
openCenter cluster schema --out cluster.schema.json --pretty=false
```

### Show schema version
```bash
openCenter cluster schema --version
```
Output:
```
Schema version: 1.0.0
```

### Pipe to jq for processing
```bash
openCenter cluster schema | jq '.properties.opencenter'
```

### Save to project schema directory
```bash
openCenter cluster schema --out schema/cluster.schema.json
```

## Output

### Pretty Formatted (Default)

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "title": "openCenter Cluster Configuration",
  "description": "Complete configuration schema for openCenter clusters",
  "properties": {
    "opencenter": {
      "type": "object",
      "properties": {
        "meta": {
          "type": "object",
          "properties": {
            "name": {
              "type": "string",
              "description": "Cluster name"
            },
            "env": {
              "type": "string",
              "enum": ["dev", "staging", "prod"],
              "description": "Environment designation"
            }
          }
        }
      }
    }
  }
}
```

### Compact Format (--pretty=false)

```json
{"$schema":"http://json-schema.org/draft-07/schema#","type":"object","title":"openCenter Cluster Configuration",...}
```

## Schema Structure

The exported schema includes:

### Metadata
- Schema version
- Title and description
- JSON Schema draft version

### Configuration Sections
- `opencenter` - Core openCenter configuration
- `iac` - Infrastructure as Code configuration
- `secrets` - Secrets management configuration

### Validation Rules
- Required fields
- Data types
- Enum values
- Pattern matching
- Min/max constraints
- Custom validation rules

### Documentation
- Field descriptions
- Examples
- Default values
- Deprecation notices

## IDE Integration

### VS Code

1. Save schema to file:
```bash
openCenter cluster schema --out schema/cluster.schema.json
```

2. Configure in `.vscode/settings.json`:
```json
{
  "yaml.schemas": {
    "./schema/cluster.schema.json": "*.opencenter.yaml"
  }
}
```

### JetBrains IDEs (IntelliJ, PyCharm, etc.)

1. Save schema to file
2. Go to Settings → Languages & Frameworks → Schemas and DTDs → JSON Schema Mappings
3. Add new mapping:
   - Name: openCenter Cluster
   - Schema file: `schema/cluster.schema.json`
   - File pattern: `*.opencenter.yaml`

### Vim/Neovim with coc.nvim

Add to `coc-settings.json`:
```json
{
  "yaml.schemas": {
    "./schema/cluster.schema.json": "*.opencenter.yaml"
  }
}
```

## Schema Validation

Use the schema to validate configuration files:

### With ajv-cli
```bash
npm install -g ajv-cli
openCenter cluster schema --out schema.json
ajv validate -s schema.json -d cluster-config.yaml
```

### With Python jsonschema
```python
import json
import yaml
from jsonschema import validate

# Load schema
with open('schema.json') as f:
    schema = json.load(f)

# Load config
with open('cluster-config.yaml') as f:
    config = yaml.safe_load(f)

# Validate
validate(instance=config, schema=schema)
```

## Schema Versioning

The schema follows semantic versioning:

- **Major version**: Breaking changes to schema structure
- **Minor version**: New fields or non-breaking changes
- **Patch version**: Documentation updates or bug fixes

Check schema version:
```bash
openCenter cluster schema --version
```

## Use Cases

### Documentation Generation
```bash
# Generate schema
openCenter cluster schema --out docs/schema.json

# Generate documentation from schema
npx @adobe/jsonschema2md -d docs/schema.json -o docs/reference/
```

### CI/CD Validation
```bash
# In CI pipeline
openCenter cluster schema --out /tmp/schema.json
ajv validate -s /tmp/schema.json -d config/*.yaml
```

### IDE Autocomplete
Save schema to project and configure IDE for autocomplete and validation.

### Configuration Testing
```bash
# Validate test configurations
openCenter cluster schema --out schema.json
for config in testdata/config/*.yaml; do
  ajv validate -s schema.json -d "$config"
done
```

## Notes

- Schema is generated from Go struct definitions
- The schema includes all configuration options and validation rules
- Use `--pretty` for human-readable output
- Schema can be used for IDE integration and validation
- The schema follows JSON Schema Draft 07 specification
- Schema version is independent of openCenter version
- Generated schema includes comprehensive documentation
- Use the schema for configuration validation in CI/CD pipelines

## See Also

- `openCenter cluster validate` - Validate cluster configuration
- `openCenter cluster init` - Initialize cluster with schema defaults
- [JSON Schema Documentation](https://json-schema.org/)
