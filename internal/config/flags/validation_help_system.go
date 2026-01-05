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
	"fmt"
	"strings"
)

// ValidationMode defines the validation mode
type ValidationMode string

const (
	ValidationModeValidateOnly ValidationMode = "validate-only" // Only validate, don't apply
	ValidationModeNormal       ValidationMode = "normal"       // Normal validation and apply
)

// ValidationResult represents the result of configuration validation
type ValidationResult struct {
	Valid   bool                      `json:"valid"`
	Errors  []ConfigValidationError   `json:"errors,omitempty"`
	Warnings []ConfigValidationWarning `json:"warnings,omitempty"`
	Summary ValidationSummary         `json:"summary"`
}

// ConfigValidationError represents a validation error
type ConfigValidationError struct {
	Type        string `json:"type"`
	Path        string `json:"path,omitempty"`
	Message     string `json:"message"`
	Suggestion  string `json:"suggestion,omitempty"`
	Example     string `json:"example,omitempty"`
}

// ConfigValidationWarning represents a validation warning
type ConfigValidationWarning struct {
	Type    string `json:"type"`
	Path    string `json:"path,omitempty"`
	Message string `json:"message"`
}

// ValidationSummary provides a summary of validation results
type ValidationSummary struct {
	TotalErrors   int `json:"total_errors"`
	TotalWarnings int `json:"total_warnings"`
	ConfigPaths   int `json:"config_paths"`
	FlagsProcessed int `json:"flags_processed"`
}

// ConfigurationValidator validates configuration and flags
type ConfigurationValidator interface {
	// ValidateConfiguration validates a complete configuration
	ValidateConfiguration(config *Configuration) (*ValidationResult, error)
	
	// ValidateFlags validates parsed flags before applying them
	ValidateFlags(flags *ParsedFlags) (*ValidationResult, error)
	
	// ValidateFlag validates a single flag
	ValidateFlag(flagName, value string) error
}

// DefaultConfigurationValidator implements the ConfigurationValidator interface
type DefaultConfigurationValidator struct{}

// NewDefaultConfigurationValidator creates a new configuration validator
func NewDefaultConfigurationValidator() *DefaultConfigurationValidator {
	return &DefaultConfigurationValidator{}
}

// ValidateConfiguration validates a complete configuration
func (v *DefaultConfigurationValidator) ValidateConfiguration(config *Configuration) (*ValidationResult, error) {
	if config == nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ConfigValidationError{
				{
					Type:    "null_config",
					Message: "Configuration cannot be null",
				},
			},
			Summary: ValidationSummary{TotalErrors: 1},
		}, nil
	}
	
	var errors []ConfigValidationError
	var warnings []ConfigValidationWarning
	
	// Validate configuration data structure
	configPaths := v.validateConfigurationData(config.Data, "", &errors, &warnings)
	
	// Validate sources
	v.validateSources(config.Sources, &errors, &warnings)
	
	// Validate metadata
	v.validateMetadata(config.Metadata, &errors, &warnings)
	
	return &ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Summary: ValidationSummary{
			TotalErrors:   len(errors),
			TotalWarnings: len(warnings),
			ConfigPaths:   configPaths,
		},
	}, nil
}

// ValidateFlags validates parsed flags before applying them
func (v *DefaultConfigurationValidator) ValidateFlags(flags *ParsedFlags) (*ValidationResult, error) {
	if flags == nil {
		return &ValidationResult{
			Valid: false,
			Errors: []ConfigValidationError{
				{
					Type:    "null_flags",
					Message: "Parsed flags cannot be null",
				},
			},
			Summary: ValidationSummary{TotalErrors: 1},
		}, nil
	}
	
	var errors []ConfigValidationError
	var warnings []ConfigValidationWarning
	flagsProcessed := 0
	
	// Validate dot notation flags
	for path, value := range flags.DotNotation {
		if err := v.validateDotNotationFlag(path, value); err != nil {
			errors = append(errors, ConfigValidationError{
				Type:    "dot_notation",
				Path:    path,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	// Validate array flags
	for _, arrayFlag := range flags.ArrayFlags {
		if err := v.validateArrayFlag(arrayFlag); err != nil {
			path := ""
			if arrayFlag.Config != nil {
				path = arrayFlag.Config.Path
			}
			errors = append(errors, ConfigValidationError{
				Type:    "array_flag",
				Path:    path,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	// Validate JSON flags
	for _, jsonFlag := range flags.JSONFlags {
		if err := v.validateJSONFlag(jsonFlag); err != nil {
			errors = append(errors, ConfigValidationError{
				Type:    "json_flag",
				Path:    jsonFlag.Path,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	// Validate YAML flags
	for _, yamlFlag := range flags.YAMLFlags {
		if err := v.validateYAMLFlag(yamlFlag); err != nil {
			errors = append(errors, ConfigValidationError{
				Type:    "yaml_flag",
				Path:    yamlFlag.Path,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	// Validate configuration file flags
	for _, configFileFlag := range flags.ConfigFileFlags {
		if err := v.validateConfigFileFlag(configFileFlag); err != nil {
			errors = append(errors, ConfigValidationError{
				Type:    "config_file",
				Path:    configFileFlag.Path,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	// Validate template variables
	for name, value := range flags.TemplateVars {
		if err := v.validateTemplateVar(name, value); err != nil {
			errors = append(errors, ConfigValidationError{
				Type:    "template_var",
				Path:    name,
				Message: err.Error(),
			})
		}
		flagsProcessed++
	}
	
	return &ValidationResult{
		Valid:    len(errors) == 0,
		Errors:   errors,
		Warnings: warnings,
		Summary: ValidationSummary{
			TotalErrors:    len(errors),
			TotalWarnings:  len(warnings),
			FlagsProcessed: flagsProcessed,
		},
	}, nil
}

// ValidateFlag validates a single flag
func (v *DefaultConfigurationValidator) ValidateFlag(flagName, value string) error {
	if flagName == "" {
		return fmt.Errorf("flag name cannot be empty")
	}
	
	if value == "" {
		return fmt.Errorf("flag value cannot be empty for flag '%s'", flagName)
	}
	
	// Basic validation - could be extended with more specific rules
	return nil
}

// validateConfigurationData validates configuration data recursively
func (v *DefaultConfigurationValidator) validateConfigurationData(data map[string]interface{}, prefix string, errors *[]ConfigValidationError, warnings *[]ConfigValidationWarning) int {
	if data == nil {
		return 0
	}
	
	pathCount := 0
	
	for key, value := range data {
		currentPath := key
		if prefix != "" {
			currentPath = prefix + "." + key
		}
		pathCount++
		
		// Validate key format
		if strings.Contains(key, " ") {
			*warnings = append(*warnings, ConfigValidationWarning{
				Type:    "key_format",
				Path:    currentPath,
				Message: "Configuration key contains spaces, which may cause issues",
			})
		}
		
		// Recursively validate nested objects
		if nestedMap, ok := value.(map[string]interface{}); ok {
			nestedPaths := v.validateConfigurationData(nestedMap, currentPath, errors, warnings)
			pathCount += nestedPaths
		}
		
		// Validate arrays
		if array, ok := value.([]interface{}); ok {
			for i, item := range array {
				if nestedMap, ok := item.(map[string]interface{}); ok {
					arrayPath := fmt.Sprintf("%s[%d]", currentPath, i)
					nestedPaths := v.validateConfigurationData(nestedMap, arrayPath, errors, warnings)
					pathCount += nestedPaths
				}
			}
		}
	}
	
	return pathCount
}

// validateSources validates configuration sources
func (v *DefaultConfigurationValidator) validateSources(sources []ConfigSource, errors *[]ConfigValidationError, warnings *[]ConfigValidationWarning) {
	if len(sources) == 0 {
		*warnings = append(*warnings, ConfigValidationWarning{
			Type:    "no_sources",
			Message: "No configuration sources specified",
		})
		return
	}
	
	for _, source := range sources {
		if source.Type == "" {
			*errors = append(*errors, ConfigValidationError{
				Type:    "source_type",
				Path:    source.Path,
				Message: "Source type cannot be empty",
			})
		}
		
		if source.Path == "" {
			*errors = append(*errors, ConfigValidationError{
				Type:    "source_path",
				Message: "Source path cannot be empty",
			})
		}
	}
}

// validateMetadata validates configuration metadata
func (v *DefaultConfigurationValidator) validateMetadata(metadata ConfigMetadata, errors *[]ConfigValidationError, warnings *[]ConfigValidationWarning) {
	if metadata.ProcessedAt.IsZero() {
		*warnings = append(*warnings, ConfigValidationWarning{
			Type:    "metadata",
			Message: "Configuration processing time not set",
		})
	}
}

// validateDotNotationFlag validates a dot notation flag
func (v *DefaultConfigurationValidator) validateDotNotationFlag(path, value string) error {
	if path == "" {
		return fmt.Errorf("dot notation path cannot be empty")
	}
	
	if strings.HasPrefix(path, ".") || strings.HasSuffix(path, ".") {
		return fmt.Errorf("dot notation path cannot start or end with a dot")
	}
	
	if strings.Contains(path, "..") {
		return fmt.Errorf("dot notation path cannot contain consecutive dots")
	}
	
	return nil
}

// validateArrayFlag validates an array flag
func (v *DefaultConfigurationValidator) validateArrayFlag(arrayFlag ArrayFlag) error {
	if arrayFlag.Config == nil {
		return fmt.Errorf("array flag config cannot be nil")
	}
	
	if arrayFlag.Config.Path == "" {
		return fmt.Errorf("array flag path cannot be empty")
	}
	
	if arrayFlag.Config.Fields == nil {
		return fmt.Errorf("array flag fields cannot be nil")
	}
	
	return nil
}

// validateJSONFlag validates a JSON flag
func (v *DefaultConfigurationValidator) validateJSONFlag(jsonFlag JSONFlag) error {
	if jsonFlag.Path == "" {
		return fmt.Errorf("JSON flag path cannot be empty")
	}
	
	if jsonFlag.Value == nil {
		return fmt.Errorf("JSON flag value cannot be nil")
	}
	
	return nil
}

// validateYAMLFlag validates a YAML flag
func (v *DefaultConfigurationValidator) validateYAMLFlag(yamlFlag YAMLFlag) error {
	if yamlFlag.Path == "" {
		return fmt.Errorf("YAML flag path cannot be empty")
	}
	
	if yamlFlag.Value == nil {
		return fmt.Errorf("YAML flag value cannot be nil")
	}
	
	return nil
}

// validateConfigFileFlag validates a configuration file flag
func (v *DefaultConfigurationValidator) validateConfigFileFlag(configFileFlag *ConfigFileFlag) error {
	if configFileFlag == nil {
		return fmt.Errorf("config file flag cannot be nil")
	}
	
	if configFileFlag.Path == "" {
		return fmt.Errorf("config file path cannot be empty")
	}
	
	if configFileFlag.Type == "" {
		return fmt.Errorf("config file type cannot be empty")
	}
	
	return nil
}

// validateTemplateVar validates a template variable
func (v *DefaultConfigurationValidator) validateTemplateVar(name, value string) error {
	if name == "" {
		return fmt.Errorf("template variable name cannot be empty")
	}
	
	if strings.Contains(name, " ") {
		return fmt.Errorf("template variable name cannot contain spaces")
	}
	
	return nil
}

// HelpSystem provides help and examples for CLI flags
type HelpSystem interface {
	// GetFlagHelp returns help text for a specific flag type
	GetFlagHelp(flagType FlagType) (string, error)
	
	// GetAllFlagHelp returns help text for all flag types
	GetAllFlagHelp() (string, error)
	
	// GetExamples returns common usage examples
	GetExamples() (string, error)
	
	// GetFlagExamples returns examples for a specific flag type
	GetFlagExamples(flagType FlagType) ([]string, error)
}

// DefaultHelpSystem implements the HelpSystem interface
type DefaultHelpSystem struct{}

// NewDefaultHelpSystem creates a new help system
func NewDefaultHelpSystem() *DefaultHelpSystem {
	return &DefaultHelpSystem{}
}

// GetFlagHelp returns help text for a specific flag type
func (h *DefaultHelpSystem) GetFlagHelp(flagType FlagType) (string, error) {
	switch flagType {
	case FlagTypeDotNotation:
		return h.getDotNotationHelp(), nil
	case FlagTypeArray:
		return h.getArrayHelp(), nil
	case FlagTypeJSON:
		return h.getJSONHelp(), nil
	case FlagTypeYAML:
		return h.getYAMLHelp(), nil
	case FlagTypeTemplate:
		return h.getTemplateHelp(), nil
	case FlagTypeFile:
		return h.getFileHelp(), nil
	case FlagTypeOutput:
		return h.getOutputHelp(), nil
	default:
		return "", fmt.Errorf("unsupported flag type: %s", flagType)
	}
}

// GetAllFlagHelp returns help text for all flag types
func (h *DefaultHelpSystem) GetAllFlagHelp() (string, error) {
	var help strings.Builder
	
	help.WriteString("Enhanced CLI Configuration Flags\n")
	help.WriteString("=================================\n\n")
	
	flagTypes := []FlagType{
		FlagTypeDotNotation,
		FlagTypeArray,
		FlagTypeJSON,
		FlagTypeYAML,
		FlagTypeTemplate,
		FlagTypeFile,
		FlagTypeOutput,
	}
	
	for _, flagType := range flagTypes {
		flagHelp, err := h.GetFlagHelp(flagType)
		if err != nil {
			continue
		}
		help.WriteString(flagHelp)
		help.WriteString("\n")
	}
	
	return help.String(), nil
}

// GetExamples returns common usage examples
func (h *DefaultHelpSystem) GetExamples() (string, error) {
	examples := `
Common Configuration Examples
=============================

1. Basic Configuration:
   --cluster.name=my-cluster --infrastructure.provider=openstack

2. Array Configuration:
   --server-pool name=compute,worker_count=3,flavor=large
   --dns-server 8.8.8.8 --dns-server 8.8.4.4

3. JSON Configuration:
   --json-set 'cluster={"name": "my-cluster", "version": "1.0"}'

4. YAML Configuration:
   --yaml-set cluster --yaml-data 'name: my-cluster\nversion: 1.0'

5. File Merging:
   --base-config base.yaml --merge-config override.yaml

6. Configuration Stack:
   --config-stack base.yaml,env.yaml,local.yaml

7. Template Variables:
   --template-var CLUSTER_NAME=production --template-var REGION=us-east-1

8. Output Formatting:
   --output-format json --dry-run
   --output-format yaml --quiet

9. Complex Example:
   --base-config base.yaml \
   --server-pool name=compute,worker_count=5 \
   --json-set 'networking={"dns_servers": ["8.8.8.8"]}' \
   --template-var ENV=production \
   --output-format yaml --dry-run
`
	
	return examples, nil
}

// GetFlagExamples returns examples for a specific flag type
func (h *DefaultHelpSystem) GetFlagExamples(flagType FlagType) ([]string, error) {
	switch flagType {
	case FlagTypeDotNotation:
		return []string{
			"--cluster.name=my-cluster",
			"--infrastructure.provider=openstack",
			"--networking.dns_servers.0=8.8.8.8",
		}, nil
	case FlagTypeArray:
		return []string{
			"--server-pool name=compute,worker_count=3,flavor=large",
			"--ssh-key ~/.ssh/id_rsa.pub",
			"--dns-server 8.8.8.8",
			"--subnet name=private,cidr=10.0.1.0/24",
		}, nil
	case FlagTypeJSON:
		return []string{
			`--json-set 'cluster={"name": "my-cluster"}'`,
			`--json-set 'networking={"dns_servers": ["8.8.8.8", "8.8.4.4"]}'`,
		}, nil
	case FlagTypeYAML:
		return []string{
			"--yaml-set cluster --yaml-data 'name: my-cluster'",
			"--yaml-file cluster=config.yaml",
		}, nil
	case FlagTypeTemplate:
		return []string{
			"--template-var CLUSTER_NAME=production",
			"--template-var REGION=us-east-1",
		}, nil
	case FlagTypeFile:
		return []string{
			"--base-config base.yaml",
			"--merge-config override.yaml",
			"--config-stack base.yaml,env.yaml,local.yaml",
		}, nil
	case FlagTypeOutput:
		return []string{
			"--output-format json",
			"--output-format yaml",
			"--dry-run",
			"--quiet",
		}, nil
	default:
		return nil, fmt.Errorf("unsupported flag type: %s", flagType)
	}
}

// Helper methods for specific flag type help

func (h *DefaultHelpSystem) getDotNotationHelp() string {
	return `Dot Notation Flags:
  Use dot notation to set nested configuration values.
  Syntax: --path.to.field=value
  
  Examples:
    --cluster.name=my-cluster
    --infrastructure.provider=openstack
    --networking.dns_servers.0=8.8.8.8`
}

func (h *DefaultHelpSystem) getArrayHelp() string {
	return `Array Flags:
  Dedicated handlers for common array types.
  
  Server Pools: --server-pool name=X,worker_count=N,flavor=Y
  SSH Keys: --ssh-key /path/to/key
  DNS Servers: --dns-server IP_ADDRESS
  Subnets: --subnet name=X,cidr=Y`
}

func (h *DefaultHelpSystem) getJSONHelp() string {
	return `JSON Flags:
  Set configuration using JSON syntax.
  Syntax: --json-set 'path={"key": "value"}'
  
  Examples:
    --json-set 'cluster={"name": "my-cluster"}'
    --json-set 'networking={"dns_servers": ["8.8.8.8"]}'`
}

func (h *DefaultHelpSystem) getYAMLHelp() string {
	return `YAML Flags:
  Set configuration using YAML syntax.
  
  Inline: --yaml-set path --yaml-data 'key: value'
  File: --yaml-file path=filename.yaml`
}

func (h *DefaultHelpSystem) getTemplateHelp() string {
	return `Template Variables:
  Define variables for template substitution.
  Syntax: --template-var NAME=value
  
  Use in config: {{.NAME}}
  
  Examples:
    --template-var CLUSTER_NAME=production
    --template-var REGION=us-east-1`
}

func (h *DefaultHelpSystem) getFileHelp() string {
	return `Configuration File Flags:
  Merge multiple configuration files.
  
  Base config: --base-config base.yaml
  Override config: --merge-config override.yaml
  Config stack: --config-stack file1.yaml,file2.yaml,file3.yaml`
}

func (h *DefaultHelpSystem) getOutputHelp() string {
	return `Output Format Flags:
  Control output format and mode.
  
  Format: --output-format [json|yaml|diff]
  Modes: --dry-run, --quiet
  
  Examples:
    --output-format json --dry-run
    --output-format yaml --quiet`
}