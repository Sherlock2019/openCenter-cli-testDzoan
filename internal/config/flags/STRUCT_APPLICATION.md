# Flag Integration Struct Application

## Overview

The Flag Integration system now supports applying configuration flags directly to Go struct fields in addition to map-based configuration. This provides type-safe configuration updates with automatic type conversion and validation.

## Features

### Type-Safe Field Updates

The system automatically converts flag values to the appropriate Go types:

- **String fields**: Accepts strings, numbers, and booleans (converted to string representation)
- **Integer fields**: Accepts integers and strings (parsed as integers)
- **Boolean fields**: Accepts booleans, strings ("true"/"false"), and integers (0=false, non-zero=true)
- **Slice fields**: Accepts slices and single values (wrapped in a slice)
- **Map fields**: Accepts maps with automatic key/value type conversion
- **Nested structs**: Supports dot-notation paths to navigate nested structures

### Automatic Field Path Conversion

Flag names are automatically converted to struct field paths:

```go
// Flag name -> Struct field path
"name"                          -> "Name"
"nested.host"                   -> "Nested.Host"
"infrastructure.cluster.name"   -> "Infrastructure.Cluster.Name"
"nested_value"                  -> "NestedValue"  // Handles underscores
```

### Nil Pointer Initialization

The system automatically initializes nil pointers when setting nested fields:

```go
type Config struct {
    Nested *NestedConfig
}

// Setting "nested.host" will automatically initialize Nested if it's nil
applyToStruct(config, "nested.host", "localhost")
```

## Usage Examples

### Basic Field Updates

```go
integration, _ := NewCLIIntegration()

config := &Config{}

// Set string field
integration.applyToStruct(config, "name", "my-cluster")

// Set integer field
integration.applyToStruct(config, "count", 42)

// Set boolean field
integration.applyToStruct(config, "enabled", true)
```

### Nested Field Updates

```go
type Config struct {
    Infrastructure struct {
        Cluster struct {
            Name string
            Size int
        }
    }
}

config := &Config{}

// Use dot notation to set nested fields
integration.applyToStruct(config, "infrastructure.cluster.name", "prod-cluster")
integration.applyToStruct(config, "infrastructure.cluster.size", 10)
```

### Slice Operations

```go
type Config struct {
    Tags []string
}

config := &Config{}

// Set entire slice
integration.applyToStruct(config, "tags", []string{"prod", "us-east"})

// Append to slice
arrayOp := &ArrayOperationFlag{
    Path:      "tags",
    Operation: "append",
    Value:     "new-tag",
}
integration.applyArrayOperationToStruct(config, arrayOp)

// Insert at index
arrayOp = &ArrayOperationFlag{
    Path:      "tags",
    Operation: "insert",
    Index:     1,
    Value:     "middle-tag",
}
integration.applyArrayOperationToStruct(config, arrayOp)

// Remove by index
arrayOp = &ArrayOperationFlag{
    Path:      "tags",
    Operation: "remove",
    Index:     0,
}
integration.applyArrayOperationToStruct(config, arrayOp)
```

### Map Operations

```go
type Config struct {
    Metadata map[string]string
}

config := &Config{
    Metadata: make(map[string]string),
}

// Set single key-value pair
mapOp := &MapFlag{
    Path:      "metadata",
    Operation: "set",
    Key:       "env",
    Value:     "production",
}
integration.applyMapOperationToStruct(config, mapOp)

// Merge multiple key-value pairs
mapOp = &MapFlag{
    Path:      "metadata",
    Operation: "merge",
    Value: map[string]interface{}{
        "region": "us-east-1",
        "team":   "platform",
    },
}
integration.applyMapOperationToStruct(config, mapOp)

// Remove a key
mapOp = &MapFlag{
    Path:      "metadata",
    Operation: "remove",
    Key:       "env",
}
integration.applyMapOperationToStruct(config, mapOp)
```

### Type Conversion Examples

```go
config := &Config{}

// String to int conversion
integration.applyToStruct(config, "count", "100")  // Parsed as integer

// Int to string conversion
integration.applyToStruct(config, "name", 42)  // Converted to "42"

// String to bool conversion
integration.applyToStruct(config, "enabled", "true")  // Parsed as boolean

// Int to bool conversion
integration.applyToStruct(config, "enabled", 1)  // Converted to true
```

## Integration with Flag Parsing

The struct application system integrates seamlessly with the flag parsing system:

```go
integration, _ := NewCLIIntegration()

config := &MyConfig{}
configMap := make(map[string]interface{})

// Parse command-line flags
args := []string{
    "--name=my-cluster",
    "--count=42",
    "--enabled=true",
}

parsed, _ := integration.parser.ParseFlags(args)

// Apply to both struct and map
integration.applyFlags(parsed, config, configMap)

// Both config struct and configMap are now updated
```

## Supported Flag Types

The following flag types support struct application:

### JSON Flags

```go
jsonFlag := JSONFlag{
    Path:  "infrastructure",
    Value: map[string]interface{}{
        "provider": "openstack",
        "region":   "us-east-1",
    },
}

integration.applyJSONFlag(jsonFlag, config, configMap)
```

### YAML Flags

```go
yamlFlag := YAMLFlag{
    Path:  "services",
    Value: map[string]interface{}{
        "enabled": true,
        "count":   3,
    },
}

integration.applyYAMLFlag(yamlFlag, config, configMap)
```

### Array Operations

```go
arrayOp := ArrayOperationFlag{
    Path:      "tags",
    Operation: "append",
    Value:     "new-tag",
}

integration.applyArrayOperation(arrayOp, config, configMap)
```

### Map Operations

```go
mapOp := MapFlag{
    Path:      "metadata",
    Operation: "set",
    Key:       "env",
    Value:     "production",
}

integration.applyMapOperation(mapOp, config, configMap)
```

## Error Handling

The system provides detailed error messages for common issues:

```go
// Field not found
err := integration.applyToStruct(config, "nonexistent", "value")
// Error: field 'Nonexistent' not found in struct 'Config'

// Type conversion error
err = integration.applyToStruct(config, "count", "not-a-number")
// Error: cannot convert string 'not-a-number' to int

// Invalid nested path
err = integration.applyToStruct(config, "name.invalid", "value")
// Error: cannot navigate through field 'name' of type string

// Array index out of range
arrayOp := &ArrayOperationFlag{
    Path:      "tags",
    Operation: "remove",
    Index:     100,
}
err = integration.applyArrayOperationToStruct(config, arrayOp)
// Error: index 100 out of range for slice of length 3
```

## Implementation Details

### Field Navigation

The `navigateToField` function traverses struct fields using reflection:

1. Splits the path by dots
2. For each path segment:
   - Finds the field by name or YAML tag
   - Handles embedded/anonymous fields
   - Initializes nil pointers automatically
   - Navigates into nested structs and maps

### Type Conversion

The `setFieldValueTyped` function handles type conversion:

1. Determines the target field type
2. Converts the value to the appropriate type
3. Handles special cases (slices, maps, pointers)
4. Returns detailed error messages for conversion failures

### CamelCase Conversion

The `toCamelCase` function converts flag names to struct field names:

1. Splits by underscores and dashes
2. Capitalizes the first letter of each word
3. Joins words without separators
4. Example: "nested_value" -> "NestedValue"

## Best Practices

### 1. Use Struct Tags for Field Mapping

Define YAML tags on struct fields to support both struct and map-based configuration:

```go
type Config struct {
    ClusterName string `yaml:"cluster_name"`
    NodeCount   int    `yaml:"node_count"`
}
```

### 2. Initialize Maps Before Use

Always initialize map fields before applying map operations:

```go
type Config struct {
    Metadata map[string]string
}

config := &Config{
    Metadata: make(map[string]string),  // Initialize before use
}
```

### 3. Handle Nil Pointers

The system automatically initializes nil pointers for nested structs, but you can also pre-initialize them:

```go
type Config struct {
    Nested *NestedConfig
}

config := &Config{
    Nested: &NestedConfig{},  // Pre-initialize if needed
}
```

### 4. Use Type-Safe Values

When possible, use the correct Go types instead of relying on type conversion:

```go
// Preferred
integration.applyToStruct(config, "count", 42)

// Works but less efficient
integration.applyToStruct(config, "count", "42")
```

### 5. Validate After Application

Always validate the configuration after applying flags:

```go
integration.applyFlags(parsed, config, configMap)

// Validate the configuration
validator := NewDefaultConfigurationValidator()
result, _ := validator.ValidateConfiguration(&Configuration{Data: configMap})

if !result.Valid {
    // Handle validation errors
}
```

## Testing

The struct application system includes comprehensive tests:

- `TestApplyToStruct_StringField`: String field updates
- `TestApplyToStruct_IntField`: Integer field updates with type conversion
- `TestApplyToStruct_BoolField`: Boolean field updates with type conversion
- `TestApplyToStruct_SliceField`: Slice field updates
- `TestApplyToStruct_MapField`: Map field updates
- `TestApplyToStruct_NestedPointerField`: Nested pointer field updates
- `TestApplyToStruct_NestedValueField`: Nested value field updates
- `TestFlagNameToFieldPath`: Field path conversion
- `TestNavigateToField`: Field navigation
- `TestSetFieldValueTyped_TypeConversions`: Type conversion
- `TestApplyArrayOperationToStruct`: Array operations
- `TestApplyMapOperationToStruct`: Map operations

Run tests with:

```bash
go test -v -run TestApplyToStruct ./internal/config/flags/
```

## Related Documentation

- [Flag Integration Overview](README.md)
- [Enhanced Flag Parser](ENHANCED_PARSER.md)
- [Validation System](VALIDATION.md)
- [Configuration Merging](MERGING.md)
