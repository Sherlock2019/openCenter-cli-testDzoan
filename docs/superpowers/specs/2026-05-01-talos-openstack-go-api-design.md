# Talos OpenStack Go API Support Design

## Purpose

Add first-class Talos support to the openCenter CLI as a v2 deployment method for OpenStack clusters. This first release is intentionally scoped to OpenStack only.

Talos is not an infrastructure provider. OpenStack remains responsible for infrastructure, and Talos is the Kubernetes bootstrap and node operating system lifecycle method layered on top of that infrastructure.

## Decisions

- Talos settings live under `deployment.talos` in the active v2 config model.
- The first implementation supports `opencenter.infrastructure.provider: openstack` only.
- The implementation is a clean break from legacy Talos shapes.
- The CLI uses Talos Go machinery/client APIs directly, not `talosctl`.
- OpenTofu continues to provision OpenStack resources.
- OpenTofu-generated artifacts provide a Talos inventory file consumed by CLI bootstrap.
- Talos bootstrap uses the existing deploy resume state file for step progress only.
- Talos machine configs are generated in memory and are ephemeral by default.
- Sensitive Talos artifacts are persisted only in cluster-owned secret paths.

## Non-Goals

- No `--type talos` compatibility alias.
- No `opencenter.infrastructure.provider: talos` support.
- No `opencenter.talos` support in v2.
- No automatic migration from legacy Talos config shapes.
- No Pulumi deployment lane in the first implementation.
- No Talos support for VMware or baremetal in the first implementation.
- No default persistence of generated machine configs.

## User Flow

```text
opencenter cluster init my-cluster --type openstack --deployment talos
opencenter cluster generate my-cluster
opencenter cluster deploy my-cluster
opencenter cluster status my-cluster
```

`cluster init` creates an OpenStack infrastructure config with `deployment.method: talos`. `cluster generate` renders OpenStack infrastructure plus Talos bootstrap artifacts. `cluster deploy` provisions OpenStack resources and then bootstraps Talos using native Go APIs.

## Configuration Model

Talos settings live only under `deployment.talos`.

```yaml
deployment:
  method: talos
  talos:
    version: v1.8.0
    kubernetes_version: 1.33.5
    endpoint: ""
    install:
      disk: /dev/sda
      image: ghcr.io/siderolabs/installer:v1.8.0
    network:
      pod_subnet: 10.42.0.0/16
      service_subnet: 10.43.0.0/16
      talos_api_port: 50000
    patches:
      static:
        - disable-cni
        - disable-kubeproxy
        - disable-node-cidr-allocator
        - ntp
```

`cluster init --deployment talos --type openstack` sets:

- `opencenter.infrastructure.provider: openstack`
- `deployment.method: talos`
- `deployment.talos` defaults
- Talos-friendly platform defaults: Cilium enabled, Calico disabled, and kube-proxy disabled or replaced

Clean break validation rules:

- `--type talos` is invalid.
- `opencenter.infrastructure.provider: talos` is invalid.
- `opencenter.talos` is invalid in v2.
- `deployment.method: talos` with a non-OpenStack provider is invalid for this release.

The user-facing error for provider misuse is:

```text
talos is a deployment method, not an infrastructure provider. Use --type openstack --deployment talos.
```

## Generation and Inventory Contract

`cluster generate` for `deployment.method: talos` renders the OpenStack infrastructure directory without Kubespray deployment blocks or Kubespray inventory artifacts.

Generated layout:

```text
infrastructure/clusters/<cluster>/
  main.tf
  variables.tf
  Makefile
  talos/
    inventory.yaml
    patches/
      disable-cni.yaml
      disable-kubeproxy.yaml
      disable-node-cidr-allocator.yaml
      ntp.yaml.tmpl
      network-subnets.yaml.tmpl
```

The generated `main.tf` remains responsible for OpenStack resources. It also writes the Talos inventory file after resource values are known. The inventory is the explicit contract between OpenStack provisioning and Talos bootstrap.

Inventory shape:

```yaml
cluster:
  name: example
  endpoint: https://10.2.128.5:6443
  talos_api_port: 50000

control_plane:
  - name: example-cp-0
    talos_api_ip: 10.2.128.10
    internal_ip: 10.2.128.10
    install_disk: /dev/sda
    cert_sans:
      - 10.2.128.5
      - 10.2.128.10

workers:
  - name: example-worker-0
    talos_api_ip: 10.2.128.20
    internal_ip: 10.2.128.20
    install_disk: /dev/sda
    labels:
      node-role.kubernetes.io/worker: ""

patch_inputs:
  dns_servers:
    - 1.1.1.1
  ntp_servers:
    - time.cloudflare.com
  pod_subnet: 10.42.0.0/16
  service_subnet: 10.43.0.0/16
```

Talos bootstrap treats `inventory.yaml` as read-only input. Missing or malformed inventory fails deploy before any Talos API call is attempted.

## Bootstrap Design

`cluster deploy` selects the Talos bootstrap path when `deployment.method == "talos"`.

The Talos provider uses the existing bootstrap step runner and state mechanism, but calls Talos Go machinery/client APIs instead of external `talosctl` commands.

Step sequence:

```text
talos-preflight
opentofu-init
opentofu-apply
talos-read-inventory
talos-generate-secrets
talos-apply-machine-configs
talos-bootstrap-controlplane
talos-export-talosconfig
talos-export-kubeconfig
talos-wait-ready
```

Step responsibilities:

- `talos-preflight`: validate `deployment.talos`, OpenStack config, inventory path expectations, and the supported Talos version.
- `opentofu-init`: initialize OpenTofu in the cluster infrastructure directory.
- `opentofu-apply`: create or update OpenStack infrastructure.
- `talos-read-inventory`: read and validate `infrastructure/clusters/<cluster>/talos/inventory.yaml`.
- `talos-generate-secrets`: create or load Talos machine secrets from the cluster-owned secrets path.
- `talos-apply-machine-configs`: generate machine configs in memory, apply static and inventory-driven patches, and apply configs to nodes with Talos Go client APIs.
- `talos-bootstrap-controlplane`: bootstrap the first control-plane node through Talos Go client APIs and skip cleanly when the cluster is already bootstrapped.
- `talos-export-talosconfig`: write Talos client config to cluster-owned secrets.
- `talos-export-kubeconfig`: write kubeconfig to the existing cluster-owned kubeconfig path.
- `talos-wait-ready`: wait for Talos health and Kubernetes API readiness.

## State and Persistence

Talos deploy uses the existing bootstrap resume state file:

```text
<state_dir>/bootstrap/<org>/<cluster>/state.json
```

The state file stores only step status, timestamp, and error text. It does not store Talos secrets, generated configs, kubeconfig, talosconfig, or inventory content.

Persistent artifacts:

- Talos machine secrets: cluster-owned secrets path, SOPS-protected.
- Talosconfig: cluster-owned secrets path, SOPS-protected.
- Kubeconfig: existing cluster-owned kubeconfig path.
- Generated machine configs: ephemeral by default.

An explicit debug or export option may write generated machine configs for troubleshooting, but this must be opt-in because machine configs contain sensitive material.

## Status and Validation

`cluster status` adds a Talos section when `deployment.method == "talos"`.

It reports:

- deployment method
- infrastructure provider
- Talos inventory presence
- Talos secrets presence
- talosconfig presence
- kubeconfig presence
- Talos API health when `--refresh` is used
- Kubernetes API readiness when `--refresh` is used

`cluster status <name>` already honors the positional cluster name in the current codebase, so that is not part of this Talos implementation.

Validation rules:

- `deployment.method: talos` requires `deployment.talos`.
- `deployment.method: talos` requires `opencenter.infrastructure.provider: openstack`.
- `opencenter.infrastructure.provider: talos` is always invalid.
- `opencenter.talos` is always invalid in v2.
- `--type talos` is invalid.

## Error Handling

- Missing inventory fails before Talos API calls.
- Malformed inventory errors include the inventory path and invalid field.
- Corrupt persisted Talos secrets fail with the secret path and recovery guidance.
- Patch parse or apply failures include patch name, node name, and node role.
- Talos API failures include node address and operation.
- Bootstrap failures are recorded in the existing deploy resume state.

## Testing Strategy

Unit and command tests should cover:

- `deployment.talos` defaults for `cluster init --deployment talos --type openstack`
- validation rejection for `--type talos`
- validation rejection for `opencenter.infrastructure.provider: talos`
- validation rejection for `opencenter.talos`
- validation rejection for Talos with non-OpenStack providers
- render tests proving Talos generation excludes Kubespray module blocks and inventory artifacts
- render tests proving Talos inventory and patch artifacts are generated
- inventory parser validation
- bootstrap step construction for Talos deployments
- bootstrap state/resume behavior using the existing state mechanism
- status output for Talos artifact presence

Integration-style tests should mock OpenTofu and Talos API clients. No test should require live OpenStack, real Talos nodes, or `talosctl`.

## Implementation Boundaries

The implementation should introduce a small Talos-specific package boundary for:

- v2 Talos config defaults and validation
- Talos inventory parsing and validation
- machine config patch assembly
- Talos Go client operations

The OpenStack bootstrap provider should not grow Talos internals. Provider selection should route to a Talos deployment provider before falling back to provider-only bootstrap behavior.

## Acceptance Criteria

- `cluster init --type openstack --deployment talos` writes a valid v2 config with `deployment.method: talos` and populated `deployment.talos`.
- `cluster init --type talos` fails with clean-break guidance.
- A config with `opencenter.infrastructure.provider: talos` fails validation.
- `cluster generate` for Talos/OpenStack renders OpenStack infrastructure plus `talos/inventory.yaml` and patch artifacts.
- Talos/OpenStack generation does not render Kubespray module blocks or Kubespray inventory artifacts.
- `cluster deploy` for Talos/OpenStack builds the Talos step sequence and uses the existing resume state file.
- Talos machine configs are not persisted by default.
- Talosconfig and machine secrets are persisted under cluster-owned secrets.
- Kubeconfig is persisted at the existing cluster-owned kubeconfig path.
- `cluster status` reports Talos artifact presence and live readiness when refreshed.
