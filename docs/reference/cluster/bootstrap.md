# `openCenter cluster bootstrap` - Run Provider-Specific Bootstrap Actions

## Synopsis
```bash
openCenter cluster bootstrap [name] [OPTIONS]
```

## Description

Run provider-specific bootstrap actions to create and configure the cluster infrastructure. This command executes the necessary steps to bring up a cluster based on the configured infrastructure provider (OpenStack, AWS, GCP, Azure, Kind, etc.).

The bootstrap process varies by provider but typically includes infrastructure provisioning, cluster creation, and initial configuration.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster (format: `cluster` or `organization/cluster`). If not provided, uses the currently active cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `--dry-run`
- **Description**: Show planned actions without executing them
- **Type**: Boolean
- **Default**: `false`

### `--kubeconfig <path>`
- **Description**: Path to kubeconfig file used by bootstrap actions
- **Type**: String
- **Default**: `./kubeconfig.yaml`

### `--log <path>`
- **Description**: Log file path (defaults to `<git_dir>/infrastructure/clusters/<name>/bootstrap.log`)
- **Type**: String
- **Default**: Auto-generated based on cluster directory

### `--container-runtime <runtime>`
- **Description**: Container runtime for Kind clusters (docker or podman)
- **Type**: String
- **Default**: Determined by `CONTAINER_RUNTIME` or `KIND_EXPERIMENTAL_PROVIDER` environment variables, falls back to `docker`
- **Valid Values**: `docker`, `podman`

### `-h, --help`
- **Description**: Display help information for this subcommand

## Provider-Specific Behavior

### OpenStack, AWS, GCP, Azure

For cloud providers, the bootstrap command runs `make` in the cluster's infrastructure directory:

**Requirements**:
- GitOps repository must be set up (`cluster setup`)
- Infrastructure directory must exist
- Makefile must be present in infrastructure directory

**Actions**:
- Executes `make` in `<git_dir>/infrastructure/clusters/<cluster>/`
- Provisions infrastructure using OpenTofu/Terraform
- Configures networking and security groups
- Creates cluster resources

### Kind

For Kind (Kubernetes in Docker) clusters:

**Requirements**:
- Docker or Podman must be installed and running
- Kind CLI must be available

**Actions**:
- Creates Kind cluster with specified configuration
- Disables default CNI for custom networking
- Exports kubeconfig for cluster access
- Configures multi-node cluster (1 control-plane + 3 workers)

**Configuration**:
```yaml
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
networking:
  disableDefaultCNI: true
  podSubnet: "10.244.0.0/16"
  serviceSubnet: "10.96.0.0/12"
nodes:
- role: control-plane
- role: worker
- role: worker
- role: worker
```

## Examples

### Bootstrap active cluster
```bash
openCenter cluster bootstrap
```

### Bootstrap specific cluster
```bash
openCenter cluster bootstrap my-cluster
```

### Dry run to see planned actions
```bash
openCenter cluster bootstrap my-cluster --dry-run
```

### Bootstrap with custom kubeconfig path
```bash
openCenter cluster bootstrap my-cluster --kubeconfig /path/to/kubeconfig
```

### Bootstrap Kind cluster with Podman
```bash
openCenter cluster bootstrap kind-cluster --container-runtime podman
```

### Bootstrap with custom log file
```bash
openCenter cluster bootstrap my-cluster --log /tmp/bootstrap.log
```

### Bootstrap Kind cluster using environment variable
```bash
CONTAINER_RUNTIME=podman openCenter cluster bootstrap kind-cluster
```

## Output

### OpenStack/Cloud Provider

```
Running make in /home/user/.config/openCenter/clusters/production/gitops/infrastructure/clusters/prod-cluster
$ make
terraform init
terraform plan
terraform apply -auto-approve
Bootstrap complete.
Log written to /home/user/.config/openCenter/clusters/production/gitops/infrastructure/clusters/prod-cluster/bootstrap.log
```

### Kind Provider

```
Creating kind cluster "my-cluster" using docker
$ KIND_EXPERIMENTAL_PROVIDER=docker kind create cluster --name my-cluster --config=-
Creating cluster "my-cluster" ...
 ✓ Ensuring node image (kindest/node:v1.31.0) 🖼
 ✓ Preparing nodes 📦 📦 📦 📦
 ✓ Writing configuration 📜
 ✓ Starting control-plane 🕹️
 ✓ Installing StorageClass 💾
 ✓ Joining worker nodes 🚜
Set kubectl context to "kind-my-cluster"
$ kind export kubeconfig --name my-cluster
Bootstrap complete.
Log written to /home/user/.config/openCenter/clusters/dev/gitops/infrastructure/clusters/my-cluster/bootstrap.log
```

### Dry Run

```
$ make
$ kind create cluster --name my-cluster --config=-
$ kind export kubeconfig --name my-cluster
Bootstrap complete.
```

## Bootstrap Log

The bootstrap log contains:

- Timestamp and cluster information
- All executed commands with environment variables
- Command output (stdout and stderr)
- Error messages and exit codes

Example log header:
```
# openCenter bootstrap log
# time: 2025-11-17T10:30:00Z
# cluster: prod-cluster
# dir: /home/user/.config/openCenter/clusters/production/gitops/infrastructure/clusters/prod-cluster
```

## Environment Variables

### CONTAINER_RUNTIME
Specifies the container runtime for Kind clusters:
```bash
export CONTAINER_RUNTIME=podman
openCenter cluster bootstrap kind-cluster
```

### KIND_EXPERIMENTAL_PROVIDER
Alternative way to specify Kind provider:
```bash
export KIND_EXPERIMENTAL_PROVIDER=podman
openCenter cluster bootstrap kind-cluster
```

### KUBECONFIG
Kubeconfig path for cluster access:
```bash
export KUBECONFIG=/path/to/kubeconfig
openCenter cluster bootstrap my-cluster
```

## Exit Codes

- `0` - Bootstrap successful
- `1` - Bootstrap failed or error occurred

## Notes

- Bootstrap must be run after `cluster setup`
- The command requires the GitOps repository to be initialized
- For cloud providers, ensure credentials are configured
- Kind clusters require Docker or Podman to be running
- Bootstrap logs are written to the infrastructure directory
- Use `--dry-run` to preview actions without execution
- The command waits for all operations to complete
- Failed bootstrap can be retried after fixing issues
- Kubeconfig is automatically exported for Kind clusters
- For cloud providers, check the Makefile for specific actions

## Troubleshooting

### GitOps directory not found
**Error**: `gitops.git_dir must be configured for provider "openstack"`

**Solution**: Run setup first:
```bash
openCenter cluster setup my-cluster
openCenter cluster bootstrap my-cluster
```

### Cluster infrastructure directory not found
**Error**: `cluster infrastructure directory not found in GitOps repository`

**Solution**: Ensure setup was successful:
```bash
openCenter cluster setup my-cluster --render
openCenter cluster bootstrap my-cluster
```

### Docker/Podman not running
**Error**: `Cannot connect to the Docker daemon`

**Solution**: Start Docker or Podman:
```bash
# Docker
sudo systemctl start docker

# Podman
podman machine start
```

### Kind cluster already exists
**Error**: `ERROR: failed to create cluster: node(s) already exist for a cluster with the name "my-cluster"`

**Solution**: Delete existing cluster:
```bash
kind delete cluster --name my-cluster
openCenter cluster bootstrap my-cluster
```

### Make command failed
**Error**: `command failed: make: *** [target] Error 1`

**Solution**: Check the bootstrap log for details:
```bash
cat <git_dir>/infrastructure/clusters/<cluster>/bootstrap.log
```

## See Also

- `openCenter cluster setup` - Setup GitOps repository before bootstrap
- `openCenter cluster preflight` - Run preflight checks before bootstrap
- `openCenter cluster validate` - Validate configuration before bootstrap
