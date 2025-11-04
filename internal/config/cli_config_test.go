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
	"strings"
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
	validator := &ConfigValidator{autoRepair: false}
	
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

func TestConfigValidatorWithResult(t *testing.T) {
	// Test without auto-repair
	validator := &ConfigValidator{autoRepair: false}
	
	invalidConfig := DefaultCLIConfig()
	invalidConfig.Logging.Level = "invalid"
	invalidConfig.Logging.Format = "invalid"
	invalidConfig.Logging.File.MaxSize = -1
	
	result := validator.ValidateWithResult(invalidConfig)
	
	if result.Valid {
		t.Error("Invalid configuration should not be valid")
	}
	
	if len(result.Errors) < 3 {
		t.Errorf("Expected at least 3 errors, got %d", len(result.Errors))
	}
	
	if len(result.Repaired) != 0 {
		t.Errorf("Expected 0 repairs without auto-repair, got %d", len(result.Repaired))
	}
}

func TestConfigValidatorAutoRepair(t *testing.T) {
	// Test with auto-repair
	validator := &ConfigValidator{autoRepair: true}
	
	invalidConfig := DefaultCLIConfig()
	invalidConfig.Logging.Level = "invalid"
	invalidConfig.Logging.Format = "invalid"
	invalidConfig.Logging.File.MaxSize = -1
	
	result := validator.ValidateWithResult(invalidConfig)
	
	if !result.Valid {
		t.Error("Configuration should be valid after auto-repair")
	}
	
	if len(result.Errors) != 0 {
		t.Errorf("Expected 0 errors after auto-repair, got %d", len(result.Errors))
	}
	
	if len(result.Repaired) < 3 {
		t.Errorf("Expected at least 3 repairs, got %d", len(result.Repaired))
	}
	
	// Verify that values were actually repaired
	defaults := DefaultCLIConfig()
	if invalidConfig.Logging.Level != defaults.Logging.Level {
		t.Errorf("Log level was not repaired to default")
	}
	
	if invalidConfig.Logging.Format != defaults.Logging.Format {
		t.Errorf("Log format was not repaired to default")
	}
	
	if invalidConfig.Logging.File.MaxSize != defaults.Logging.File.MaxSize {
		t.Errorf("MaxSize was not repaired to default")
	}
}

func TestConfigValidatorPathValidation(t *testing.T) {
	validator := &ConfigValidator{autoRepair: false}
	
	// Test empty paths
	invalidConfig := DefaultCLIConfig()
	invalidConfig.Paths.ConfigDir = ""
	invalidConfig.Paths.ClustersDir = ""
	
	result := validator.ValidateWithResult(invalidConfig)
	
	if result.Valid {
		t.Error("Configuration with empty paths should not be valid")
	}
	
	if len(result.Errors) < 2 {
		t.Errorf("Expected at least 2 path errors, got %d", len(result.Errors))
	}
}

func TestConfigValidatorWarnings(t *testing.T) {
	validator := &ConfigValidator{autoRepair: false}
	
	// Test configuration that should generate warnings
	config := DefaultCLIConfig()
	config.Behavior.AutoConfirm = true
	config.Behavior.DryRun = false
	config.Defaults.Provider = "unknown-provider"
	config.Defaults.Environment = "unknown-env"
	
	result := validator.ValidateWithResult(config)
	
	if len(result.Warnings) == 0 {
		t.Error("Expected warnings for potentially problematic configuration")
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

	// Test repaired error
	repairedErr := &ConfigError{
		Type:     "validation",
		Field:    "logging.level",
		Value:    "warn",
		Message:  "repaired to default",
		Repaired: true,
	}

	expectedRepairedError := "[AUTO-REPAIRED] validation error in field 'logging.level': repaired to default"
	if repairedErr.Error() != expectedRepairedError {
		t.Errorf("Repaired ConfigError.Error() = %s, expected %s", repairedErr.Error(), expectedRepairedError)
	}

	repairedSuggestions := repairedErr.Suggestions()
	if len(repairedSuggestions) == 0 || repairedSuggestions[0] != "Configuration was automatically repaired with default values" {
		t.Error("Repaired ConfigError should provide repair confirmation")
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

func TestConfigManagerValidation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Test validation of valid config
	result := cm.ValidateConfig()
	if !result.Valid {
		t.Error("Default configuration should be valid")
	}

	// Test validation summary
	summary := cm.GetValidationSummary()
	if !strings.Contains(summary, "✓ Configuration is valid") {
		t.Error("Validation summary should indicate valid configuration")
	}
}

func TestConfigManagerRepair(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("Failed to create config manager: %v", err)
	}

	// Introduce invalid values
	cm.config.Logging.Level = "invalid"
	cm.config.Logging.Format = "invalid"
	cm.config.Logging.File.MaxSize = -1

	// Test repair
	result, err := cm.RepairConfig()
	if err != nil {
		t.Errorf("RepairConfig should not return error: %v", err)
	}

	if !result.Valid {
		t.Error("Configuration should be valid after repair")
	}

	if len(result.Repaired) == 0 {
		t.Error("Expected repairs to be made")
	}

	// Verify values were actually repaired
	defaults := DefaultCLIConfig()
	if cm.config.Logging.Level != defaults.Logging.Level {
		t.Error("Log level should be repaired to default")
	}
}

func TestConfigManagerGracefulDegradation(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	// Create config file with invalid content
	invalidYAML := `
logging:
  level: invalid_level
  format: invalid_format
  file:
    maxSize: -1
paths:
  configDir: ""
  clustersDir: ""
`
	
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0600); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	// ConfigManager should handle this gracefully with auto-repair
	cm, err := NewConfigManager(configPath)
	if err != nil {
		t.Fatalf("ConfigManager should handle invalid config gracefully: %v", err)
	}

	// Configuration should be valid after auto-repair
	result := cm.ValidateConfig()
	if !result.Valid {
		t.Error("Configuration should be valid after auto-repair during load")
	}

	// Check that default values were applied
	defaults := DefaultCLIConfig()
	if cm.config.Logging.Level != defaults.Logging.Level {
		t.Error("Invalid log level should be repaired to default")
	}
}