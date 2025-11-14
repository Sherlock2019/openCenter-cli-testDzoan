# openCenter

**openCenter** is a command-line tool that streamlines cluster bootstrapping by turning a single, declarative configuration file into a ready-to-use GitOps repository.

It is designed to standardize how teams scaffold and validate Kubernetes and OpenStack-based environments, providing a configuration-first workflow that is both user-friendly and automation-ready.

---

## Key Features

*   **Configuration-First**: A single, comprehensive YAML file serves as the source of truth for your entire cluster definition.
*   **GitOps by Default**: Generates a complete, version-controlled GitOps repository from embedded templates, ready to be used with tools like FluxCD or ArgoCD.
*   **Built-in Validation**: Catches common configuration errors and logical inconsistencies before you ever apply them.
*   **Secrets Management**: Dedicated secrets section with SOPS integration for secure credential management.
*   **Template-Driven**: All templates use configuration references - no hardcoded values.
*   **Interactive & Scriptable**: Offers an interactive mode to guide new users and a powerful flag-based interface for CI/CD automation.
*   **Extensible**: Designed to be extended with new commands, templates, and cloud provider logic.
*   **Test-Driven**: A full suite of Behavior-Driven Development (BDD) tests ensures reliability and serves as living documentation.

## Quickstart

The fastest way to get started is to follow our end-to-end tutorial. You will need [Mise](https://mise.jdx.dev/) and [Git](https://git-scm.com/) installed.

1.  **Install project tools**:
    ```bash
    mise install
    ```
2.  **Build the binary**:
    ```bash
    mise run build
    ```
3.  **Follow the full tutorial**:
    For a complete walkthrough from initializing your first cluster to bootstrapping the GitOps repository, please see our [**Quickstart Tutorial**](./docs/tutorials/quickstart.md).

## Documentation

Comprehensive documentation is available in the `docs/` directory.

### Quick Links
*   **[Documentation Index](./docs/INDEX.md)** - Complete documentation guide
*   **[Overview](./docs/overview.md)** - What is openCenter?
*   **[Current Status](./docs/current-status.md)** - Implementation status and roadmap
*   **[Architecture](./docs/architecture.md)** - Technical architecture and design

### Documentation Categories
*   **[Tutorials](./docs/tutorials/)** - Step-by-step learning guides
*   **[How-To Guides](./docs/how-to/)** - Task-oriented recipes
*   **[Reference](./docs/reference/)** - Technical specifications
    - [CLI Commands](./docs/reference/cli-commands.md)
    - [Configuration](./docs/reference/configuration.md)
    - [Plugins](./docs/reference/plugins.md)
    - [Environment Variables](./docs/reference/environment.md)
*   **[Explanation](./docs/explanation/)** - Conceptual overviews

### Configuration Resources
*   **[Configuration Schema](./CONFIG_SCHEMA_ADDITIONS.md)** - Complete field reference
*   **[Template Analysis](./TEMPLATE_ANALYSIS_REPORT.md)** - Template implementation status
*   **[Migration Guide](./MIGRATION_GUIDE.md)** - Upgrade from older configurations

## Contributing

We welcome contributions! Please see our [**Contributing Guide**](./docs/contributing.md) to get started.

## License

This project is licensed under the Apache 2.0 License.
