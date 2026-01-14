package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCompositionIntegration_CompleteWorkflow tests the complete composition workflow
// from template registration through composition to final rendering
func TestCompositionIntegration_CompleteWorkflow(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()

	// Create base cluster template
	baseTemplate := filepath.Join(tmpDir, "base-cluster.tmpl")
	baseContent := `apiVersion: v1
kind: Cluster
metadata:
  name: {{ .ClusterName }}
  organization: {{ .Organization }}
spec:
  provider: {{ .Provider }}
  version: {{ .KubernetesVersion }}`
	err := os.WriteFile(baseTemplate, []byte(baseContent), 0644)
	require.NoError(t, err)

	// Create networking overlay
	networkingOverlay := filepath.Join(tmpDir, "networking.tmpl")
	networkingContent := `networking:
  podCIDR: {{ .PodCIDR }}
  serviceCIDR: {{ .ServiceCIDR }}
  cni: {{ .CNI }}`
	err = os.WriteFile(networkingOverlay, []byte(networkingContent), 0644)
	require.NoError(t, err)

	// Create storage overlay
	storageOverlay := filepath.Join(tmpDir, "storage.tmpl")
	storageContent := `storage:
  storageClass: {{ .StorageClass }}
  volumeSize: {{ .VolumeSize }}`
	err = os.WriteFile(storageOverlay, []byte(storageContent), 0644)
	require.NoError(t, err)

	// Setup engine and registry
	engine := NewGoTemplateEngine()
	registry := NewInMemoryTemplateRegistry()

	// Register templates
	err = registry.RegisterTemplate(TemplateDefinition{
		Name:     "base-cluster",
		Path:     baseTemplate,
		Type:     TemplateTypeBase,
		Provider: "openstack",
		Metadata: TemplateMetadata{
			Description: "Base cluster configuration",
			Version:     "1.0.0",
			Priority:    0,
		},
	})
	require.NoError(t, err)

	err = registry.RegisterTemplate(TemplateDefinition{
		Name:     "networking",
		Path:     networkingOverlay,
		Type:     TemplateTypeOverlay,
		Services: []string{"calico"},
		Metadata: TemplateMetadata{
			Description: "Networking configuration overlay",
			Version:     "1.0.0",
			Priority:    2,
		},
	})
	require.NoError(t, err)

	err = registry.RegisterTemplate(TemplateDefinition{
		Name:     "storage",
		Path:     storageOverlay,
		Type:     TemplateTypeOverlay,
		Services: []string{"ceph"},
		Metadata: TemplateMetadata{
			Description: "Storage configuration overlay",
			Version:     "1.0.0",
			Priority:    1,
		},
	})
	require.NoError(t, err)

	// Create composition
	composition := TemplateComposition{
		BaseTemplate: baseTemplate,
		Overlays: []TemplateOverlay{
			{
				Name:     "networking",
				Path:     networkingOverlay,
				Priority: 2,
			},
			{
				Name:     "storage",
				Path:     storageOverlay,
				Priority: 1,
			},
		},
	}

	// Prepare data
	data := map[string]interface{}{
		"ClusterName":       "production-cluster",
		"Organization":      "acme-corp",
		"Provider":          "openstack",
		"KubernetesVersion": "1.28.0",
		"PodCIDR":           "10.244.0.0/16",
		"ServiceCIDR":       "10.96.0.0/12",
		"CNI":               "calico",
		"StorageClass":      "ceph-rbd",
		"VolumeSize":        "100Gi",
	}

	// Compose templates
	composer := NewDefaultTemplateComposer(engine, registry)
	result, err := composer.Compose(context.Background(), composition, data)
	require.NoError(t, err)

	// Verify result contains all expected content
	resultStr := string(result)

	// Base template content
	assert.Contains(t, resultStr, "apiVersion: v1")
	assert.Contains(t, resultStr, "kind: Cluster")
	assert.Contains(t, resultStr, "name: production-cluster")
	assert.Contains(t, resultStr, "organization: acme-corp")
	assert.Contains(t, resultStr, "provider: openstack")
	assert.Contains(t, resultStr, "version: 1.28.0")

	// Networking overlay content (priority 2, applied first)
	assert.Contains(t, resultStr, "networking:")
	assert.Contains(t, resultStr, "podCIDR: 10.244.0.0/16")
	assert.Contains(t, resultStr, "serviceCIDR: 10.96.0.0/12")
	assert.Contains(t, resultStr, "cni: calico")

	// Storage overlay content (priority 1, applied second)
	assert.Contains(t, resultStr, "storage:")
	assert.Contains(t, resultStr, "storageClass: ceph-rbd")
	assert.Contains(t, resultStr, "volumeSize: 100Gi")

	// Verify order: base, then networking (priority 2), then storage (priority 1)
	baseIdx := indexOfSubstring(resultStr, "kind: Cluster")
	networkingIdx := indexOfSubstring(resultStr, "networking:")
	storageIdx := indexOfSubstring(resultStr, "storage:")

	assert.True(t, baseIdx < networkingIdx, "base should come before networking overlay")
	assert.True(t, networkingIdx < storageIdx, "networking overlay should come before storage overlay")
}

// TestCompositionIntegration_ConditionalOverlays tests conditional overlay application
func TestCompositionIntegration_ConditionalOverlays(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()

	// Create base template
	baseTemplate := filepath.Join(tmpDir, "base.tmpl")
	baseContent := `environment: {{ .Environment }}`
	err := os.WriteFile(baseTemplate, []byte(baseContent), 0644)
	require.NoError(t, err)

	// Create production overlay
	prodOverlay := filepath.Join(tmpDir, "prod.tmpl")
	prodContent := `replicas: 3
monitoring: enabled`
	err = os.WriteFile(prodOverlay, []byte(prodContent), 0644)
	require.NoError(t, err)

	// Create development overlay
	devOverlay := filepath.Join(tmpDir, "dev.tmpl")
	devContent := `replicas: 1
debug: enabled`
	err = os.WriteFile(devOverlay, []byte(devContent), 0644)
	require.NoError(t, err)

	// Setup engine and registry
	engine := NewGoTemplateEngine()
	registry := NewInMemoryTemplateRegistry()
	composer := NewDefaultTemplateComposer(engine, registry)

	// Test production environment
	t.Run("production environment", func(t *testing.T) {
		composition := TemplateComposition{
			BaseTemplate: baseTemplate,
			Overlays: []TemplateOverlay{
				{
					Name:     "production",
					Path:     prodOverlay,
					Priority: 1,
					Conditions: []RenderCondition{
						{
							Type:  ConditionTypeEquals,
							Field: "Environment",
							Value: "production",
						},
					},
				},
				{
					Name:     "development",
					Path:     devOverlay,
					Priority: 1,
					Conditions: []RenderCondition{
						{
							Type:  ConditionTypeEquals,
							Field: "Environment",
							Value: "development",
						},
					},
				},
			},
		}

		data := map[string]interface{}{
			"Environment": "production",
		}

		result, err := composer.Compose(context.Background(), composition, data)
		require.NoError(t, err)

		resultStr := string(result)
		assert.Contains(t, resultStr, "environment: production")
		assert.Contains(t, resultStr, "replicas: 3")
		assert.Contains(t, resultStr, "monitoring: enabled")
		assert.NotContains(t, resultStr, "debug: enabled")
	})

	// Test development environment
	t.Run("development environment", func(t *testing.T) {
		composition := TemplateComposition{
			BaseTemplate: baseTemplate,
			Overlays: []TemplateOverlay{
				{
					Name:     "production",
					Path:     prodOverlay,
					Priority: 1,
					Conditions: []RenderCondition{
						{
							Type:  ConditionTypeEquals,
							Field: "Environment",
							Value: "production",
						},
					},
				},
				{
					Name:     "development",
					Path:     devOverlay,
					Priority: 1,
					Conditions: []RenderCondition{
						{
							Type:  ConditionTypeEquals,
							Field: "Environment",
							Value: "development",
						},
					},
				},
			},
		}

		data := map[string]interface{}{
			"Environment": "development",
		}

		result, err := composer.Compose(context.Background(), composition, data)
		require.NoError(t, err)

		resultStr := string(result)
		assert.Contains(t, resultStr, "environment: development")
		assert.Contains(t, resultStr, "replicas: 1")
		assert.Contains(t, resultStr, "debug: enabled")
		assert.NotContains(t, resultStr, "monitoring: enabled")
	})
}

// TestCompositionIntegration_WithPatches tests composition with patches
func TestCompositionIntegration_WithPatches(t *testing.T) {
	// Create test directory
	tmpDir := t.TempDir()

	// Create base template
	baseTemplate := filepath.Join(tmpDir, "base.tmpl")
	baseContent := `version: 1.0.0
features:
  - feature1
  - feature2`
	err := os.WriteFile(baseTemplate, []byte(baseContent), 0644)
	require.NoError(t, err)

	// Setup engine and registry
	engine := NewGoTemplateEngine()
	registry := NewInMemoryTemplateRegistry()
	composer := NewDefaultTemplateComposer(engine, registry)

	// Create composition with patches
	composition := TemplateComposition{
		BaseTemplate: baseTemplate,
		Patches: []TemplatePatch{
			{
				Operation: "replace",
				Path:      "version: 1.0.0",
				Value:     "version: 2.0.0",
			},
			{
				Operation: "add",
				Path:      "features",
				Value:     "  - feature3",
			},
		},
	}

	result, err := composer.Compose(context.Background(), composition, map[string]interface{}{})
	require.NoError(t, err)

	resultStr := string(result)
	assert.Contains(t, resultStr, "version: 2.0.0")
	assert.NotContains(t, resultStr, "version: 1.0.0")
	assert.Contains(t, resultStr, "feature1")
	assert.Contains(t, resultStr, "feature2")
	assert.Contains(t, resultStr, "feature3")
}
