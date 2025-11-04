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

package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultCLIConfig(t *testing.T) {
	config := DefaultCLIConfig()

	// Test default values
	if config.Logging.Level != "warn" {
		t.Errorf("Expected default log level 'warn', got '%s'", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("Expected default log format 'text', got '%s'", config.Logging.Format)
	}

	if config.Logging.Output != "stderr" {
		t.Errorf("Expected default log output 'stderr', got '%s'", config.Logging.Output)
	}

	if config.Behavior.AutoConfirm != false {
		t.Errorf("Expected default autoConfirm false, got %v", config.Behavior.AutoConfirm)
	}

	if config.Defaults.Provider != "openstack" {
		t.Errorf("Expected default provider 'openstack', got '%s'", config.Defaults.Provider)
	}
}

func TestExpandPath(t *testing.T) {
	// Test tilde expansion
	home, _ := os.UserHomeDir()
	
	tests := []struct {
		input    string
		expected string
	}{
		{"~/test", filepath.Join(home, "test")},
		{"~", home},
		{"/absolute/path", "/absolute/path"},
		{"relative/path", "relative/path"},
	}

	for _, test := range tests {
		result := ExpandPath(test.input)
		if result != test.expected {
			t.Errorf("ExpandPath(%s) = %s, expected %s", test.input, result, test.expected)
		}
	}
}

func TestConfigValidator(t *testing.T) {
	validator := &ConfigValidator{}
	
	// Test valid configuration
	validConfig := DefaultCLIConfig()
	if err := validator.Validate(validConfig); err != nil {
		t.Errorf("Valid configuration should not produce error: %v", err)
	}

	// Test invalid log level
	invalidConfig := DefaultCLIConfig()
	invalidConfig.Logging.Level = "invalid"
	if err := validator.Validate(invalidConfig); err == nil {
		t.Error("Invalid log level should produce error")
	}

	// Test invalid log format
	invalidConfig = DefaultCLIConfig()
	invalidConfig.Logging.Format = "invalid"
	if err := validator.Validate(invalidConfig); err == nil {
		t.Error("Invalid log format should produce error")
	}
}

func TestConfigManagerSetGetValue(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test setting and getting values
	tests := []struct {
		key   string
		value interface{}
	}{
		{"logging.level", "debug"},
		{"logging.format", "json"},
		{"behavior.verbose", true},
		{"defaults.provider", "aws"},
		{"logging.file.maxSize", 200},
	}

	for _, test := range tests {
		// Set value
		if err := cm.SetValue(test.key, test.value); err != nil {
			t.Errorf("Failed to set %s = %v: %v", test.key, test.value, err)
			continue
		}

		// Get value
		result, err := cm.GetValue(test.key)
		if err != nil {
			t.Errorf("Failed to get %s: %v", test.key, err)
			continue
		}

		if result != test.value {
			t.Errorf("GetValue(%s) = %v, expected %v", test.key, result, test.value)
		}
	}
}

func TestConfigError(t *testing.T) {
	err := &ConfigError{
		Type:    "validation",
		Field:   "logging.level",
		Value:   "invalid",
		Message: "invalid log level",
	}

	expectedError := "validation error in field 'logging.level': invalid log level"
	if err.Error() != expectedError {
		t.Errorf("ConfigError.Error() = %s, expected %s", err.Error(), expectedError)
	}

	suggestions := err.Suggestions()
	if len(suggestions) == 0 {
		t.Error("ConfigError should provide suggestions")
	}
}

func TestConfigManagerMergeWithDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Create partial config
	partialConfig := &CLIConfig{
		Logging: LoggingConfig{
			Level: "debug",
		},
		Behavior: BehaviorConfig{
			Verbose: true,
		},
	}

	// Merge with defaults
	merged := cm.mergeWithDefaults(partialConfig)

	// Check that specified values are preserved
	if merged.Logging.Level != "debug" {
		t.Errorf("Expected merged log level 'debug', got '%s'", merged.Logging.Level)
	}

	if merged.Behavior.Verbose != true {
		t.Errorf("Expected merged verbose true, got %v", merged.Behavior.Verbose)
	}

	// Check that default values are filled in
	if merged.Logging.Format != "text" {
		t.Errorf("Expected default log format 'text', got '%s'", merged.Logging.Format)
	}

	if merged.Defaults.Provider != "openstack" {
		t.Errorf("Expected default provider 'openstack', got '%s'", merged.Defaults.Provider)
	}
}