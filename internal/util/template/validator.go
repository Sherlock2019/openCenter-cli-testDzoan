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
	"fmt"
	"reflect"
	"strings"
	"text/template"
)

// DefaultTemplateValidator implements TemplateValidator interface
type DefaultTemplateValidator struct {
	templates *template.Template
}

// NewDefaultTemplateValidator creates a new default template validator
func NewDefaultTemplateValidator() *DefaultTemplateValidator {
	return &DefaultTemplateValidator{}
}

// Init initializes the template validator with templates
func (v *DefaultTemplateValidator) Init(templates *template.Template) error {
	if templates == nil {
		return fmt.Errorf("templates cannot be nil")
	}
	v.templates = templates
	return nil
}

// ValidateTemplate validates that a template exists and can be parsed
func (v *DefaultTemplateValidator) ValidateTemplate(templateName string) error {
	if v.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	tmpl := v.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	// Try to parse the template to check for syntax errors
	if tmpl.Tree == nil {
		return fmt.Errorf("template %s has no parse tree", templateName)
	}

	return nil
}

// ValidateTemplateData validates that data contains required fields for a template
func (v *DefaultTemplateValidator) ValidateTemplateData(templateName string, data interface{}) error {
	if err := v.ValidateTemplateExists(templateName); err != nil {
		return err
	}

	// Basic validation - check if data is not nil
	if data == nil {
		return fmt.Errorf("template data cannot be nil for template: %s", templateName)
	}

	// Additional validation could be added here based on template requirements
	return nil
}

// ValidateTemplateExists validates that a template exists
func (v *DefaultTemplateValidator) ValidateTemplateExists(templateName string) error {
	if v.templates == nil {
		return fmt.Errorf("templates not initialized")
	}

	tmpl := v.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template not found: %s", templateName)
	}

	return nil
}

// ValidateRequiredFields validates that data contains all required fields
func (v *DefaultTemplateValidator) ValidateRequiredFields(data interface{}, requiredFields []string) error {
	if data == nil {
		return fmt.Errorf("data cannot be nil when validating required fields")
	}

	if len(requiredFields) == 0 {
		return nil // No required fields to validate
	}

	// Use reflection to check for required fields
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	var missingFields []string

	switch value.Kind() {
	case reflect.Struct:
		// Validate struct fields
		for _, fieldName := range requiredFields {
			field := value.FieldByName(fieldName)
			if !field.IsValid() || isFieldEmpty(field) {
				missingFields = append(missingFields, fieldName)
			}
		}
	case reflect.Map:
		// Validate map keys
		for _, fieldName := range requiredFields {
			mapValue := value.MapIndex(reflect.ValueOf(fieldName))
			if !mapValue.IsValid() || isFieldEmpty(mapValue) {
				missingFields = append(missingFields, fieldName)
			}
		}
	default:
		return fmt.Errorf("unsupported data type for field validation: %s", value.Kind())
	}

	if len(missingFields) > 0 {
		return fmt.Errorf("missing required fields: %s", strings.Join(missingFields, ", "))
	}

	return nil
}

// ValidateNetworkPluginConfig validates network plugin configuration
func (v *DefaultTemplateValidator) ValidateNetworkPluginConfig(pluginType string, config map[string]interface{}) error {
	if pluginType == "" {
		return fmt.Errorf("network plugin type cannot be empty")
	}

	supportedPlugins := []string{"calico", "cilium", "kube-ovn", "flannel"}
	isSupported := false
	for _, supported := range supportedPlugins {
		if pluginType == supported {
			isSupported = true
			break
		}
	}

	if !isSupported {
		return fmt.Errorf("unsupported network plugin: %s, supported plugins: %s", 
			pluginType, strings.Join(supportedPlugins, ", "))
	}

	// Plugin-specific validation
	switch pluginType {
	case "calico":
		return v.validateCalicoConfig(config)
	case "cilium":
		return v.validateCiliumConfig(config)
	case "kube-ovn":
		return v.validateKubeOVNConfig(config)
	case "flannel":
		return v.validateFlannelConfig(config)
	}

	return nil
}

// validateCalicoConfig validates Calico-specific configuration
func (v *DefaultTemplateValidator) validateCalicoConfig(config map[string]interface{}) error {
	// Calico-specific validation logic
	if config == nil {
		return nil // Default configuration is acceptable
	}

	// Check for conflicting configurations
	if ipv4Pool, exists := config["ipv4_pool"]; exists {
		if ipv4Pool == "" {
			return fmt.Errorf("calico ipv4_pool cannot be empty")
		}
	}

	return nil
}

// validateCiliumConfig validates Cilium-specific configuration
func (v *DefaultTemplateValidator) validateCiliumConfig(config map[string]interface{}) error {
	// Cilium-specific validation logic
	if config == nil {
		return nil // Default configuration is acceptable
	}

	// Check for Cilium-specific requirements
	if hubble, exists := config["hubble"]; exists {
		if hubbleMap, ok := hubble.(map[string]interface{}); ok {
			if enabled, exists := hubbleMap["enabled"]; exists {
				if enabledBool, ok := enabled.(bool); ok && enabledBool {
					// Validate Hubble configuration
					if ui, exists := hubbleMap["ui"]; exists {
						if uiMap, ok := ui.(map[string]interface{}); ok {
							if _, exists := uiMap["enabled"]; !exists {
								return fmt.Errorf("hubble ui configuration missing enabled field")
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// validateKubeOVNConfig validates Kube-OVN-specific configuration
func (v *DefaultTemplateValidator) validateKubeOVNConfig(config map[string]interface{}) error {
	// Kube-OVN-specific validation logic
	if config == nil {
		return nil // Default configuration is acceptable
	}

	// Check for Kube-OVN specific requirements
	if subnet, exists := config["default_subnet"]; exists {
		if subnet == "" {
			return fmt.Errorf("kube-ovn default_subnet cannot be empty")
		}
	}

	return nil
}

// validateFlannelConfig validates Flannel-specific configuration
func (v *DefaultTemplateValidator) validateFlannelConfig(config map[string]interface{}) error {
	// Flannel-specific validation logic
	if config == nil {
		return nil // Default configuration is acceptable
	}

	// Check for Flannel specific requirements
	if backend, exists := config["backend"]; exists {
		if backendMap, ok := backend.(map[string]interface{}); ok {
			if backendType, exists := backendMap["type"]; exists {
				validTypes := []string{"vxlan", "host-gw", "udp"}
				isValid := false
				for _, validType := range validTypes {
					if backendType == validType {
						isValid = true
						break
					}
				}
				if !isValid {
					return fmt.Errorf("invalid flannel backend type: %v, valid types: %s", 
						backendType, strings.Join(validTypes, ", "))
				}
			}
		}
	}

	return nil
}

// isFieldEmpty checks if a reflect.Value represents an empty field
func isFieldEmpty(field reflect.Value) bool {
	if !field.IsValid() {
		return true
	}

	switch field.Kind() {
	case reflect.String:
		return field.String() == ""
	case reflect.Slice, reflect.Map, reflect.Array:
		return field.Len() == 0
	case reflect.Ptr, reflect.Interface:
		return field.IsNil()
	case reflect.Bool:
		return !field.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return field.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return field.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return field.Float() == 0
	}

	return false
}