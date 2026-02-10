# GitOps Template Parsing Fix

## Summary

All GitOps template parsing errors have been successfully resolved. Templates now parse correctly without "bad character U+002D '-'" errors.

## Problem Description

Multiple GitOps tests were failing with template parsing errors indicating "bad character U+002D '-'". This error suggested invalid Go template syntax, specifically issues with hyphenated field names.

### Failed Tests

1. `TestRenderInfrastructureClusterRendersConfigValues`
2. `TestRenderClusterAppsRendersClusterName`
3. `TestRenderClusterAppsSkipsDisabledServices`
4. `TestRenderClusterAppsAtomic`
5. `TestShouldSkipFile_DisabledServiceSources`
6. `TestKubeletRotateServerCertsRendering`
7. `TestKubeletRotateServerCertsDefaultValue`

### Error Pattern

```
bad character U+002D '-' in template parsing
```

This error occurs when Go templates try to access struct fields with hyphens, which is invalid syntax.

## Root Causes

1. **Invalid struct field paths**: Templates were using non-existent configuration paths
2. **Hyphenated field access**: Attempting to access fields like `.Openstack-ccm` (invalid)
3. **Typos in template syntax**: `|| defauult` instead of `| default`
4. **Incorrect configuration structure**: Using `.OpenCenter.OpenStack.*` instead of `.OpenCenter.Infrastructure.Cloud.OpenStack.*`

## Solution

### Fixed Templates

#### 1. `openstack-ccm/helm-values/override-values.yaml.tpl`

**Issues**:
- Used `.OpenCenter.OpenStack.*` instead of `.OpenCenter.Infrastructure.Cloud.OpenStack.*`
- Had invalid hyphenated field access (`.Openstack-ccm`)
- Typo: `|| defauult` instead of `| default`

**Fixes**:
```yaml
# Before (INCORRECT)
auth-url: {{ .OpenCenter.OpenStack.AuthURL }}
domain-name: {{ .OpenCenter.OpenStack.Domain || defauult "rackspace_cloud_domain" }}

# After (CORRECT)
auth-url: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL }}
domain-name: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Domain | default "rackspace_cloud_domain" }}
```

#### 2. `openstack-csi/helm-values/override-values.yaml.tpl`

**Issues**:
- Incorrect OpenStack configuration paths
- Missing Infrastructure.Cloud prefix

**Fixes**:
```yaml
# Before (INCORRECT)
auth-url = {{ .OpenCenter.OpenStack.AuthURL }}

# After (CORRECT)
auth-url = {{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL }}
```

#### 3. `velero/velero-credentials.yaml.tpl`

**Issues**:
- Incorrect OpenStack paths
- Missing Infrastructure.Cloud prefix

**Fixes**:
```yaml
# Before (INCORRECT)
OS_AUTH_URL: {{ .OpenCenter.OpenStack.AuthURL }}

# After (CORRECT)
OS_AUTH_URL: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.AuthURL }}
```

#### 4. `velero/helm-values/override-values.yaml.tpl`

**Issues**:
- Used `.OpenCenter.Cluster.Name` instead of `.OpenCenter.Meta.Name`
- Incorrect OpenStack region path

**Fixes**:
```yaml
# Before (INCORRECT)
bucket: {{ .OpenCenter.Cluster.Name }}-velero
region: {{ .OpenCenter.OpenStack.Region }}

# After (CORRECT)
bucket: {{ .OpenCenter.Meta.Name }}-velero
region: {{ .OpenCenter.Infrastructure.Cloud.OpenStack.Region }}
```

## Verification

### Template Parsing
✅ All templates now parse without errors
✅ No "bad character U+002D '-'" errors
✅ Valid Go template syntax throughout

### Remaining Test Failures

Some GitOps tests still fail, but for different reasons:
- Missing `main.tf` file (separate issue)
- Missing `sources.yaml` file (separate issue)

These are **not** template parsing errors - they're missing file issues that need separate investigation.

## Impact Assessment

### Before Fix
- ❌ 7 GitOps tests failing with template parsing errors
- ❌ Templates couldn't be rendered
- ❌ Cluster setup blocked
- ❌ Invalid Go template syntax

### After Fix
- ✅ Template parsing errors eliminated
- ✅ Templates render successfully
- ✅ Valid Go template syntax
- ⚠️ Some tests still fail (different issues)

## Lessons Learned

1. **Go template field access**: Cannot use hyphens in field names (`.field-name` is invalid)
2. **Configuration structure**: Must match actual struct hierarchy
3. **Template syntax**: Use `| default` not `|| default`
4. **Testing**: Template parsing errors are distinct from missing file errors
5. **Validation**: Always verify template syntax before deployment

## Related Issues

The following test failures remain but are **unrelated** to template parsing:
- Missing `main.tf` template file
- Missing `sources.yaml` file
- File path resolution issues

These require separate investigation and fixes.

## Next Steps

1. ✅ Template parsing fixed (DONE)
2. ⚠️ Investigate missing file issues (TODO)
3. ⚠️ Fix remaining GitOps test failures (TODO)

---

**Fixed**: 2026-02-04
**Verified**: Template parsing successful
**Status**: ✅ Template syntax resolved
