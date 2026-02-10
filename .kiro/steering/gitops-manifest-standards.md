---
inclusion: always
---

# GitOps Manifest Standards

Standards and patterns for generating correct FluxCD Kustomization and GitRepository manifests in opencenter-cli.

## FluxCD Kustomization Standards

### Required Structure

All Kustomization manifests must follow this pattern:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: service-name
  namespace: flux-system
spec:
  interval: 5m
  path: ./applications/base/services/service-name
  prune: true
  wait: true
  sourceRef:
    kind: GitRepository
    name: opencenter-service-name
    namespace: flux-system
  healthChecks:
    - apiVersion: helm.toolkit.fluxcd.io/v2
      kind: HelmRelease
      name: service-name
      namespace: service-namespace
```

### Indentation Rules

**Critical**: Use 2-space indentation throughout. Common mistakes:

```yaml
# WRONG - no indentation
dependsOn:
- name: sources
namespace: flux-system

# CORRECT - proper 2-space nesting
dependsOn:
  - name: sources
    namespace: flux-system
```

### Interval Timing

- **Base kustomizations**: `interval: 5m` (fast reconciliation)
- **GitRepository sources**: `interval: 15m` (less frequent polling)

### Override Kustomizations

When generating overlay-specific kustomizations:

```yaml
apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: service-name-override
  namespace: flux-system
spec:
  interval: 5m
  path: ./applications/overlays/{{ .ClusterName }}/services/service-name
  prune: true
  dependsOn:
    - name: service-name
      namespace: flux-system
  sourceRef:
    kind: GitRepository
    name: opencenter-service-name
    namespace: flux-system
  healthChecks:
    - apiVersion: helm.toolkit.fluxcd.io/v2
      kind: HelmRelease
      name: service-name
      namespace: service-namespace
```

**Key requirements**:
- Must have `dependsOn` referencing base kustomization
- Path must use cluster name variable: `./applications/overlays/{{ .ClusterName }}/services/...`
- Include `healthChecks` for deployment monitoring
- Never hardcode cluster names like `dev-cluster` or `stage-cluster`

### Decryption Blocks

For services requiring SOPS decryption:

```yaml
spec:
  decryption:
    provider: sops
    secretRef:
      name: sops-age
```

**Indentation critical**: `secretRef` must be indented under `decryption`

### Wait Flags

Add `wait: true` to kustomizations that:
- Deploy infrastructure components (cert-manager, gateway-api)
- Have dependencies that must complete before proceeding
- Deploy CRDs that other resources depend on

## GitRepository Standards

### Repository URL Format

**Critical**: Repository name capitalization matters

```yaml
# CORRECT
url: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
                                        ^
                                  (capital C)

# WRONG
url: ssh://git@github.com/rackerlabs/opencenter-gitops-base.git
```

### Reference Type

Use branch-based references for active development:

```yaml
# CORRECT - branch reference
ref:
  branch: main

# WRONG - tag reference (unless pinning specific version)
ref:
  tag: v0.1.0
```

### Standard GitRepository Pattern

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-service-name
  namespace: flux-system
spec:
  interval: 15m
  url: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
  ref:
    branch: main
  secretRef:
    name: opencenter-base
```

### Config Repository Pattern

For service-specific configuration repositories:

```yaml
apiVersion: source.toolkit.fluxcd.io/v1
kind: GitRepository
metadata:
  name: opencenter-service-config
  namespace: flux-system
spec:
  interval: 15m
  url: ssh://git@github.com/rackerlabs/openCenter-gitops-base.git
  ref:
    branch: main
  secretRef:
    name: opencenter-base
  include:
    - repository:
        name: opencenter-service
```

**Key points**:
- Use `opencenter-base` secret (not `flux-system`)
- Proper indentation in `include` section
- Reference base repository, not example repos

## Service-Specific Patterns

### cert-manager

**Secrets encoding**:
```yaml
# CORRECT - base64 encoded
data:
  access-key-id: YWJjZGVmZ2g=
  secret-access-key: MTIzNDU2Nzg=

# WRONG - plaintext
data:
  access-key-id: abcdefgh
  secret-access-key: 12345678
```

**SOPS configuration**: Only encrypt `data` field, not `type` or `metadata`

**Issuer selectors**: Use full cluster domain
```yaml
# CORRECT
selector:
  dnsNames:
    - "*.fcc.k8s-qa.ord1.k8s.opencenter.cloud"

# WRONG
selector:
  dnsNames:
    - "*.k8s-qa.farmcreditfunding.com"
```

**kustomization.yaml secretGenerator**:
```yaml
# CORRECT indentation
secretGenerator:
  - name: cert-manager-values-override
    type: Opaque
    files: [override.yaml=helm-values/override-values.yaml]
    options:
      disableNameSuffixHash: true

# WRONG - options not indented
secretGenerator:
  - name: cert-manager-values-override
    type: Opaque
    files: [override.yaml=helm-values/override-values.yaml]
  options:
  disableNameSuffixHash: true
```

### gateway

**Hostname format**: Include organization prefix
```yaml
# CORRECT
hostname: auth.fcc.k8s-qa.ord1.k8s.opencenter.cloud

# WRONG - missing org prefix
hostname: auth.k8s-qa.ord1.k8s.opencenter.cloud
```

**HTTPRoute listeners**: Always include port 80 with redirect
```yaml
parentRefs:
  - name: gateway-name
    namespace: gateway-system
    sectionName: http
  - name: gateway-name
    namespace: gateway-system
    sectionName: https
```

### vsphere-csi

**Helm values requirements**:
```yaml
storageClass:
  default:
    datastoreURL: "ds:///vmfs/volumes/1375553-san-fc-hlu1-Gold"

snapshotter:
  image:
    registry: registry.k8s.io
    repository: sig-storage/csi-snapshotter
    tag: v8.2.0
```

**Storage class formatting**: No gaps between fields
```yaml
# CORRECT
allowVolumeExpansion: true
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: san-fc-hlu1-gold-retain
parameters:
  datastoreurl: ds:///vmfs/volumes/1375553-san-fc-hlu1-Gold
```

### metallb

**IPAddressPool**: Verify subnet matches node network
```yaml
spec:
  addresses:
    - 172.23.0.6-172.23.0.8  # Must match actual node subnet
```

## Template Generation Rules

### Path Variables

Always use template variables for cluster-specific paths:

```go
// CORRECT
path := fmt.Sprintf("./applications/overlays/%s/services/%s", clusterName, serviceName)

// WRONG - hardcoded cluster name
path := "./applications/overlays/dev-cluster/services/harbor"
```

### YAML Marshaling

When generating YAML from structs:

```go
// Use yaml.v3 with proper indentation
data, err := yaml.Marshal(kustomization)
if err != nil {
    return fmt.Errorf("marshaling kustomization: %w", err)
}

// Verify indentation before writing
if !validateIndentation(data) {
    return fmt.Errorf("invalid YAML indentation")
}
```

### Template Validation

Before writing manifests:

1. Validate YAML syntax
2. Check indentation (2 spaces)
3. Verify cluster name substitution
4. Confirm repository URLs
5. Check ref type (branch vs tag)

## Testing Generated Manifests

### Validation Commands

```bash
# YAML syntax validation
yamllint applications/overlays/{{ .ClusterName }}/services/fluxcd/*.yaml

# FluxCD validation
flux check

# Kustomization validation
flux get kustomizations -n flux-system

# Health check
flux get helmreleases --all-namespaces
```

### Common Issues Checklist

- [ ] All YAML uses 2-space indentation
- [ ] No hardcoded cluster names (dev-cluster, stage-cluster)
- [ ] Repository URLs use correct capitalization (openCenter)
- [ ] GitRepository refs use `branch: main` not `tag: v0.1.0`
- [ ] Override kustomizations have `dependsOn` on base
- [ ] Decryption blocks properly indented
- [ ] Secrets are base64 encoded
- [ ] Hostnames include organization prefix
- [ ] Intervals set correctly (5m for kustomizations, 15m for sources)
- [ ] `wait: true` on infrastructure components

## Code Review Requirements

When reviewing GitOps manifest generation code:

1. **Template correctness**: Verify all variables are substituted
2. **Indentation**: Check YAML marshaling produces 2-space indents
3. **Path construction**: Ensure cluster name is parameterized
4. **Repository references**: Confirm correct capitalization
5. **Dependency ordering**: Verify `dependsOn` chains are correct
6. **Health checks**: Confirm `healthChecks` are included where needed
7. **Testing**: Require actual manifest generation test with validation

## Related Documentation

- Service Registry Patterns: `.kiro/steering/service-registry-patterns.md`
- FluxCD Documentation: https://fluxcd.io/docs/
- Kustomize Documentation: https://kustomize.io/
