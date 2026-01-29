package registry

import (
	"reflect"
	"sync"
)

var (
	// serviceRegistry maps service names to their configuration types
	serviceRegistry = make(map[string]reflect.Type)
	registryLock    sync.RWMutex
)

// RegisterServiceConfig registers a service configuration struct for a given service name.
// The config argument should be an instance of the struct (not a pointer).
func RegisterServiceConfig(name string, config interface{}) {
	registryLock.Lock()
	defer registryLock.Unlock()
	serviceRegistry[name] = reflect.TypeOf(config)
}

// GetServiceConfigType returns the reflection type of the configuration struct for the given service.
// It returns nil if the service is not registered.
func GetServiceConfigType(name string) reflect.Type {
	registryLock.RLock()
	defer registryLock.RUnlock()
	return serviceRegistry[name]
}

// GetRegisteredServices returns a list of all registered service names.
func GetRegisteredServices() []string {
	registryLock.RLock()
	defer registryLock.RUnlock()
	keys := make([]string, 0, len(serviceRegistry))
	for k := range serviceRegistry {
		keys = append(keys, k)
	}
	return keys
}

// IsRegistered checks if a service is registered in the service registry.
func IsRegistered(name string) bool {
	registryLock.RLock()
	defer registryLock.RUnlock()
	_, exists := serviceRegistry[name]
	return exists
}
