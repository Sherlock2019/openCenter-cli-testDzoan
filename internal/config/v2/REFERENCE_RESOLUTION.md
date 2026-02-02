# Reference Resolution in v2 Configuration

## Overview

The v2 configuration loader supports three types of references that can be used in configuration values:

1. **Environment Variable References**: `${env:VARIABLE_NAME}`
2. **File References**: `${file:/path/to/file}`
3. **Config Path References**: `${ref:path.to.value}` (planned for future implementation)

## Environment Variable References

Environment variable references allow you to inject values from the environment into your configuration.

### Syntax

```yaml
${env:VARIABLE_NAME}
```

### Examples

```yaml
schema_version: "2.0"
opencenter:
  meta:
    name: "${env:CLUSTER_NAME}"
    organization: "${env:ORG_NAME}"
    env: "production"
    region: "ord1"
  gitops:
    git_url: "${env:GIT_REPO_URL}"
```

### Behavior

- If the environment variable is not set, an error is returned
- Empty environment variable values are treated as errors
- Environment variable values are cached for performance
- Multiple references to the same variable use the cached value

### Usage

```bash
# Set environment variables
export CLUSTER_NAME="prod-cluster"
export ORG_NAME="my-org"
export GIT_REPO_URL="git@github.com:my-org/gitops.git"

# Load configuration (references will be resolved)
opencenter cluster validate prod-cluster
```

## File References

File references allow you to inject file contents into your configuration.

### Syntax

```yaml
${file:/path/to/file}
```

### Examples

```yaml
schema_version: "2.0"
secrets:
  global:
    aws_access_key: "${file:/secrets/aws-access-key.txt}"
    aws_secret_key: "${file:/secrets/aws-secret-key.txt}"
opencenter:
  meta:
    name: "my-cluster"
    organization: "my-org"
    env: "production"
    region: "ord1"
```

### Behavior

- File contents are read and trimmed (leading/trailing whitespace removed)
- If the file doesn't exist or can't be read, an error is returned
- File contents are cached for performance
- Multiple references to the same file use the cached value

### Usage

```bash
# Create secret files
echo "AKIAIOSFODNN7EXAMPLE" > /secrets/aws-access-key.txt
echo "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY" > /secrets/aws-secret-key.txt

# Load configuration (file contents will be injected)
opencenter cluster validate my-cluster
```

## Config Path References (Planned)

Config path references allow you to reference other values within the same configuration.

### Syntax

```yaml
${ref:path.to.value}
```

### Status

This feature is planned but not yet fully implemented. The resolver detects circular references but does not yet support path lookup within the configuration structure.

### Future Example

```yaml
schema_version: "2.0"
opencenter:
  meta:
    name: "my-cluster"
    organization: "my-org"
    env: "production"
    region: "ord1"
  cluster:
    name: "${ref:opencenter.meta.name}"  # References meta.name
    fqdn: "${ref:opencenter.meta.name}.${ref:opencenter.meta.region}.example.com"
```

## Multiple References

You can use multiple references in a single string value:

```yaml
opencenter:
  meta:
    name: "${env:CLUSTER_NAME}"
    organization: "${env:ORG_NAME}"
  cluster:
    fqdn: "${env:CLUSTER_NAME}.${env:REGION}.example.com"
```

## Caching

The reference resolver caches resolved values to improve performance:

- Environment variables are cached with key `env:VARIABLE_NAME`
- File contents are cached with key `file:/path/to/file`
- Cache is per-resolver instance (one per config load)

## Error Handling

The resolver provides clear error messages for common issues:

```
Error: environment variable ${env:MISSING_VAR} is not set or empty
Error: failed to read file ${file:/nonexistent/file.txt}: no such file or directory
Error: circular reference detected: ${ref:opencenter.meta.name} at path 'opencenter.cluster.name'
```

## Circular Reference Detection

The resolver detects circular references in config path references:

```yaml
# This would cause an error:
opencenter:
  meta:
    name: "${ref:opencenter.cluster.name}"  # References cluster.name
  cluster:
    name: "${ref:opencenter.meta.name}"     # References meta.name (circular!)
```

Error: `circular reference detected: ${ref:opencenter.meta.name} at path 'opencenter.cluster.name'`

## Maximum Depth Protection

The resolver limits recursion depth to prevent infinite loops:

- Maximum depth: 10 levels
- Applies to nested structures (maps, slices, structs)
- Prevents stack overflow from deeply nested configurations

## Implementation Details

### Resolution Order

1. Parse configuration YAML
2. Normalize field values
3. **Resolve references** (this step)
4. Apply defaults
5. Validate configuration
6. Freeze (mark as immutable)

### Type Support

The resolver works with:
- String fields (primary use case)
- Map values (ServiceMap, etc.)
- Slice elements
- Nested structs

### Performance

- Caching reduces redundant file reads and environment variable lookups
- Reflection-based traversal is efficient for typical configuration sizes
- Maximum depth limit prevents performance degradation from pathological cases

## Testing

The reference resolver is tested with:
- Unit tests for each reference type
- Property-based tests for correctness properties
- Integration tests with the full config loader pipeline

See `resolver_test.go` and `resolver_property_test.go` for examples.

## Future Enhancements

1. **Config Path References**: Full implementation of `${ref:path.to.value}`
2. **Default Values**: Support `${env:VAR:-default}` syntax
3. **Transformations**: Support `${env:VAR|upper}` for value transformations
4. **Validation**: Validate reference syntax at parse time
5. **IDE Support**: Provide schema hints for reference completion
