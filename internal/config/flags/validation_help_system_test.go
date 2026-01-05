// Copyright 2025 Victor Palma <victor.palma@rackspace.com>
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flags

import (
	"strings"
	"testing"
	"time"
)

func TestDefaultConfigurationValidator_ValidateConfiguration(t *testing.T) {
	validator := NewDefaultConfigurationValidator()
	
	tests := []struct {
		name           string
		config         *Configuration
		expectValid    bool
		expectErrors   int
		expectWarnings int
		checkFunc      func(*testing.T, *ValidationResult)
	}{
		{
			name:        "nil configuration",
			config:      nil,
			expectValid: false,
			expectErrors: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if len(result.Errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(result.Errors))
				}
				if result.Errors[0].Type != "null_config" {
					t.Errorf("Expected null_config error, got %s", result.Errors[0].Type)
				}
			},
		},
		{
			name: "valid configuration",
			config: &Configuration{
				Data: map[string]interface{}{
					"cluster": map[string]interface{}{
						"name": "test-cluster",
					},
					"infrastructure": map[string]interface{}{
						"provider": "openstack",
					},
				},
				Sources: []ConfigSource{
					{Type: SourceCLI, Path: "test", Priority: 1},
				},
				Metadata: ConfigMetadata{
					ProcessedAt: time.Now(),
				},
			},
			expectValid: true,
			expectErrors: 0,
		},
		{
			name: "configuration with warnings",
			config: &Configuration{
				Data: map[string]interface{}{
					"cluster name": "test-cluster", // Space in key should generate warning
					"infrastructure": map[string]interface{}{
						"provider": "openstack",
					},
				},
				Sources: []ConfigSource{
					{Type: SourceCLI, Path: "test", Priority: 1},
				},
				Metadata: ConfigMetadata{
					ProcessedAt: time.Now(),
				},
			},
			expectValid:    true,
			expectWarnings: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if len(result.Warnings) != 1 {
					t.Errorf("Expected 1 warning, got %d", len(result.Warnings))
				}
				if result.Warnings[0].Type != "key_format" {
					t.Errorf("Expected key_format warning, got %s", result.Warnings[0].Type)
				}
			},
		},
		{
			name: "configuration with invalid sources",
			config: &Configuration{
				Data: map[string]interface{}{
					"cluster": map[string]interface{}{
						"name": "test-cluster",
					},
				},
				Sources: []ConfigSource{
					{Type: "", Path: "test", Priority: 1}, // Empty type should generate error
				},
				Metadata: ConfigMetadata{
					ProcessedAt: time.Now(),
				},
			},
			expectValid:  false,
			expectErrors: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if len(result.Errors) != 1 {
					t.Errorf("Expected 1 error, got %d", len(result.Errors))
				}
				if result.Errors[0].Type != "source_type" {
					t.Errorf("Expected source_type error, got %s", result.Errors[0].Type)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateConfiguration(tt.config)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}
			
			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
			}
			
			if len(result.Warnings) != tt.expectWarnings {
				t.Errorf("Expected %d warnings, got %d", tt.expectWarnings, len(result.Warnings))
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestDefaultConfigurationValidator_ValidateFlags(t *testing.T) {
	validator := NewDefaultConfigurationValidator()
	
	tests := []struct {
		name         string
		flags        *ParsedFlags
		expectValid  bool
		expectErrors int
		checkFunc    func(*testing.T, *ValidationResult)
	}{
		{
			name:         "nil flags",
			flags:        nil,
			expectValid:  false,
			expectErrors: 1,
		},
		{
			name: "valid flags",
			flags: &ParsedFlags{
				DotNotation: map[string]string{
					"cluster.name": "test-cluster",
				},
				ArrayFlags: []ArrayFlag{
					{
						Type: "server-pool",
						Config: &ArrayConfig{
							Path:   "infrastructure.server_pools",
							Fields: map[string]interface{}{"name": "compute"},
						},
					},
				},
				JSONFlags: []JSONFlag{
					{
						Path:  "networking",
						Value: map[string]interface{}{"dns_servers": []string{"8.8.8.8"}},
					},
				},
				YAMLFlags: []YAMLFlag{
					{
						Path:  "cluster",
						Value: map[string]interface{}{"version": "1.0"},
					},
				},
				TemplateVars: map[string]string{
					"CLUSTER_NAME": "production",
				},
				ConfigFileFlags: []*ConfigFileFlag{
					{
						Path:     "/path/to/config.yaml",
						Type:     "yaml",
						Priority: 1,
					},
				},
			},
			expectValid: true,
		},
		{
			name: "invalid dot notation",
			flags: &ParsedFlags{
				DotNotation: map[string]string{
					".invalid.path": "value", // Path starting with dot
				},
			},
			expectValid:  false,
			expectErrors: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if result.Errors[0].Type != "dot_notation" {
					t.Errorf("Expected dot_notation error, got %s", result.Errors[0].Type)
				}
			},
		},
		{
			name: "invalid array flag",
			flags: &ParsedFlags{
				ArrayFlags: []ArrayFlag{
					{
						Type:   "server-pool",
						Config: nil, // Nil config should generate error
					},
				},
			},
			expectValid:  false,
			expectErrors: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if result.Errors[0].Type != "array_flag" {
					t.Errorf("Expected array_flag error, got %s", result.Errors[0].Type)
				}
			},
		},
		{
			name: "invalid template variable",
			flags: &ParsedFlags{
				TemplateVars: map[string]string{
					"INVALID NAME": "value", // Space in name should generate error
				},
			},
			expectValid:  false,
			expectErrors: 1,
			checkFunc: func(t *testing.T, result *ValidationResult) {
				if result.Errors[0].Type != "template_var" {
					t.Errorf("Expected template_var error, got %s", result.Errors[0].Type)
				}
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateFlags(tt.flags)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if result.Valid != tt.expectValid {
				t.Errorf("Expected valid=%v, got %v", tt.expectValid, result.Valid)
			}
			
			if len(result.Errors) != tt.expectErrors {
				t.Errorf("Expected %d errors, got %d", tt.expectErrors, len(result.Errors))
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestDefaultConfigurationValidator_ValidateFlag(t *testing.T) {
	validator := NewDefaultConfigurationValidator()
	
	tests := []struct {
		name      string
		flagName  string
		value     string
		expectErr bool
	}{
		{
			name:     "valid flag",
			flagName: "cluster.name",
			value:    "test-cluster",
		},
		{
			name:      "empty flag name",
			flagName:  "",
			value:     "value",
			expectErr: true,
		},
		{
			name:      "empty flag value",
			flagName:  "cluster.name",
			value:     "",
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFlag(tt.flagName, tt.value)
			
			if tt.expectErr && err == nil {
				t.Error("Expected error, but got none")
			}
			
			if !tt.expectErr && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestDefaultHelpSystem_GetFlagHelp(t *testing.T) {
	helpSystem := NewDefaultHelpSystem()
	
	tests := []struct {
		name      string
		flagType  FlagType
		expectErr bool
		checkFunc func(*testing.T, string)
	}{
		{
			name:     "dot notation help",
			flagType: FlagTypeDotNotation,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "Dot Notation Flags") {
					t.Error("Help should contain 'Dot Notation Flags'")
				}
				if !strings.Contains(help, "--path.to.field=value") {
					t.Error("Help should contain syntax example")
				}
			},
		},
		{
			name:     "array help",
			flagType: FlagTypeArray,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "Array Flags") {
					t.Error("Help should contain 'Array Flags'")
				}
				if !strings.Contains(help, "--server-pool") {
					t.Error("Help should contain server-pool example")
				}
			},
		},
		{
			name:     "JSON help",
			flagType: FlagTypeJSON,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "JSON Flags") {
					t.Error("Help should contain 'JSON Flags'")
				}
				if !strings.Contains(help, "--json-set") {
					t.Error("Help should contain json-set example")
				}
			},
		},
		{
			name:     "YAML help",
			flagType: FlagTypeYAML,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "YAML Flags") {
					t.Error("Help should contain 'YAML Flags'")
				}
				if !strings.Contains(help, "--yaml-set") {
					t.Error("Help should contain yaml-set example")
				}
			},
		},
		{
			name:     "template help",
			flagType: FlagTypeTemplate,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "Template Variables") {
					t.Error("Help should contain 'Template Variables'")
				}
				if !strings.Contains(help, "--template-var") {
					t.Error("Help should contain template-var example")
				}
			},
		},
		{
			name:     "file help",
			flagType: FlagTypeFile,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "Configuration File Flags") {
					t.Error("Help should contain 'Configuration File Flags'")
				}
				if !strings.Contains(help, "--base-config") {
					t.Error("Help should contain base-config example")
				}
			},
		},
		{
			name:     "output help",
			flagType: FlagTypeOutput,
			checkFunc: func(t *testing.T, help string) {
				if !strings.Contains(help, "Output Format Flags") {
					t.Error("Help should contain 'Output Format Flags'")
				}
				if !strings.Contains(help, "--output-format") {
					t.Error("Help should contain output-format example")
				}
			},
		},
		{
			name:      "unsupported flag type",
			flagType:  FlagType("unsupported"),
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			help, err := helpSystem.GetFlagHelp(tt.flagType)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if help == "" {
				t.Error("Help text should not be empty")
			}
			
			if tt.checkFunc != nil {
				tt.checkFunc(t, help)
			}
		})
	}
}

func TestDefaultHelpSystem_GetAllFlagHelp(t *testing.T) {
	helpSystem := NewDefaultHelpSystem()
	
	help, err := helpSystem.GetAllFlagHelp()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if help == "" {
		t.Error("Help text should not be empty")
	}
	
	// Check that help contains sections for all flag types
	expectedSections := []string{
		"Enhanced CLI Configuration Flags",
		"Dot Notation Flags",
		"Array Flags",
		"JSON Flags",
		"YAML Flags",
		"Template Variables",
		"Configuration File Flags",
		"Output Format Flags",
	}
	
	for _, section := range expectedSections {
		if !strings.Contains(help, section) {
			t.Errorf("Help should contain section '%s'", section)
		}
	}
}

func TestDefaultHelpSystem_GetExamples(t *testing.T) {
	helpSystem := NewDefaultHelpSystem()
	
	examples, err := helpSystem.GetExamples()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
		return
	}
	
	if examples == "" {
		t.Error("Examples should not be empty")
	}
	
	// Check that examples contain common patterns
	expectedPatterns := []string{
		"Common Configuration Examples",
		"Basic Configuration",
		"Array Configuration",
		"JSON Configuration",
		"YAML Configuration",
		"File Merging",
		"Configuration Stack",
		"Template Variables",
		"Output Formatting",
		"Complex Example",
	}
	
	for _, pattern := range expectedPatterns {
		if !strings.Contains(examples, pattern) {
			t.Errorf("Examples should contain pattern '%s'", pattern)
		}
	}
}

func TestDefaultHelpSystem_GetFlagExamples(t *testing.T) {
	helpSystem := NewDefaultHelpSystem()
	
	tests := []struct {
		name         string
		flagType     FlagType
		expectErr    bool
		minExamples  int
		checkExample func(*testing.T, []string)
	}{
		{
			name:        "dot notation examples",
			flagType:    FlagTypeDotNotation,
			minExamples: 2,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--cluster.name=") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain cluster.name example")
				}
			},
		},
		{
			name:        "array examples",
			flagType:    FlagTypeArray,
			minExamples: 3,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--server-pool") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain server-pool example")
				}
			},
		},
		{
			name:        "JSON examples",
			flagType:    FlagTypeJSON,
			minExamples: 1,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--json-set") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain json-set example")
				}
			},
		},
		{
			name:        "YAML examples",
			flagType:    FlagTypeYAML,
			minExamples: 1,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--yaml-set") || strings.Contains(example, "--yaml-file") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain yaml example")
				}
			},
		},
		{
			name:        "template examples",
			flagType:    FlagTypeTemplate,
			minExamples: 1,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--template-var") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain template-var example")
				}
			},
		},
		{
			name:        "file examples",
			flagType:    FlagTypeFile,
			minExamples: 2,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--base-config") || strings.Contains(example, "--config-stack") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain file flag example")
				}
			},
		},
		{
			name:        "output examples",
			flagType:    FlagTypeOutput,
			minExamples: 2,
			checkExample: func(t *testing.T, examples []string) {
				found := false
				for _, example := range examples {
					if strings.Contains(example, "--output-format") || strings.Contains(example, "--dry-run") {
						found = true
						break
					}
				}
				if !found {
					t.Error("Should contain output flag example")
				}
			},
		},
		{
			name:      "unsupported flag type",
			flagType:  FlagType("unsupported"),
			expectErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			examples, err := helpSystem.GetFlagExamples(tt.flagType)
			
			if tt.expectErr {
				if err == nil {
					t.Error("Expected error, but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(examples) < tt.minExamples {
				t.Errorf("Expected at least %d examples, got %d", tt.minExamples, len(examples))
			}
			
			if tt.checkExample != nil {
				tt.checkExample(t, examples)
			}
		})
	}
}

func TestCLIIntegration_ValidationAndHelp(t *testing.T) {
	integration, err := NewCLIIntegration()
	if err != nil {
		t.Fatalf("Failed to create CLI integration: %v", err)
	}
	
	// Test help methods
	t.Run("GetAllFlagHelp", func(t *testing.T) {
		help, err := integration.GetAllFlagHelp()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if help == "" {
			t.Error("Help should not be empty")
		}
	})
	
	t.Run("GetExamples", func(t *testing.T) {
		examples, err := integration.GetExamples()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if examples == "" {
			t.Error("Examples should not be empty")
		}
	})
	
	t.Run("GetFlagHelp", func(t *testing.T) {
		help, err := integration.GetFlagHelp(FlagTypeDotNotation)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if help == "" {
			t.Error("Help should not be empty")
		}
	})
	
	t.Run("GetFlagExamples", func(t *testing.T) {
		examples, err := integration.GetFlagExamples(FlagTypeArray)
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if len(examples) == 0 {
			t.Error("Examples should not be empty")
		}
	})
}