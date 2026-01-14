package template

import (
	"fmt"
	"strings"
)

// ValidateTemplateDefinition validates a template definition for completeness and correctness
func ValidateTemplateDefinition(template TemplateDefinition) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}

	if template.Path == "" {
		return fmt.Errorf("template path is required for template %s", template.Name)
	}

	if template.Type == "" {
		return fmt.Errorf("template type is required for template %s", template.Name)
	}

	// Validate template type
	validTypes := map[TemplateType]bool{
		TemplateTypeInfrastructure: true,
		TemplateTypeService:        true,
		TemplateTypeBase:           true,
		TemplateTypeOverlay:        true,
	}

	if !validTypes[template.Type] {
		return fmt.Errorf("invalid template type %s for template %s", template.Type, template.Name)
	}

	// Validate conditions
	for i, condition := range template.Conditions {
		if err := ValidateRenderCondition(condition); err != nil {
			return fmt.Errorf("invalid condition %d for template %s: %w", i, template.Name, err)
		}
	}

	return nil
}

// ValidateRenderCondition validates a render condition
func ValidateRenderCondition(condition RenderCondition) error {
	if condition.Type == "" {
		return fmt.Errorf("condition type is required")
	}

	validTypes := map[ConditionType]bool{
		ConditionTypeEquals:      true,
		ConditionTypeNotEquals:   true,
		ConditionTypeContains:    true,
		ConditionTypeExists:      true,
		ConditionTypeGreaterThan: true,
		ConditionTypeLessThan:    true,
	}

	if !validTypes[condition.Type] {
		return fmt.Errorf("invalid condition type: %s", condition.Type)
	}

	// Field is required for all condition types
	if condition.Field == "" {
		return fmt.Errorf("condition field is required")
	}

	// Value is required for all types except Exists
	if condition.Type != ConditionTypeExists && condition.Value == nil {
		return fmt.Errorf("condition value is required for type %s", condition.Type)
	}

	return nil
}

// MatchesProvider checks if a template matches a given provider
func MatchesProvider(template TemplateDefinition, provider string) bool {
	// Empty provider means universal template
	if template.Provider == "" {
		return true
	}
	return strings.EqualFold(template.Provider, provider)
}

// MatchesService checks if a template is associated with a given service
func MatchesService(template TemplateDefinition, service string) bool {
	for _, svc := range template.Services {
		if strings.EqualFold(svc, service) {
			return true
		}
	}
	return false
}

// HasTag checks if a template has a specific tag
func HasTag(template TemplateDefinition, tag string) bool {
	for _, t := range template.Metadata.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// FilterByTags returns templates that have all specified tags
func FilterByTags(templates []TemplateDefinition, tags []string) []TemplateDefinition {
	if len(tags) == 0 {
		return templates
	}

	var result []TemplateDefinition
	for _, template := range templates {
		hasAllTags := true
		for _, tag := range tags {
			if !HasTag(template, tag) {
				hasAllTags = false
				break
			}
		}
		if hasAllTags {
			result = append(result, template)
		}
	}

	return result
}

// FilterByType returns templates of a specific type
func FilterByType(templates []TemplateDefinition, templateType TemplateType) []TemplateDefinition {
	var result []TemplateDefinition
	for _, template := range templates {
		if template.Type == templateType {
			result = append(result, template)
		}
	}
	return result
}

// GetTemplateDependencyGraph builds a dependency graph for templates
func GetTemplateDependencyGraph(templates []TemplateDefinition) map[string][]string {
	graph := make(map[string][]string)
	for _, template := range templates {
		graph[template.Name] = template.Dependencies
	}
	return graph
}

// HasCircularDependencies checks if there are circular dependencies in a set of templates
func HasCircularDependencies(templates []TemplateDefinition) (bool, string) {
	graph := GetTemplateDependencyGraph(templates)
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var hasCycle func(name string) (bool, string)
	hasCycle = func(name string) (bool, string) {
		if visited[name] {
			return false, ""
		}

		if visiting[name] {
			return true, name
		}

		visiting[name] = true

		for _, dep := range graph[name] {
			if cycle, cycleName := hasCycle(dep); cycle {
				return true, cycleName
			}
		}

		visiting[name] = false
		visited[name] = true

		return false, ""
	}

	for name := range graph {
		if cycle, cycleName := hasCycle(name); cycle {
			return true, cycleName
		}
	}

	return false, ""
}
