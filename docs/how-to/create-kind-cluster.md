---
id: create-kind-cluster
title: "Create an openCenter Cluster with Kind"
sidebar_label: Create Kind Cluster
description: Create a local openCenter Kubernetes cluster using Kind with optional local Gitea for GitOps.
doc_type: how-to
audience: "platform engineers, developers"
tags: [kind, cluster, local, gitea, bootstrap, gitops]
---

# Create an openCenter Cluster with Kind

**Purpose:** For platform engineers and developers, shows how to create a local openCenter cluster using Kind, covering Gitea setup, cluster initialization, and FluxCD bootstrap.

This guide walks through the full lifecycle: local Gitea server, cluster init, setup, bootstrap, and FluxCD reconciliation against a local Git repository.

## Prerequisites

- openCenter CLI built (`mise run build`)
- Podman or Docker installed and running
- `kind` CLI installed (`mise install` handles this)
- `flux` CLI installed (for the FluxCD bootstrap step)
- 8 GB RAM available minimum

Verify tools:

```bash
mise run build
./bin/opencenter version
podman --version   # or docker --version
kind --version
flux --version
```

## Step 1: Start Local Gitea Server

openCenter ships scripts in `hack/gitea-local/` to run a disposable Gitea instance. The `mise` tasks wrap these scripts.

Start Gitea and create a user, token, and repository in one shot:

```bash
mise run gitea-up
```

This runs two tasks sequentially:

1. `gitea-setup` — pulls the Gitea container image, generates self-signed TLS certs, and starts the container on ports 3000 (HTTP), 3001 (HTTPS), and 2222 (SSH).
2. `gitea-configure` — creates an admin token, a `newuser` account, a user API token, and a `test-repo` repository.

After completion, the script prints a summary with tokens and the repository URL. Token files are saved to the repo root:

| File | Content |
|---|---|
| `.gitea_admin_token` | Admin API token |
| `.gitea_newuser_token` | User API token |

Verify Gitea is running:

```bash
curl -sk https://localhost:3001/api/v1/version
# {"version":"1.24.x"}
```

## Step 2: Initialize the Cluster Configuration

Create a Kind cluster configuration using the CLI:

```bash
./bin/opencenter cluster init my-cluster \
  --org local \
  --type kind
```

This creates:

- Configuration file at the customer GitOps directory (e.g., `customers/local/.my-cluster-config.yaml`)
- SOPS Age encryption keys in `customers/local/secrets/age/`
- SSH key pair in `customers/local/secrets/ssh/`

### Adjust Configuration Defaults

Edit the generated config to set values appropriate for a local Kind cluster:

```bash
$EDITOR ~/.config/opencenter/clusters/local/.my-cluster-config.yaml
```

Key fields to check:

| Field | Recommended Value | Why |
|---|---|---|
| `kubernetes.version` | A version with a published `kindest/node` image (e.g., `1.33.7`) | Kind pulls `kindest/node:v<version>`. If the tag doesn't exist, bootstrap fails. |
| `kubernetes.api_port` | `6443` | Port 443 requires root privileges on macOS/Linux. |
| `compute.master_count` | `1` | Single control plane is sufficient for local dev. |
| `compute.worker_count` | `2` | Two workers give enough room for service scheduling. |

Check available `kindest/node` tags if unsure:

```bash
# Using podman
podman search --list-tags docker.io/kindest/node --limit 200 | grep "v1.3"

# Using docker
docker image ls kindest/node
```

### Point GitOps URLs to Local Gitea (Optional)

If you want FluxCD to reconcile from the local Gitea, update the `gitops` section in the config:

```yaml
gitops:
  git_url: https://localhost:3001/newuser/test-repo.git
```

This step is optional. Without it, the cluster still boots; FluxCD just won't have a reachable Git source until you configure one.

## Step 3: Select the Cluster

Make the new cluster active so subsequent commands target it:

```bash
./bin/opencenter cluster select my-cluster
```

## Step 4: Validate Configuration

Catch errors before deploying:

```bash
./bin/opencenter cluster validate my-cluster
```

Validation checks schema compliance, required fields, network CIDR validity, and SOPS key presence. Fix any reported errors before proceeding.

## Step 5: Generate the GitOps Repository

Generate infrastructure templates, FluxCD manifests, and application overlays:

```bash
./bin/opencenter cluster setup my-cluster --force
```

`--force` overwrites any previously generated files. The command creates the full directory tree under the customer GitOps directory:

```
customers/local/
├── applications/overlays/my-cluster/
│   ├── flux-system/           # FluxCD bootstrap manifests
│   ├── services/              # Platform service Kustomizations
│   └── managed-services/      # Customer application Kustomizations
├── infrastructure/clusters/my-cluster/
│   ├── kind-config.yaml       # Kind cluster definition
│   └── kubeconfig.yaml        # Written during bootstrap
└── secrets/
    ├── age/                   # SOPS Age keys
    └── ssh/                   # SSH key pair
```

Verify the generated `kind-config.yaml` matches your expectations:

```bash
cat customers/local/infrastructure/clusters/my-cluster/kind-config.yaml
```

## Step 6: Bootstrap the Cluster

Create the Kind cluster and provision infrastructure:

```bash
./bin/opencenter cluster bootstrap my-cluster --container-runtime podman
```

Replace `podman` with `docker` if you use Docker.

This command:

1. Creates the Kind cluster using the generated `kind-config.yaml`.
2. Exports the kubeconfig to the infrastructure directory.

Bootstrap takes 1–3 minutes depending on whether the `kindest/node` image is already cached.

Verify the cluster is running:

```bash
export KUBECONFIG=customers/local/infrastructure/clusters/my-cluster/kubeconfig.yaml
kubectl get nodes
```

Expected output (1 control-plane + 2 workers):

```
NAME                       STATUS   ROLES           AGE   VERSION
my-cluster-control-plane   Ready    control-plane   30s   v1.33.7
my-cluster-worker          Ready    <none>          20s   v1.33.7
my-cluster-worker2         Ready    <none>          20s   v1.33.7
```

## Step 7: Push GitOps Content to Gitea

Commit the generated manifests and push to the local Gitea repository:

```bash
cd customers/local

# Add Gitea as remote (use token-based auth for HTTPS)
git remote add origin \
  https://newuser:$(cat ../../openCenter-cli/.gitea_newuser_token)@localhost:3001/newuser/test-repo.git

# Push (force because Gitea has an initial commit)
GIT_SSL_NO_VERIFY=true git push --force origin main
```

`GIT_SSL_NO_VERIFY=true` is needed because Gitea uses a self-signed certificate.

## Step 8: Bootstrap FluxCD Against Local Gitea

Install FluxCD and point it at the local Gitea repo:

```bash
flux bootstrap git \
  --url=https://localhost:3001/newuser/test-repo.git \
  --branch=main \
  --path=applications/overlays/my-cluster \
  --token-auth \
  --password=$(cat ../../openCenter-cli/.gitea_newuser_token) \
  --username=newuser \
  --ca-file=../../openCenter-cli/gitea/gitea/certs/ca.pem
```

This installs the FluxCD controllers (source-controller, kustomize-controller, helm-controller, notification-controller) and creates the `flux-system` GitRepository and Kustomization.

### Fix In-Cluster Connectivity to Gitea

FluxCD runs inside the Kind cluster. The source-controller cannot reach `localhost:3001` because that resolves to the container's own loopback, not the host.

Connect the Gitea container to the Kind network:

```bash
podman network connect kind gitea    # or: docker network connect kind gitea
```

Get the Gitea container's IP on the Kind network:

```bash
podman inspect gitea \
  --format '{{range $name, $net := .NetworkSettings.Networks}}{{$name}}: {{$net.IPAddress}}{{"\n"}}{{end}}'
# Example output:
# kind: 10.89.0.9
# podman: 10.88.0.6
```

The TLS certificate must include this IP as a Subject Alternative Name. Regenerate the certificate if needed (see [Troubleshooting](#tls-certificate-does-not-cover-container-ip) below), then restart Gitea:

```bash
podman restart gitea   # or: docker restart gitea
```

After restart, the IP may change. Re-check it and update the FluxCD GitRepository:

```bash
GITEA_IP=$(podman inspect gitea \
  --format '{{(index .NetworkSettings.Networks "kind").IPAddress}}')

# Patch the GitRepository URL
kubectl -n flux-system patch gitrepository flux-system \
  --type=merge \
  -p "{\"spec\":{\"url\":\"https://${GITEA_IP}:3001/newuser/test-repo.git\"}}"

# Update the CA cert in the flux-system secret
CA_B64=$(base64 < ../../openCenter-cli/gitea/gitea/certs/ca.pem)
kubectl -n flux-system patch secret flux-system \
  --type=merge \
  -p "{\"data\":{\"ca.crt\":\"${CA_B64}\"}}"
```

Also update `gotk-sync.yaml` in the repo so FluxCD doesn't revert the URL on the next reconciliation:

```bash
sed -i '' "s|url: https://localhost:3001|url: https://${GITEA_IP}:3001|" \
  applications/overlays/my-cluster/flux-system/gotk-sync.yaml

git add -A && git commit -m "fix: use gitea container IP for in-cluster access"
GIT_SSL_NO_VERIFY=true git push origin main
```

Trigger reconciliation:

```bash
flux reconcile source git flux-system -n flux-system
```

## Verification

Check that FluxCD is reconciling from the local Gitea:

```bash
flux get sources git -n flux-system
# READY: True, stored artifact for revision 'main@sha1:...'

flux get kustomizations -n flux-system
# flux-system should show READY: True
```

Check cluster health:

```bash
kubectl get nodes
kubectl get pods -n flux-system
```

## Cleanup

Tear everything down when you're done:

```bash
# Delete the Kind cluster
KIND_EXPERIMENTAL_PROVIDER=podman kind delete cluster --name my-cluster

# Destroy Gitea and its data
mise run gitea-cleanup
```

## Troubleshooting

### `kindest/node` Image Not Found

**Error:** `manifest unknown` when pulling `kindest/node:v<version>`

**Cause:** The Kubernetes version in your config doesn't have a published Kind image.

**Fix:** List available tags and pick one that exists:

```bash
podman search --list-tags docker.io/kindest/node --limit 200 | grep "v1.3"
```

Update `kubernetes.version` in your cluster config, then re-run `opencenter cluster setup --force` and `opencenter cluster bootstrap --restart`.

### Port 443 Permission Denied

**Error:** `listen tcp 127.0.0.1:443: bind: permission denied`

**Cause:** `api_port` is set to 443, which requires root.

**Fix:** Set `kubernetes.api_port: 6443` in your cluster config, re-run setup and bootstrap.

### FluxCD Cannot Reach Gitea (`connection refused`)

**Error:** `dial tcp [::1]:3001: connect: connection refused`

**Cause:** The source-controller resolves `localhost` to the container's loopback, not the host.

**Fix:** Connect Gitea to the Kind network and patch the GitRepository URL to use the container IP. See [Step 8](#step-8-bootstrap-fluxcd-against-local-gitea).

### TLS Certificate Does Not Cover Container IP

**Error:** `x509: certificate is valid for 127.0.0.1, not 10.89.0.x`

**Cause:** The self-signed cert was generated for `localhost` only.

**Fix:** Regenerate the certificate with the container IP as a SAN. Include several IPs in the range to survive container restarts:

```bash
CERT_DIR="gitea/gitea/certs"

openssl genrsa -out "$CERT_DIR/ca-key.pem" 2048
openssl req -new -x509 -days 365 -key "$CERT_DIR/ca-key.pem" \
  -out "$CERT_DIR/ca.pem" -subj "/CN=Gitea-CA"

openssl genrsa -out "$CERT_DIR/key.pem" 2048
openssl req -new -key "$CERT_DIR/key.pem" \
  -out /tmp/server.csr -subj "/CN=localhost"

printf 'authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage=digitalSignature,nonRepudiation,keyEncipherment,dataEncipherment
subjectAltName=DNS:localhost,DNS:gitea,IP:127.0.0.1,IP:::1,IP:10.89.0.7,IP:10.89.0.8,IP:10.89.0.9,IP:10.89.0.10
' > /tmp/server.ext

openssl x509 -req -days 365 -in /tmp/server.csr \
  -CA "$CERT_DIR/ca.pem" -CAkey "$CERT_DIR/ca-key.pem" \
  -CAcreateserial -out "$CERT_DIR/cert.pem" -extfile /tmp/server.ext

rm -f /tmp/server.csr /tmp/server.ext "$CERT_DIR/ca-key.pem"
podman restart gitea
```

Then update the CA cert in the `flux-system` secret and reconcile.

### Kustomization Fails with `PodMonitor` CRD Missing

**Error:** `no matches for kind "PodMonitor" in version "monitoring.coreos.com/v1"`

**Cause:** The generated overlay includes a PodMonitor resource, but the prometheus-operator CRDs are not installed.

**Fix:** Remove the PodMonitor reference from the services kustomization:

```bash
sed -i '' '/podmonitor.yaml/d' \
  applications/overlays/my-cluster/services/fluxcd/kustomization.yaml

git add -A && git commit -m "fix: remove PodMonitor (no prometheus-operator CRDs)"
GIT_SSL_NO_VERIFY=true git push origin main
flux reconcile source git flux-system
```

### Service Kustomizations Show `GitRepository not found`

**Cause:** Service kustomizations reference `GitRepository` sources (e.g., `opencenter-cert-manager`) that point to the `openCenter-gitops-base` repository. These sources are not available in the local Gitea.

**Fix:** This is expected for a local-only setup. To resolve it, either:

- Mirror `openCenter-gitops-base` into the local Gitea and create matching `GitRepository` sources.
- Or point the service sources at the upstream GitHub repository (requires internet access from the cluster).

---

**Evidence:**

- Gitea setup scripts: `hack/gitea-local/setup-gitea.sh`, `hack/gitea-local/configure-gitea-user-tokens.sh`
- Mise tasks: `.mise.toml` (`gitea-setup`, `gitea-configure`, `gitea-up`, `gitea-cleanup`)
- Cluster init command: `cmd/cluster_init.go:35-97`
- Cluster setup command: `cmd/cluster_setup.go:28-60`
- Cluster bootstrap command: `cmd/cluster_bootstrap.go:35-70`
- Kind config generation: `customers/local/infrastructure/clusters/*/kind-config.yaml`
- FluxCD bootstrap: `flux bootstrap git` CLI
