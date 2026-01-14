package template

import (
	"fmt"
	"sort"
	"sync"
)

// TemplateRegistry manages template definitions with metadata and dependency resolution
type TemplateRegistry interface {
	RegisterTemplate(template TemplateDefinition) error
	GetTemplate(name string) (TemplateDefinition, error)
	GetTemplatesForProvider(provider string) []TemplateDefinition
	GetTemplatesForService(service string) []TemplateDefinition
	GetTemplatesForEnabledServices(enabledServices []string) []TemplateDefinition
	ResolveTemplateDependencies(templates []string) ([]TemplateDefinition, error)
	ListTemplates() []TemplateDefinition
	UnregisterTemplate(name string) error
}

// TemplateType represents the category of a template
type TemplateType string

const (
	TemplateTypeInfrastructure TemplateType = "infrastructure"
	TemplateTypeService        TemplateType = "service"
	TemplateTypeBase           TemplateType = "base"
	TemplateTypeOverlay        TemplateType = "overlay"
)

// ConditionType represents the type of rendering condition
type ConditionType string

const (
	ConditionTypeEquals       ConditionType = "equals"
	ConditionTypeNotEquals    ConditionType = "not_equals"
	ConditionTypeContains     ConditionType = "contains"
	ConditionTypeExists       ConditionType = "exists"
	ConditionTypeGreaterThan  ConditionType = "greater_than"
	ConditionTypeLessThan     ConditionType = "less_than"
)

// RenderCondition defines a condition for template rendering
type RenderCondition struct {
	Type     ConditionType
	Field    string
	Operator string
	Value    interface{}
}

// TemplateDefinition represents a template with its metadata
type TemplateDefinition struct {
	Name         string
	Path         string
	Type         TemplateType
	Provider     string
	Services     []string
	Dependencies []string
	Conditions   []RenderCondition
	Metadata     TemplateMetadata
}

// TemplateMetadata holds additional information about a template
type TemplateMetadata struct {
	Description string
	Version     string
	Author      string
	Tags        []string
	Priority    int
}

// InMemoryTemplateRegistry is an in-memory implementation of TemplateRegistry
type InMemoryTemplateRegistry struct {
	templates map[string]TemplateDefinition
	mu        sync.RWMutex
}

// NewInMemoryTemplateRegistry creates a new in-memory template registry
func NewInMemoryTemplateRegistry() *InMemoryTemplateRegistry {
	return &InMemoryTemplateRegistry{
		templates: make(map[string]TemplateDefinition),
	}
}

// RegisterTemplate registers a template in the registry
// This implements Property 7: Template Dependency Validation
// **Validates: Requirements 2.2**
func (r *InMemoryTemplateRegistry) RegisterTemplate(template TemplateDefinition) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate basic template fields
	if template.Name == "" {
		return fmt.Errorf("template name cannot be empty")
	}

	if template.Path == "" {
		return fmt.Errorf("template path cannot be empty for template %s", template.Name)
	}

	// Validate dependencies
	if err := r.validateDependencies(template); err != nil {
		return fmt.Errorf("invalid dependencies for template %s: %w", template.Name, err)
	}

	// Validate render conditions
	if err := r.validateConditions(template); err != nil {
		return fmt.Errorf("invalid conditions for template %s: %w", template.Name, err)
	}

	r.templates[template.Name] = template
	return nil
}

// validateDependencies checks that template dependencies are valid
func (r *InMemoryTemplateRegistry) validateDependencies(template TemplateDefinition) error {
	// Check for self-dependency
	for _, dep := range template.Dependencies {
		if dep == template.Name {
			return fmt.Errorf("template cannot depend on itself")
		}
	}

	// Check for empty dependency names
	for _, dep := range template.Dependencies {
		if dep == "" {
			return fmt.Errorf("dependency name cannot be empty")
		}
	}

	// Check for duplicate dependencies
	seen := make(map[string]bool)
	for _, dep := range template.Dependencies {
		if seen[dep] {
			return fmt.Errorf("duplicate dependency: %s", dep)
		}
		seen[dep] = true
	}

	// Note: We don't check if dependencies exist yet because they might be registered later
	// Circular dependency detection happens during resolution, not registration
	return nil
}

// validateConditions checks that render conditions are properly formed
func (r *InMemoryTemplateRegistry) validateConditions(template TemplateDefinition) error {
	for i, condition := range template.Conditions {
		// Validate condition type
		if !isValidConditionType(condition.Type) {
			return fmt.Errorf("condition %d has invalid type: %s", i, condition.Type)
		}

		// Validate field name is not empty for conditions that require it
		if requiresField(condition.Type) && condition.Field == "" {
			return fmt.Errorf("condition %d of type %s requires a field name", i, condition.Type)
		}

		// Validate value is provided for conditions that require it
		if requiresValue(condition.Type) && condition.Value == nil {
			return fmt.Errorf("condition %d of type %s requires a value", i, condition.Type)
		}
	}

	return nil
}

// isValidConditionType checks if a condition type is valid
func isValidConditionType(condType ConditionType) bool {
	switch condType {
	case ConditionTypeEquals,
		ConditionTypeNotEquals,
		ConditionTypeContains,
		ConditionTypeExists,
		ConditionTypeGreaterThan,
		ConditionTypeLessThan:
		return true
	default:
		return false
	}
}

// requiresField returns true if the condition type requires a field name
func requiresField(condType ConditionType) bool {
	// All condition types require a field
	return true
}

// requiresValue returns true if the condition type requires a value
func requiresValue(condType ConditionType) bool {
	// Exists condition doesn't require a value, all others do
	return condType != ConditionTypeExists
}

// GetTemplate retrieves a template by name
func (r *InMemoryTemplateRegistry) GetTemplate(name string) (TemplateDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	template, exists := r.templates[name]
	if !exists {
		return TemplateDefinition{}, fmt.Errorf("template %s not found", name)
	}

	return template, nil
}

// GetTemplatesForProvider returns all templates compatible with a provider
func (r *InMemoryTemplateRegistry) GetTemplatesForProvider(provider string) []TemplateDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []TemplateDefinition
	for _, template := range r.templates {
		// Include templates with no provider specified (universal) or matching provider
		if template.Provider == "" || template.Provider == provider {
			result = append(result, template)
		}
	}

	// Sort by priority (higher priority first) and then by name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.Priority != result[j].Metadata.Priority {
			return result[i].Metadata.Priority > result[j].Metadata.Priority
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// GetTemplatesForService returns all templates associated with a service
func (r *InMemoryTemplateRegistry) GetTemplatesForService(service string) []TemplateDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []TemplateDefinition
	for _, template := range r.templates {
		for _, svc := range template.Services {
			if svc == service {
				result = append(result, template)
				break
			}
		}
	}

	// Sort by priority and name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.Priority != result[j].Metadata.Priority {
			return result[i].Metadata.Priority > result[j].Metadata.Priority
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// GetTemplatesForEnabledServices returns all templates that are either:
// 1. Not associated with any service (universal templates), or
// 2. Associated with at least one enabled service
// This implements Property 9: Service-Based Template Filtering
// **Validates: Requirements 2.4**
func (r *InMemoryTemplateRegistry) GetTemplatesForEnabledServices(enabledServices []string) []TemplateDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Create a set of enabled services for O(1) lookup
	enabledSet := make(map[string]bool)
	for _, service := range enabledServices {
		enabledSet[service] = true
	}

	var result []TemplateDefinition
	for _, template := range r.templates {
		// Include templates with no services (universal templates)
		if len(template.Services) == 0 {
			result = append(result, template)
			continue
		}

		// Include templates if at least one of their services is enabled
		hasEnabledService := false
		for _, svc := range template.Services {
			if enabledSet[svc] {
				hasEnabledService = true
				break
			}
		}

		if hasEnabledService {
			result = append(result, template)
		}
	}

	// Sort by priority and name
	sort.Slice(result, func(i, j int) bool {
		if result[i].Metadata.Priority != result[j].Metadata.Priority {
			return result[i].Metadata.Priority > result[j].Metadata.Priority
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// ResolveTemplateDependencies resolves template dependencies in correct order
func (r *InMemoryTemplateRegistry) ResolveTemplateDependencies(templates []string) ([]TemplateDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	visited := make(map[string]bool)
	resolved := make([]TemplateDefinition, 0)
	visiting := make(map[string]bool)

	var resolve func(name string) error
	resolve = func(name string) error {
		if visited[name] {
			return nil
		}

		if visiting[name] {
			return fmt.Errorf("circular dependency detected involving template %s", name)
		}

		template, exists := r.templates[name]
		if !exists {
			return fmt.Errorf("template %s not found", name)
		}

		visiting[name] = true

		// Resolve dependencies first
		for _, dep := range template.Dependencies {
			if err := resolve(dep); err != nil {
				return err
			}
		}

		visiting[name] = false
		visited[name] = true
		resolved = append(resolved, template)

		return nil
	}

	// Resolve each requested template
	for _, name := range templates {
		if err := resolve(name); err != nil {
			return nil, err
		}
	}

	return resolved, nil
}

// ListTemplates returns all registered templates
func (r *InMemoryTemplateRegistry) ListTemplates() []TemplateDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]TemplateDefinition, 0, len(r.templates))
	for _, template := range r.templates {
		result = append(result, template)
	}

	// Sort by name for consistent ordering
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

// UnregisterTemplate removes a template from the registry
func (r *InMemoryTemplateRegistry) UnregisterTemplate(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.templates[name]; !exists {
		return fmt.Errorf("template %s not found", name)
	}

	// Check if any other templates depend on this one
	for _, template := range r.templates {
		for _, dep := range template.Dependencies {
			if dep == name {
				return fmt.Errorf("cannot unregister template %s: template %s depends on it", name, template.Name)
			}
		}
	}

	delete(r.templates, name)
	return nil
}
