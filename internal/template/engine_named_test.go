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
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestExecuteTemplate tests executing a named template from a collection.
func TestExecuteTemplate(t *testing.T) {
	// Create temporary directory with multiple templates
	tmpDir := t.TempDir()

	// Create template files
	templates := map[string]string{
		"base.tmpl":   "Base: {{.Value}}",
		"header.tmpl": "Header: {{.Title}}",
		"footer.tmpl": "Footer: {{.Copyright}}",
	}

	// Create engine
	engine := NewGoTemplateEngine()

	// Load each template file
	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write template %s: %v", name, err)
		}
		if err := engine.LoadFromFile(path); err != nil {
			t.Fatalf("failed to load template %s: %v", name, err)
		}
	}

	// Execute named template using full path
	basePath := filepath.Join(tmpDir, "base.tmpl")
	data := map[string]string{"Value": "Test"}
	result, err := engine.ExecuteTemplate(basePath, data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	expected := "Base: Test"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestExecuteTemplateWithFS tests executing named templates from embedded FS.
func TestExecuteTemplateWithFS(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Debug: List all loaded templates
	engine.mu.RLock()
	if engine.rootTemplate != nil {
		t.Logf("Loaded templates:")
		for _, tmpl := range engine.rootTemplate.Templates() {
			t.Logf("  - %s", tmpl.Name())
		}
	}
	engine.mu.RUnlock()

	// Execute named template - try with just the filename
	data := map[string]string{"Name": "ExecuteTest"}
	result, err := engine.ExecuteTemplate("simple.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	// Verify result contains expected content
	if !strings.Contains(string(result), "ExecuteTest") {
		t.Errorf("result should contain 'ExecuteTest', got: %s", string(result))
	}
}

// TestExecuteTemplateToWriter tests executing a named template to a writer.
func TestExecuteTemplateToWriter(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Execute to buffer
	var buf bytes.Buffer
	data := map[string]string{"Name": "WriterTest"}
	if err := engine.ExecuteTemplateToWriter("simple.tmpl", data, &buf); err != nil {
		t.Fatalf("ExecuteTemplateToWriter failed: %v", err)
	}

	// Verify result
	if !strings.Contains(buf.String(), "WriterTest") {
		t.Errorf("result should contain 'WriterTest', got: %s", buf.String())
	}
}

// TestExecuteTemplateNotLoaded tests error when no templates are loaded.
func TestExecuteTemplateNotLoaded(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Try to execute without loading templates
	_, err := engine.ExecuteTemplate("nonexistent.tmpl", nil)
	if err == nil {
		t.Error("expected error when executing template without loading")
	}

	if !strings.Contains(err.Error(), "no templates loaded") {
		t.Errorf("expected 'no templates loaded' error, got: %v", err)
	}
}

// TestExecuteTemplateNotFound tests error when template name doesn't exist.
func TestExecuteTemplateNotFound(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Try to execute non-existent template
	_, err := engine.ExecuteTemplate("nonexistent.tmpl", nil)
	if err == nil {
		t.Error("expected error when executing non-existent template")
	}
}

// TestGetTemplate tests retrieving a parsed template by name.
func TestGetTemplate(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Get template by name
	tmpl, err := engine.GetTemplate("simple.tmpl")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if tmpl == nil {
		t.Error("expected non-nil template")
	}

	// Verify template name
	if tmpl.Name() != "simple.tmpl" {
		t.Errorf("expected template name 'simple.tmpl', got %q", tmpl.Name())
	}
}

// TestGetTemplateNotFound tests error when template is not found.
func TestGetTemplateNotFound(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Try to get non-existent template
	_, err := engine.GetTemplate("nonexistent.tmpl")
	if err == nil {
		t.Error("expected error when getting non-existent template")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

// TestExecuteTemplateWithComplexData tests executing templates with complex data structures.
func TestExecuteTemplateWithComplexData(t *testing.T) {
	// Create temporary directory with template
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "complex.tmpl")

	// Template similar to inventory.tmpl
	templateContent := `[master]
{{- range $i, $e := until (int .Counts.Master) }}
master-{{ $i }}
{{- end }}

[worker]
{{- range $i, $e := until (int .Counts.Worker) }}
worker-{{ $i }}
{{- end }}`

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Create engine and load template
	engine := NewGoTemplateEngine()
	if err := engine.LoadFromFile(templatePath); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Execute template with complex data
	data := map[string]interface{}{
		"Counts": map[string]interface{}{
			"Master": 2,
			"Worker": 3,
		},
	}

	result, err := engine.ExecuteTemplate(templatePath, data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	output := string(result)

	// Verify master nodes
	if !strings.Contains(output, "master-0") {
		t.Error("missing master-0")
	}
	if !strings.Contains(output, "master-1") {
		t.Error("missing master-1")
	}

	// Verify worker nodes
	if !strings.Contains(output, "worker-0") {
		t.Error("missing worker-0")
	}
	if !strings.Contains(output, "worker-1") {
		t.Error("missing worker-1")
	}
	if !strings.Contains(output, "worker-2") {
		t.Error("missing worker-2")
	}
}

// TestExecuteTemplateWithCustomFunctions tests that custom functions work with ExecuteTemplate.
func TestExecuteTemplateWithCustomFunctions(t *testing.T) {
	// Create temporary directory with template
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "custom.tmpl")

	// Template using custom function
	templateContent := "Result: {{customFunc .Value}}"

	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	// Create engine and register custom function
	engine := NewGoTemplateEngine()
	engine.RegisterFunction("customFunc", func(s string) string {
		return "custom_" + s
	})

	// Load template after registering function
	if err := engine.LoadFromFile(templatePath); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Execute template
	data := map[string]string{"Value": "test"}
	result, err := engine.ExecuteTemplate(templatePath, data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	expected := "Result: custom_test"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}
