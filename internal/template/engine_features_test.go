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
	"reflect"
	"sort"
	"strings"
	"testing"
)

// TestAllExistingTemplateFeatures is a comprehensive test that verifies
// the Go template engine supports ALL features used in the existing codebase.
// This test validates the acceptance criterion:
// "Go template engine supports all existing template features"
func TestAllExistingTemplateFeatures(t *testing.T) {
	t.Run("Feature1_SprigFunctions", testSprigFunctions)
	t.Run("Feature2_UntilFunction", testUntilFunction)
	t.Run("Feature3_IntFunction", testIntFunction)
	t.Run("Feature4_RangeWithIndex", testRangeWithIndex)
	t.Run("Feature5_TrimSpaceActions", testTrimSpaceActions)
	t.Run("Feature6_NestedDataAccess", testNestedDataAccess)
	t.Run("Feature7_NamedTemplateExecution", testNamedTemplateExecution)
	t.Run("Feature8_MultipleTemplateCollection", testMultipleTemplateCollection)
	t.Run("Feature9_CustomFunctionRegistration", testCustomFunctionRegistration)
	t.Run("Feature10_TemplateFromEmbeddedFS", testTemplateFromEmbeddedFS)
	t.Run("Feature11_UpperFunction", testUpperFunction)
	t.Run("Feature12_PipelineOperations", testPipelineOperations)
}

// testSprigFunctions verifies that Sprig functions are available by default.
func testSprigFunctions(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Test various Sprig functions used in the codebase
	tests := []struct {
		name     string
		template string
		data     interface{}
		expected string
	}{
		{
			name:     "upper",
			template: "{{.Value | upper}}",
			data:     map[string]string{"Value": "hello"},
			expected: "HELLO",
		},
		{
			name:     "lower",
			template: "{{.Value | lower}}",
			data:     map[string]string{"Value": "HELLO"},
			expected: "hello",
		},
		{
			name:     "trim",
			template: "{{.Value | trim}}",
			data:     map[string]string{"Value": "  hello  "},
			expected: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderString(context.Background(), tt.name, tt.template, tt.data)
			if err != nil {
				t.Fatalf("RenderString failed: %v", err)
			}

			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

// testUntilFunction verifies the 'until' function used in inventory.tmpl.
func testUntilFunction(t *testing.T) {
	engine := NewGoTemplateEngine()

	// This pattern is used in internal/provision/templates/inventory.tmpl
	template := `{{- range $i, $e := until (int .Count) }}
node-{{ $i }}
{{- end }}`

	data := map[string]interface{}{"Count": 3}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	output := string(result)
	if !strings.Contains(output, "node-0") {
		t.Error("missing node-0")
	}
	if !strings.Contains(output, "node-1") {
		t.Error("missing node-1")
	}
	if !strings.Contains(output, "node-2") {
		t.Error("missing node-2")
	}
}

// testIntFunction verifies the 'int' type conversion function.
func testIntFunction(t *testing.T) {
	engine := NewGoTemplateEngine()

	template := `{{int .Value}}`
	data := map[string]interface{}{"Value": "42"}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if string(result) != "42" {
		t.Errorf("expected '42', got %q", string(result))
	}
}

// testRangeWithIndex verifies range iteration with index variable.
func testRangeWithIndex(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Pattern used in inventory.tmpl
	template := `{{- range $i, $e := until (int .Count) }}
item-{{ $i }}
{{- end }}`

	data := map[string]interface{}{"Count": 2}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(result)), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d", len(lines))
	}
}

// testTrimSpaceActions verifies {{- and -}} trim space actions.
func testTrimSpaceActions(t *testing.T) {
	engine := NewGoTemplateEngine()

	// The {{- trims whitespace before the action
	// The -}} trims whitespace after the action
	template := `{{- range $i, $e := until 2 -}}
line-{{ $i }}
{{- end -}}`

	data := map[string]interface{}{}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	// With proper trimming, we should get "line-0\nline-1" without extra whitespace
	output := string(result)
	if strings.HasPrefix(output, "\n") || strings.HasSuffix(output, "\n") {
		t.Errorf("template should not have leading/trailing newlines, got: %q", output)
	}
}

// testNestedDataAccess verifies accessing nested data structures.
func testNestedDataAccess(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Pattern used in inventory.tmpl: .IAC.Counts.master
	template := `{{.IAC.Counts.Master}}`

	data := map[string]interface{}{
		"IAC": map[string]interface{}{
			"Counts": map[string]interface{}{
				"Master": 3,
			},
		},
	}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if string(result) != "3" {
		t.Errorf("expected '3', got %q", string(result))
	}
}

// testNamedTemplateExecution verifies ExecuteTemplate functionality.
func testNamedTemplateExecution(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Execute named template
	data := map[string]string{"Name": "Test"}
	result, err := engine.ExecuteTemplate("simple.tmpl", data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed: %v", err)
	}

	if !strings.Contains(string(result), "Test") {
		t.Errorf("result should contain 'Test', got: %s", string(result))
	}
}

// testMultipleTemplateCollection verifies loading multiple templates together.
func testMultipleTemplateCollection(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load all templates from embedded FS
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Verify multiple templates are loaded
	expectedTemplates := []string{
		"simple.tmpl",
		"cluster-config.tmpl",
		"inventory.tmpl",
		"service-manifest.tmpl",
	}

	for _, name := range expectedTemplates {
		tmpl, err := engine.GetTemplate(name)
		if err != nil {
			t.Errorf("template %s not found: %v", name, err)
		}
		if tmpl == nil {
			t.Errorf("template %s is nil", name)
		}
	}
}

// testCustomFunctionRegistration verifies custom function registration.
func testCustomFunctionRegistration(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Register custom functions similar to provision package
	engine.RegisterFunction("hcl", func(v interface{}) string {
		return fmt.Sprintf("hcl(%v)", v)
	})

	engine.RegisterFunction("sortedKeys", func(m map[string]interface{}) []string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		return keys
	})

	// Test custom function
	template := `{{hcl .Value}}`
	data := map[string]string{"Value": "test"}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if !strings.Contains(string(result), "hcl(test)") {
		t.Errorf("expected 'hcl(test)', got %q", string(result))
	}
}

// testTemplateFromEmbeddedFS verifies loading templates from embedded filesystem.
func testTemplateFromEmbeddedFS(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Load from embedded FS (testTemplates is defined in engine_integration_test.go)
	if err := engine.LoadFromFS(testTemplates, "testdata/*.tmpl"); err != nil {
		t.Fatalf("LoadFromFS failed: %v", err)
	}

	// Verify templates are loaded
	tmpl, err := engine.GetTemplate("simple.tmpl")
	if err != nil {
		t.Fatalf("GetTemplate failed: %v", err)
	}

	if tmpl == nil {
		t.Error("expected non-nil template")
	}
}

// testUpperFunction verifies the upper function from Sprig.
func testUpperFunction(t *testing.T) {
	engine := NewGoTemplateEngine()

	template := `{{.Name | upper}}`
	data := map[string]string{"Name": "hello"}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if string(result) != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", string(result))
	}
}

// testPipelineOperations verifies pipeline operations with multiple functions.
func testPipelineOperations(t *testing.T) {
	engine := NewGoTemplateEngine()

	template := `{{.Value | trim | upper}}`
	data := map[string]string{"Value": "  hello  "}

	result, err := engine.RenderString(context.Background(), "test", template, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	if string(result) != "HELLO" {
		t.Errorf("expected 'HELLO', got %q", string(result))
	}
}

// TestCompatibilityWithProvisionTemplates tests compatibility with actual provision templates.
func TestCompatibilityWithProvisionTemplates(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Register custom functions used in provision package
	engine.RegisterFunction("hcl", hclRender)
	engine.RegisterFunction("sortedKeys", sortedKeys)

	// Test inventory.tmpl pattern
	inventoryTemplate := `[master]
{{- range $i, $e := until (int .IAC.Counts.master) }}
test-master-{{ $i }}
{{- end }}

[worker]
{{- range $i, $e := until (int .IAC.Counts.worker) }}
test-worker-{{ $i }}
{{- end }}`

	data := map[string]interface{}{
		"IAC": map[string]interface{}{
			"Counts": map[string]interface{}{
				"master": 2,
				"worker": 3,
			},
		},
	}

	result, err := engine.RenderString(context.Background(), "inventory", inventoryTemplate, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	output := string(result)

	// Verify master nodes
	if !strings.Contains(output, "test-master-0") {
		t.Error("missing test-master-0")
	}
	if !strings.Contains(output, "test-master-1") {
		t.Error("missing test-master-1")
	}

	// Verify worker nodes
	if !strings.Contains(output, "test-worker-0") {
		t.Error("missing test-worker-0")
	}
	if !strings.Contains(output, "test-worker-1") {
		t.Error("missing test-worker-1")
	}
	if !strings.Contains(output, "test-worker-2") {
		t.Error("missing test-worker-2")
	}
}

// TestCompatibilityWithAnsibleTemplates tests compatibility with ansible.cfg.tmpl.
func TestCompatibilityWithAnsibleTemplates(t *testing.T) {
	engine := NewGoTemplateEngine()

	// ansible.cfg.tmpl is a simple template
	ansibleTemplate := `[defaults]
inventory = inventory
host_key_checking = False`

	data := map[string]interface{}{}

	result, err := engine.RenderString(context.Background(), "ansible", ansibleTemplate, data)
	if err != nil {
		t.Fatalf("RenderString failed: %v", err)
	}

	output := string(result)
	if !strings.Contains(output, "[defaults]") {
		t.Error("missing [defaults] section")
	}
	if !strings.Contains(output, "inventory = inventory") {
		t.Error("missing inventory setting")
	}
}

// TestLoadFromFileAndExecuteTemplate tests the workflow used in provision packages.
func TestLoadFromFileAndExecuteTemplate(t *testing.T) {
	// Create temporary directory with templates
	tmpDir := t.TempDir()

	// Create template files similar to provision/templates
	templates := map[string]string{
		"ansible.cfg.tmpl": `[defaults]
inventory = inventory
host_key_checking = False`,
		"inventory.tmpl": `[master]
{{- range $i, $e := until (int .Counts.Master) }}
master-{{ $i }}
{{- end }}`,
	}

	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write template %s: %v", name, err)
		}
	}

	// Create engine and load templates
	engine := NewGoTemplateEngine()

	// Load all templates
	for name := range templates {
		path := filepath.Join(tmpDir, name)
		if err := engine.LoadFromFile(path); err != nil {
			t.Fatalf("LoadFromFile failed for %s: %v", name, err)
		}
	}

	// Execute ansible.cfg.tmpl
	ansiblePath := filepath.Join(tmpDir, "ansible.cfg.tmpl")
	result1, err := engine.ExecuteTemplate(ansiblePath, nil)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed for ansible.cfg: %v", err)
	}

	if !strings.Contains(string(result1), "[defaults]") {
		t.Error("ansible.cfg template not rendered correctly")
	}

	// Execute inventory.tmpl
	inventoryPath := filepath.Join(tmpDir, "inventory.tmpl")
	data := map[string]interface{}{
		"Counts": map[string]interface{}{
			"Master": 2,
		},
	}

	result2, err := engine.ExecuteTemplate(inventoryPath, data)
	if err != nil {
		t.Fatalf("ExecuteTemplate failed for inventory: %v", err)
	}

	if !strings.Contains(string(result2), "master-0") {
		t.Error("inventory template not rendered correctly")
	}
}

// Helper functions from provision package for testing

func hclRender(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "null"
	case string:
		s := strings.TrimSpace(t)
		if isExpr(s) {
			return s
		}
		return fmt.Sprintf("\"%s\"", escapeQuotes(s))
	case bool:
		if t {
			return "true"
		}
		return "false"
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%v", t)
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", t)
	case float32, float64:
		return fmt.Sprintf("%v", t)
	case []interface{}:
		parts := make([]string, 0, len(t))
		for _, e := range t {
			parts = append(parts, hclRender(e))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, k := range keys {
			parts = append(parts, fmt.Sprintf("%s = %s", k, hclRender(t[k])))
		}
		if len(parts) == 0 {
			return "{}"
		}
		return "{ " + strings.Join(parts, " ") + " }"
	default:
		rv := reflect.ValueOf(v)
		switch rv.Kind() {
		case reflect.Map:
			iter := rv.MapRange()
			tmp := map[string]interface{}{}
			for iter.Next() {
				k := fmt.Sprintf("%v", iter.Key().Interface())
				tmp[k] = iter.Value().Interface()
			}
			return hclRender(tmp)
		case reflect.Slice, reflect.Array:
			n := rv.Len()
			parts := make([]string, 0, n)
			for i := 0; i < n; i++ {
				parts = append(parts, hclRender(rv.Index(i).Interface()))
			}
			return "[" + strings.Join(parts, ", ") + "]"
		}
		return fmt.Sprintf("\"%v\"", v)
	}
}

func isExpr(s string) bool {
	if strings.HasPrefix(s, "${") && strings.HasSuffix(s, "}") {
		return true
	}
	if strings.Contains(s, "local.") || strings.Contains(s, "var.") || strings.Contains(s, "module.") {
		return true
	}
	if i := strings.Index(s, "("); i > 0 {
		return true
	}
	return false
}

func escapeQuotes(s string) string {
	return strings.ReplaceAll(s, "\"", "\\\"")
}

func sortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
