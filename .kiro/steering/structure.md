# Project Structure

## Root Layout

```
opencenter-cli/
├── cmd/                    # CLI commands (Cobra)
├── internal/               # Internal packages (not importable)
├── docs/                   # Documentation
├── tests/                  # BDD test scenarios
├── schema/                 # JSON schema definitions
├── testdata/               # Test fixtures and data
├── hack/                   # Scripts and utilities
├── bin/                    # Compiled binaries (gitignored)
├── third-party/            # External dependencies/submodules
├── main.go                 # Entry point
├── go.mod                  # Go module definition
└── .mise.toml              # Build tasks and tool versions
```

## Command Layer (`cmd/`)

Each command is a separate file following the pattern `cmd/<command>_<subcommand>.go`:

- `cluster_*.go`: Cluster lifecycle commands (init, validate, setup, bootstrap, etc.)
- `sops_*.go`: Secrets management commands
- `config_*.go`: Configuration commands
- `plugins_*.go`: Plugin management
- `root.go`: Root command and global flags

**Naming Convention**: `newCluster<Action>Cmd()` returns `*cobra.Command`

## Internal Packages (`internal/`)

### Configuration (`internal/config/`)
Core configuration management:
- `config.go`: Main Config struct and types
- `schema.go`: JSON schema generation
- `validator.go`: Validation logic (schema + business rules)
- `loader.go`: Configuration loading from YAML
- `manager.go`: Configuration lifecycle management
- `path_resolver.go`: Organization-based path resolution
- `migrator.go`: Schema migration between versions
- `defaults/`: Default configuration templates per provider

### GitOps (`internal/gitops/`)
GitOps repository scaffolding:
- `copy.go`: Template copying and rendering logic
- `embed.go`: Embedded template management (`//go:embed`)
- `gitops-base-dir/`: Base repository structure (embedded)
- `templates/`: Cluster-specific templates (embedded)
  - `cluster-apps-base/`: Application manifests
  - `infrastructure-cluster-template/`: Infrastructure configs

### Secrets (`internal/sops/`)
SOPS and Age key management:
- `manager.go`: SOPS manager interface
- `keys.go`: Age key generation and storage
- `encrypt.go`: Encryption/decryption operations
- `git.go`: Git integration for encrypted files
- `validator.go`: SOPS configuration validation

### Providers (`internal/cloud/`, `internal/provision/`)
Cloud provider adapters:
- `internal/cloud/openstack/`: OpenStack preflight checks
- `internal/provision/`: Terraform/OpenTofu provisioning
- `internal/ansible/`: Ansible provisioning (Kubespray)
- `internal/talos/`: Talos Linux provider (Pulumi-based)

### Utilities (`internal/util/`)
Shared utility packages:
- `crypto/`: Key generation and management
- `errors/`: Error handling and aggregation
- `files/`: File operations (atomic writes, backups)
- `paths/`: Path resolution and validation
- `security/`: Security utilities (credential masking, audit logging)
- `template/`: Template engine and validation

## Documentation (`docs/`)

Organized following Diátaxis framework:
- `reference/`: API and CLI reference documentation
- `dev/`: Developer guides and architecture docs
- `providers/`: Provider-specific documentation
- `templates/`: Documentation templates

## Testing (`tests/`)

BDD tests using Cucumber/Gherkin:
- `features/*.feature`: Gherkin test scenarios
- `features/steps/`: Step definitions in Go

**Tag Convention**: Use `@wip` for work-in-progress scenarios

## Configuration Storage

User configurations stored in organization-based structure:

```
~/.config/opencenter/clusters/
└── <organization>/
    ├── <cluster>/
    │   └── .<cluster>-config.yaml
    ├── secrets/
    │   ├── age/
    │   │   └── <cluster>-key.txt
    │   └── ssh/
    │       └── <cluster>-<env>-<region>
    └── gitops/
        ├── applications/
        │   └── overlays/<cluster>/
        └── infrastructure/
            └── clusters/<cluster>/
```

## Code Organization Principles

1. **Separation of Concerns**: Each package has a single, well-defined responsibility
2. **Dependency Injection**: Avoid global state, pass dependencies explicitly
3. **Interface-Based Design**: Define interfaces in consumer packages
4. **Embedded Resources**: Templates and defaults embedded in binary via `//go:embed`
5. **Error Wrapping**: Use `fmt.Errorf` with `%w` for error context

## File Naming Conventions

- Commands: `<noun>_<verb>.go` (e.g., `cluster_init.go`)
- Tests: `<name>_test.go` (unit), `<name>_property_test.go` (property)
- Interfaces: `interfaces.go` in each package
- Documentation: `doc.go` for package documentation

## Task Management with Mise

**Critical Rule**: Always use mise tasks instead of raw commands.

### When to Create a Mise Task

Create a new task in `.mise.toml` when:
- Adding a new build target or compilation step
- Creating a new test suite or test category
- Adding deployment or infrastructure commands
- Implementing validation or verification steps
- Creating utility scripts or helper commands

### Task Naming Conventions

- Use kebab-case: `test-integration`, `build-linux`, `deploy-local`
- Prefix related tasks: `gitea-setup`, `gitea-configure`, `gitea-cleanup`
- Use descriptive names that indicate the action and target

### Example Task Patterns

```toml
# Single command task
format = "gofmt -w ."

# Multi-step task with array
test-all = [
  "mise run test",
  "mise run godog"
]

# Script task with heredoc
build-custom = '''
#!/usr/bin/env bash
set -e
echo "Building custom target..."
go build -o bin/custom ./cmd/custom
'''
```

### Discovering Tasks

Users can see all available tasks with: `mise tasks`
