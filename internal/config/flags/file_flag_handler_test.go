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
	"os"
	"path/filepath"
	"testing"
)

func TestFileFlagHandler_CanHandle(t *testing.T) {
	handler := NewFileFlagHandler()

	tests := []struct {
		name     string
		flagName string
		expected bool
	}{
		{"base-config flag", "base-config", true},
		{"base-config with path", "base-config-cluster", true},
		{"merge-config flag", "merge-config", true},
		{"merge-config with path", "merge-config-override", true},
		{"config-stack flag", "config-stack", true},
		{"config-stack with path", "config-stack-all", true},
		{"other flag", "server-pool", false},
		{"json flag", "json-set", false},
		{"yaml flag", "yaml-set", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.CanHandle(tt.flagName)
			if result != tt.expected {
				t.Errorf("CanHandle(%s) = %v, expected %v", tt.flagName, result, tt.expected)
			}
		})
	}
}

func TestFileFlagHandler_GetFlagType(t *testing.T) {
	handler := NewFileFlagHandler()

	if handler.GetFlagType() != FlagTypeFile {
		t.Errorf("GetFlagType() = %v, expected %v", handler.GetFlagType(), FlagTypeFile)
	}
}

func TestFileFlagHandler_ParseFlag_SingleFile(t *testing.T) {
	handler := NewFileFlagHandler()

	// Create a temporary YAML file
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write some YAML content
	yamlContent := `
opencenter:
  cluster:
    name: test-cluster
  infrastructure:
    provider: openstack
`
	if _, err := tmpFile.WriteString(yamlContent); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name      string
		flagName  string
		value     string
		expectErr bool
		checkFunc func(*testing.T, interface{})
	}{
		{
			name:     "valid base-config file",
			flagName: "base-config",
			value:    tmpFile.Name(),
			checkFunc: func(t *testing.T, result interface{}) {
				configFlag, ok := result.(*ConfigFileFlag)
				if !ok {
					t.Errorf("Expected *ConfigFileFlag, got %T", result)
					return
				}
				if configFlag.MergeType != ConfigFileMergeBase {
					t.Errorf("Expected merge type %v, got %v", ConfigFileMergeBase, configFlag.MergeType)
				}
				if configFlag.Priority != 1 {
					t.Errorf("Expected priority 1, got %d", configFlag.Priority)
				}
				if configFlag.Type != "yaml" {
					t.Errorf("Expected type yaml, got %s", configFlag.Type)
				}
			},
		},
		{
			name:     "valid merge-config file",
			flagName: "merge-config",
			value:    tmpFile.Name(),
			checkFunc: func(t *testing.T, result interface{}) {
				configFlag, ok := result.(*ConfigFileFlag)
				if !ok {
					t.Errorf("Expected *ConfigFileFlag, got %T", result)
					return
				}
				if configFlag.MergeType != ConfigFileMergeOverride {
					t.Errorf("Expected merge type %v, got %v", ConfigFileMergeOverride, configFlag.MergeType)
				}
				if configFlag.Priority != 2 {
					t.Errorf("Expected priority 2, got %d", configFlag.Priority)
				}
			},
		},
		{
			name:      "empty file path",
			flagName:  "base-config",
			value:     "",
			expectErr: true,
		},
		{
			name:      "non-existent file",
			flagName:  "base-config",
			value:     "/non/existent/file.yaml",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.ParseFlag(tt.flagName, tt.value)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestFileFlagHandler_ParseFlag_ConfigStack(t *testing.T) {
	handler := NewFileFlagHandler()

	// Create temporary files
	tmpFile1, err := os.CreateTemp("", "test-config1-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp file 1: %v", err)
	}
	defer os.Remove(tmpFile1.Name())
	tmpFile1.WriteString("config1: value1")
	tmpFile1.Close()

	tmpFile2, err := os.CreateTemp("", "test-config2-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file 2: %v", err)
	}
	defer os.Remove(tmpFile2.Name())
	tmpFile2.WriteString(`{"config2": "value2"}`)
	tmpFile2.Close()

	tests := []struct {
		name      string
		flagName  string
		value     string
		expectErr bool
		checkFunc func(*testing.T, interface{})
	}{
		{
			name:     "valid config stack",
			flagName: "config-stack",
			value:    tmpFile1.Name() + "," + tmpFile2.Name(),
			checkFunc: func(t *testing.T, result interface{}) {
				configFlags, ok := result.([]*ConfigFileFlag)
				if !ok {
					t.Errorf("Expected []*ConfigFileFlag, got %T", result)
					return
				}
				if len(configFlags) != 2 {
					t.Errorf("Expected 2 config flags, got %d", len(configFlags))
					return
				}

				// Check first file
				if configFlags[0].Type != "yaml" {
					t.Errorf("Expected first file type yaml, got %s", configFlags[0].Type)
				}
				if configFlags[0].Priority != 1 {
					t.Errorf("Expected first file priority 1, got %d", configFlags[0].Priority)
				}
				if configFlags[0].MergeType != ConfigFileMergeStack {
					t.Errorf("Expected first file merge type %v, got %v", ConfigFileMergeStack, configFlags[0].MergeType)
				}

				// Check second file
				if configFlags[1].Type != "json" {
					t.Errorf("Expected second file type json, got %s", configFlags[1].Type)
				}
				if configFlags[1].Priority != 2 {
					t.Errorf("Expected second file priority 2, got %d", configFlags[1].Priority)
				}
			},
		},
		{
			name:     "config stack with spaces",
			flagName: "config-stack",
			value:    " " + tmpFile1.Name() + " , " + tmpFile2.Name() + " ",
			checkFunc: func(t *testing.T, result interface{}) {
				configFlags, ok := result.([]*ConfigFileFlag)
				if !ok {
					t.Errorf("Expected []*ConfigFileFlag, got %T", result)
					return
				}
				if len(configFlags) != 2 {
					t.Errorf("Expected 2 config flags, got %d", len(configFlags))
				}
			},
		},
		{
			name:      "empty config stack",
			flagName:  "config-stack",
			value:     "",
			expectErr: true,
		},
		{
			name:      "config stack with non-existent file",
			flagName:  "config-stack",
			value:     tmpFile1.Name() + ",/non/existent/file.yaml",
			expectErr: true,
		},
		{
			name:      "config stack with only commas",
			flagName:  "config-stack",
			value:     ",,,",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := handler.ParseFlag(tt.flagName, tt.value)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, result)
			}
		})
	}
}

func TestFileFlagHandler_LoadConfigurationFile(t *testing.T) {
	handler := NewFileFlagHandler()

	// Create temporary YAML file
	tmpYAMLFile, err := os.CreateTemp("", "test-config-*.yaml")
	if err != nil {
		t.Fatalf("Failed to create temp YAML file: %v", err)
	}
	defer os.Remove(tmpYAMLFile.Name())

	yamlContent := `
opencenter:
  cluster:
    name: test-cluster
  infrastructure:
    provider: openstack
    server_pools:
      - name: compute
        worker_count: 3
`
	tmpYAMLFile.WriteString(yamlContent)
	tmpYAMLFile.Close()

	// Create temporary JSON file
	tmpJSONFile, err := os.CreateTemp("", "test-config-*.json")
	if err != nil {
		t.Fatalf("Failed to create temp JSON file: %v", err)
	}
	defer os.Remove(tmpJSONFile.Name())

	jsonContent := `{
		"opencenter": {
			"cluster": {
				"name": "json-cluster"
			},
			"networking": {
				"dns_servers": ["8.8.8.8", "8.8.4.4"]
			}
		}
	}`
	tmpJSONFile.WriteString(jsonContent)
	tmpJSONFile.Close()

	tests := []struct {
		name       string
		configFlag *ConfigFileFlag
		expectErr  bool
		checkFunc  func(*testing.T, *Configuration)
	}{
		{
			name: "load YAML file",
			configFlag: &ConfigFileFlag{
				Path:      tmpYAMLFile.Name(),
				Type:      "yaml",
				Priority:  1,
				MergeType: ConfigFileMergeBase,
			},
			checkFunc: func(t *testing.T, config *Configuration) {
				if config.Data == nil {
					t.Error("Expected non-nil data")
					return
				}

				// Check nested structure
				opencenter, ok := config.Data["opencenter"].(map[string]interface{})
				if !ok {
					t.Error("Expected opencenter to be a map")
					return
				}

				cluster, ok := opencenter["cluster"].(map[string]interface{})
				if !ok {
					t.Error("Expected cluster to be a map")
					return
				}

				if cluster["name"] != "test-cluster" {
					t.Errorf("Expected cluster name 'test-cluster', got %v", cluster["name"])
				}

				// Check sources
				if len(config.Sources) != 1 {
					t.Errorf("Expected 1 source, got %d", len(config.Sources))
				} else {
					source := config.Sources[0]
					if source.Type != SourceFile {
						t.Errorf("Expected source type %v, got %v", SourceFile, source.Type)
					}
					if source.Priority != 1 {
						t.Errorf("Expected source priority 1, got %d", source.Priority)
					}
				}
			},
		},
		{
			name: "load JSON file",
			configFlag: &ConfigFileFlag{
				Path:      tmpJSONFile.Name(),
				Type:      "json",
				Priority:  2,
				MergeType: ConfigFileMergeOverride,
			},
			checkFunc: func(t *testing.T, config *Configuration) {
				if config.Data == nil {
					t.Error("Expected non-nil data")
					return
				}

				// Check nested structure
				opencenter, ok := config.Data["opencenter"].(map[string]interface{})
				if !ok {
					t.Error("Expected opencenter to be a map")
					return
				}

				cluster, ok := opencenter["cluster"].(map[string]interface{})
				if !ok {
					t.Error("Expected cluster to be a map")
					return
				}

				if cluster["name"] != "json-cluster" {
					t.Errorf("Expected cluster name 'json-cluster', got %v", cluster["name"])
				}
			},
		},
		{
			name: "non-existent file",
			configFlag: &ConfigFileFlag{
				Path:      "/non/existent/file.yaml",
				Type:      "yaml",
				Priority:  1,
				MergeType: ConfigFileMergeBase,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := handler.LoadConfigurationFile(tt.configFlag)

			if tt.expectErr {
				if err == nil {
					t.Errorf("Expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, config)
			}
		})
	}
}

func TestFileFlagHandler_DetectFileType(t *testing.T) {
	handler := NewFileFlagHandler()

	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{"YAML file with .yaml extension", "/path/to/config.yaml", "yaml"},
		{"YAML file with .yml extension", "/path/to/config.yml", "yaml"},
		{"JSON file", "/path/to/config.json", "json"},
		{"File without extension", "/path/to/config", "yaml"},          // Default to YAML
		{"File with unknown extension", "/path/to/config.txt", "yaml"}, // Default to YAML
		{"Uppercase extensions", "/path/to/CONFIG.YAML", "yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handler.detectFileType(tt.filePath)
			if result != tt.expected {
				t.Errorf("detectFileType(%s) = %s, expected %s", tt.filePath, result, tt.expected)
			}
		})
	}
}

func TestFileFlagHandler_Integration(t *testing.T) {
	// Create temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "file-flag-test-")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create base config file
	baseConfigPath := filepath.Join(tmpDir, "base.yaml")
	baseConfig := `
opencenter:
  cluster:
    name: base-cluster
  infrastructure:
    provider: openstack
    server_pools:
      - name: control
        worker_count: 3
`
	if err := os.WriteFile(baseConfigPath, []byte(baseConfig), 0644); err != nil {
		t.Fatalf("Failed to write base config: %v", err)
	}

	// Create override config file
	overrideConfigPath := filepath.Join(tmpDir, "override.yaml")
	overrideConfig := `
opencenter:
  cluster:
    name: override-cluster
  infrastructure:
    server_pools:
      - name: compute
        worker_count: 5
`
	if err := os.WriteFile(overrideConfigPath, []byte(overrideConfig), 0644); err != nil {
		t.Fatalf("Failed to write override config: %v", err)
	}

	handler := NewFileFlagHandler()

	// Test parsing multiple file flags
	baseFlag, err := handler.ParseFlag("base-config", baseConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse base config flag: %v", err)
	}

	overrideFlag, err := handler.ParseFlag("merge-config", overrideConfigPath)
	if err != nil {
		t.Fatalf("Failed to parse override config flag: %v", err)
	}

	// Test loading configurations
	baseConfigFlag := baseFlag.(*ConfigFileFlag)
	baseConfiguration, err := handler.LoadConfigurationFile(baseConfigFlag)
	if err != nil {
		t.Fatalf("Failed to load base configuration: %v", err)
	}

	overrideConfigFlag := overrideFlag.(*ConfigFileFlag)
	overrideConfiguration, err := handler.LoadConfigurationFile(overrideConfigFlag)
	if err != nil {
		t.Fatalf("Failed to load override configuration: %v", err)
	}

	// Test merging configurations
	merger := NewDefaultConfigurationMerger()
	mergedConfig, err := merger.MergeConfigurations([]Configuration{*baseConfiguration, *overrideConfiguration})
	if err != nil {
		t.Fatalf("Failed to merge configurations: %v", err)
	}

	// Verify merged result
	opencenter, ok := mergedConfig.Data["opencenter"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected opencenter to be a map")
	}

	cluster, ok := opencenter["cluster"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected cluster to be a map")
	}

	// Override should win
	if cluster["name"] != "override-cluster" {
		t.Errorf("Expected cluster name 'override-cluster', got %v", cluster["name"])
	}

	// Check that both sources are tracked
	if len(mergedConfig.Sources) != 2 {
		t.Errorf("Expected 2 sources, got %d", len(mergedConfig.Sources))
	}
}
