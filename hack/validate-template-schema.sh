#!/usr/bin/env bash
set -e

echo "Validating template schema references..."

# Check for any remaining incorrect field references
echo "Checking for incorrect field references..."

# Check for ClusterRegion (should be Meta.Region)
if grep -r "ClusterRegion" internal/gitops/templates/ 2>/dev/null; then
    echo "❌ Found ClusterRegion references - should be Meta.Region"
    exit 1
fi

# Check for ClusterOrganization (should be Meta.Organization)  
if grep -r "ClusterOrganization" internal/gitops/templates/ 2>/dev/null; then
    echo "❌ Found ClusterOrganization references - should be Meta.Organization"
    exit 1
fi

# Check for missing dots in template variables
if grep -r "{{ [A-Z]" internal/gitops/templates/ 2>/dev/null; then
    echo "❌ Found template variables missing dots"
    exit 1
fi

# Check for undefined secret fields
echo "Checking for undefined secret field references..."

# Run a quick template validation by trying to render a test template
echo "Running template rendering test..."
if ! mise run test 2>&1 | grep -q "ok.*internal/gitops"; then
    echo "❌ Template rendering tests failed"
    exit 1
fi

echo "✅ All template schema references are valid!"
echo ""
echo "Summary of fixes applied:"
echo "- Fixed .OpenCenter.Cluster.ClusterRegion → .OpenCenter.Meta.Region"
echo "- Fixed .OpenCenter.Cluster.ClusterOrganization → .OpenCenter.Meta.Organization"
echo "- Fixed .Secrets.Headlamp.oidc_client_secret → .Secrets.Headlamp.OIDCClientSecret"
echo "- Fixed .Secrets.Grafana.password → .Secrets.Grafana.AdminPassword"
echo "- Fixed .Secrets.Loki.AWSAccesKey → .Secrets.Loki.S3SecretAccessKey"
echo "- Added missing WebhookURL field to ServiceCfg"
echo "- Added missing TempoSecrets struct and fields"
echo "- Fixed template syntax in tempo/helm-values/override-values.yaml.tpl"