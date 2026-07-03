---
id: gitops-configuration
title: "GitOps Configuration Reference"
sidebar_label: GitOps Config
description: Complete reference of opencenter.gitops.* fields with types, defaults, and validation rules.
doc_type: reference
audience: "platform engineers, operators"
tags: [gitops, configuration, fluxcd, repository, auth]
---

> **Purpose:** For platform engineers and operators, documents every field in the `opencenter.gitops.*` configuration section with types, defaults, and validation rules.

The `opencenter.gitops` section configures the cluster GitOps repository, authentication, FluxCD reconciliation, and overlay units.
This section is required.

## Full Structure

```yaml
opencenter:
  gitops:
    repository:
      url: "ssh://git@github.com/my-org/my-cluster-gitops.git"
      branch: "main"
      path: ""
      local_dir: "/path/to/local/checkout"
      secret_name: "flux-system"
    base_repo:
      url: "https://git@github.com/opencenter-cloud/openCenter-gitops-base.git"
      release: "2026.01"
      branch: ""
    auth:
      ssh:
        private_key: "~/.config/opencenter/clusters/my-org/secrets/ssh/my-cluster-key"
        public_key: "~/.config/opencenter/clusters/my-org/secrets/ssh/my-cluster-key.pub"
      token:
        provider: "github"
        token: ""
        token_file: "~/.config/opencenter/clusters/my-org/secrets/git-token.txt"
        owner: "my-org"
        organization: ""
    flux:
      interval: "5m"
      prune: true
    overlay_units:
      customer_managed:
        enabled: false
        repository_url: ""
        repository_name: ""
        branch: ""
        secret_name: ""
        flux_name_prefix: ""
        interval: ""
        emit_secret: false
        kustomizations:
          - name: "apps"
            path: "/clusters/my-cluster/apps"
            depends_on:
              - "sources"
      sops:
        enabled: false
        rules:
          - path_regex: "secrets/.*\\.yaml$"
            encrypted_regex: "^(data|stringData)$"
            age_recipients:
              - "age1..."
```

> Configure either `auth.ssh` or `auth.token`, not both.

## opencenter.gitops.repository

Cluster-specific GitOps repository settings.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `url` | string (URI) | _(placeholder)_ | Remote repository URL. Must use `https://` or `ssh://` scheme. **Required.** |
| `branch` | string | `"main"` | Target branch for FluxCD to track. |
| `path` | string | `""` | Subdirectory within the repository for this cluster's manifests. Empty means repository root. |
| `local_dir` | string | _(derived)_ | Local filesystem path for the repository checkout. Resolved by the path resolver during `cluster init`. |
| `secret_name` | string | `"flux-system"` | Kubernetes Secret name used by FluxCD for repository authentication. |

### Validation

- `url` is required and must parse as a valid URI with scheme `https` or `ssh`.
- If `url` is empty or still contains the `cluster init` placeholder, readiness checks fail.

## opencenter.gitops.base_repo

Upstream platform services template repository (openCenter-gitops-base).

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `url` | string (URI) | `"https://git@github.com/opencenter-cloud/openCenter-gitops-base.git"` | Base repository containing security-hardened Helm values and Kustomize bases. |
| `release` | string | `"2026.01"` | Release tag to pin the base repository at. Takes precedence over `branch` when both are set. |
| `branch` | string | `""` | Branch to track. Use as an alternative to `release` for development or testing. |

### Validation

- `url`, when provided, must be a valid URI.

## opencenter.gitops.auth

Authentication for FluxCD to access the GitOps repository.
Exactly one method (SSH or token) must be configured.
The CLI setting `cluster_defaults.gitops_auth_method` (default: `token`) determines which auth block `cluster init` pre-populates.

### opencenter.gitops.auth.ssh

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `private_key` | string | `""` | Path to the SSH private key file (ed25519 recommended). |
| `public_key` | string | `""` | Path to the corresponding SSH public key file. Used as a deploy key on the Git provider. |

#### Validation

- Both `private_key` and `public_key` are required when SSH auth is configured.
- The repository URL must use the `ssh://` scheme.

### opencenter.gitops.auth.token

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `provider` | string | _(none)_ | Git hosting provider. **Required when token auth is configured.** Allowed values: `github`, `gitlab`, `gitea`. |
| `token` | string | `""` | Inline personal access token value. Mutually exclusive with `token_file`; one of the two is required. |
| `token_file` | string | `""` | Path to a file containing the access token. Preferred over inline `token` for security. |
| `owner` | string | _(extracted from URL)_ | Repository owner (user or organization). If omitted, extracted from `repository.url`. |
| `organization` | string | `""` | Git organization used as the username in authenticated HTTPS URLs (`https://<organization>:<token>@host/path`). |

#### Validation

- `provider` is required and must be one of `github`, `gitlab`, `gitea`.
- `provider` must match the repository host: `github.com` → `github`, `gitlab.com` → `gitlab`, any other host → `gitea`.
- At least one of `token` or `token_file` must be set (value must not be empty or a placeholder).
- The repository URL must use the `https://` scheme.

## opencenter.gitops.auth Mutual Exclusion

| Repository Scheme | Required Auth | Notes |
|-------------------|---------------|-------|
| `ssh://` | `auth.ssh` | `auth.token` must be absent. |
| `https://` | `auth.token` | `auth.ssh` must be absent. |

Configuring both `auth.ssh` and `auth.token` simultaneously is a validation error.

## opencenter.gitops.flux

FluxCD reconciliation behavior.

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `interval` | string | `"5m"` | How often FluxCD checks the repository for changes. Go duration format (e.g., `1m`, `5m`, `15m`). |
| `prune` | boolean | `true` | Whether FluxCD removes resources from the cluster that are no longer present in Git. |

## opencenter.gitops.overlay_units

Overlay units extend the generated GitOps repository with customer-managed applications and SOPS encryption rules.

### opencenter.gitops.overlay_units.customer_managed

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Enable a separate customer-managed application repository. |
| `repository_url` | string | `""` | Git URL for the customer application repository. Must use `ssh://` or `https://` (not `http://`). |
| `repository_name` | string | `""` | Logical name for the repository. **Required when enabled.** |
| `branch` | string | `""` | Branch to track in the customer repository. |
| `secret_name` | string | `""` | Kubernetes Secret name for authenticating to the customer repository. |
| `flux_name_prefix` | string | `""` | Prefix for generated FluxCD Kustomization resource names. |
| `interval` | string | `""` | Reconciliation interval for customer-managed Kustomizations. Inherits from `gitops.flux.interval` when empty. |
| `emit_secret` | boolean | `false` | Generate a Kubernetes Secret manifest containing SSH deploy key credentials. |
| `kustomizations` | array | `[]` | List of Kustomization entry points in the customer repository. |

#### kustomizations[] items

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Logical name for the Kustomization (used in FluxCD resource naming). |
| `path` | string | Path within the customer repository. Must start with `/`. |
| `depends_on` | string[] | List of Kustomization names that must reconcile before this one. |

#### Validation

- `repository_url` must use `ssh://` or `https://`. Plain `http://` is rejected.
- `repository_name` is required when `enabled: true`.
- `emit_secret: true` requires `secrets.overlay_units.customer_managed.identity` to be set.
- `emit_secret: true` is rejected when `repository_url` uses `https://` (only SSH deploy keys are emitted).
- Each `kustomizations[].path` must start with `/`.

### opencenter.gitops.overlay_units.sops

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `enabled` | boolean | `false` | Generate a `.sops.yaml` file in the customer-managed overlay. |
| `rules` | array | `[]` | SOPS creation rules. **At least one rule is required when enabled.** |

#### rules[] items

| Field | Type | Description |
|-------|------|-------------|
| `path_regex` | string | Regex matching files to encrypt (e.g., `secrets/.*\.yaml$`). |
| `encrypted_regex` | string | Regex matching YAML keys to encrypt (e.g., `^(data\|stringData)$`). |
| `age_recipients` | string[] | Age public keys for encryption. Each entry must be non-empty. |

#### Validation

- At least one rule is required when `sops.enabled: true`.
- Each `age_recipients[]` entry must be non-empty.

## Related Secrets

The `secrets.overlay_units.customer_managed` section stores credentials for the customer-managed overlay unit:

| Field | Description |
|-------|-------------|
| `identity` | SSH private key content for the customer-managed repository deploy key. |
| `identity_pub` | Corresponding SSH public key. |
| `known_hosts` | SSH known_hosts entry for the customer repository host. |

These are required when `overlay_units.customer_managed.emit_secret: true` with an SSH repository URL.

## Examples

### Token Authentication (GitHub)

```yaml
opencenter:
  gitops:
    repository:
      url: "https://github.com/my-org/my-cluster-gitops.git"
      branch: "main"
    auth:
      token:
        provider: "github"
        token_file: "~/.config/opencenter/clusters/my-org/secrets/git-token.txt"
        owner: "my-org"
    flux:
      interval: "5m"
      prune: true
```

### SSH Authentication

```yaml
opencenter:
  gitops:
    repository:
      url: "ssh://git@github.com/my-org/my-cluster-gitops.git"
      branch: "main"
    auth:
      ssh:
        private_key: "~/.config/opencenter/clusters/my-org/secrets/ssh/my-cluster-key"
        public_key: "~/.config/opencenter/clusters/my-org/secrets/ssh/my-cluster-key.pub"
    flux:
      interval: "5m"
      prune: true
```

### Customer-Managed Overlay

```yaml
opencenter:
  gitops:
    overlay_units:
      customer_managed:
        enabled: true
        repository_url: "ssh://git@github.com/my-org/customer-apps.git"
        repository_name: "customer-apps"
        branch: "main"
        emit_secret: true
        kustomizations:
          - name: "apps"
            path: "/clusters/my-cluster/apps"
            depends_on:
              - "sources"
          - name: "monitoring"
            path: "/clusters/my-cluster/monitoring"
            depends_on:
              - "apps"
      sops:
        enabled: true
        rules:
          - path_regex: "secrets/.*\\.yaml$"
            encrypted_regex: "^(data|stringData)$"
            age_recipients:
              - "age1ql3z7hjy54pw3hyww5ayyfg7zqgvc7w3j2elw8zmrj2kg5sfn9aqmcac8p"
```

## See Also

- [Configuration Schema Reference](configuration-schema.md) — full cluster config structure
- [Default Values Reference](default-values.md) — all defaults by provider
- [Flux Bootstrap Methods](../operations/flux-bootstrap-methods.md) — step-by-step SSH and token setup
- [GitOps Workflow](../concepts/gitops-workflow.md) — architecture and reconciliation model
