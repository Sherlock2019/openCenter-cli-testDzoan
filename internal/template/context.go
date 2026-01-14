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
	"text/template"
)

// RenderContext provides comprehensive context for template rendering operations.
// It includes configuration data, metadata, custom functions, and rendering options.
type RenderContext struct {
	// Data is the primary data object passed to the template
	Data interface{}

	// Metadata contains additional contextual information
	Metadata map[string]interface{}

	// Functions contains custom template functions
	Functions template.FuncMap

	// Options contains rendering options
	Options RenderOptions
}

// RenderOptions configures template rendering behavior.
type RenderOptions struct {
	// StrictMode causes rendering to fail on missing keys
	StrictMode bool

	// MissingKeyBehavior controls how missing keys are handled
	MissingKeyBehavior MissingKeyBehavior

	// ValidateBeforeRender validates template syntax before rendering
	ValidateBeforeRender bool

	// IncludeLineNumbers includes line numbers in error messages
	IncludeLineNumbers bool
}

// MissingKeyBehavior defines how to handle missing keys in templates.
type MissingKeyBehavior int

const (
	// MissingKeyError causes an error when a key is missing
	MissingKeyError MissingKeyBehavior = iota

	// MissingKeyZero returns the zero value for missing keys
	MissingKeyZero

	// MissingKeyInvalid returns "<no value>" for missing keys (default)
	MissingKeyInvalid
)

// NewRenderContext creates a new render context with the given data.
func NewRenderContext(data interface{}) *RenderContext {
	return &RenderContext{
		Data:      data,
		Metadata:  make(map[string]interface{}),
		Functions: make(template.FuncMap),
		Options: RenderOptions{
			StrictMode:           false,
			MissingKeyBehavior:   MissingKeyInvalid,
			ValidateBeforeRender: true,
			IncludeLineNumbers:   true,
		},
	}
}

// WithMetadata adds metadata to the render context.
// Returns the context for method chaining.
func (rc *RenderContext) WithMetadata(key string, value interface{}) *RenderContext {
	rc.Metadata[key] = value
	return rc
}

// WithFunction adds a custom function to the render context.
// Returns the context for method chaining.
func (rc *RenderContext) WithFunction(name string, fn interface{}) *RenderContext {
	rc.Functions[name] = fn
	return rc
}

// WithFunctions adds multiple custom functions to the render context.
// Returns the context for method chaining.
func (rc *RenderContext) WithFunctions(funcMap template.FuncMap) *RenderContext {
	for name, fn := range funcMap {
		rc.Functions[name] = fn
	}
	return rc
}

// WithStrictMode enables or disables strict mode.
// In strict mode, missing keys cause rendering to fail.
// Returns the context for method chaining.
func (rc *RenderContext) WithStrictMode(strict bool) *RenderContext {
	rc.Options.StrictMode = strict
	if strict {
		rc.Options.MissingKeyBehavior = MissingKeyError
	}
	return rc
}

// WithMissingKeyBehavior sets the missing key behavior.
// Returns the context for method chaining.
func (rc *RenderContext) WithMissingKeyBehavior(behavior MissingKeyBehavior) *RenderContext {
	rc.Options.MissingKeyBehavior = behavior
	return rc
}

// WithValidation enables or disables pre-render validation.
// Returns the context for method chaining.
func (rc *RenderContext) WithValidation(validate bool) *RenderContext {
	rc.Options.ValidateBeforeRender = validate
	return rc
}

// ToTemplateData converts the render context to a map suitable for template rendering.
// This provides a consistent structure for accessing data and metadata in templates.
func (rc *RenderContext) ToTemplateData() map[string]interface{} {
	return map[string]interface{}{
		"Data":     rc.Data,
		"Metadata": rc.Metadata,
	}
}

// Validate validates the render context for completeness.
// Returns an error if the context is invalid.
func (rc *RenderContext) Validate() error {
	if rc.Data == nil {
		return fmt.Errorf("render context data cannot be nil")
	}

	// Validate custom functions
	for name, fn := range rc.Functions {
		if fn == nil {
			return fmt.Errorf("function %s cannot be nil", name)
		}
	}

	return nil
}

// Clone creates a deep copy of the render context.
// This is useful for creating isolated contexts for concurrent rendering.
func (rc *RenderContext) Clone() *RenderContext {
	clone := &RenderContext{
		Data:      rc.Data,
		Metadata:  make(map[string]interface{}),
		Functions: make(template.FuncMap),
		Options:   rc.Options,
	}

	// Copy metadata
	for k, v := range rc.Metadata {
		clone.Metadata[k] = v
	}

	// Copy functions
	for k, v := range rc.Functions {
		clone.Functions[k] = v
	}

	return clone
}

// ContextBuilder provides a fluent interface for building render contexts.
type ContextBuilder struct {
	context *RenderContext
}

// NewContextBuilder creates a new context builder.
func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		context: NewRenderContext(nil),
	}
}

// WithData sets the primary data for the context.
func (cb *ContextBuilder) WithData(data interface{}) *ContextBuilder {
	cb.context.Data = data
	return cb
}

// WithMetadata adds metadata to the context.
func (cb *ContextBuilder) WithMetadata(key string, value interface{}) *ContextBuilder {
	cb.context.WithMetadata(key, value)
	return cb
}

// WithFunction adds a custom function to the context.
func (cb *ContextBuilder) WithFunction(name string, fn interface{}) *ContextBuilder {
	cb.context.WithFunction(name, fn)
	return cb
}

// WithFunctions adds multiple custom functions to the context.
func (cb *ContextBuilder) WithFunctions(funcMap template.FuncMap) *ContextBuilder {
	cb.context.WithFunctions(funcMap)
	return cb
}

// WithStrictMode enables or disables strict mode.
func (cb *ContextBuilder) WithStrictMode(strict bool) *ContextBuilder {
	cb.context.WithStrictMode(strict)
	return cb
}

// WithValidation enables or disables pre-render validation.
func (cb *ContextBuilder) WithValidation(validate bool) *ContextBuilder {
	cb.context.WithValidation(validate)
	return cb
}

// Build builds and validates the render context.
func (cb *ContextBuilder) Build() (*RenderContext, error) {
	if err := cb.context.Validate(); err != nil {
		return nil, fmt.Errorf("invalid render context: %w", err)
	}
	return cb.context, nil
}
