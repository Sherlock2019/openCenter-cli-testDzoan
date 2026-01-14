package template

import (
	"fmt"
	"sort"
)

// DependencyResolver handles template dependency resolution
type DependencyResolver struct {
	registry TemplateRegistry
}

// NewDependencyResolver creates a new dependency resolver
func NewDependencyResolver(registry TemplateRegistry) *DependencyResolver {
	return &DependencyResolver{
		registry: registry,
	}
}

// ResolveDependencies resolves all dependencies for a set of templates
// Returns templates in dependency order (dependencies first)
func (r *DependencyResolver) ResolveDependencies(templateNames []string) ([]TemplateDefinition, error) {
	return r.registry.ResolveTemplateDependencies(templateNames)
}

// GetDependencyTree builds a complete dependency tree for a template
func (r *DependencyResolver) GetDependencyTree(templateName string) (map[string][]string, error) {
	tree := make(map[string][]string)
	visited := make(map[string]bool)

	var buildTree func(name string) error
	buildTree = func(name string) error {
		if visited[name] {
			return nil
		}

		template, err := r.registry.GetTemplate(name)
		if err != nil {
			return err
		}

		visited[name] = true
		tree[name] = template.Dependencies

		for _, dep := range template.Dependencies {
			if err := buildTree(dep); err != nil {
				return err
			}
		}

		return nil
	}

	if err := buildTree(templateName); err != nil {
		return nil, err
	}

	return tree, nil
}

// ValidateDependencies checks if all dependencies for a template exist
func (r *DependencyResolver) ValidateDependencies(templateName string) error {
	template, err := r.registry.GetTemplate(templateName)
	if err != nil {
		return err
	}

	for _, dep := range template.Dependencies {
		if _, err := r.registry.GetTemplate(dep); err != nil {
			return fmt.Errorf("dependency %s for template %s not found: %w", dep, templateName, err)
		}

		// Recursively validate dependencies
		if err := r.ValidateDependencies(dep); err != nil {
			return err
		}
	}

	return nil
}

// GetAllDependencies returns all direct and transitive dependencies for a template
func (r *DependencyResolver) GetAllDependencies(templateName string) ([]string, error) {
	allDeps := make(map[string]bool)
	visited := make(map[string]bool)

	var collectDeps func(name string) error
	collectDeps = func(name string) error {
		if visited[name] {
			return nil
		}

		template, err := r.registry.GetTemplate(name)
		if err != nil {
			return err
		}

		visited[name] = true

		for _, dep := range template.Dependencies {
			allDeps[dep] = true
			if err := collectDeps(dep); err != nil {
				return err
			}
		}

		return nil
	}

	if err := collectDeps(templateName); err != nil {
		return nil, err
	}

	// Convert map to sorted slice
	result := make([]string, 0, len(allDeps))
	for dep := range allDeps {
		result = append(result, dep)
	}
	sort.Strings(result)

	return result, nil
}

// GetDependents returns all templates that depend on the given template
func (r *DependencyResolver) GetDependents(templateName string) []string {
	var dependents []string

	for _, template := range r.registry.ListTemplates() {
		for _, dep := range template.Dependencies {
			if dep == templateName {
				dependents = append(dependents, template.Name)
				break
			}
		}
	}

	sort.Strings(dependents)
	return dependents
}

// TopologicalSort performs a topological sort on templates based on dependencies
func (r *DependencyResolver) TopologicalSort(templateNames []string) ([]TemplateDefinition, error) {
	// Build in-degree map
	inDegree := make(map[string]int)
	graph := make(map[string][]string)
	templates := make(map[string]TemplateDefinition)

	// Initialize for requested templates
	for _, name := range templateNames {
		template, err := r.registry.GetTemplate(name)
		if err != nil {
			return nil, err
		}
		templates[name] = template
		inDegree[name] = 0
		graph[name] = template.Dependencies
	}

	// Calculate in-degrees
	for name, deps := range graph {
		for _, dep := range deps {
			// Ensure dependency is in our set
			if _, exists := templates[dep]; !exists {
				depTemplate, err := r.registry.GetTemplate(dep)
				if err != nil {
					return nil, fmt.Errorf("dependency %s not found: %w", dep, err)
				}
				templates[dep] = depTemplate
				inDegree[dep] = 0
				graph[dep] = depTemplate.Dependencies
			}
			inDegree[name]++
		}
	}

	// Find all nodes with in-degree 0
	queue := make([]string, 0)
	for name, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]TemplateDefinition, 0)
	visited := 0

	for len(queue) > 0 {
		// Sort queue for deterministic ordering
		sort.Strings(queue)

		current := queue[0]
		queue = queue[1:]

		result = append(result, templates[current])
		visited++

		// Reduce in-degree for dependents
		for name, deps := range graph {
			for _, dep := range deps {
				if dep == current {
					inDegree[name]--
					if inDegree[name] == 0 {
						queue = append(queue, name)
					}
				}
			}
		}
	}

	if visited != len(templates) {
		return nil, fmt.Errorf("circular dependency detected in templates")
	}

	return result, nil
}

// CanSafelyRemove checks if a template can be safely removed without breaking dependencies
func (r *DependencyResolver) CanSafelyRemove(templateName string) (bool, []string) {
	dependents := r.GetDependents(templateName)
	return len(dependents) == 0, dependents
}
