# Documentation Summary

This document provides a high-level summary of the openCenter documentation.

## What is openCenter?

openCenter is a command-line tool that streamlines Kubernetes and OpenStack cluster bootstrapping by transforming a single, declarative YAML configuration file into a ready-to-use GitOps repository. It provides a configuration-first workflow with built-in validation, secrets management, and multi-provider support.

## Current Status (v0.0.1)

openCenter is in active development with core functionality implemented:

### ✅ Implemented
- Complete CLI with 20+ commands
- YAML-based configuration with JSON schema
- Organization-based multi-tenancy
- GitOps repository scaffolding
- SOPS-based secrets management
- OpenStack, AWS, and Kind provider support
- Comprehensive BDD and unit test coverage
- Mise-based build system

### 🔄 In Progress
- Enhanced cloud provider validation
- Improved error messages
- Additional service integrations
- Documentation expansion

### 📋 Planned
- Interactive configuration wizard
- Cluster health monitoring
- Automated upgrades
- Multi-cluster management
- Disaster recovery workflows

## Key Features

1. **Configuration-First Workflow**
   - Single YAML file as source of truth
   - JSON schema for IDE integration
   - Comprehensive validation

2. **GitOps by Default**
   - Automatic repository scaffolding
   - Template-driven manifest generation
   - FluxCD integration

3. **Secrets Management**
   - SOPS with Age encryption
   - Automatic key generation
   - Key rotation support

4. **Multi-Provider Support**
   - OpenStack (primary)
   - AWS (basic support)
   - Kind (local development)
   - VMware (partial)

5. **Organization-Based Multi-Tenancy**
   - Isolated cluster configurations
   - Shared SOPS keys
   - Organization-wide policies

## Architecture Highlights

### Modular Design
- **CLI Layer:** Cobra-based command structure
- **Configuration Management:** Schema-driven validation
- **GitOps Scaffolding:** Template-based generation
- **Secrets Management:** SOPS integration
- **Provider Adapters:** Cloud-specific logic
- **Plugin System:** Extensible architecture

### Key Design Principles
- Configuration as Code
- GitOps Native
- Security First
- Provider Agnostic
- Testability
- Extensibility
- User Experience

## Documentation Structure

Following the [Diátaxis](https://diataxis.fr/) framework:

### 📚 Tutorials (Learning-Oriented)
- [Quickstart Tutorial](tutorials/quickstart.md)
- Multi-Cluster Setup (Coming Soon)
- Production Deployment (Coming Soon)

### 🔧 How-To Guides (Task-Oriented)
- [Configure a Cluster](how-to/configure-cluster.md)
- [Manage Secrets](how-to/manage-secrets.md)
- [Setup GitOps](how-to/setup-gitops.md)
- [Use Plugins](how-to/plugins.md)

### 📖 Reference (Information-Oriented)
- [CLI Commands](reference/cli-commands.md) - Complete command reference
- [Configuration](reference/configuration.md) - Configuration file reference
- [Plugins](reference/plugins.md) - Plugin development
- [Environment Variables](reference/environment.md) - Environment configuration

### 💡 Explanation (Understanding-Oriented)
- [Overview](overview.md) - High-level overview
- [Architecture](architecture.md) - Technical architecture
- [Current Status](current-status.md) - Implementation status

## Quick Start

```bash
# Install tools
mise install

# Build openCenter
mise run build

# Initialize a cluster
./bin/openCenter cluster init my-cluster

# Validate configuration
./bin/openCenter cluster validate my-cluster

# Setup GitOps repository
./bin/openCenter cluster setup my-cluster

# Bootstrap the cluster
./bin/openCenter cluster bootstrap my-cluster
```

## Common Use Cases

### 1. Development Cluster
```bash
# Create Kind cluster for local development
openCenter cluster init dev-cluster \
  --opencenter.infrastructure.provider=kind \
  --opencenter.cluster.kubernetes.master_count=1 \
  --opencenter.cluster.kubernetes.worker_count=2

openCenter cluster setup dev-cluster
openCenter cluster bootstrap dev-cluster
```

### 2. Production OpenStack Cluster
```bash
# Initialize production cluster
openCenter cluster init prod-cluster \
  --opencenter.meta.env=prod \
  --opencenter.meta.organization=production \
  --opencenter.infrastructure.provider=openstack \
  --opencenter.cluster.kubernetes.version=1.31.4 \
  --opencenter.cluster.kubernetes.master_count=3 \
  --opencenter.cluster.kubernetes.worker_count=5

# Validate and setup
openCenter cluster validate prod-cluster
openCenter cluster setup prod-cluster --render
openCenter cluster bootstrap prod-cluster
```

### 3. Multi-Organization Setup
```bash
# Organization 1
openCenter cluster init org1-cluster \
  --opencenter.meta.organization=org1

# Organization 2
openCenter cluster init org2-cluster \
  --opencenter.meta.organization=org2

# List all clusters
openCenter cluster list
```

## Technology Stack

- **Language:** Go 1.24+
- **CLI Framework:** Cobra
- **Configuration:** YAML with JSON schema
- **Testing:** Godog (BDD) + Go testing
- **Build System:** Mise
- **Secrets:** SOPS + Age
- **IaC:** OpenTofu/Terraform
- **GitOps:** FluxCD

## Security Features

- SOPS encryption for all secrets
- Age key-based encryption
- Secure file permissions (0600)
- No plaintext secrets in configuration
- Git-friendly encrypted files
- Key rotation support
- Organization-wide key management

## Performance Characteristics

- **Configuration Load:** <100ms
- **Validation:** <500ms
- **Schema Generation:** <200ms
- **Template Rendering:** <2s
- **Repository Setup:** <5s
- **Key Generation:** <1s
- **Encryption:** ~100ms per file

## Known Limitations

- VMware provider validation incomplete
- Some BDD tests fail intermittently
- Large configuration files slow to validate
- Windows support limited
- Azure and GCP providers not yet implemented

## Next Steps

### Short Term (1-2 weeks)
- Complete VMware provider validation
- Fix intermittent test failures
- Improve error messages
- Complete documentation

### Medium Term (1-2 months)
- Interactive configuration wizard
- Cluster health monitoring
- Enhanced connectivity validation
- Expand test coverage

### Long Term (3-6 months)
- Automated upgrade system
- Multi-cluster management
- Disaster recovery workflows
- Cost optimization features
- Azure and GCP support

## Getting Help

- **Documentation:** [docs/INDEX.md](INDEX.md)
- **Issues:** https://github.com/rackerlabs/openCenter-cli/issues
- **Verbose Logging:** Use `--verbose` flag
- **Debug Config:** Use `--generate-debug-config` flag

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

### Areas for Contribution
- Provider implementations
- Documentation improvements
- Test coverage expansion
- Bug fixes
- Feature enhancements
- Plugin development

## License

Apache 2.0 License - See [LICENSE](../LICENSE) for details.

## Version Information

- **Documentation Version:** 1.0.0
- **openCenter Version:** 0.0.1
- **Last Updated:** November 7, 2025
- **Go Version:** 1.24+
- **Kubernetes Versions:** 1.29.x - 1.31.x

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [FluxCD Documentation](https://fluxcd.io/docs/)
- [SOPS Documentation](https://github.com/mozilla/sops)
- [Age Encryption](https://age-encryption.org/)
- [OpenTofu Documentation](https://opentofu.org/docs/)
- [Diátaxis Framework](https://diataxis.fr/)

---

For the complete documentation index, see [INDEX.md](INDEX.md).
