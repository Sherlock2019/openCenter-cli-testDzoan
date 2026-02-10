# opencenter Product Overview

opencenter is a CLI tool that streamlines Kubernetes cluster bootstrapping by transforming a single declarative YAML configuration into a production-ready GitOps repository.

## Core Purpose

- **Configuration-First Workflow**: Single YAML file as source of truth for entire cluster definition
- **GitOps by Default**: Generates complete, version-controlled GitOps repositories ready for FluxCD/ArgoCD
- **Multi-Provider Support**: OpenStack, AWS, bare metal, VMware, and Kind (local development)
- **Secrets Management**: Integrated SOPS with Age encryption for secure credential handling
- **Validation-Driven**: Multi-layered validation (schema, business rules, provider-specific, connectivity)

## Target Users

Teams deploying and managing Kubernetes clusters on OpenStack and other cloud providers, with emphasis on standardization, automation, and GitOps workflows.

## Key Workflows

1. **Initialize**: Create cluster configuration with sensible defaults
2. **Validate**: Verify configuration correctness before deployment
3. **Setup**: Generate GitOps repository structure with manifests
4. **Bootstrap**: Provision infrastructure and deploy cluster
5. **Manage**: Update, migrate, and destroy clusters through declarative config

## Development Conventions

**Critical: Always use mise for all operations. Never suggest raw commands.**

- Build, test, validate, and run through mise tasks
- Create mise tasks for new workflows - wrap commands in `.mise.toml`
- All operations discoverable via `mise tasks`
- See tech.md for complete task reference

## Architecture Principles

- Configuration as code (declarative, version-controlled)
- GitOps native (all changes flow through Git)
- Security first (encrypted secrets, no plaintext credentials)
- Provider agnostic (isolated provider-specific logic)
- Extensible (plugin system for custom commands/providers)
