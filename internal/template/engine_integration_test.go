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
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/*.tmpl
var testTemplates embed.FS

// TestIntegrationWithEmbeddedFS tests loading templates from an embedded filesystem.
func TestIntegrationWithEmbeddedFS(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Verify templates were loaded and cached
	if len(engine.cache) == 0 {
		t.Error("no templates were cached after LoadFromFS")
	}

	// Try to render a loaded template
	data := map[string]string{"Name": "Integration"}
	result, err := engine.Render(context.Background(), "testdata/simple.tmpl", data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	if !strings.Contains(string(result), "Integration") {
		t.Errorf("rendered output doesn't contain expected data: %s", string(result))
	}
}

// TestIntegrationCompleteWorkflow tests a complete workflow similar to existing usage.
func TestIntegrationCompleteWorkflow(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Step 1: Register custom functions (like existing code does)
	engine.RegisterFunction("customFunc", func(s string) string {
		return "custom_" + s
	})

	// Step 2: Create a template with Sprig functions and custom functions
	templateContent := `
Cluster: {{.ClusterName | upper}}
Organization: {{.Organization | lower}}
Custom: {{customFunc .Value}}
Masters:
{{- range $i, $e := until (int .MasterCount) }}
  - master-{{ $i }}
{{- end }}
`

	data := map[string]interface{}{
		"ClusterName":  "test-cluster",
		"Organization": "TEST-ORG",
		"Value":        "test",
		"MasterCount":  3,
	}

	// Step 3: Render the template
	result, err := engine.RenderString(context.Background(), "workflow", templateContent, data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	output := string(result)

	// Step 4: Verify output
	if !strings.Contains(output, "TEST-CLUSTER") {
		t.Error("missing uppercase cluster name")
	}
	if !strings.Contains(output, "test-org") {
		t.Error("missing lowercase organization")
	}
	if !strings.Contains(output, "custom_test") {
		t.Error("missing custom function output")
	}
	if !strings.Contains(output, "master-0") || !strings.Contains(output, "master-2") {
		t.Error("missing master nodes")
	}
}

// TestIntegrationFileSystemAndCaching tests the complete file system workflow with caching.
func TestIntegrationFileSystemAndCaching(t *testing.T) {
	// Create temporary directory with templates
	tmpDir := t.TempDir()

	// Create multiple template files
	templates := map[string]string{
		"base.tmpl":    "Base: {{.Value}}",
		"complex.tmpl": "Complex: {{.Value | upper}}",
		"list.tmpl": `List:
{{- range .Items }}
- {{.}}
{{- end }}`,
	}

	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write template %s: %v", name, err)
		}
	}

	// Create engine
	engine := NewGoTemplateEngine()

	// Render each template and verify caching
	data := map[string]interface{}{
		"Value": "test",
		"Items": []string{"item1", "item2", "item3"},
	}

	for name := range templates {
		path := filepath.Join(tmpDir, name)

		// First render - should parse and cache
		result1, err := engine.Render(context.Background(), path, data)
		if err != nil {
			t.Fatalf("first render of %s failed: %v", name, err)
		}

		// Second render - should use cache
		result2, err := engine.Render(context.Background(), path, data)
		if err != nil {
			t.Fatalf("second render of %s failed: %v", name, err)
		}

		// Results should be identical
		if string(result1) != string(result2) {
			t.Errorf("cached render of %s produced different result", name)
		}
	}

	// Verify all templates are cached
	if len(engine.cache) != len(templates) {
		t.Errorf("expected %d cached templates, got %d", len(templates), len(engine.cache))
	}
}

// TestIntegrationErrorHandling tests error handling in various scenarios.
func TestIntegrationErrorHandling(t *testing.T) {
	engine := NewGoTemplateEngine()

	tests := []struct {
		name        string
		template    string
		data        interface{}
		shouldError bool
		errorMsg    string
	}{
		{
			name:        "invalid syntax",
			template:    "Hello {{.Name}",
			data:        map[string]string{"Name": "World"},
			shouldError: true,
			errorMsg:    "bad character", // Changed from "parse" to match actual error
		},
		{
			name:        "missing field",
			template:    "Hello {{.MissingField}}",
			data:        map[string]string{"Name": "World"},
			shouldError: false, // Go templates don't error on missing fields by default
			errorMsg:    "",
		},
		{
			name:        "valid template",
			template:    "Hello {{.Name}}",
			data:        map[string]string{"Name": "World"},
			shouldError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.RenderString(context.Background(), tt.name, tt.template, tt.data)

			if tt.shouldError {
				if err == nil {
					t.Error("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

// TestIntegrationConcurrentRendering tests concurrent template rendering.
func TestIntegrationConcurrentRendering(t *testing.T) {
	engine := NewGoTemplateEngine()

	templateContent := "Hello {{.Name}}!"
	data := map[string]string{"Name": "World"}

	// Render concurrently
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			result, err := engine.RenderString(context.Background(), "concurrent", templateContent, data)
			if err != nil {
				errors <- err
			} else {
				expected := "Hello World!"
				if string(result) != expected {
					errors <- err
				}
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("concurrent rendering error: %v", err)
	}
}

// TestIntegrationSprigCompatibility tests compatibility with all commonly used Sprig functions.
func TestIntegrationSprigCompatibility(t *testing.T) {
	engine := NewGoTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     interface{}
		contains string
	}{
		{
			name:     "upper function",
			template: "{{.Value | upper}}",
			data:     map[string]string{"Value": "hello"},
			contains: "HELLO",
		},
		{
			name:     "lower function",
			template: "{{.Value | lower}}",
			data:     map[string]string{"Value": "HELLO"},
			contains: "hello",
		},
		{
			name:     "trim function",
			template: "{{.Value | trim}}",
			data:     map[string]string{"Value": "  hello  "},
			contains: "hello",
		},
		{
			name:     "default function",
			template: "{{.Missing | default \"fallback\"}}",
			data:     map[string]string{},
			contains: "fallback",
		},
		{
			name:     "until function",
			template: "{{range until 3}}{{.}}{{end}}",
			data:     map[string]string{},
			contains: "012",
		},
		{
			name:     "int function",
			template: "{{int .Value}}",
			data:     map[string]string{"Value": "42"},
			contains: "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderString(context.Background(), tt.name, tt.template, tt.data)
			if err != nil {
				t.Fatalf("render failed: %v", err)
			}

			if !strings.Contains(string(result), tt.contains) {
				t.Errorf("expected output to contain %q, got %q", tt.contains, string(result))
			}
		})
	}
}
