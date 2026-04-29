# Cluster Deploy Dry-Run Plan Design

Date: 2026-04-28
Status: Approved for implementation planning

## Context

`opencenter cluster deploy --dry-run` currently loads and validates a cluster configuration, then skips the provider bootstrap path before provider steps are built. The resulting output says deploy completed with zero steps, which does not help users understand what a real deploy would do.

The improved experience should make dry-run a plan-only preview. It should show the resolved cluster identity, important paths, runtime artifact paths, and provider-specific steps that would run, while making it clear that no commands are executed and prerequisites are not fully simulated.

## Goals

- Make `opencenter cluster deploy --dry-run` verbose enough to explain the deploy plan.
- Show resolved paths where commands would run and where key files would be read or written.
- Build the dry-run plan from the configured provider type.
- Preserve the safety guarantee that dry-run does not provision infrastructure, create clusters, write files, update status, mutate Git repositories, or contact remote systems.
- Keep real deploy behavior unchanged except for refactors needed to expose plan metadata.

## Non-Goals

- Dry-run will not fully validate local prerequisites such as installed binaries, generated files, Docker or Podman availability, local Gitea state, OpenStack credentials, or cloud API connectivity.
- Dry-run will not execute provider step closures with a fake runner.
- Dry-run will not estimate remote resource changes beyond listing the commands and actions that a real deploy would attempt.
- This design does not add deploy implementations for providers that are not currently implemented.

## User Experience

Dry-run output should start with explicit plan-only language:

```text
Deploy plan only (dry-run)
No commands will be run, no files will be written, and prerequisites are not fully validated.
```

The summary should show:

- Cluster identifier, including organization when available.
- Provider.
- Cluster config path.
- GitOps directory.
- Infrastructure cluster directory.
- Cluster-owned kubeconfig path.
- Bootstrap log path that would be used.
- Bootstrap resume state path that would be used.
- Step filtering information when `--step`, `--from-step`, or `--restart` changes the plan.

Environment output should list only variable names or safe values. Secret-bearing values, especially provider credentials, must be redacted.

Example shape:

```text
Deploy plan only (dry-run)
No commands will be run, no files will be written, and prerequisites are not fully validated.

Cluster: opencenter/demo
Provider: kind
Config: /.../clusters/opencenter/.demo-config.yaml
GitOps dir: /.../clusters/opencenter
Cluster dir: /.../clusters/opencenter/infrastructure/clusters/demo
Kubeconfig: /.../clusters/opencenter/infrastructure/clusters/demo/kubeconfig.yaml
Log file would be: /.../state/logs/bootstrap/opencenter/demo/bootstrap-20260428T120000Z.log
Resume state would be: /.../state/bootstrap/opencenter/demo/state.json

Steps that would run:
  1. kind-create
     Action: Create Kind cluster
     Working dir: /.../clusters/opencenter/infrastructure/clusters/demo
     Commands:
       - kind get clusters
      - kind create cluster --name demo <kind config flag> /.../kind-config.yaml
     Reads:
       - /.../kind-config.yaml
     Writes:
       - local Kind cluster "demo"
     Notes:
       - Plan only; kind availability and config file existence were not checked.
```

Dry-run output must not use success wording such as `Deploy complete` because the deploy was not performed.

## Architecture

Keep dry-run planning inside the existing `internal/cluster` bootstrap flow.

Add plan metadata to `bootstrapStep` alongside the current executable closure. Provider step builders will still return `[]bootstrapStep`, but each step will include:

- `ID`
- `Description`
- `Run`
- `Plan`

The plan metadata should use small structs, for example:

- `bootstrapStepPlan`
- `bootstrapCommandPlan`
- `bootstrapPathPlan`
- `BootstrapPlan`

`BootstrapService.Bootstrap` should change dry-run behavior to:

1. Resolve cluster paths.
2. Load the cluster configuration.
3. Validate schema version and basic bootstrap config.
4. Resolve runtime paths without creating log or state files.
5. Build provider steps.
6. Apply `--step`, `--from-step`, and `--restart` filtering.
7. Populate `BootstrapResult.Plan`.
8. Return without invoking any step `Run` closure.

The CLI command should format `result.Plan` when `opts.DryRun` is true. Real deploy output should remain on the existing progress and completion path.

## Provider Behavior

Provider type must drive the planned steps and details.

### Kind

Kind dry-run should plan these steps:

- `kind-create`
  - Commands: `kind get clusters`, then `kind create cluster --name <cluster> <kind config flag> <clusterDir>/kind-config.yaml` when the cluster is absent
  - Reads: `<clusterDir>/kind-config.yaml`
  - Writes: local Kind cluster
- `kind-export-kubeconfig`
  - Command: `kind export kubeconfig --name <cluster> --kubeconfig <kubeconfig>`
  - Writes: cluster-owned kubeconfig directory and file
- `gitea-attach-kind`
  - Action: attach local Gitea to the Kind network
  - Notes that local Gitea status is not checked in plan-only dry-run
- `flux-bootstrap`
  - Action or command equivalent for bootstrapping Flux from local Gitea
- `gitea-rebase`
  - Action: rebase local checkout with Flux bootstrap commits
  - Working directory should show the resolved GitOps repo directory
- `gitops-push`
  - Action: push GitOps repository to local Gitea
- `flux-verify`
  - Commands: `flux check`, `flux get sources git -n flux-system`, `flux get kustomizations -n flux-system`
  - Environment: `KUBECONFIG=<kubeconfig>`

### OpenStack

OpenStack dry-run should plan these steps:

- `openstack-preflight`
  - Action: validate configured OpenStack credentials and bootstrap prerequisites
  - Notes that credentials are not fully verified in plan-only dry-run
- `opentofu-init`
  - Working directory: `<gitopsDir>/infrastructure/clusters/<cluster>`
  - Command: `<opentofu path or opentofu> init`
  - Environment: `KUBECONFIG`, `PATH`, and OpenStack credential variable names with secret values redacted
- `opentofu-apply`
  - Working directory: `<gitopsDir>/infrastructure/clusters/<cluster>`
  - Command: `<opentofu path or opentofu> apply -auto-approve`
  - Environment: `KUBECONFIG`, `PATH`, and OpenStack credential variable names with secret values redacted
- `openstack-normalize-kubeconfig`
  - Reads: kubeconfig candidates under the infrastructure cluster directory
  - Writes: cluster-owned kubeconfig path

The OpenStack provider currently checks the infrastructure directory and extracts credentials before returning steps. Those checks should move into the real execution closures, especially `openstack-preflight`, so dry-run can still describe the plan when credentials or generated infrastructure files are missing.

### Planned And Unsupported Providers

- `aws`, `gcp`, and `azure` should keep the current planned-provider rejection.
- Providers without a deploy implementation should fail with a clear message that deploy planning is not available for that provider.
- Empty or unknown providers should fail through validation-style errors.

## Locking And Side Effects

Dry-run should not acquire a deploy lock, prompt to break a lock, or break a lock. It is a read-only planning operation.

Dry-run should not:

- Create bootstrap log files.
- Create or delete bootstrap state files.
- Update cluster status.
- Check or auto-commit GitOps repository changes.
- Verify Git remotes.
- Execute any provider command.
- Call any step `Run` closure.

If lock inspection is added later, it should be non-invasive and reported as a note only.

## Error Handling

Dry-run should still fail for:

- Missing active cluster when no cluster argument is provided.
- Missing or unloadable cluster configuration.
- Invalid schema version.
- Basic bootstrap config validation failures, such as missing provider or missing Kind config block for a Kind cluster.
- Planned providers that are not available.
- Unsupported providers with no deploy plan.
- Invalid `--step` or `--from-step` values.

Dry-run should not fail solely because:

- A planned command binary is not installed.
- A referenced generated file is missing.
- Docker or Podman is unavailable.
- Local Gitea is not running.
- OpenStack credentials are incomplete.
- OpenTofu cannot be contacted or initialized.

Those cases should be reflected as warnings or notes when they are relevant to the provider plan.

## Testing

Add or update tests for:

- Kind dry-run prints provider-specific step IDs, commands, working directories, and path details.
- OpenStack dry-run prints OpenTofu working directories and commands without requiring credentials or an infrastructure directory.
- `--step` filters the dry-run plan to one step.
- `--from-step` filters the dry-run plan from the selected step onward.
- Invalid `--step` or `--from-step` values return a useful error.
- Dry-run does not create bootstrap log or state files.
- Dry-run does not acquire or break deploy locks.
- Dry-run does not invoke provider step `Run` closures.
- Real deploy behavior still executes steps and preserves progress, status, logging, and resume state behavior.

## Documentation

Update the generated or reference documentation for `opencenter cluster deploy` so `--dry-run` is described as a plan-only preview. The documentation should state that dry-run is not a full simulation and does not validate every local or remote prerequisite.
