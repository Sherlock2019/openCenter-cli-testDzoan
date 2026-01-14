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
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"
)

// TestRenderString tests rendering a template from a string.
func TestRenderString(t *testing.T) {
	engine := NewGoTemplateEngine()

	templateContent := "Hello {{.Name}}!"
	data := map[string]string{"Name": "World"}

	result, err := engine.RenderString(context.Background(), "test", templateContent, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	expected := "Hello World!"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestRenderStringWithSprigFunctions tests that Sprig functions are available by default.
func TestRenderStringWithSprigFunctions(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Test using Sprig's upper function
	templateContent := "{{.Name | upper}}"
	data := map[string]string{"Name": "hello"}

	result, err := engine.RenderString(context.Background(), "test", templateContent, data)
	if err != nil {
		t.Fatalf("RenderString with Sprig function failed: %v", err)
	}

	expected := "HELLO"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestRenderStringWithUntilFunction tests the 'until' function used in existing templates.
func TestRenderStringWithUntilFunction(t *testing.T) {
	engine := NewGoTemplateEngine()

	// This mimics the pattern used in inventory.tmpl
	templateContent := `{{- range $i, $e := until (int .Count) }}
item-{{ $i }}
{{- end }}`
	data := map[string]interface{}{"Count": 3}

	result, err := engine.RenderString(context.Background(), "test", templateContent, data)
	if err != nil {
		t.Fatalf("RenderString with until function failed: %v", err)
	}

	// Should produce: item-0\nitem-1\nitem-2
	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d", len(lines))
	}

	for i, line := range lines {
		expected := "item-" + string(rune('0'+i))
		if line != expected {
			t.Errorf("line %d: expected %q, got %q", i, expected, line)
		}
	}
}

// TestRenderToWriter tests rendering directly to a writer.
func TestRenderToWriter(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")
	outputPath := filepath.Join(tmpDir, "output.txt")

	// Write template file
	templateContent := "Hello {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Create engine and render
	engine := NewGoTemplateEngine()
	data := map[string]string{"Name": "World"}

	// Open output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		t.Fatalf("failed to create output file: %v", err)
	}
	defer outFile.Close()

	// Render to writer
	if err := engine.RenderToWriter(context.Background(), templatePath, data, outFile); err != nil {
		t.Fatalf("RenderToWriter failed: %v", err)
	}

	// Close file to flush
	outFile.Close()

	// Read and verify output
	output, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	expected := "Hello World!"
	if string(output) != expected {
		t.Errorf("expected %q, got %q", expected, string(output))
	}
}

// TestRenderFromFile tests rendering a template from a file.
func TestRenderFromFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")

	// Write template file
	templateContent := "Hello {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Create engine and render
	engine := NewGoTemplateEngine()
	data := map[string]string{"Name": "World"}

	result, err := engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	expected := "Hello World!"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestRenderWithCaching tests that caching works correctly.
func TestRenderWithCaching(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")

	// Write template file
	templateContent := "Hello {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Create engine with caching enabled
	engine := NewGoTemplateEngine()
	data := map[string]string{"Name": "World"}

	// First render - should parse and cache
	result1, err := engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	// Verify template is cached
	if len(engine.cache) != 1 {
		t.Errorf("expected 1 cached template, got %d", len(engine.cache))
	}

	// Second render - should use cache
	result2, err := engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	// Results should be identical
	if string(result1) != string(result2) {
		t.Errorf("cached render produced different result")
	}

	// Cache size should still be 1
	if len(engine.cache) != 1 {
		t.Errorf("expected 1 cached template after second render, got %d", len(engine.cache))
	}
}

// TestRenderWithoutCaching tests that caching can be disabled.
func TestRenderWithoutCaching(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")

	// Write template file
	templateContent := "Hello {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Create engine with caching disabled
	engine := NewGoTemplateEngine()
	engine.SetCacheEnabled(false)

	data := map[string]string{"Name": "World"}

	// First render
	_, err := engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("first render failed: %v", err)
	}

	// Cache should be empty
	if len(engine.cache) != 0 {
		t.Errorf("expected empty cache, got %d entries", len(engine.cache))
	}

	// Second render
	_, err = engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("second render failed: %v", err)
	}

	// Cache should still be empty
	if len(engine.cache) != 0 {
		t.Errorf("expected empty cache after second render, got %d entries", len(engine.cache))
	}
}

// TestRegisterFunctions tests registering multiple functions at once.
func TestRegisterFunctions(t *testing.T) {
	engine := NewGoTemplateEngine()

	funcs := template.FuncMap{
		"func1": func(s string) string { return "1_" + s },
		"func2": func(s string) string { return "2_" + s },
		"func3": func(s string) string { return "3_" + s },
	}

	engine.RegisterFunctions(funcs)

	// Verify all functions were registered
	for name := range funcs {
		if _, ok := engine.funcMap[name]; !ok {
			t.Errorf("function %s was not registered", name)
		}
	}

	// Test using the registered functions
	templateContent := "{{func1 .Value}} {{func2 .Value}} {{func3 .Value}}"
	data := map[string]string{"Value": "test"}

	result, err := engine.RenderString(context.Background(), "test", templateContent, data)
	if err != nil {
		t.Fatalf("render with custom functions failed: %v", err)
	}

	expected := "1_test 2_test 3_test"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestLoadFromFile tests loading a template from a file.
func TestLoadFromFile(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")

	// Write template file
	templateContent := "Hello {{.Name}}!"
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Create engine and load template
	engine := NewGoTemplateEngine()
	if err := engine.LoadFromFile(templatePath); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify template is cached
	if len(engine.cache) != 1 {
		t.Errorf("expected 1 cached template, got %d", len(engine.cache))
	}

	// Render using the loaded template
	data := map[string]string{"Name": "World"}
	result, err := engine.Render(context.Background(), templatePath, data)
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}

	expected := "Hello World!"
	if string(result) != expected {
		t.Errorf("expected %q, got %q", expected, string(result))
	}
}

// TestValidateTemplate tests template validation.
func TestValidateTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewGoTemplateEngine()

	tests := []struct {
		name        string
		content     string
		shouldError bool
		description string
	}{
		{
			name:        "valid_simple",
			content:     "Hello {{.Name}}!",
			shouldError: false,
			description: "simple valid template",
		},
		{
			name:        "valid_with_range",
			content:     "{{range .Items}}{{.}}{{end}}",
			shouldError: false,
			description: "valid template with range",
		},
		{
			name:        "valid_with_if",
			content:     "{{if .Enabled}}yes{{else}}no{{end}}",
			shouldError: false,
			description: "valid template with if/else",
		},
		{
			name:        "valid_with_sprig",
			content:     "{{.Name | upper}}",
			shouldError: false,
			description: "valid template with Sprig function",
		},
		{
			name:        "invalid_missing_closing_brace",
			content:     "Hello {{.Name}!",
			shouldError: true,
			description: "missing closing braces",
		},
		{
			name:        "invalid_unclosed_range",
			content:     "{{range .Items}}{{.}}",
			shouldError: true,
			description: "unclosed range block",
		},
		{
			name:        "invalid_unclosed_if",
			content:     "{{if .Enabled}}yes",
			shouldError: true,
			description: "unclosed if block",
		},
		{
			name:        "invalid_mismatched_end",
			content:     "{{range .Items}}{{.}}{{end}}{{end}}",
			shouldError: true,
			description: "extra end statement",
		},
		{
			name:        "invalid_undefined_function",
			content:     "{{.Name | nonexistentfunc}}",
			shouldError: true,
			description: "undefined function call",
		},
		{
			name:        "valid_nested_blocks",
			content:     "{{range .Items}}{{if .Active}}{{.Name}}{{end}}{{end}}",
			shouldError: false,
			description: "nested range and if blocks",
		},
		{
			name:        "valid_with_comments",
			content:     "{{/* comment */}}Hello {{.Name}}",
			shouldError: false,
			description: "template with comments",
		},
		{
			name:        "invalid_bad_comment",
			content:     "{{/* unclosed comment",
			shouldError: true,
			description: "unclosed comment",
		},
		{
			name:        "invalid_bad_action",
			content:     "{{.Name",
			shouldError: true,
			description: "incomplete action",
		},
		{
			name:        "invalid_bad_pipe",
			content:     "{{.Name | | upper}}",
			shouldError: true,
			description: "double pipe operator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write template file
			templatePath := filepath.Join(tmpDir, tt.name+".tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}

			// Validate template
			err := engine.ValidateTemplate(templatePath)

			if tt.shouldError && err == nil {
				t.Errorf("%s: expected validation error but got none", tt.description)
			}

			if !tt.shouldError && err != nil {
				t.Errorf("%s: unexpected validation error: %v", tt.description, err)
			}
		})
	}
}

// TestValidateTemplateBeforeRender verifies that syntax errors are caught during validation
// before any rendering attempt, ensuring safe template processing.
func TestValidateTemplateBeforeRender(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewGoTemplateEngine()

	// Create a template with syntax error
	invalidPath := filepath.Join(tmpDir, "invalid.tmpl")
	if err := os.WriteFile(invalidPath, []byte("Hello {{.Name}!"), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Validation should catch the error
	err := engine.ValidateTemplate(invalidPath)
	if err == nil {
		t.Fatal("ValidateTemplate should have caught syntax error")
	}

	// Attempting to render should also fail with the same type of error
	_, renderErr := engine.Render(context.Background(), invalidPath, map[string]string{"Name": "World"})
	if renderErr == nil {
		t.Fatal("Render should have failed with syntax error")
	}

	// Both errors should indicate a parsing/syntax issue
	if !strings.Contains(err.Error(), "parse") && !strings.Contains(err.Error(), "template") {
		t.Errorf("validation error should mention parsing: %v", err)
	}
	if !strings.Contains(renderErr.Error(), "parse") && !strings.Contains(renderErr.Error(), "template") {
		t.Errorf("render error should mention parsing: %v", renderErr)
	}
}

// TestValidateTemplateWithRenderString tests that RenderString also validates syntax.
func TestValidateTemplateWithRenderString(t *testing.T) {
	engine := NewGoTemplateEngine()

	tests := []struct {
		name        string
		content     string
		shouldError bool
	}{
		{
			name:        "valid",
			content:     "Hello {{.Name}}!",
			shouldError: false,
		},
		{
			name:        "invalid_syntax",
			content:     "Hello {{.Name}!",
			shouldError: true,
		},
		{
			name:        "invalid_unclosed_range",
			content:     "{{range .Items}}{{.}}",
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := engine.RenderString(context.Background(), tt.name, tt.content, map[string]interface{}{
				"Name":  "World",
				"Items": []string{"a", "b"},
			})

			if tt.shouldError && err == nil {
				t.Error("expected error for invalid template syntax")
			}

			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error for valid template: %v", err)
			}
		})
	}
}

// TestRenderComplexTemplate tests rendering a complex template similar to existing ones.
func TestRenderComplexTemplate(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Template similar to inventory.tmpl
	templateContent := `[master]
{{- range $i, $e := until (int .Counts.Master) }}
test-master-{{ $i }}
{{- end }}

[worker]
{{- range $i, $e := until (int .Counts.Worker) }}
test-worker-{{ $i }}
{{- end }}`

	data := map[string]interface{}{
		"Counts": map[string]interface{}{
			"Master": 3,
			"Worker": 2,
		},
	}

	result, err := engine.RenderString(context.Background(), "test", templateContent, data)
	if err != nil {
		t.Fatalf("render complex template failed: %v", err)
	}

	output := string(result)

	// Verify master nodes
	if !strings.Contains(output, "test-master-0") {
		t.Error("missing test-master-0")
	}
	if !strings.Contains(output, "test-master-1") {
		t.Error("missing test-master-1")
	}
	if !strings.Contains(output, "test-master-2") {
		t.Error("missing test-master-2")
	}

	// Verify worker nodes
	if !strings.Contains(output, "test-worker-0") {
		t.Error("missing test-worker-0")
	}
	if !strings.Contains(output, "test-worker-1") {
		t.Error("missing test-worker-1")
	}
}

// TestRenderToWriterWithBuffer tests rendering to a buffer.
func TestRenderToWriterWithBuffer(t *testing.T) {
	engine := NewGoTemplateEngine()

	templateContent := "Hello {{.Name}}!"
	data := map[string]string{"Name": "World"}

	// Create a temporary file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	// Render to buffer
	var buf bytes.Buffer
	if err := engine.RenderToWriter(context.Background(), templatePath, data, &buf); err != nil {
		t.Fatalf("RenderToWriter failed: %v", err)
	}

	expected := "Hello World!"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

// TestRenderStringCancellation tests that RenderString respects context cancellation.
func TestRenderStringCancellation(t *testing.T) {
	engine := NewGoTemplateEngine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := engine.RenderString(ctx, "test", "Hello {{.Name}}", nil)
	if err == nil {
		t.Error("expected error when rendering with cancelled context")
	}
}

// TestRenderToWriterCancellation tests that RenderToWriter respects context cancellation.
func TestRenderToWriterCancellation(t *testing.T) {
	engine := NewGoTemplateEngine()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	var buf bytes.Buffer
	err := engine.RenderToWriter(ctx, "test.tmpl", nil, &buf)
	if err == nil {
		t.Error("expected error when rendering with cancelled context")
	}
}
