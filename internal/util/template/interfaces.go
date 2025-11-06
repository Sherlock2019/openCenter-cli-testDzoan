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
	"io"
	"text/template"
)

// TemplateRenderer interface for rendering templates
type TemplateRenderer interface {
	RenderTemplate(templateName string, data interface{}) (string, error)
	RenderTemplateToWriter(templateName string, data interface{}, writer io.Writer) error
	GetTemplate(templateName string) (*template.Template, error)
	ListTemplates() []string
}

// TemplateValidator interface for validating templates and data
type TemplateValidator interface {
	ValidateTemplate(templateName string) error
	ValidateTemplateData(templateName string, data interface{}) error
	ValidateTemplateExists(templateName string) error
	ValidateRequiredFields(data interface{}, requiredFields []string) error
}

// TemplateEngine interface combining rendering and validation
type TemplateEngine interface {
	TemplateRenderer
	TemplateValidator
	Init() error
	AddFunctions(funcMap template.FuncMap) error
}

// NetworkPluginHandler interface for handling network plugin templates
type NetworkPluginHandler interface {
	ValidateNetworkPlugin(pluginType string, config map[string]interface{}) error
	RenderNetworkPluginConfig(pluginType string, config map[string]interface{}) (string, error)
	GetSupportedPlugins() []string
	GetRequiredFields(pluginType string) []string
}

// TemplateData represents data passed to templates
type TemplateData struct {
	Data   interface{}
	Config map[string]interface{}
}

// ValidationError represents a template validation error
type ValidationError struct {
	Field   string
	Message string
	Value   interface{}
}

func (e ValidationError) Error() string {
	return e.Message
}