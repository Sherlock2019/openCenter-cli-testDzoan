# Validate GitOps Manifests

Guide for validating generated GitOps manifests to catch common configuration issues before deployment.

## Overview

The `cluster validate-manifests` command checks generated GitOps manifests for common issues documented in production deployments. It validates:

- YAML formatting and indentation
- FluxCD Kustomization configurations
- GitRepository source definitions
- Service-specific configurations
- Security best practices

## Prerequisites

- Cluster configuration initialized with `cluster init`
- GitOps repository generated with `cluster setup`

## Basic Usage

### Validate Current Cluster

```bash
opencenter cluster validate-manifests
```

### Validate Specific Cluster

```bash
opencenter cluster validate-manifests my-cluster
```

## Validation Checks

### FluxCD Kustomization Manifests

**Location**: `applications/overlays/{cluster}/services/fluxcd/*.yaml`

**Checks**:
- ✅ Proper 2-space indentation (no tabs)
- ✅ Interval set to `5m` for fast reconciliation
- ✅ No hardcoded cluster names (`dev-cluster`, `stage-cluster`)
- ✅ `dependsOn` blocks properly indented
- ✅ `decryption` blocks properly indented
- ✅ `wait: true` on infrastructure components

**Example Issues**:

```yaml
# ❌ WRONG - no indentation
dependsOn:
- name: sources
namespace: flux-system

# ✅ CORRECT - proper 2-space nesting
dependsOn:
  - name: sources
    namespace: flux-system
```

```yaml
# ❌ WRONG - hardcoded cluster name
path: ./applications/overlays/dev-cluster/services/harbor

# ✅ CORRECT - uses actual cluster name
path: ./applications/overlays/k8s-qa/services/harbor
```

### GitRepository Sources

**Location**: `applications/overlays/{cluster}/services/sources/*.yaml`

**Checks**:
- ✅ Repository URL uses correct capitalization (`openCenter-gitops-base`)
- ✅ Interval set to `15m` for less frequent polling
- ✅ Uses `branch: main` not `tag: v0.1.0`
- ✅ Proper indentation in `ref` and `secretRef` blocks
- ✅ Uses `opencenter-base` secret (not `flux-system`)

**Example Issues**:

```yaml
# ❌ WRONG - lowercase 'c'
url: ssh://git@github.com/rackerlabs/opencenter-gitops-base.git

# ✅ CORRECT - capital 'C'
url: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
```

```yaml
# ❌ WRONG - tag-based reference
ref:
  tag: v0.1.0

# ✅ CORRECT - branch-based reference
ref:
  branch: main
```

### cert-manager Configuration

**Location**: `applications/overlays/{cluster}/services/cert-manager/`

**Checks**:
- ✅ Secrets are base64 encoded (not plaintext)
- ✅ SOPS encrypts only `data` field (not `type` or `metadata`)
- ✅ Issuer selectors use correct domain format
- ✅ `kustomization.yaml` secretGenerator properly indented

**Example Issues**:

```yaml
# ❌ WRONG - plaintext secret
data:
  access-key-id: abcdefgh
  secret-access-key: 12345678

# ✅ CORRECT - base64 encoded
data:
  access-key-id: YWJjZGVmZ2g=
  secret-access-key: MTIzNDU2Nzg=
```

```yaml
# ❌ WRONG - incorrect domain
selector:
  dnsNames:
    - "*.k8s-qa.farmcreditfunding.com"

# ✅ CORRECT - full cluster domain
selector:
  dnsNames:
    - "*.fcc.k8s-qa.ord1.k8s.opencenter.cloud"
```

### Gateway Configuration

**Location**: `applications/overlays/{cluster}/services/gateway/`

**Checks**:
- ✅ Hostnames include organization prefix
- ✅ HTTPRoute has both port 80 and 443 listeners
- ✅ Port 80 redirects to 443

**Example Issues**:

```yaml
# ❌ WRONG - missing org prefix
hostname: auth.k8s-qa.ord1.k8s.opencenter.cloud

# ✅ CORRECT - includes org prefix
hostname: auth.fcc.k8s-qa.ord1.k8s.opencenter.cloud
```

```yaml
# ❌ WRONG - missing port 80 listener
parentRefs:
  - name: gateway-name
    namespace: gateway-system
    sectionName: https

# ✅ CORRECT - both http and https
parentRefs:
  - name: gateway-name
    namespace: gateway-system
    sectionName: http
  - name: gateway-name
    namespace: gateway-system
    sectionName: https
```

### vSphere CSI Configuration

**Location**: `applications/overlays/{cluster}/services/vsphere-csi/`

**Checks**:
- ✅ Snapshotter version is `v8.2.0` (not `v3.3.0`)
- ✅ Registry is `registry.k8s.io` (not `registry.k8s.io/csi-vsphere`)
- ✅ StorageClass includes `datastoreURL` parameter
- ✅ No formatting gaps in StorageClass definitions

**Example Issues**:

```yaml
# ❌ WRONG - old snapshotter version
snapshotter:
  image:
    tag: v3.3.0

# ✅ CORRECT - current version
snapshotter:
  image:
    tag: v8.2.0
```

```yaml
# ❌ WRONG - incorrect registry
snapshotter:
  image:
    registry: registry.k8s.io/csi-vsphere

# ✅ CORRECT - base registry
snapshotter:
  image:
    registry: registry.k8s.io
```

### MetalLB Configuration

**Location**: `applications/overlays/{cluster}/services/metallb/`

**Checks**:
- ✅ IPAddressPool addresses match node network subnet
- ✅ Valid IP range format (IP-IP or CIDR)

**Example Issues**:

```yaml
# ❌ WRONG - incorrect subnet
spec:
  addresses:
    - 172.23.0.6-172.23.0.8  # Doesn't match node subnet

# ✅ CORRECT - matches node network
spec:
  addresses:
    - 10.2.128.6-10.2.128.8
```

### Headlamp Configuration

**Location**: `applications/overlays/{cluster}/services/headlamp/`

**Checks**:
- ✅ HTTPRoute URL includes organization prefix
- ✅ Dashboard hostname format correct

**Example Issues**:

```yaml
# ❌ WRONG - missing org prefix
hostname: dashboard.k8s-qa.ord1.k8s.opencenter.cloud

# ✅ CORRECT - includes org prefix
hostname: dashboard.fcc.k8s-qa.ord1.k8s.opencenter.cloud
```

## Interpreting Results

### Successful Validation

```bash
$ opencenter cluster validate-manifests k8s-qa
Validating GitOps manifests in: /path/to/gitops/repo

✅ All manifests validated successfully
```

### Failed Validation

```bash
$ opencenter cluster validate-manifests k8s-qa
Validating GitOps manifests in: /path/to/gitops/repo

❌ Validation failed:

manifest validation failed with 5 errors:
applications/overlays/k8s-qa/services/fluxcd/harbor.yaml: interval should be 5m, got 15m
applications/overlays/k8s-qa/services/fluxcd/loki.yaml: contains hardcoded cluster name (dev-cluster or stage-cluster)
applications/overlays/k8s-qa/services/sources/opencenter-harbor.yaml: repository URL should use 'openCenter-gitops-base' (capital C)
applications/overlays/k8s-qa/services/cert-manager/secret.yaml: secret field 'access-key-id' is not base64 encoded
applications/overlays/k8s-qa/services/vsphere-csi/helm-values.yaml: snapshotter version should be v8.2.0, not v3.3.0
```

## Fixing Common Issues

### Fix Indentation

Use `yamllint` to check and fix indentation:

```bash
# Check indentation
yamllint applications/overlays/k8s-qa/services/fluxcd/*.yaml

# Auto-fix with yq (if available)
yq eval -i '.' file.yaml
```

### Fix Hardcoded Cluster Names

Search and replace hardcoded names:

```bash
# Find hardcoded cluster names
grep -r "dev-cluster\|stage-cluster" applications/overlays/k8s-qa/

# Replace with actual cluster name
find applications/overlays/k8s-qa/ -type f -name "*.yaml" \
  -exec sed -i 's/dev-cluster/k8s-qa/g' {} \;
```

### Fix Repository URLs

```bash
# Fix repository capitalization
find applications/overlays/k8s-qa/services/sources/ -type f -name "*.yaml" \
  -exec sed -i 's/opencenter-gitops-base/openCenter-gitops-base/g' {} \;
```

### Fix Intervals

```bash
# Fix Kustomization intervals (should be 5m)
find applications/overlays/k8s-qa/services/fluxcd/ -type f -name "*.yaml" \
  -exec sed -i 's/interval: 15m/interval: 5m/g' {} \;

# GitRepository intervals should remain 15m
```

### Encode Secrets

```bash
# Base64 encode a secret value
echo -n "my-secret-value" | base64

# Decode to verify
echo "bXktc2VjcmV0LXZhbHVl" | base64 -d
```

## Integration with CI/CD

Add manifest validation to your CI/CD pipeline:

```yaml
# .github/workflows/validate.yml
name: Validate Manifests

on:
  pull_request:
    paths:
      - 'applications/**/*.yaml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Install opencenter CLI
        run: |
          curl -L https://github.com/rackerlabs/opencenter-cli/releases/latest/download/opencenter-linux-amd64 -o opencenter
          chmod +x opencenter
          sudo mv opencenter /usr/local/bin/
      
      - name: Validate manifests
        run: opencenter cluster validate-manifests
```

## Pre-commit Hook

Add validation as a pre-commit hook:

```bash
# .git/hooks/pre-commit
#!/bin/bash

# Validate manifests before commit
if ! opencenter cluster validate-manifests; then
    echo "❌ Manifest validation failed. Fix errors before committing."
    exit 1
fi

echo "✅ Manifest validation passed"
```

Make it executable:

```bash
chmod +x .git/hooks/pre-commit
```

## Troubleshooting

### Validation Fails on Fresh Setup

If validation fails immediately after `cluster setup`, the templates may need updates. Check:

1. Template versions in `internal/gitops/templates/`
2. Default values in `internal/config/config.go`
3. Service configurations in `internal/config/services/`

### False Positives

Some validation checks may flag intentional configurations. Review the error and:

1. Verify the configuration is correct for your environment
2. If needed, temporarily disable specific checks (not recommended)
3. Report false positives as issues

### Performance Issues

For large repositories, validation may be slow. Optimize by:

1. Validating only changed files in CI/CD
2. Running validation in parallel for multiple clusters
3. Caching validation results

## Related Commands

- `cluster validate` - Validate cluster configuration (not manifests)
- `cluster setup` - Generate GitOps repository
- `cluster render` - Render templates without writing files

## Related Documentation

- [GitOps Manifest Standards](../../.kiro/steering/gitops-manifest-standards.md)
- [Service Registry Patterns](../../.kiro/steering/service-registry-patterns.md)
- [Lessons Learned](../../testdata/lessons-learned.md)
