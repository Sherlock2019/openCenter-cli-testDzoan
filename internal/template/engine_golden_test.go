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
	"flag"
	"os"
	"path/filepath"
	"testing"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

// TestTemplateRenderingGolden validates template rendering against golden files.
func TestTemplateRenderingGolden(t *testing.T) {
	engine := NewGoTemplateEngine()

	testCases := []struct {
		name         string
		templatePath string
		data         interface{}
		goldenFile   string
	}{
		{
			name:         "simple template",
			templatePath: "testdata/simple.tmpl",
			data: map[string]interface{}{
				"Name":         "John Doe",
				"Organization": "Acme Corp",
			},
			goldenFile: "testdata/golden/simple.golden",
		},
		{
			name:         "cluster config template",
			templatePath: "testdata/cluster-config.tmpl",
			data: map[string]interface{}{
				"ClusterName":       "test-cluster",
				"Organization":      "test-org",
				"Provider":          "openstack",
				"KubernetesVersion": "1.28.0",
				"MasterCount":       3,
				"WorkerCount":       5,
				"PodCIDR":           "10.244.0.0/16",
				"ServiceCIDR":       "10.96.0.0/12",
			},
			goldenFile: "testdata/golden/cluster-config.golden",
		},
		{
			name:         "inventory template",
			templatePath: "testdata/inventory.tmpl",
			data: map[string]interface{}{
				"ClusterName": "prod-cluster",
				"MasterCount": 3,
				"WorkerCount": 2,
				"MasterIPs":   []string{"192.168.1.10", "192.168.1.11", "192.168.1.12"},
				"WorkerIPs":   []string{"192.168.1.20", "192.168.1.21"},
			},
			goldenFile: "testdata/golden/inventory.golden",
		},
		{
			name:         "service manifest template",
			templatePath: "testdata/service-manifest.tmpl",
			data: map[string]interface{}{
				"ServiceName": "web-service",
				"Namespace":   "production",
				"ServiceType": "loadbalancer",
				"Labels": map[string]string{
					"environment": "production",
					"team":        "platform",
				},
				"Ports": []map[string]interface{}{
					{
						"Name":       "http",
						"Port":       80,
						"TargetPort": 8080,
						"Protocol":   "tcp",
					},
					{
						"Name":       "https",
						"Port":       443,
						"TargetPort": 8443,
						"Protocol":   "tcp",
					},
				},
			},
			goldenFile: "testdata/golden/service-manifest.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Render the template
			result, err := engine.Render(context.Background(), tc.templatePath, tc.data)
			if err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			// Update golden file if flag is set
			if *updateGolden {
				goldenDir := filepath.Dir(tc.goldenFile)
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden directory: %v", err)
				}
				if err := os.WriteFile(tc.goldenFile, result, 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated golden file: %s", tc.goldenFile)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v\nRun with -update-golden to create it", tc.goldenFile, err)
			}

			// Compare result with golden file
			if string(result) != string(golden) {
				t.Errorf("rendered output does not match golden file\nGot:\n%s\n\nExpected:\n%s\n\nRun with -update-golden to update", string(result), string(golden))
			}
		})
	}
}

// TestTemplateRenderingGoldenWithCustomFunctions tests golden file validation with custom functions.
func TestTemplateRenderingGoldenWithCustomFunctions(t *testing.T) {
	engine := NewGoTemplateEngine()

	// Register custom functions
	engine.RegisterFunction("multiply", func(a, b int) int {
		return a * b
	})
	engine.RegisterFunction("concat", func(strs ...string) string {
		result := ""
		for _, s := range strs {
			result += s
		}
		return result
	})

	testCases := []struct {
		name         string
		templatePath string
		data         interface{}
		goldenFile   string
	}{
		{
			name:         "template with custom functions",
			templatePath: "testdata/custom-functions.tmpl",
			data: map[string]interface{}{
				"Value1": 5,
				"Value2": 10,
				"Prefix": "cluster",
				"Name":   "prod",
				"Suffix": "01",
			},
			goldenFile: "testdata/golden/custom-functions.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create template file if it doesn't exist (for testing purposes)
			templateContent := `Result: {{multiply .Value1 .Value2}}
Name: {{concat .Prefix "-" .Name "-" .Suffix}}`

			templateDir := filepath.Dir(tc.templatePath)
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatalf("failed to create template directory: %v", err)
			}
			if err := os.WriteFile(tc.templatePath, []byte(templateContent), 0644); err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}
			defer os.Remove(tc.templatePath)

			// Render the template
			result, err := engine.Render(context.Background(), tc.templatePath, tc.data)
			if err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			// Update golden file if flag is set
			if *updateGolden {
				goldenDir := filepath.Dir(tc.goldenFile)
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden directory: %v", err)
				}
				if err := os.WriteFile(tc.goldenFile, result, 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated golden file: %s", tc.goldenFile)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v\nRun with -update-golden to create it", tc.goldenFile, err)
			}

			// Compare result with golden file
			if string(result) != string(golden) {
				t.Errorf("rendered output does not match golden file\nGot:\n%s\n\nExpected:\n%s\n\nRun with -update-golden to update", string(result), string(golden))
			}
		})
	}
}

// TestTemplateRenderingGoldenEdgeCases tests edge cases with golden files.
func TestTemplateRenderingGoldenEdgeCases(t *testing.T) {
	engine := NewGoTemplateEngine()

	testCases := []struct {
		name         string
		templatePath string
		data         interface{}
		goldenFile   string
	}{
		{
			name:         "empty data",
			templatePath: "testdata/empty-data.tmpl",
			data:         map[string]interface{}{},
			goldenFile:   "testdata/golden/empty-data.golden",
		},
		{
			name:         "nil values",
			templatePath: "testdata/nil-values.tmpl",
			data: map[string]interface{}{
				"Value": nil,
			},
			goldenFile: "testdata/golden/nil-values.golden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create template files for edge cases
			var templateContent string
			switch tc.name {
			case "empty data":
				templateContent = "No data provided"
			case "nil values":
				templateContent = "Value: {{if .Value}}{{.Value}}{{else}}not set{{end}}"
			}

			templateDir := filepath.Dir(tc.templatePath)
			if err := os.MkdirAll(templateDir, 0755); err != nil {
				t.Fatalf("failed to create template directory: %v", err)
			}
			if err := os.WriteFile(tc.templatePath, []byte(templateContent), 0644); err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}
			defer os.Remove(tc.templatePath)

			// Render the template
			result, err := engine.Render(context.Background(), tc.templatePath, tc.data)
			if err != nil {
				t.Fatalf("failed to render template: %v", err)
			}

			// Update golden file if flag is set
			if *updateGolden {
				goldenDir := filepath.Dir(tc.goldenFile)
				if err := os.MkdirAll(goldenDir, 0755); err != nil {
					t.Fatalf("failed to create golden directory: %v", err)
				}
				if err := os.WriteFile(tc.goldenFile, result, 0644); err != nil {
					t.Fatalf("failed to update golden file: %v", err)
				}
				t.Logf("updated golden file: %s", tc.goldenFile)
				return
			}

			// Read golden file
			golden, err := os.ReadFile(tc.goldenFile)
			if err != nil {
				t.Fatalf("failed to read golden file %s: %v\nRun with -update-golden to create it", tc.goldenFile, err)
			}

			// Compare result with golden file
			if string(result) != string(golden) {
				t.Errorf("rendered output does not match golden file\nGot:\n%s\n\nExpected:\n%s\n\nRun with -update-golden to update", string(result), string(golden))
			}
		})
	}
}
