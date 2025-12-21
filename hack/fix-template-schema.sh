#!/usr/bin/env bash
set -e

echo "Fixing template schema references..."

# Fix .OpenCenter.Cluster.ClusterRegion -> .OpenCenter.Meta.Region
find internal/gitops/templates -name "*.tpl" -type f -exec sed -i '' 's/\.OpenCenter\.Cluster\.ClusterRegion/\.OpenCenter\.Meta\.Region/g' {} \;

# Fix .OpenCenter.Cluster.ClusterOrganization -> .OpenCenter.Meta.Organization  
find internal/gitops/templates -name "*.tpl" -type f -exec sed -i '' 's/\.OpenCenter\.Cluster\.ClusterOrganization/\.OpenCenter\.Meta\.Organization/g' {} \;

# Fix cert-manager LetsEncryptEmail field
find internal/gitops/templates -name "*.tpl" -type f -exec sed -i '' 's/\.LetsEncryptEmail/\.Email/g' {} \;

# Fix any remaining ClusterRegion references
find internal/gitops/templates -name "*.tpl" -type f -exec sed -i '' 's/ClusterRegion/Region/g' {} \;

echo "Template schema references fixed!"