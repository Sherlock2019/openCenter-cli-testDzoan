package template

import (
	"fmt"
	"testing"
)

// TestConflictResolutionErrorMessages demonstrates the enhanced error messages
// for conflict resolution in template composition.
//
// This test is for documentation purposes and shows the improved error messages
// that help users understand and resolve conflicts.
func TestConflictResolutionErrorMessages(t *testing.T) {
	engine := NewGoTemplateEngine()
	registry := NewInMemoryTemplateRegistry()
	composer := NewDefaultTemplateComposer(engine, registry)

	// Register base template
	err := registry.RegisterTemplate(TemplateDefinition{
		Name:     "base-template",
		Path:     "/tmp/base.tmpl",
		Type:     TemplateTypeBase,
		Provider: "openstack",
	})
	if err != nil {
		t.Fatalf("Failed to register base template: %v", err)
	}

	t.Run("Duplicate Overlay Names Error", func(t *testing.T) {
		composition := TemplateComposition{
			BaseTemplate: "base-template",
			Overlays: []TemplateOverlay{
				{Name: "my-overlay", Path: "/tmp/overlay1.tmpl", Priority: 1},
				{Name: "my-overlay", Path: "/tmp/overlay2.tmpl", Priority: 2},
			},
		}

		err := composer.ValidateComposition(composition)
		if err != nil {
			fmt.Println("\n=== Duplicate Overlay Names Error ===")
			fmt.Println(err.Error())
			fmt.Println()
		}
	})

	t.Run("Provider Conflict Error", func(t *testing.T) {
		// Register AWS overlay
		err := registry.RegisterTemplate(TemplateDefinition{
			Name:     "aws-overlay",
			Path:     "/tmp/aws-overlay.tmpl",
			Type:     TemplateTypeOverlay,
			Provider: "aws",
		})
		if err != nil {
			t.Fatalf("Failed to register AWS overlay: %v", err)
		}

		composition := TemplateComposition{
			BaseTemplate: "base-template",
			Overlays: []TemplateOverlay{
				{Name: "aws-overlay", Path: "/tmp/aws-overlay.tmpl", Priority: 1},
			},
		}

		err = composer.ValidateComposition(composition)
		if err != nil {
			fmt.Println("\n=== Provider Conflict Error ===")
			fmt.Println(err.Error())
			fmt.Println()
		}
	})

	t.Run("Type Mismatch Error", func(t *testing.T) {
		// Register infrastructure overlay
		err := registry.RegisterTemplate(TemplateDefinition{
			Name:     "infra-overlay",
			Path:     "/tmp/infra-overlay.tmpl",
			Type:     TemplateTypeInfrastructure,
			Provider: "openstack",
		})
		if err != nil {
			t.Fatalf("Failed to register infrastructure overlay: %v", err)
		}

		// Register service base
		err = registry.RegisterTemplate(TemplateDefinition{
			Name:     "service-base",
			Path:     "/tmp/service-base.tmpl",
			Type:     TemplateTypeService,
			Provider: "openstack",
		})
		if err != nil {
			t.Fatalf("Failed to register service base: %v", err)
		}

		composition := TemplateComposition{
			BaseTemplate: "service-base",
			Overlays: []TemplateOverlay{
				{Name: "infra-overlay", Path: "/tmp/infra-overlay.tmpl", Priority: 1},
			},
		}

		err = composer.ValidateComposition(composition)
		if err != nil {
			fmt.Println("\n=== Type Mismatch Error ===")
			fmt.Println(err.Error())
			fmt.Println()
		}
	})

	t.Run("Multiple Provider Conflict Error", func(t *testing.T) {
		// Register OpenStack overlay
		err := registry.RegisterTemplate(TemplateDefinition{
			Name:     "openstack-overlay",
			Path:     "/tmp/openstack-overlay.tmpl",
			Type:     TemplateTypeOverlay,
			Provider: "openstack",
		})
		if err != nil {
			t.Fatalf("Failed to register OpenStack overlay: %v", err)
		}

		// Register GCP overlay
		err = registry.RegisterTemplate(TemplateDefinition{
			Name:     "gcp-overlay",
			Path:     "/tmp/gcp-overlay.tmpl",
			Type:     TemplateTypeOverlay,
			Provider: "gcp",
		})
		if err != nil {
			t.Fatalf("Failed to register GCP overlay: %v", err)
		}

		composition := TemplateComposition{
			BaseTemplate: "base-template",
			Overlays: []TemplateOverlay{
				{Name: "openstack-overlay", Path: "/tmp/openstack-overlay.tmpl", Priority: 1},
				{Name: "gcp-overlay", Path: "/tmp/gcp-overlay.tmpl", Priority: 2},
			},
		}

		err = composer.ValidateComposition(composition)
		if err != nil {
			fmt.Println("\n=== Multiple Provider Conflict Error ===")
			fmt.Println(err.Error())
			fmt.Println()
		}
	})
}
