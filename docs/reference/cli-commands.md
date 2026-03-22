---
id: cli-commands
title: "CLI Commands Reference"
sidebar_label: CLI Commands
description: Complete reference of all openCenter CLI commands, flags, and options.
doc_type: reference
audience: "all users"
tags: [cli, commands, flags, reference]
---

# CLI Commands Reference

**Purpose:** Complete reference of the shipped `opencenter` command tree, generated from the live Cobra command graph.

Use `go run -tags tools ./cmd/docs` to refresh the generated inventory in [`docs/reference/opencenter`](/Users/victor.palma/projects/openCenter-cloud/openCenter-cli/docs/reference/opencenter).

## Global Flags

| Flag | Description |
|------|-------------|
| `--config` | Alternative cluster configuration file path |
| `--config-dir` | Configuration directory override |
| `--dry-run` | Print planned actions without executing them |
| `--log-level` | Set log level: `debug`, `info`, `warn`, `error` |
| `--set` | Override configuration values using dot notation |
| `--show-active` | Display the current active cluster |
| `-h, --help` | Show command help |
| `-v, --version` | Show version information |

## `--set` Dot-Notation Examples

```bash
# Override provider and VMware metadata for a one-off validation
opencenter cluster validate prod-cluster \
  --set opencenter.infrastructure.provider=vmware \
  --set opencenter.infrastructure.cloud.vmware.datacenter=DC1 \
  --set opencenter.infrastructure.cloud.vmware.network=dvpg-prod

# Override a service value during bootstrap
opencenter cluster bootstrap prod-cluster \
  --set opencenter.services.cert-manager.email=platform@example.com
```

## Root Commands

| Command | Purpose |
|---------|---------|
| `opencenter cluster` | Cluster lifecycle, validation, rendering, drift, services, backup, and key management |
| `opencenter completion` | Shell completion scripts for `bash`, `fish`, `powershell`, and `zsh` |
| `opencenter config` | CLI defaults and local IDE configuration |
| `opencenter plugins` | External plugin discovery |
| `opencenter secrets` | Secret encryption, sync, validation, and key operations |
| `opencenter shell-init` | Session-scoped shell integration for active-cluster selection |
| `opencenter version` | Version and build metadata |
| `opencenter <external-plugin>` | Dynamically discovered plugin entrypoints such as `rmpk` |

## Cluster Commands

### Lifecycle and Validation

| Command |
|---------|
| `opencenter cluster audit-log` |
| `opencenter cluster bootstrap` |
| `opencenter cluster check-keys` |
| `opencenter cluster current` |
| `opencenter cluster destroy` |
| `opencenter cluster edit` |
| `opencenter cluster env` |
| `opencenter cluster info` |
| `opencenter cluster init` |
| `opencenter cluster install-hooks` |
| `opencenter cluster list` |
| `opencenter cluster lock` |
| `opencenter cluster preflight` |
| `opencenter cluster render` |
| `opencenter cluster revoke-key` |
| `opencenter cluster rotate-keys` |
| `opencenter cluster select` |
| `opencenter cluster setup` |
| `opencenter cluster status` |
| `opencenter cluster unlock` |
| `opencenter cluster update` |
| `opencenter cluster validate` |
| `opencenter cluster validate-manifests` |

### Backup

| Command |
|---------|
| `opencenter cluster backup create` |
| `opencenter cluster backup delete` |
| `opencenter cluster backup list` |
| `opencenter cluster backup restore` |
| `opencenter cluster backup schedule` |

### Configuration

| Command |
|---------|
| `opencenter cluster config export-effective` |
| `opencenter cluster config update` |

### Drift Detection

| Command |
|---------|
| `opencenter cluster drift detect` |
| `opencenter cluster drift reconcile` |
| `opencenter cluster drift schedule` |

### Encryption Keys

| Command |
|---------|
| `opencenter cluster keys` |
| `opencenter cluster keys list` |

### Services

| Command |
|---------|
| `opencenter cluster service disable` |
| `opencenter cluster service enable` |
| `opencenter cluster service options` |
| `opencenter cluster service status` |

## Config Commands

| Command |
|---------|
| `opencenter config edit` |
| `opencenter config get` |
| `opencenter config ide` |
| `opencenter config path` |
| `opencenter config reset` |
| `opencenter config set` |
| `opencenter config view` |

## Secrets Commands

### Core Operations

| Command |
|---------|
| `opencenter secrets decrypt` |
| `opencenter secrets delete` |
| `opencenter secrets describe` |
| `opencenter secrets encrypt` |
| `opencenter secrets get` |
| `opencenter secrets list` |
| `opencenter secrets login` |
| `opencenter secrets set` |
| `opencenter secrets status` |
| `opencenter secrets sync` |
| `opencenter secrets validate` |

### Secret Keys

| Command |
|---------|
| `opencenter secrets keys backup` |
| `opencenter secrets keys generate` |
| `opencenter secrets keys rotate` |
| `opencenter secrets keys validate` |

## Completion Commands

| Command |
|---------|
| `opencenter completion bash` |
| `opencenter completion fish` |
| `opencenter completion powershell` |
| `opencenter completion zsh` |

## GA Notes

- Canonical infrastructure provider names are `openstack`, `vmware`, `kind`, and `baremetal`.
- `vsphere` remains accepted as a compatibility alias for existing configuration files, but documentation now uses `vmware`.
- AWS-backed integrations such as Route53 and S3 credential flows remain supported where services use them, but AWS is not a GA infrastructure provider.
