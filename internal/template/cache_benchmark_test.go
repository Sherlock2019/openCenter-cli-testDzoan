/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// BenchmarkTemplateCaching benchmarks the performance improvement from template caching.
// This benchmark compares cached vs non-cached template rendering to demonstrate
// measurable performance improvements.
func BenchmarkTemplateCaching(b *testing.B) {
	// Create a temporary directory for test templates
	tmpDir := b.TempDir()

	// Create test templates with varying complexity
	templates := map[string]string{
		"simple.tmpl": `Hello, {{.Name}}!`,
		"medium.tmpl": `
{{- range .Items }}
Item: {{.Name}} - {{.Value}}
{{- end }}
Total: {{.Total}}
`,
		"complex.tmpl": `
{{- define "header" -}}
=== {{.Title}} ===
{{- end -}}

{{- template "header" . }}

{{- range $idx, $item := .Items }}
{{- if gt $idx 0 }}
---
{{- end }}
Item #{{add $idx 1}}: {{$item.Name}}
  Description: {{$item.Description}}
  Price: ${{$item.Price}}
  {{- if $item.Tags }}
  Tags: {{join ", " $item.Tags}}
  {{- end }}
{{- end }}

Total Items: {{len .Items}}
Grand Total: ${{.GrandTotal}}
`,
	}

	// Write templates to disk
	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write template %s: %v", name, err)
		}
	}

	// Test data for different template complexities
	testData := map[string]interface{}{
		"simple.tmpl": map[string]string{
			"Name": "World",
		},
		"medium.tmpl": map[string]interface{}{
			"Items": []map[string]interface{}{
				{"Name": "Item1", "Value": 100},
				{"Name": "Item2", "Value": 200},
				{"Name": "Item3", "Value": 300},
			},
			"Total": 600,
		},
		"complex.tmpl": map[string]interface{}{
			"Title": "Product Catalog",
			"Items": []map[string]interface{}{
				{
					"Name":        "Widget A",
					"Description": "A high-quality widget",
					"Price":       29.99,
					"Tags":        []string{"popular", "new", "featured"},
				},
				{
					"Name":        "Widget B",
					"Description": "An affordable widget",
					"Price":       19.99,
					"Tags":        []string{"budget", "reliable"},
				},
				{
					"Name":        "Widget C",
					"Description": "A premium widget",
					"Price":       49.99,
					"Tags":        []string{"premium", "featured"},
				},
			},
			"GrandTotal": 99.97,
		},
	}

	// Benchmark with caching enabled
	b.Run("WithCache", func(b *testing.B) {
		for templateName := range templates {
			b.Run(templateName, func(b *testing.B) {
				engine := NewGoTemplateEngine()
				engine.SetCacheEnabled(true)
				templatePath := filepath.Join(tmpDir, templateName)
				data := testData[templateName]
				ctx := context.Background()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := engine.Render(ctx, templatePath, data)
					if err != nil {
						b.Fatalf("render failed: %v", err)
					}
				}
			})
		}
	})

	// Benchmark with caching disabled
	b.Run("WithoutCache", func(b *testing.B) {
		for templateName := range templates {
			b.Run(templateName, func(b *testing.B) {
				engine := NewGoTemplateEngine()
				engine.SetCacheEnabled(false)
				templatePath := filepath.Join(tmpDir, templateName)
				data := testData[templateName]
				ctx := context.Background()

				b.ResetTimer()
				for i := 0; i < b.N; i++ {
					_, err := engine.Render(ctx, templatePath, data)
					if err != nil {
						b.Fatalf("render failed: %v", err)
					}
				}
			})
		}
	})
}

// BenchmarkCacheHitRate benchmarks the cache hit rate with multiple templates.
// This demonstrates that caching provides consistent performance benefits
// when rendering the same templates multiple times.
func BenchmarkCacheHitRate(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple templates
	numTemplates := 10
	for i := 0; i < numTemplates; i++ {
		templatePath := filepath.Join(tmpDir, fmt.Sprintf("template%d.tmpl", i))
		content := fmt.Sprintf("Template %d: {{.Value}}", i)
		if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write template: %v", err)
		}
	}

	engine := NewGoTemplateEngine()
	engine.SetCacheEnabled(true)
	ctx := context.Background()
	data := map[string]int{"Value": 42}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Render templates in round-robin fashion to test cache hits
		templateIdx := i % numTemplates
		templatePath := filepath.Join(tmpDir, fmt.Sprintf("template%d.tmpl", templateIdx))
		_, err := engine.Render(ctx, templatePath, data)
		if err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}
}

// BenchmarkConcurrentCachedRendering benchmarks concurrent template rendering with caching.
// This demonstrates that the cache is thread-safe and provides performance benefits
// under concurrent load.
func BenchmarkConcurrentCachedRendering(b *testing.B) {
	tmpDir := b.TempDir()

	// Create a test template
	templatePath := filepath.Join(tmpDir, "concurrent.tmpl")
	content := `Hello, {{.Name}}! Your ID is {{.ID}}.`
	if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
		b.Fatalf("failed to write template: %v", err)
	}

	engine := NewGoTemplateEngine()
	engine.SetCacheEnabled(true)
	ctx := context.Background()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			data := map[string]interface{}{
				"Name": fmt.Sprintf("User%d", i),
				"ID":   i,
			}
			_, err := engine.Render(ctx, templatePath, data)
			if err != nil {
				b.Fatalf("render failed: %v", err)
			}
			i++
		}
	})
}

// BenchmarkCacheMemoryOverhead benchmarks the memory overhead of template caching.
// This helps ensure that caching doesn't consume excessive memory.
func BenchmarkCacheMemoryOverhead(b *testing.B) {
	tmpDir := b.TempDir()

	// Create templates of varying sizes
	templateSizes := []int{100, 1000, 10000}

	for _, size := range templateSizes {
		b.Run(fmt.Sprintf("TemplateSize%d", size), func(b *testing.B) {
			// Create a template with the specified size
			templatePath := filepath.Join(tmpDir, fmt.Sprintf("large%d.tmpl", size))
			content := "{{.Value}}"
			for i := 0; i < size; i++ {
				content += fmt.Sprintf("\nLine %d: {{.Value}}", i)
			}
			if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
				b.Fatalf("failed to write template: %v", err)
			}

			engine := NewGoTemplateEngine()
			engine.SetCacheEnabled(true)
			ctx := context.Background()
			data := map[string]int{"Value": 42}

			b.ResetTimer()
			b.ReportAllocs()

			for i := 0; i < b.N; i++ {
				_, err := engine.Render(ctx, templatePath, data)
				if err != nil {
					b.Fatalf("render failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkRenderString benchmarks rendering templates from strings.
// This is useful for comparing performance when templates are not file-based.
func BenchmarkRenderString(b *testing.B) {
	engine := NewGoTemplateEngine()
	ctx := context.Background()

	templates := map[string]string{
		"simple":  `Hello, {{.Name}}!`,
		"complex": `{{range .Items}}{{.Name}}: {{.Value}}\n{{end}}`,
	}

	testData := map[string]interface{}{
		"simple": map[string]string{"Name": "World"},
		"complex": map[string]interface{}{
			"Items": []map[string]interface{}{
				{"Name": "A", "Value": 1},
				{"Name": "B", "Value": 2},
				{"Name": "C", "Value": 3},
			},
		},
	}

	for name, tmpl := range templates {
		b.Run(name, func(b *testing.B) {
			data := testData[name]
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, err := engine.RenderString(ctx, name, tmpl, data)
				if err != nil {
					b.Fatalf("render failed: %v", err)
				}
			}
		})
	}
}

// BenchmarkCacheClearPerformance benchmarks the performance of cache clearing.
// This ensures that cache management operations are efficient.
func BenchmarkCacheClearPerformance(b *testing.B) {
	tmpDir := b.TempDir()

	// Create multiple templates and populate cache
	numTemplates := 100
	engine := NewGoTemplateEngine()
	engine.SetCacheEnabled(true)
	ctx := context.Background()

	for i := 0; i < numTemplates; i++ {
		templatePath := filepath.Join(tmpDir, fmt.Sprintf("template%d.tmpl", i))
		content := fmt.Sprintf("Template %d: {{.Value}}", i)
		if err := os.WriteFile(templatePath, []byte(content), 0644); err != nil {
			b.Fatalf("failed to write template: %v", err)
		}

		// Render once to populate cache
		_, err := engine.Render(ctx, templatePath, map[string]int{"Value": i})
		if err != nil {
			b.Fatalf("render failed: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		engine.ClearCache()
	}
}
