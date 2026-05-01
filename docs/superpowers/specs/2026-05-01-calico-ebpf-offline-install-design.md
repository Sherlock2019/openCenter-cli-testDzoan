# Calico eBPF Offline Install Design

## Context

OpenStack deploy now installs the selected CNI after kubeconfig normalization. The current Calico path uses the Tigera Helm chart and native v3 CRD flow that was added as an experimental bootstrap path. The desired behavior is to make Calico follow Tigera's self-managed on-premises installation flow, use eBPF mode, and keep every required manifest bundled in the CLI so OpenStack clusters can install Calico without internet access.

The Calico version for this design is pinned to `v3.32.0`, matching the current Tigera documentation version selected during design. Offline installation is deterministic: the CLI installs the bundled Calico assets it ships with instead of downloading manifests at deploy time.

## Goals

- Replace the OpenStack Calico installer with the Tigera on-premises operator manifest flow.
- Enable Calico eBPF mode through `custom-resources-bpf.yaml`.
- Bundle all Calico manifests required by the install path into the CLI binary.
- Keep Cilium and Kube-OVN behavior unchanged.
- Preserve the existing OpenStack step name, resume behavior, dry-run output, `--step`, and `--from-step` support.
- Fail clearly when a config asks for a Calico version that is not bundled.

## Non-Goals

- Do not make Kubespray install Calico.
- Do not require Helm for Calico after this change.
- Do not add Flux ownership for the initial Calico bootstrap in this change.
- Do not support live downloads as a fallback for missing Calico assets.
- Do not change Cilium or Kube-OVN install methods.
- Do not introduce multi-version Calico asset management yet.

## Sources

The bundled assets come from Calico `v3.32.0`:

- Tigera on-premises install documentation: https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises
- v3 CRDs: https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/v3_projectcalico_org.yaml
- Tigera operator: https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/tigera-operator.yaml
- eBPF custom resources: https://raw.githubusercontent.com/projectcalico/calico/v3.32.0/manifests/custom-resources-bpf.yaml

## Architecture

Calico installation remains owned by the existing OpenStack bootstrap step:

1. `openstack-preflight`
2. `opentofu-init`
3. `opentofu-apply`
4. `openstack-normalize-kubeconfig`
5. `openstack-install-network-plugin`
6. optional Flux bootstrap, when configured

Only the Calico branch inside `openstack-install-network-plugin` changes. Instead of building Helm chart commands, it will use embedded Calico manifests. The network plugin selector and validation rules remain the source of truth for deciding that Calico is the enabled plugin.

The Calico installer should be isolated behind the same network-plugin installer interface used by the current code. That keeps OpenStack step orchestration, dry-run planning, state persistence, and fake-runner tests stable while changing the Calico implementation underneath.

## Bundled Assets

Add a Calico asset package under `internal/cluster`, for example:

- `internal/cluster/assets/calico/v3.32.0/v3_projectcalico_org.yaml`
- `internal/cluster/assets/calico/v3.32.0/tigera-operator.yaml`
- `internal/cluster/assets/calico/v3.32.0/custom-resources-bpf.yaml`

Embed the files with `go:embed`. At runtime, the installer writes them into a scoped temporary directory and applies the temp files with `kubectl --kubeconfig <path> apply -f <file>`.

The repository should include a small checksum or verification test so future changes can detect accidental asset drift. The deploy path itself must not fetch or verify against the network.

## Calico Install Flow

The OpenStack Calico install sequence should be:

1. Validate that the resolved Calico version is empty, `3.32.0`, or `v3.32.0`.
2. Verify that the Kubernetes API supports `MutatingAdmissionPolicy`, which Tigera requires for native `projectcalico.org/v3` CRDs.
3. Write the embedded manifests to a temp directory.
4. Patch the eBPF custom resources manifest with OpenCenter cluster settings.
5. Apply `v3_projectcalico_org.yaml`.
6. Apply `tigera-operator.yaml`.
7. Apply the patched `custom-resources-bpf.yaml`.
8. Wait for the Tigera operator deployment.
9. Wait for Calico health using `tigerastatus` and Calico pod readiness.

The v3 CRD manifest follows Tigera's native `projectcalico.org/v3` CRD path. The implementation must not also apply `v1_crd_projectcalico_org.yaml`; Tigera documents these as mutually exclusive CRD paths.

`custom-resources-bpf.yaml` already enables:

- `linuxDataplane: BPF`
- `bpfNetworkBootstrap: Enabled`
- `kubeProxyManagement: Enabled`

The implementation should patch the default IP pool CIDR from the cluster pod CIDR. It should preserve the bundled eBPF fields unless the existing config has a direct equivalent. Service CIDR should not be forced into this manifest unless the Tigera installation API field is explicitly supported by the bundled operator resource.

## Version Behavior

The default OpenStack Calico version becomes `3.32.0`. Config values accepted for the bundled installer:

- empty version
- `3.32.0`
- `v3.32.0`

Any other Calico version should fail before applying manifests with an error like:

`OpenStack Calico offline installer bundles v3.32.0; configure calico.version: 3.32.0 or add bundled assets for the requested version`

This avoids pretending an offline installer can satisfy arbitrary versions.

## Install Method Semantics

OpenStack still accepts the existing CNI install methods:

- empty value, treated as `helm`
- `helm`
- `kustomize-helm`

For Calico, both accepted methods route to the bundled on-prem manifest installer because the requested Calico path is manifest-based and offline. The method remains accepted so existing OpenStack configs using `helm` or `kustomize-helm` do not break when they select Calico.

`kubespray` remains invalid for OpenStack CNI installation and should continue returning the targeted migration error that recommends `helm` or `kustomize-helm`.

## Dry Run and Resume

Dry-run output for `openstack-install-network-plugin` should describe the bundled manifest operations rather than Helm commands. It should show:

- selected plugin: `calico`
- bundled version: `v3.32.0`
- eBPF mode enabled
- embedded manifests to apply
- readiness checks to run

The existing saved-state behavior remains unchanged. If earlier steps completed, a rerun should resume at `openstack-install-network-plugin`. If Calico install fails, rerunning should reapply the manifests idempotently and repeat readiness checks.

## Error Handling

The Calico installer should fail fast when:

- the kubeconfig path is missing
- `kubectl` is not available
- Calico version is not the bundled version
- the Kubernetes API does not support `MutatingAdmissionPolicy`
- embedded assets cannot be written
- manifest patching fails
- any `kubectl apply` command fails
- readiness polling times out

Errors should include the plugin name, bundled version, and failing phase. Command failures should preserve stderr through the existing bootstrap log path.

## Testing

Add or update tests for:

- Calico default version resolves to `3.32.0`.
- `v3.32.0` and `3.32.0` are accepted.
- another Calico version fails with a bundled-version error.
- the fake runner sees `kubectl apply` for `v3_projectcalico_org.yaml`, operator, and patched eBPF custom resources in order.
- Helm commands are no longer emitted for Calico.
- Cilium and Kube-OVN command plans remain unchanged.
- dry-run for `--step openstack-install-network-plugin` shows the bundled Calico eBPF plan.
- `--from-step openstack-install-network-plugin` still starts at the CNI step.
- readiness uses Tigera/Calico checks rather than Helm release status.
- embedded asset tests prove the expected Calico files are present.

Run the relevant verification after implementation:

```sh
go test ./internal/config/v2 ./internal/cluster ./cmd
mise run schema-verify
```

## Documentation

Update OpenStack deploy and networking docs to state:

- OpenStack Calico uses the bundled Tigera on-premises eBPF manifest flow.
- The bundled Calico version is `v3.32.0`.
- Calico install does not need internet access for manifests.
- Image pulls still depend on whatever image registry access or mirroring the target cluster has.
- Kubespray does not install OpenStack CNIs.
- Only one network plugin may be enabled.

Troubleshooting should add a note for version mismatch errors and Calico readiness failures involving `tigerastatus`.

## Open Questions Resolved

The version strategy is resolved as Option 1: pin to Calico `v3.32.0` and bundle those assets for offline installation.
