# Service Rendering Test Suite

Comprehensive test suite to validate service configuration rendering and YAML compliance.

## Purpose

Ensures all service configurations render correctly with proper YAML formatting and complete field sets. Catches missing service registrations that would cause services to render only BaseConfig fields instead of their full configuration.

## Tests

### TestServiceRendering

Main test that validates overall config rendering:

- **YAML Syntax Validation**: Verifies generated YAML is syntactically valid
- **YAML Linting**: Runs yamllint to ensure compliance with YAML standards
- **Service Field Count Validation**: Checks each service has more than BaseConfig fields (12 fields)
- **Service Registration Completeness**: Verifies all services are registered in the service registry
- **No Trailing Spaces**: Ensures no lines end with whitespace
- **Proper Indentation**: Validates 2-space indentation throughout

### TestSpecificServiceRendering

Tests individual service configurations with expected field validation:

- **tempo**: Storage configuration (swift/s3), volume settings
- **loki**: Logging backend configuration, swift integration
- **kube-prometheus-stack**: Prometheus, Grafana, Alertmanager volumes
- **cert-manager**: Email, region, Let's Encrypt server
- **keycloak**: Realm, client ID, frontend URL

### TestYAMLLintCompliance

Tests multiple configuration scenarios:

- **minimal config**: Basic cluster configuration
- **all services enabled**: Full service stack
- **openstack provider**: Provider-specific configuration

## Running Tests

```bash
# Run all service rendering tests
go test -v -run "TestServiceRendering|TestSpecificServiceRendering|TestYAMLLintCompliance" ./internal/config

# Run specific test
go test -v -run TestServiceRendering ./internal/config

# Run with yamllint (requires yamllint installed)
go test -v -run TestYAMLLintCompliance ./internal/config
```

## Requirements

- Go 1.25+
- yamllint (optional, tests skip if not available)

Install yamllint:
```bash
# macOS
brew install yamllint

# Ubuntu/Debian
apt-get install yamllint

# Python pip
pip install yamllint
```

## Common Issues

### Service Only Shows BaseConfig Fields

**Symptom**: Service renders with only 11-12 fields (enabled, namespace, hostname, etc.)

**Cause**: Service not registered in service registry

**Fix**: Add init() function to service file:

```go
func init() {
    registry.RegisterServiceConfig("service-name", ServiceConfig{})
}
```

### YAML Indentation Errors

**Symptom**: yamllint reports "wrong indentation: expected 2 but found 4"

**Cause**: Using yaml.Marshal() instead of custom encoder with 2-space indent

**Fix**: Use marshalYAMLWithIndent() helper function

### Missing Expected Fields

**Symptom**: Test fails with "service 'X' missing expected field: Y"

**Cause**: Field name mismatch between test and actual struct

**Fix**: Check struct definition for correct YAML tag names

## Integration with CI/CD

Add to CI pipeline to catch service registration issues:

```yaml
- name: Test Service Rendering
  run: go test -v -run TestServiceRendering ./internal/config
```

## Related Documentation

- [Service Registry Patterns](.kiro/steering/service-registry-patterns.md)
- [GitOps Manifest Standards](.kiro/steering/gitops-manifest-standards.md)
