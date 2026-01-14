package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInMemoryTemplateRegistry(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.templates)
	assert.Equal(t, 0, len(registry.templates))
}

func TestRegisterTemplate(t *testing.T) {
	tests := []struct {
		name        string
		template    TemplateDefinition
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid template",
			template: TemplateDefinition{
				Name: "test-template",
				Path: "/path/to/template",
				Type: TemplateTypeBase,
			},
			expectError: false,
		},
		{
			name: "empty name",
			template: TemplateDefinition{
				Path: "/path/to/template",
				Type: TemplateTypeBase,
			},
			expectError: true,
			errorMsg:    "template name cannot be empty",
		},
		{
			name: "empty path",
			template: TemplateDefinition{
				Name: "test-template",
				Type: TemplateTypeBase,
			},
			expectError: true,
			errorMsg:    "template path cannot be empty",
		},
		{
			name: "self-dependency",
			template: TemplateDefinition{
				Name:         "test-template",
				Path:         "/path/to/template",
				Type:         TemplateTypeBase,
				Dependencies: []string{"test-template"},
			},
			expectError: true,
			errorMsg:    "cannot depend on itself",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			registry := NewInMemoryTemplateRegistry()
			err := registry.RegisterTemplate(tt.template)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRegistryGetTemplate(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	template := TemplateDefinition{
		Name: "test-template",
		Path: "/path/to/template",
		Type: TemplateTypeBase,
		Metadata: TemplateMetadata{
			Description: "Test template",
			Version:     "1.0.0",
		},
	}

	err := registry.RegisterTemplate(template)
	require.NoError(t, err)

	t.Run("existing template", func(t *testing.T) {
		retrieved, err := registry.GetTemplate("test-template")
		assert.NoError(t, err)
		assert.Equal(t, template.Name, retrieved.Name)
		assert.Equal(t, template.Path, retrieved.Path)
		assert.Equal(t, template.Type, retrieved.Type)
		assert.Equal(t, template.Metadata.Description, retrieved.Metadata.Description)
	})

	t.Run("non-existing template", func(t *testing.T) {
		_, err := registry.GetTemplate("non-existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestGetTemplatesForProvider(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	templates := []TemplateDefinition{
		{
			Name:     "openstack-template",
			Path:     "/path/to/openstack",
			Type:     TemplateTypeInfrastructure,
			Provider: "openstack",
			Metadata: TemplateMetadata{Priority: 10},
		},
		{
			Name:     "aws-template",
			Path:     "/path/to/aws",
			Type:     TemplateTypeInfrastructure,
			Provider: "aws",
			Metadata: TemplateMetadata{Priority: 5},
		},
		{
			Name:     "universal-template",
			Path:     "/path/to/universal",
			Type:     TemplateTypeBase,
			Provider: "", // Universal
			Metadata: TemplateMetadata{Priority: 1},
		},
		{
			Name:     "another-openstack",
			Path:     "/path/to/another",
			Type:     TemplateTypeService,
			Provider: "openstack",
			Metadata: TemplateMetadata{Priority: 10},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	t.Run("openstack provider", func(t *testing.T) {
		result := registry.GetTemplatesForProvider("openstack")
		assert.Len(t, result, 3) // 2 openstack + 1 universal

		// Check that openstack templates are included
		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}
		assert.Contains(t, names, "openstack-template")
		assert.Contains(t, names, "universal-template")
		assert.Contains(t, names, "another-openstack")
	})

	t.Run("aws provider", func(t *testing.T) {
		result := registry.GetTemplatesForProvider("aws")
		assert.Len(t, result, 2) // 1 aws + 1 universal

		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}
		assert.Contains(t, names, "aws-template")
		assert.Contains(t, names, "universal-template")
	})

	t.Run("priority ordering", func(t *testing.T) {
		result := registry.GetTemplatesForProvider("openstack")
		// Higher priority should come first
		assert.True(t, result[0].Metadata.Priority >= result[len(result)-1].Metadata.Priority)
	})
}

func TestGetTemplatesForService(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	templates := []TemplateDefinition{
		{
			Name:     "prometheus-template",
			Path:     "/path/to/prometheus",
			Type:     TemplateTypeService,
			Services: []string{"prometheus", "monitoring"},
			Metadata: TemplateMetadata{Priority: 10},
		},
		{
			Name:     "grafana-template",
			Path:     "/path/to/grafana",
			Type:     TemplateTypeService,
			Services: []string{"grafana", "monitoring"},
			Metadata: TemplateMetadata{Priority: 5},
		},
		{
			Name:     "loki-template",
			Path:     "/path/to/loki",
			Type:     TemplateTypeService,
			Services: []string{"loki", "logging"},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	t.Run("monitoring service", func(t *testing.T) {
		result := registry.GetTemplatesForService("monitoring")
		assert.Len(t, result, 2)

		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}
		assert.Contains(t, names, "prometheus-template")
		assert.Contains(t, names, "grafana-template")
	})

	t.Run("logging service", func(t *testing.T) {
		result := registry.GetTemplatesForService("logging")
		assert.Len(t, result, 1)
		assert.Equal(t, "loki-template", result[0].Name)
	})

	t.Run("non-existing service", func(t *testing.T) {
		result := registry.GetTemplatesForService("non-existing")
		assert.Len(t, result, 0)
	})
}

func TestResolveTemplateDependencies(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	templates := []TemplateDefinition{
		{
			Name:         "base",
			Path:         "/path/to/base",
			Type:         TemplateTypeBase,
			Dependencies: []string{},
		},
		{
			Name:         "middleware",
			Path:         "/path/to/middleware",
			Type:         TemplateTypeService,
			Dependencies: []string{"base"},
		},
		{
			Name:         "app",
			Path:         "/path/to/app",
			Type:         TemplateTypeService,
			Dependencies: []string{"middleware", "base"},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	t.Run("resolve single template", func(t *testing.T) {
		result, err := registry.ResolveTemplateDependencies([]string{"base"})
		assert.NoError(t, err)
		assert.Len(t, result, 1)
		assert.Equal(t, "base", result[0].Name)
	})

	t.Run("resolve with dependencies", func(t *testing.T) {
		result, err := registry.ResolveTemplateDependencies([]string{"app"})
		assert.NoError(t, err)
		assert.Len(t, result, 3)

		// Dependencies should come before dependents
		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}

		baseIdx := indexOf(names, "base")
		middlewareIdx := indexOf(names, "middleware")
		appIdx := indexOf(names, "app")

		assert.True(t, baseIdx < middlewareIdx)
		assert.True(t, middlewareIdx < appIdx)
	})

	t.Run("non-existing template", func(t *testing.T) {
		_, err := registry.ResolveTemplateDependencies([]string{"non-existing"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestResolveTemplateDependencies_CircularDependency(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Create circular dependency: A -> B -> C -> A
	templates := []TemplateDefinition{
		{
			Name:         "template-a",
			Path:         "/path/to/a",
			Type:         TemplateTypeBase,
			Dependencies: []string{"template-b"},
		},
		{
			Name:         "template-b",
			Path:         "/path/to/b",
			Type:         TemplateTypeService,
			Dependencies: []string{"template-c"},
		},
		{
			Name:         "template-c",
			Path:         "/path/to/c",
			Type:         TemplateTypeService,
			Dependencies: []string{"template-a"},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	_, err := registry.ResolveTemplateDependencies([]string{"template-a"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "circular dependency")
}

func TestListTemplates(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	templates := []TemplateDefinition{
		{Name: "template-c", Path: "/c", Type: TemplateTypeBase},
		{Name: "template-a", Path: "/a", Type: TemplateTypeBase},
		{Name: "template-b", Path: "/b", Type: TemplateTypeBase},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	result := registry.ListTemplates()
	assert.Len(t, result, 3)

	// Should be sorted by name
	assert.Equal(t, "template-a", result[0].Name)
	assert.Equal(t, "template-b", result[1].Name)
	assert.Equal(t, "template-c", result[2].Name)
}

func TestUnregisterTemplate(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	templates := []TemplateDefinition{
		{
			Name: "base",
			Path: "/path/to/base",
			Type: TemplateTypeBase,
		},
		{
			Name:         "dependent",
			Path:         "/path/to/dependent",
			Type:         TemplateTypeService,
			Dependencies: []string{"base"},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	t.Run("cannot unregister with dependents", func(t *testing.T) {
		err := registry.UnregisterTemplate("base")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "depends on it")
	})

	t.Run("can unregister without dependents", func(t *testing.T) {
		err := registry.UnregisterTemplate("dependent")
		assert.NoError(t, err)

		_, err = registry.GetTemplate("dependent")
		assert.Error(t, err)
	})

	t.Run("non-existing template", func(t *testing.T) {
		err := registry.UnregisterTemplate("non-existing")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})
}

func TestConcurrentAccess(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Register initial template
	err := registry.RegisterTemplate(TemplateDefinition{
		Name: "base",
		Path: "/path/to/base",
		Type: TemplateTypeBase,
	})
	require.NoError(t, err)

	// Concurrent reads and writes
	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = registry.RegisterTemplate(TemplateDefinition{
				Name: "template-" + string(rune(i)),
				Path: "/path/" + string(rune(i)),
				Type: TemplateTypeBase,
			})
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = registry.GetTemplatesForProvider("openstack")
			_, _ = registry.GetTemplate("base")
		}
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Should not panic and should have templates
	templates := registry.ListTemplates()
	assert.NotEmpty(t, templates)
}

// Helper function
func indexOf(slice []string, item string) int {
	for i, v := range slice {
		if v == item {
			return i
		}
	}
	return -1
}

// TestTemplateRegistryManagesMetadataCorrectly validates the acceptance criterion:
// "Template registry manages all template metadata correctly"
// This is a comprehensive test that validates Property 6 from the design document.
func TestTemplateRegistryManagesMetadataCorrectly(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Test 1: Register template with complete metadata
	t.Run("stores complete metadata", func(t *testing.T) {
		template := TemplateDefinition{
			Name:         "complete-template",
			Path:         "/path/to/complete",
			Type:         TemplateTypeInfrastructure,
			Provider:     "openstack",
			Services:     []string{"prometheus", "grafana"},
			Dependencies: []string{"base-template"},
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: "openstack",
				},
			},
			Metadata: TemplateMetadata{
				Description: "Complete template with all metadata",
				Version:     "1.2.3",
				Author:      "Test Author",
				Tags:        []string{"monitoring", "infrastructure"},
				Priority:    10,
			},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Retrieve and verify all metadata is preserved
		retrieved, err := registry.GetTemplate("complete-template")
		require.NoError(t, err)

		assert.Equal(t, template.Name, retrieved.Name)
		assert.Equal(t, template.Path, retrieved.Path)
		assert.Equal(t, template.Type, retrieved.Type)
		assert.Equal(t, template.Provider, retrieved.Provider)
		assert.Equal(t, template.Services, retrieved.Services)
		assert.Equal(t, template.Dependencies, retrieved.Dependencies)
		assert.Equal(t, len(template.Conditions), len(retrieved.Conditions))
		assert.Equal(t, template.Metadata.Description, retrieved.Metadata.Description)
		assert.Equal(t, template.Metadata.Version, retrieved.Metadata.Version)
		assert.Equal(t, template.Metadata.Author, retrieved.Metadata.Author)
		assert.Equal(t, template.Metadata.Tags, retrieved.Metadata.Tags)
		assert.Equal(t, template.Metadata.Priority, retrieved.Metadata.Priority)
	})

	// Test 2: Retrieve by name
	t.Run("retrieves by name correctly", func(t *testing.T) {
		template := TemplateDefinition{
			Name: "named-template",
			Path: "/path/to/named",
			Type: TemplateTypeService,
			Metadata: TemplateMetadata{
				Description: "Template for name retrieval test",
			},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		retrieved, err := registry.GetTemplate("named-template")
		require.NoError(t, err)
		assert.Equal(t, "named-template", retrieved.Name)
		assert.Equal(t, "Template for name retrieval test", retrieved.Metadata.Description)
	})

	// Test 3: Filter by provider
	t.Run("filters by provider correctly", func(t *testing.T) {
		templates := []TemplateDefinition{
			{
				Name:     "aws-template-1",
				Path:     "/aws/1",
				Type:     TemplateTypeInfrastructure,
				Provider: "aws",
			},
			{
				Name:     "aws-template-2",
				Path:     "/aws/2",
				Type:     TemplateTypeService,
				Provider: "aws",
			},
		}

		for _, tmpl := range templates {
			err := registry.RegisterTemplate(tmpl)
			require.NoError(t, err)
		}

		awsTemplates := registry.GetTemplatesForProvider("aws")
		awsCount := 0
		for _, tmpl := range awsTemplates {
			if tmpl.Provider == "aws" {
				awsCount++
			}
		}
		assert.GreaterOrEqual(t, awsCount, 2, "Should have at least 2 AWS templates")
	})

	// Test 4: Filter by service
	t.Run("filters by service correctly", func(t *testing.T) {
		template := TemplateDefinition{
			Name:     "monitoring-template",
			Path:     "/monitoring",
			Type:     TemplateTypeService,
			Services: []string{"prometheus", "grafana", "alertmanager"},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		prometheusTemplates := registry.GetTemplatesForService("prometheus")
		found := false
		for _, tmpl := range prometheusTemplates {
			if tmpl.Name == "monitoring-template" {
				found = true
				assert.Contains(t, tmpl.Services, "prometheus")
			}
		}
		assert.True(t, found, "Should find monitoring-template when filtering by prometheus service")
	})

	// Test 5: Metadata persistence across operations
	t.Run("metadata persists across operations", func(t *testing.T) {
		template := TemplateDefinition{
			Name: "persistent-template",
			Path: "/persistent",
			Type: TemplateTypeBase,
			Metadata: TemplateMetadata{
				Description: "Original description",
				Version:     "1.0.0",
				Tags:        []string{"tag1", "tag2"},
			},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Perform various operations
		_ = registry.ListTemplates()
		_ = registry.GetTemplatesForProvider("")

		// Verify metadata is still intact
		retrieved, err := registry.GetTemplate("persistent-template")
		require.NoError(t, err)
		assert.Equal(t, "Original description", retrieved.Metadata.Description)
		assert.Equal(t, "1.0.0", retrieved.Metadata.Version)
		assert.Equal(t, []string{"tag1", "tag2"}, retrieved.Metadata.Tags)
	})

	// Test 6: Multiple templates with different metadata
	t.Run("manages multiple templates with different metadata", func(t *testing.T) {
		templates := []TemplateDefinition{
			{
				Name:     "template-a",
				Path:     "/a",
				Type:     TemplateTypeBase,
				Provider: "openstack",
				Metadata: TemplateMetadata{
					Priority: 10,
					Tags:     []string{"base"},
				},
			},
			{
				Name:     "template-b",
				Path:     "/b",
				Type:     TemplateTypeService,
				Provider: "aws",
				Services: []string{"service1"},
				Metadata: TemplateMetadata{
					Priority: 5,
					Tags:     []string{"service"},
				},
			},
			{
				Name: "template-c",
				Path: "/c",
				Type: TemplateTypeOverlay,
				Metadata: TemplateMetadata{
					Priority: 1,
					Tags:     []string{"overlay"},
				},
			},
		}

		for _, tmpl := range templates {
			err := registry.RegisterTemplate(tmpl)
			require.NoError(t, err)
		}

		// Verify each template maintains its unique metadata
		for _, original := range templates {
			retrieved, err := registry.GetTemplate(original.Name)
			require.NoError(t, err)
			assert.Equal(t, original.Type, retrieved.Type)
			assert.Equal(t, original.Provider, retrieved.Provider)
			assert.Equal(t, original.Services, retrieved.Services)
			assert.Equal(t, original.Metadata.Priority, retrieved.Metadata.Priority)
			assert.Equal(t, original.Metadata.Tags, retrieved.Metadata.Tags)
		}
	})

	// Test 7: Conditions are preserved
	t.Run("preserves render conditions", func(t *testing.T) {
		template := TemplateDefinition{
			Name: "conditional-template",
			Path: "/conditional",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "environment",
					Value: "production",
				},
				{
					Type:  ConditionTypeExists,
					Field: "feature_flag",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		retrieved, err := registry.GetTemplate("conditional-template")
		require.NoError(t, err)
		assert.Len(t, retrieved.Conditions, 2)
		assert.Equal(t, ConditionTypeEquals, retrieved.Conditions[0].Type)
		assert.Equal(t, "environment", retrieved.Conditions[0].Field)
		assert.Equal(t, "production", retrieved.Conditions[0].Value)
		assert.Equal(t, ConditionTypeExists, retrieved.Conditions[1].Type)
		assert.Equal(t, "feature_flag", retrieved.Conditions[1].Field)
	})
}

// TestProperty6_TemplateMetadataCompleteness validates Property 6 from the design document:
// "For any registered template, it should have complete metadata and be retrievable by name, provider, or service"
// **Validates: Requirements 2.1**
func TestProperty6_TemplateMetadataCompleteness(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Register a template with complete metadata
	template := TemplateDefinition{
		Name:         "test-template",
		Path:         "/path/to/template",
		Type:         TemplateTypeInfrastructure,
		Provider:     "openstack",
		Services:     []string{"prometheus", "grafana"},
		Dependencies: []string{"base"},
		Conditions: []RenderCondition{
			{Type: ConditionTypeEquals, Field: "env", Value: "prod"},
		},
		Metadata: TemplateMetadata{
			Description: "Test template",
			Version:     "1.0.0",
			Author:      "Test",
			Tags:        []string{"monitoring"},
			Priority:    5,
		},
	}

	err := registry.RegisterTemplate(template)
	require.NoError(t, err)

	// Test 1: Retrievable by name
	t.Run("retrievable by name", func(t *testing.T) {
		retrieved, err := registry.GetTemplate("test-template")
		require.NoError(t, err)
		assert.Equal(t, template.Name, retrieved.Name)
		assert.NotNil(t, retrieved.Metadata)
	})

	// Test 2: Retrievable by provider
	t.Run("retrievable by provider", func(t *testing.T) {
		templates := registry.GetTemplatesForProvider("openstack")
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == "test-template" {
				found = true
				assert.NotNil(t, tmpl.Metadata)
				break
			}
		}
		assert.True(t, found, "Template should be retrievable by provider")
	})

	// Test 3: Retrievable by service
	t.Run("retrievable by service", func(t *testing.T) {
		templates := registry.GetTemplatesForService("prometheus")
		found := false
		for _, tmpl := range templates {
			if tmpl.Name == "test-template" {
				found = true
				assert.NotNil(t, tmpl.Metadata)
				break
			}
		}
		assert.True(t, found, "Template should be retrievable by service")
	})

	// Test 4: Metadata completeness
	t.Run("has complete metadata", func(t *testing.T) {
		retrieved, err := registry.GetTemplate("test-template")
		require.NoError(t, err)

		// Verify all metadata fields are preserved
		assert.NotEmpty(t, retrieved.Name)
		assert.NotEmpty(t, retrieved.Path)
		assert.NotEmpty(t, retrieved.Type)
		assert.NotEmpty(t, retrieved.Provider)
		assert.NotEmpty(t, retrieved.Services)
		assert.NotEmpty(t, retrieved.Dependencies)
		assert.NotEmpty(t, retrieved.Conditions)
		assert.NotEmpty(t, retrieved.Metadata.Description)
		assert.NotEmpty(t, retrieved.Metadata.Version)
		assert.NotEmpty(t, retrieved.Metadata.Author)
		assert.NotEmpty(t, retrieved.Metadata.Tags)
	})
}

// TestGetTemplatesForEnabledServices tests the service filtering functionality
func TestGetTemplatesForEnabledServices(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Register templates with various service associations
	templates := []TemplateDefinition{
		{
			Name:     "prometheus-template",
			Path:     "/path/to/prometheus",
			Type:     TemplateTypeService,
			Services: []string{"prometheus"},
			Metadata: TemplateMetadata{Priority: 10},
		},
		{
			Name:     "grafana-template",
			Path:     "/path/to/grafana",
			Type:     TemplateTypeService,
			Services: []string{"grafana"},
			Metadata: TemplateMetadata{Priority: 8},
		},
		{
			Name:     "monitoring-template",
			Path:     "/path/to/monitoring",
			Type:     TemplateTypeService,
			Services: []string{"prometheus", "grafana"},
			Metadata: TemplateMetadata{Priority: 9},
		},
		{
			Name:     "loki-template",
			Path:     "/path/to/loki",
			Type:     TemplateTypeService,
			Services: []string{"loki"},
			Metadata: TemplateMetadata{Priority: 7},
		},
		{
			Name:     "universal-template",
			Path:     "/path/to/universal",
			Type:     TemplateTypeBase,
			Services: []string{}, // No services - universal
			Metadata: TemplateMetadata{Priority: 5},
		},
		{
			Name:     "keycloak-template",
			Path:     "/path/to/keycloak",
			Type:     TemplateTypeService,
			Services: []string{"keycloak"},
			Metadata: TemplateMetadata{Priority: 6},
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	t.Run("only prometheus enabled", func(t *testing.T) {
		enabledServices := []string{"prometheus"}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		// Should include: prometheus-template, monitoring-template (has prometheus), universal-template
		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}

		assert.Contains(t, names, "prometheus-template")
		assert.Contains(t, names, "monitoring-template")
		assert.Contains(t, names, "universal-template")
		assert.NotContains(t, names, "grafana-template")
		assert.NotContains(t, names, "loki-template")
		assert.NotContains(t, names, "keycloak-template")
	})

	t.Run("prometheus and grafana enabled", func(t *testing.T) {
		enabledServices := []string{"prometheus", "grafana"}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}

		assert.Contains(t, names, "prometheus-template")
		assert.Contains(t, names, "grafana-template")
		assert.Contains(t, names, "monitoring-template")
		assert.Contains(t, names, "universal-template")
		assert.NotContains(t, names, "loki-template")
		assert.NotContains(t, names, "keycloak-template")
	})

	t.Run("no services enabled", func(t *testing.T) {
		enabledServices := []string{}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		// Should only include universal templates
		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}

		assert.Contains(t, names, "universal-template")
		assert.NotContains(t, names, "prometheus-template")
		assert.NotContains(t, names, "grafana-template")
		assert.NotContains(t, names, "monitoring-template")
		assert.NotContains(t, names, "loki-template")
		assert.NotContains(t, names, "keycloak-template")
	})

	t.Run("all services enabled", func(t *testing.T) {
		enabledServices := []string{"prometheus", "grafana", "loki", "keycloak"}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		// Should include all templates
		assert.Len(t, result, 6)
	})

	t.Run("non-existent service enabled", func(t *testing.T) {
		enabledServices := []string{"non-existent-service"}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		// Should only include universal templates
		names := make([]string, len(result))
		for i, tmpl := range result {
			names[i] = tmpl.Name
		}

		assert.Contains(t, names, "universal-template")
		assert.Len(t, result, 1)
	})

	t.Run("priority ordering is maintained", func(t *testing.T) {
		enabledServices := []string{"prometheus", "grafana"}
		result := registry.GetTemplatesForEnabledServices(enabledServices)

		// Verify templates are sorted by priority (higher first)
		for i := 0; i < len(result)-1; i++ {
			if result[i].Metadata.Priority == result[i+1].Metadata.Priority {
				// If priorities are equal, should be sorted by name
				assert.True(t, result[i].Name < result[i+1].Name)
			} else {
				assert.True(t, result[i].Metadata.Priority >= result[i+1].Metadata.Priority)
			}
		}
	})
}

// TestProperty9_ServiceBasedTemplateFiltering validates Property 9 from the design document:
// "For any disabled service, its associated templates should not be included in template resolution"
// **Validates: Requirements 2.4**
func TestProperty9_ServiceBasedTemplateFiltering(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	// Register templates with service associations
	templates := []TemplateDefinition{
		{
			Name:     "enabled-service-template",
			Path:     "/enabled",
			Type:     TemplateTypeService,
			Services: []string{"enabled-service"},
		},
		{
			Name:     "disabled-service-template",
			Path:     "/disabled",
			Type:     TemplateTypeService,
			Services: []string{"disabled-service"},
		},
		{
			Name:     "multi-service-template",
			Path:     "/multi",
			Type:     TemplateTypeService,
			Services: []string{"enabled-service", "disabled-service"},
		},
		{
			Name:     "universal-template",
			Path:     "/universal",
			Type:     TemplateTypeBase,
			Services: []string{}, // No services
		},
	}

	for _, tmpl := range templates {
		err := registry.RegisterTemplate(tmpl)
		require.NoError(t, err)
	}

	// Test with only enabled-service enabled
	enabledServices := []string{"enabled-service"}
	result := registry.GetTemplatesForEnabledServices(enabledServices)

	names := make([]string, len(result))
	for i, tmpl := range result {
		names[i] = tmpl.Name
	}

	// Should include templates associated with enabled services
	assert.Contains(t, names, "enabled-service-template", "Template for enabled service should be included")
	assert.Contains(t, names, "multi-service-template", "Template with at least one enabled service should be included")
	assert.Contains(t, names, "universal-template", "Universal template should always be included")

	// Should NOT include templates only associated with disabled services
	assert.NotContains(t, names, "disabled-service-template", "Template for disabled service should be excluded")
}

// TestServiceFilteringEdgeCases tests edge cases for service filtering
func TestServiceFilteringEdgeCases(t *testing.T) {
	registry := NewInMemoryTemplateRegistry()

	t.Run("template with empty services list is always included", func(t *testing.T) {
		template := TemplateDefinition{
			Name:     "no-services",
			Path:     "/no-services",
			Type:     TemplateTypeBase,
			Services: []string{},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Test with no enabled services
		result := registry.GetTemplatesForEnabledServices([]string{})
		assert.Len(t, result, 1)
		assert.Equal(t, "no-services", result[0].Name)

		// Test with some enabled services
		result = registry.GetTemplatesForEnabledServices([]string{"some-service"})
		assert.Len(t, result, 1)
		assert.Equal(t, "no-services", result[0].Name)
	})

	t.Run("template with nil services list is always included", func(t *testing.T) {
		template := TemplateDefinition{
			Name:     "nil-services",
			Path:     "/nil-services",
			Type:     TemplateTypeBase,
			Services: nil,
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		result := registry.GetTemplatesForEnabledServices([]string{})
		found := false
		for _, tmpl := range result {
			if tmpl.Name == "nil-services" {
				found = true
				break
			}
		}
		assert.True(t, found, "Template with nil services should be included")
	})

	t.Run("case sensitivity in service names", func(t *testing.T) {
		template := TemplateDefinition{
			Name:     "case-sensitive",
			Path:     "/case",
			Type:     TemplateTypeService,
			Services: []string{"MyService"},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Exact match should work
		result := registry.GetTemplatesForEnabledServices([]string{"MyService"})
		found := false
		for _, tmpl := range result {
			if tmpl.Name == "case-sensitive" {
				found = true
				break
			}
		}
		assert.True(t, found, "Exact case match should find template")

		// Different case should not match
		result = registry.GetTemplatesForEnabledServices([]string{"myservice"})
		found = false
		for _, tmpl := range result {
			if tmpl.Name == "case-sensitive" {
				found = true
				break
			}
		}
		assert.False(t, found, "Different case should not match")
	})

	t.Run("duplicate services in enabled list", func(t *testing.T) {
		template := TemplateDefinition{
			Name:     "duplicate-test",
			Path:     "/duplicate",
			Type:     TemplateTypeService,
			Services: []string{"service1"},
		}

		err := registry.RegisterTemplate(template)
		require.NoError(t, err)

		// Enabled services with duplicates
		result := registry.GetTemplatesForEnabledServices([]string{"service1", "service1", "service1"})

		// Should still only return the template once
		count := 0
		for _, tmpl := range result {
			if tmpl.Name == "duplicate-test" {
				count++
			}
		}
		assert.Equal(t, 1, count, "Template should only appear once even with duplicate enabled services")
	})
}

// TestProperty7_TemplateDependencyValidation validates Property 7 from the design document:
// "For any template registration, invalid dependencies should be rejected and valid dependencies should be accepted"
// **Validates: Requirements 2.2**
func TestProperty7_TemplateDependencyValidation(t *testing.T) {
	t.Run("valid dependencies are accepted", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "valid-deps",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"dep1", "dep2", "dep3"},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Valid dependencies should be accepted")
	})

	t.Run("self-dependency is rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "self-dep",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"self-dep"},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Self-dependency should be rejected")
		assert.Contains(t, err.Error(), "cannot depend on itself")
	})

	t.Run("empty dependency name is rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "empty-dep",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"dep1", "", "dep2"},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Empty dependency name should be rejected")
		assert.Contains(t, err.Error(), "dependency name cannot be empty")
	})

	t.Run("duplicate dependencies are rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "duplicate-deps",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"dep1", "dep2", "dep1"},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Duplicate dependencies should be rejected")
		assert.Contains(t, err.Error(), "duplicate dependency")
	})

	t.Run("dependencies can be registered in any order", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		// Register dependent before dependency
		dependent := TemplateDefinition{
			Name:         "dependent",
			Path:         "/path/to/dependent",
			Type:         TemplateTypeService,
			Dependencies: []string{"base"},
		}

		err := registry.RegisterTemplate(dependent)
		assert.NoError(t, err, "Should allow registering dependent before dependency")

		// Register dependency after dependent
		base := TemplateDefinition{
			Name: "base",
			Path: "/path/to/base",
			Type: TemplateTypeBase,
		}

		err = registry.RegisterTemplate(base)
		assert.NoError(t, err, "Should allow registering dependency after dependent")
	})

	t.Run("multiple valid dependencies", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "multi-deps",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"base", "middleware", "utils", "config"},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Multiple valid dependencies should be accepted")
	})
}

// TestConditionValidation tests validation of render conditions
func TestConditionValidation(t *testing.T) {
	t.Run("valid conditions are accepted", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "valid-conditions",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: "openstack",
				},
				{
					Type:  ConditionTypeExists,
					Field: "feature_flag",
				},
				{
					Type:  ConditionTypeGreaterThan,
					Field: "version",
					Value: 1.0,
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Valid conditions should be accepted")
	})

	t.Run("invalid condition type is rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "invalid-condition-type",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionType("invalid_type"),
					Field: "provider",
					Value: "openstack",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Invalid condition type should be rejected")
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("condition without required field is rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "no-field",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "", // Empty field
					Value: "openstack",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Condition without field should be rejected")
		assert.Contains(t, err.Error(), "requires a field name")
	})

	t.Run("condition without required value is rejected", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "no-value",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: nil, // No value
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Condition without required value should be rejected")
		assert.Contains(t, err.Error(), "requires a value")
	})

	t.Run("exists condition without value is accepted", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "exists-no-value",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeExists,
					Field: "feature_flag",
					Value: nil, // Exists doesn't need a value
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Exists condition without value should be accepted")
	})

	t.Run("all condition types are validated", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		validConditionTypes := []ConditionType{
			ConditionTypeEquals,
			ConditionTypeNotEquals,
			ConditionTypeContains,
			ConditionTypeExists,
			ConditionTypeGreaterThan,
			ConditionTypeLessThan,
		}

		for _, condType := range validConditionTypes {
			template := TemplateDefinition{
				Name: "test-" + string(condType),
				Path: "/path/to/template",
				Type: TemplateTypeService,
				Conditions: []RenderCondition{
					{
						Type:  condType,
						Field: "test_field",
						Value: "test_value",
					},
				},
			}

			// Exists doesn't need a value
			if condType == ConditionTypeExists {
				template.Conditions[0].Value = nil
			}

			err := registry.RegisterTemplate(template)
			assert.NoError(t, err, "Valid condition type %s should be accepted", condType)
		}
	})

	t.Run("multiple conditions are all validated", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name: "multi-conditions",
			Path: "/path/to/template",
			Type: TemplateTypeService,
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: "openstack",
				},
				{
					Type:  ConditionType("invalid"), // Invalid type
					Field: "test",
					Value: "value",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Should reject template with any invalid condition")
		assert.Contains(t, err.Error(), "invalid type")
	})

	t.Run("empty conditions list is accepted", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:       "no-conditions",
			Path:       "/path/to/template",
			Type:       TemplateTypeService,
			Conditions: []RenderCondition{},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Template with no conditions should be accepted")
	})

	t.Run("nil conditions list is accepted", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:       "nil-conditions",
			Path:       "/path/to/template",
			Type:       TemplateTypeService,
			Conditions: nil,
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Template with nil conditions should be accepted")
	})
}

// TestDependencyAndConditionValidationTogether tests that both validations work together
func TestDependencyAndConditionValidationTogether(t *testing.T) {
	t.Run("template with valid dependencies and conditions", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "complete-template",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"base", "middleware"},
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: "openstack",
				},
				{
					Type:  ConditionTypeExists,
					Field: "feature_flag",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.NoError(t, err, "Template with valid dependencies and conditions should be accepted")
	})

	t.Run("template with invalid dependencies and valid conditions", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "invalid-deps",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"base", "base"}, // Duplicate
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "provider",
					Value: "openstack",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Should reject template with invalid dependencies")
		assert.Contains(t, err.Error(), "duplicate dependency")
	})

	t.Run("template with valid dependencies and invalid conditions", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "invalid-conditions",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"base", "middleware"},
			Conditions: []RenderCondition{
				{
					Type:  ConditionTypeEquals,
					Field: "", // Empty field
					Value: "openstack",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Should reject template with invalid conditions")
		assert.Contains(t, err.Error(), "requires a field name")
	})

	t.Run("template with both invalid dependencies and conditions", func(t *testing.T) {
		registry := NewInMemoryTemplateRegistry()

		template := TemplateDefinition{
			Name:         "all-invalid",
			Path:         "/path/to/template",
			Type:         TemplateTypeService,
			Dependencies: []string{"all-invalid"}, // Self-dependency
			Conditions: []RenderCondition{
				{
					Type:  ConditionType("invalid"),
					Field: "test",
					Value: "value",
				},
			},
		}

		err := registry.RegisterTemplate(template)
		assert.Error(t, err, "Should reject template with invalid dependencies and conditions")
		// Should fail on dependencies first (checked first in RegisterTemplate)
		assert.Contains(t, err.Error(), "invalid dependencies")
	})
}
