# Development Guide

## Language & Runtime

- **Go 1.25.2**: Primary language for CLI implementation
- **Mise**: Tool version management and task automation (replaces Make)

## Coding Standards

### Go Conventions
- Formatting enforced with `gofmt` - run `mise run fmt` before committing
- Standard Go naming: `CamelCase` for exported, `mixedCase` for locals
- Test files: `*_test.go` (unit), `*_property_test.go` (property-based)
- Test functions: `TestXxx` naming convention

### Import Organization
1. Standard library
2. External dependencies  
3. Internal packages

Example:
```go
import (
    "fmt"
    "os"
    
    "github.com/spf13/cobra"
    "gopkg.in/yaml.v3"
    
    "github.com/rackerlabs/opencenter-cli/internal/config"
)
```

### Commit Messages
- Prefer Conventional Commits: `feat:`, `fix:`, `docs:`, `refactor:`
- Occasional imperative style acceptable: `Add feature X`
- PRs need: description, test commands run, docs updates

## Core Dependencies

### CLI Framework
- **Cobra**: Command structure, flag parsing, help generation
- **Bubble Tea**: Interactive TUI components (when needed)
- **Lipgloss**: Terminal styling

### Configuration & Templating
- **gopkg.in/yaml.v3**: YAML parsing and marshaling
- **text/template**: Go template engine for file generation
- **Sprig v3**: Extended template functions

### Security & Secrets
- **SOPS**: Secrets encryption/decryption
- **Age (filippo.io/age)**: Encryption backend for SOPS
- **golang.org/x/crypto**: SSH key generation

### Cloud Providers
- **gophercloud**: OpenStack API client
- **AWS SDK** (planned): AWS integration

### Testing
- **godog**: BDD testing framework (Cucumber for Go)
- **gopter**: Property-based testing
- **testify**: Assertions and test utilities

## Build System (Mise)

Common tasks defined in `.mise.toml`:

```bash
# Build binary with version info
mise run build

# Build for Linux
mise run build-linux

# Build for all platforms
mise run build-all

# Run unit tests
mise run test

# Run BDD tests
mise run godog

# Run WIP scenarios only
mise run godog-wip

# Format code
mise run fmt

# Tidy dependencies
mise run tidy

# Upgrade all dependencies
mise run upgrade-deps

# Generate JSON schema
mise run schema

# Validate configuration
mise run validate

# Run preflight checks
mise run preflight

# Generate CLI documentation
mise run docs-gen

# Comprehensive schema verification
mise run schema-verify

# Clean build artifacts
mise run clean
```

## Infrastructure Tools

- **OpenTofu/Terraform**: Infrastructure provisioning
- **Ansible**: Configuration management (Kubespray provider)
- **Pulumi**: Infrastructure as code (Talos provider)
- **FluxCD/ArgoCD**: GitOps continuous delivery

## Development Workflow

**Always use mise for task execution. Never use raw commands.**

1. Install tools: `mise install`
2. Build binary: `mise run build`
3. Run tests: `mise run test && mise run godog`
4. Format code: `mise run fmt`
5. Validate changes: `mise run schema-verify` (for schema changes)

### Pre-Commit Checklist

Before committing code, always run:

```bash
# Build to verify compilation
mise run build

# Format code
mise run fmt

# Run tests
mise run test

# Run service rendering tests (if config changes)
go test -v -run TestServiceRendering ./internal/config
```

**Critical**: Always run `mise run build` before committing to catch compilation errors early.

### Creating New Mise Tasks

When adding new functionality that requires running commands:

1. **Always create a mise task** in `.mise.toml` under `[tasks]`
2. Use descriptive task names (e.g., `test-integration`, `deploy-local`)
3. Document the task purpose with a comment
4. Chain related tasks using task dependencies

Example task definition:
```toml
[tasks]
# Run integration tests with local cluster
test-integration = [
  "mise run kind-cluster-no-cni",
  "go test ./tests/integration/... -v"
]
```

**Never suggest raw bash/go commands** - always wrap them in mise tasks for consistency and discoverability.

## Version Management

Build information is injected at compile time via ldflags:
- `version`: Git tag or "dev"
- `gitCommit`: Current commit hash
- `gitBranch`: Current branch name
- `gitTag`: Exact tag if on tagged commit
- `buildDate`: ISO 8601 timestamp

## Testing Strategy

- **Unit Tests**: Standard Go tests in `internal/` packages (`mise run test`)
- **BDD Tests**: Gherkin scenarios in `tests/features/` (`mise run godog`)
- **Property Tests**: Generative testing for critical logic (gopter)
- **Integration Tests**: Full workflow validation

### Test Execution
- `mise run test`: Run unit tests (`go test ./internal/...`)
- `mise run godog`: Run BDD tests (non-@wip scenarios)
- `mise run godog-wip`: Run only @wip scenarios during development
- Keep fixtures in `testdata/` and reuse existing patterns

### Test Organization
- Gherkin features: `tests/features/*.feature`
- Step definitions: `tests/features/steps/`
- Use `@wip` tag for work-in-progress scenarios

## Security & Configuration

- **SOPS**: Secrets encryption - never commit plaintext secrets
- **YAML Linting**: Checked with `.yamllint` - run `yamllint` for config changes
- **Credential Masking**: Use `internal/security` utilities for logging
- See `.sops.yaml` for encryption configuration
