# Talos OpenStack External Management Remediation Design

## Purpose

Remediate the current Talos OpenStack deployment path so `opencenter cluster deploy` works from outside the OpenStack tenant network while keeping Talos management access restricted to explicitly allowed source CIDRs.

The immediate failure is caused by conflating two network surfaces:

- Kubernetes API access belongs on the cluster VIP or load balancer.
- Talos API access belongs on per-node management addresses.

The current Talos client wrapper has been changed to connect directly to a target node when one is supplied, which is the right direction. The remediation should complete that design by making direct-node Talos management access explicit in configuration, generated OpenStack infrastructure, generated inventory, client behavior, validation, and tests.

## References

- Rackspace: Running Talos on OpenStack Flex, 2024-11-04
- NETWAYS: Talos Linux on OpenStack Step-by-Step, 2024-08-22
- OneUptime: Set Up Talos Linux on OpenStack, 2026-03-03
- Sidero Talos documentation: `talosctl` endpoints and nodes
- Sidero Talos documentation: network connectivity

## Decisions

- External-first OpenStack Talos deploys are supported.
- The Kubernetes API endpoint uses HTTPS on port `443`.
- Talos management traffic uses node-specific addresses on the Talos API port, default `50000`.
- Talos management ingress is allowed only from configured CIDRs.
- The Kubernetes API VIP/load balancer must not forward Talos API traffic.
- Initial machine config application uses direct maintenance-mode node connections, equivalent to `talosctl apply-config --insecure --nodes 198.51.100.11`.
- Post-apply Talos operations use authenticated Talos client config and direct node endpoints when targeting a specific node.
- Inventory is the contract between OpenStack provisioning and Talos bootstrap.

## Non-Goals

- Do not add a bastion relay in this remediation.
- Do not proxy Talos API through the Kubernetes API VIP.
- Do not expose Talos API to `0.0.0.0/0` by default.
- Do not require live OpenStack or live Talos nodes for unit tests.
- Do not switch to Cluster API or a management-cluster model.

## Configuration

Add external management CIDRs under `deployment.talos.network`.

```yaml
deployment:
  method: talos
  talos:
    endpoint: "https://203.0.113.10:443"
    network:
      pod_subnet: 10.42.0.0/16
      service_subnet: 10.43.0.0/16
      talos_api_port: 50000
      management_cidrs:
        - "198.51.100.25/32"
```

Validation rules:

- `deployment.method: talos` with OpenStack requires at least one `management_cidrs` entry.
- Every `management_cidrs` entry must be a valid CIDR.
- `deployment.talos.endpoint` for OpenStack Talos must use HTTPS and port `443` when set explicitly.
- Empty `deployment.talos.endpoint` remains allowed before OpenTofu renders the concrete endpoint, but generated inventory must contain a concrete `https://...:443` endpoint.
- `deployment.talos.network.talos_api_port` defaults to `50000` and must remain a valid TCP port.

## OpenStack Generation

The OpenStack Talos template should render two separate access surfaces.

Kubernetes API:

- Exposed by the existing VIP or load balancer on `443`.
- Controlled by the existing Kubernetes API ACL settings.
- Written to inventory as `cluster.endpoint`.

Talos management API:

- Exposed on each node management address, default port `50000`.
- Controlled by `deployment.talos.network.management_cidrs`.
- Written to inventory as each node's `talos_api_ip`.
- Not forwarded by the Kubernetes API VIP or load balancer.

Generated inventory shape:

```yaml
cluster:
  name: example
  endpoint: https://203.0.113.10:443
  talos_api_port: 50000

control_plane:
  - name: example-cp-1
    talos_api_ip: 198.51.100.11
    internal_ip: 10.2.128.11
    install_disk: /dev/vda
    cert_sans:
      - 203.0.113.10
      - 198.51.100.11
      - 10.2.128.11

workers:
  - name: example-worker-1
    talos_api_ip: 198.51.100.21
    internal_ip: 10.2.128.21
    install_disk: /dev/vda
```

Generation must use per-node floating IPs or explicitly configured externally reachable management IPs for `talos_api_ip`. It must preserve tenant/private addresses in `internal_ip`.

If the upstream OpenStack module cannot create or expose per-node management floating IPs and security rules, remediation will add those resources in the generated OpenStack Talos Terraform wrapper. The important contract is that `talos_api_ip` must be reachable from the operator CIDRs before Talos API calls run.

## Talos Client Behavior

Split Talos operations into initial apply mode and post-apply authenticated mode.

Initial apply mode:

- Used by `talos-apply-machine-configs`.
- Connects directly to each inventory node's `talos_api_ip`.
- Uses the inventory Talos API port, default `50000`.
- Uses maintenance-mode insecure TLS behavior equivalent to `talosctl apply-config --insecure`.
- Does not use `WithNodes` proxy metadata.
- Does not use the Kubernetes API endpoint.

Post-apply authenticated mode:

- Used by bootstrap, kubeconfig export, and health checks.
- Uses generated Talos client config and cluster secrets.
- When targeting one node, uses that node's management address as the endpoint.
- Avoids routing through the Kubernetes API VIP.
- May include control-plane management addresses in generated talosconfig endpoints for later operator use.

Error handling:

- Include operation name, node name, node management IP, and port.
- Classify timeouts, connection refusals, and DNS errors as reachability failures. Classify TLS handshake and authorization errors as TLS/authentication failures. Preserve the original wrapped error for anything that does not match those categories.
- Tell the operator to check `management_cidrs`, security group rules, and per-node floating IP reachability for connection timeouts.

## Deployment Flow

The existing step sequence stays intact, but the step responsibilities become stricter.

1. `talos-preflight` validates OpenStack credentials, Talos config, and management CIDRs.
2. `opentofu-init` initializes OpenTofu.
3. `opentofu-apply` provisions node management access, security rules, Kubernetes API access on `443`, and writes inventory.
4. `talos-read-inventory` validates the inventory contract, including `cluster.endpoint` on `443` and every node having both management and internal IPs.
5. `talos-generate-secrets` creates or loads machine secrets.
6. `talos-apply-machine-configs` applies configs through direct maintenance-mode node connections.
7. `talos-bootstrap-controlplane` bootstraps the first control-plane node through its management IP.
8. `talos-export-talosconfig` writes talosconfig with Talos management endpoints.
9. `talos-export-kubeconfig` writes kubeconfig for Kubernetes API endpoint `443`.
10. `talos-wait-ready` checks Talos node health and Kubernetes API readiness separately.

## Testing Strategy

Unit and render tests should cover:

- Talos OpenStack defaults include `management_cidrs` only when supplied by the user or CLI defaults; validation rejects empty values for deployable configs.
- Generated Talos OpenStack `main.tf` renders Kubernetes API endpoint port `443`.
- Generated OpenStack security rules allow Talos API port `50000` only from configured management CIDRs.
- Generated inventory distinguishes `talos_api_ip` from `internal_ip`.
- Inventory validation rejects OpenStack Talos `cluster.endpoint` values on `:6443`.
- Initial machine config application creates a direct maintenance-mode client and does not set Talos proxy node metadata.
- Bootstrap, kubeconfig export, and health checks use direct Talos management endpoints instead of the Kubernetes API VIP.
- Error messages include node name, management IP, port, and operation.

Integration-style tests should use mocked OpenTofu and mocked Talos clients. No test should require live OpenStack, live Talos nodes, or `talosctl`.

## Acceptance Criteria

- A generated OpenStack Talos cluster exposes Kubernetes API on `443`.
- A generated OpenStack Talos cluster exposes Talos API only on node management addresses and only to configured CIDRs.
- `opencenter cluster deploy` applies machine configs without attempting to proxy Talos requests through the Kubernetes API VIP.
- Deploy failures during Talos API phases identify the node management IP being used.
- The repo has tests that fail on the previous VIP/proxy behavior and pass with direct-node external management behavior.
