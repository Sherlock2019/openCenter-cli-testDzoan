package config

import (
	"encoding/json"
	"fmt"
	"reflect"

	"gopkg.in/yaml.v3"

	"github.com/rackerlabs/openCenter-cli/internal/config/registry"
	"github.com/rackerlabs/openCenter-cli/internal/config/services"
)

// ServiceMap handles polymorphic unmarshalling of service configurations.
// It maps service names to their specific configuration structs.
type ServiceMap map[string]any

// UnmarshalYAML implements the yaml.Unmarshaler interface.
// It merges services from YAML with existing services in the map, preserving
// services that aren't defined in the YAML (for default service population).
func (sm *ServiceMap) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return fmt.Errorf("expected map for Services, got %v", node.Kind)
	}

	// Initialize map if nil, but don't replace existing services
	if *sm == nil {
		*sm = make(ServiceMap)
	}

	// Iterate over keys and values from YAML
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		valNode := node.Content[i+1]

		var serviceName string
		if err := keyNode.Decode(&serviceName); err != nil {
			return err
		}

		// Look up registered type
		configType := registry.GetServiceConfigType(serviceName)
		if configType == nil {
			// If not registered, use generic map or skip?
			// Using DefaultServiceConfig as fallback for unknown services might be safer if we want to preserve them
			// For now, let's use BaseConfig to at least capture enabled status
			configType = reflect.TypeOf(services.BaseConfig{})
		}

		// Create a new instance of the config type
		configPtr := reflect.New(configType).Interface()

		// Unmarshal into the specific struct
		if err := valNode.Decode(configPtr); err != nil {
			return fmt.Errorf("failed to decode config for service %s: %w", serviceName, err)
		}

		// Store in map, overwriting any existing service with the same name
		// This allows YAML to override defaults while preserving services not in YAML
		(*sm)[serviceName] = configPtr
	}

	return nil
}

// MarshalYAML implements the yaml.Marshaler interface.
func (sm ServiceMap) MarshalYAML() (interface{}, error) {
	return (map[string]any)(sm), nil
}

// MarshalJSON implements the json.Marshaler interface.
func (sm ServiceMap) MarshalJSON() ([]byte, error) {
	return json.Marshal((map[string]any)(sm))
}
