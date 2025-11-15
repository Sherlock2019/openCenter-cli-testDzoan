# `openCenter cluster preflight` - Run Preflight Checks

## Synopsis
```bash
openCenter cluster preflight [name]
```

## Description

Run preflight checks for required tools and provider-specific requirements. This command verifies that all necessary tools are installed and accessible, and performs provider-specific validation checks before cluster deployment.

If no cluster name is provided, runs preflight checks for the currently active cluster.

## Arguments

### `[name]`
- **Required/Optional**: Optional
- **Description**: Name of the cluster (format: `cluster` or `organization/cluster`). If not provided, uses the currently active cluster
- **Example**: `my-cluster` or `production/my-cluster`

## Options

### `-h, --help`
- **Description**: Display help information for this subcommand

## Preflight Checks

### Common Tool Checks

The command checks for the following required tools:

- **git** - Version control for GitOps repository
- **kubectl** - Kubernetes command-line tool
- **talosctl** - Talos Linux CLI (for Talos-based clusters)

### Provider-Specific Checks

#### OpenStack Provider
- OpenStack CLI tools availability
- Authentication endpoint accessibility
- Cloud credentials validation
- Network connectivity to OpenStack API

#### AWS Provider
- AWS CLI availability
- AWS credentials configuration
- Region accessibility
- IAM permissions validation

#### GCP Provider
- gcloud CLI availability
- GCP credentials configuration
- Project and region validation
- API enablement checks

#### Azure Provider
- Azure CLI availability
- Azure credentials configuration
- Subscription and resource group validation
- Service principal permissions

#### Kind Provider
- Docker or Podman availability
- Container runtime accessibility
- Kind CLI availability
- Local network configuration

## Examples

### Run preflight for active cluster
```bash
openCenter cluster preflight
```
Runs preflight checks for the currently active cluster.

### Run preflight for specific cluster
```bash
openCenter cluster preflight my-cluster
```
Runs preflight checks for the specified cluster.

### Run preflight for OpenStack cluster
```bash
openCenter cluster preflight production/openstack-cluster
```
Runs OpenStack-specific preflight checks.

### Run preflight for Kind cluster
```bash
openCenter cluster preflight dev/kind-cluster
```
Runs Kind-specific preflight checks.

## Output

### Successful Preflight

```
git: OK
kubectl: OK
talosctl: OK
OpenStack: Checking auth endpoint https://openstack.example.com:5000/v3
OpenStack: Authentication successful
OpenStack: Network connectivity OK
Preflight complete.
```

### Failed Preflight

```
git: OK
kubectl: MISSING
talosctl: MISSING
OpenStack: Checking auth endpoint https://openstack.example.com:5000/v3
OpenStack: ERROR - Connection timeout
Preflight complete.
```

## Tool Requirements

### Required Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| git | GitOps repository management | `apt install git` or `brew install git` |
| kubectl | Kubernetes cluster management | [Install kubectl](https://kubernetes.io/docs/tasks/tools/) |
| talosctl | Talos Linux management | [Install talosctl](https://www.talos.dev/latest/introduction/getting-started/) |

### Provider-Specific Tools

#### OpenStack
- `openstack` CLI
- `python-openstackclient`

#### AWS
- `aws` CLI
- AWS credentials configured

#### GCP
- `gcloud` CLI
- GCP credentials configured

#### Azure
- `az` CLI
- Azure credentials configured

#### Kind
- `kind` CLI
- `docker` or `podman`

## Exit Codes

- `0` - Preflight checks completed (may include warnings)
- `1` - Error running preflight checks

## Notes

- Preflight checks do not fail on missing tools, they report status
- Provider-specific checks are performed based on the infrastructure provider
- The command checks tool availability in the system PATH
- OpenStack checks validate authentication endpoint accessibility
- Kind checks verify container runtime availability
- Missing tools are reported as "MISSING" in the output
- The command does not install missing tools automatically
- Run preflight checks before `cluster setup` and `cluster bootstrap`
- Preflight checks help identify issues early in the deployment process

## Troubleshooting

### kubectl not found
**Output**: `kubectl: MISSING`

**Solution**: Install kubectl:
```bash
# macOS
brew install kubectl

# Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

### talosctl not found
**Output**: `talosctl: MISSING`

**Solution**: Install talosctl:
```bash
curl -sL https://talos.dev/install | sh
```

### OpenStack authentication failed
**Output**: `OpenStack: ERROR - Authentication failed`

**Solution**: Check OpenStack credentials:
```bash
# Verify credentials are set
env | grep OS_

# Test authentication
openstack token issue
```

### Container runtime not available
**Output**: `docker: MISSING`

**Solution**: Install Docker or Podman:
```bash
# Docker
curl -fsSL https://get.docker.com | sh

# Podman
brew install podman  # macOS
apt install podman   # Linux
```

## See Also

- `openCenter cluster validate` - Validate cluster configuration
- `openCenter cluster setup` - Setup GitOps repository
- `openCenter cluster bootstrap` - Bootstrap cluster
