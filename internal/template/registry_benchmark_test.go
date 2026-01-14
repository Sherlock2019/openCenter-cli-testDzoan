package template

import (
	"fmt"
	"testing"
)

func BenchmarkRegisterTemplate(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		template := TemplateDefinition{
			Name: fmt.Sprintf("template-%d", i),
			Path: fmt.Sprintf("/path/to/template-%d", i),
			Type: TemplateTypeBase,
			Metadata: TemplateMetadata{
				Description: "Benchmark template",
				Version:     "1.0.0",
			},
		}
		_ = registry.RegisterTemplate(template)
	}
}

func BenchmarkGetTemplate(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate registry
	for i := 0; i < 1000; i++ {
		template := TemplateDefinition{
			Name: fmt.Sprintf("template-%d", i),
			Path: fmt.Sprintf("/path/to/template-%d", i),
			Type: TemplateTypeBase,
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.GetTemplate(fmt.Sprintf("template-%d", i%1000))
	}
}

func BenchmarkGetTemplatesForProvider(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate with mixed providers
	providers := []string{"openstack", "aws", "azure", ""}
	for i := 0; i < 1000; i++ {
		template := TemplateDefinition{
			Name:     fmt.Sprintf("template-%d", i),
			Path:     fmt.Sprintf("/path/to/template-%d", i),
			Type:     TemplateTypeBase,
			Provider: providers[i%len(providers)],
			Metadata: TemplateMetadata{Priority: i % 10},
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetTemplatesForProvider("openstack")
	}
}

func BenchmarkGetTemplatesForService(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate with services
	services := [][]string{
		{"prometheus", "monitoring"},
		{"grafana", "monitoring"},
		{"loki", "logging"},
		{"elasticsearch", "logging"},
	}

	for i := 0; i < 1000; i++ {
		template := TemplateDefinition{
			Name:     fmt.Sprintf("template-%d", i),
			Path:     fmt.Sprintf("/path/to/template-%d", i),
			Type:     TemplateTypeService,
			Services: services[i%len(services)],
			Metadata: TemplateMetadata{Priority: i % 10},
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.GetTemplatesForService("monitoring")
	}
}

func BenchmarkResolveTemplateDependencies(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Create a dependency chain: base -> layer1 -> layer2 -> ... -> layer10
	_ = registry.RegisterTemplate(TemplateDefinition{
		Name: "base",
		Path: "/path/to/base",
		Type: TemplateTypeBase,
	})

	for i := 1; i <= 10; i++ {
		template := TemplateDefinition{
			Name:         fmt.Sprintf("layer%d", i),
			Path:         fmt.Sprintf("/path/to/layer%d", i),
			Type:         TemplateTypeService,
			Dependencies: []string{fmt.Sprintf("layer%d", i-1)},
		}
		if i == 1 {
			template.Dependencies = []string{"base"}
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = registry.ResolveTemplateDependencies([]string{"layer10"})
	}
}

func BenchmarkListTemplates(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate registry
	for i := 0; i < 1000; i++ {
		template := TemplateDefinition{
			Name: fmt.Sprintf("template-%d", i),
			Path: fmt.Sprintf("/path/to/template-%d", i),
			Type: TemplateTypeBase,
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = registry.ListTemplates()
	}
}

func BenchmarkConcurrentReads(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate registry
	for i := 0; i < 100; i++ {
		template := TemplateDefinition{
			Name:     fmt.Sprintf("template-%d", i),
			Path:     fmt.Sprintf("/path/to/template-%d", i),
			Type:     TemplateTypeBase,
			Provider: "openstack",
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			_, _ = registry.GetTemplate(fmt.Sprintf("template-%d", i%100))
			_ = registry.GetTemplatesForProvider("openstack")
			i++
		}
	})
}

func BenchmarkConcurrentWrites(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			template := TemplateDefinition{
				Name: fmt.Sprintf("template-%d-%d", b.N, i),
				Path: fmt.Sprintf("/path/to/template-%d-%d", b.N, i),
				Type: TemplateTypeBase,
			}
			_ = registry.RegisterTemplate(template)
			i++
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	registry := NewInMemoryTemplateRegistry()

	// Pre-populate
	for i := 0; i < 100; i++ {
		template := TemplateDefinition{
			Name:     fmt.Sprintf("template-%d", i),
			Path:     fmt.Sprintf("/path/to/template-%d", i),
			Type:     TemplateTypeBase,
			Provider: "openstack",
		}
		_ = registry.RegisterTemplate(template)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Mix of operations
		switch i % 4 {
		case 0:
			_, _ = registry.GetTemplate(fmt.Sprintf("template-%d", i%100))
		case 1:
			_ = registry.GetTemplatesForProvider("openstack")
		case 2:
			_ = registry.ListTemplates()
		case 3:
			template := TemplateDefinition{
				Name: fmt.Sprintf("new-template-%d", i),
				Path: fmt.Sprintf("/path/to/new-%d", i),
				Type: TemplateTypeBase,
			}
			_ = registry.RegisterTemplate(template)
		}
	}
}
