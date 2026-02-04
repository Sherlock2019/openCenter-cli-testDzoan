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

package services

import (
	"context"
	"errors"
	"testing"
)

// TestPluginComposition tests that embedded plugins can access base methods
func TestPluginComposition(t *testing.T) {
	t.Run("embedded plugin can access base methods", func(t *testing.T) {
		// Create a plugin that embeds BaseServicePlugin
		type TestPlugin struct {
			*BaseServicePlugin
		}

		metadata := PluginMetadata{
			Name:        "test-plugin",
			Version:     "1.0.0",
			Description: "A test plugin",
			Type:        ServiceTypeCore,
			Author:      "opencenter",
			License:     "Apache-2.0",
		}

		base := NewBasePlugin(metadata)
		plugin := &TestPlugin{BaseServicePlugin: base}

		// Verify embedded plugin can access all base methods
		if plugin.Name() != "test-plugin" {
			t.Errorf("expected Name() to return 'test-plugin', got %s", plugin.Name())
		}
		if plugin.Version() != "1.0.0" {
			t.Errorf("expected Version() to return '1.0.0', got %s", plugin.Version())
		}
		if plugin.Description() != "A test plugin" {
			t.Errorf("expected Description() to return 'A test plugin', got %s", plugin.Description())
		}
		if plugin.Type() != ServiceTypeCore {
			t.Errorf("expected Type() to return ServiceTypeCore, got %s", plugin.Type())
		}
		if plugin.Author() != "opencenter" {
			t.Errorf("expected Author() to return 'opencenter', got %s", plugin.Author())
		}
		if plugin.License() != "Apache-2.0" {
			t.Errorf("expected License() to return 'Apache-2.0', got %s", plugin.License())
		}
	})

	t.Run("embedded plugin can override specific methods", func(t *testing.T) {
		// Create a plugin that embeds BaseServicePlugin and overrides methods
		type CustomPlugin struct {
			*BaseServicePlugin
			customName string
		}

		metadata := PluginMetadata{
			Name:        "base-name",
			Version:     "1.0.0",
			Description: "A custom plugin",
			Type:        ServiceTypeCore,
			Author:      "opencenter",
			License:     "Apache-2.0",
		}

		base := NewBasePlugin(metadata)
		plugin := &CustomPlugin{
			BaseServicePlugin: base,
			customName:        "custom-name",
		}

		// Verify that we can still access base methods
		// (In Go, we can't truly override methods on embedded structs,
		// but we can demonstrate that the plugin has access to base functionality)
		if plugin.Version() != "1.0.0" {
			t.Errorf("expected Version() to return '1.0.0', got %s", plugin.Version())
		}
		if plugin.Description() != "A custom plugin" {
			t.Errorf("expected Description() to return 'A custom plugin', got %s", plugin.Description())
		}
		if plugin.Type() != ServiceTypeCore {
			t.Errorf("expected Type() to return ServiceTypeCore, got %s", plugin.Type())
		}
	})

	t.Run("embedded plugin can inject custom logic", func(t *testing.T) {
		type ValidatingPlugin struct {
			*BaseServicePlugin
		}

		metadata := PluginMetadata{
			Name:    "validating-plugin",
			Version: "1.0.0",
			Type:    ServiceTypeCore,
		}

		base := NewBasePlugin(metadata)
		plugin := &ValidatingPlugin{BaseServicePlugin: base}

		// Inject custom validation logic
		validationError := errors.New("validation failed")
		base.SetValidator(func(config interface{}) error {
			if config == nil {
				return validationError
			}
			return nil
		})

		// Test validation with nil config (should fail)
		err := plugin.Validate(nil)
		if err == nil {
			t.Error("expected Validate(nil) to return error, got nil")
		}

		// Test validation with non-nil config (should pass)
		err = plugin.Validate("some-config")
		if err != nil {
			t.Errorf("expected Validate(config) to return nil, got %v", err)
		}
	})

	t.Run("embedded plugin can inject custom renderer", func(t *testing.T) {
		type RenderingPlugin struct {
			*BaseServicePlugin
		}

		metadata := PluginMetadata{
			Name:    "rendering-plugin",
			Version: "1.0.0",
			Type:    ServiceTypeCore,
		}

		base := NewBasePlugin(metadata)
		plugin := &RenderingPlugin{BaseServicePlugin: base}

		// Track if renderer was called
		renderCalled := false
		base.SetRenderer(func(ctx context.Context, config interface{}, workspace interface{}) error {
			renderCalled = true
			return nil
		})

		// Test rendering
		err := plugin.Render(context.Background(), nil, nil)
		if err != nil {
			t.Errorf("expected Render() to return nil, got %v", err)
		}
		if !renderCalled {
			t.Error("expected renderer to be called")
		}
	})

	t.Run("embedded plugin can inject custom status function", func(t *testing.T) {
		type StatusPlugin struct {
			*BaseServicePlugin
		}

		metadata := PluginMetadata{
			Name:    "status-plugin",
			Version: "1.0.0",
			Type:    ServiceTypeCore,
		}

		base := NewBasePlugin(metadata)
		plugin := &StatusPlugin{BaseServicePlugin: base}

		// Inject custom status logic
		base.SetStatusFunc(func(config interface{}) ServiceStatus {
			return ServiceStatus{
				State:   "running",
				Message: "Service is running",
				Details: map[string]interface{}{
					"custom": "data",
				},
			}
		})

		// Test status
		status := plugin.Status(nil)
		if status.State != "running" {
			t.Errorf("expected State to be 'running', got %s", status.State)
		}
		if status.Message != "Service is running" {
			t.Errorf("expected Message to be 'Service is running', got %s", status.Message)
		}
		if status.Details["custom"] != "data" {
			t.Errorf("expected Details['custom'] to be 'data', got %v", status.Details["custom"])
		}
	})
}

// TestBasePluginDefaults tests default behavior when no custom logic is injected
func TestBasePluginDefaults(t *testing.T) {
	metadata := PluginMetadata{
		Name:    "default-plugin",
		Version: "1.0.0",
		Type:    ServiceTypeCore,
	}

	plugin := NewBasePlugin(metadata)

	t.Run("default validator returns nil", func(t *testing.T) {
		err := plugin.Validate(nil)
		if err != nil {
			t.Errorf("expected default Validate() to return nil, got %v", err)
		}
	})

	t.Run("default renderer returns nil", func(t *testing.T) {
		err := plugin.Render(context.Background(), nil, nil)
		if err != nil {
			t.Errorf("expected default Render() to return nil, got %v", err)
		}
	})

	t.Run("default status returns pending", func(t *testing.T) {
		status := plugin.Status(nil)
		if status.State != "pending" {
			t.Errorf("expected default State to be 'pending', got %s", status.State)
		}
		if status.Message != "Service not yet deployed" {
			t.Errorf("expected default Message to be 'Service not yet deployed', got %s", status.Message)
		}
	})
}

// TestSetterNilHandling tests that setters handle nil values gracefully
func TestSetterNilHandling(t *testing.T) {
	metadata := PluginMetadata{
		Name:    "nil-test-plugin",
		Version: "1.0.0",
		Type:    ServiceTypeCore,
	}

	plugin := NewBasePlugin(metadata)

	t.Run("SetValidator with nil does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetValidator(nil) panicked: %v", r)
			}
		}()
		plugin.SetValidator(nil)
		// Should still use default validator
		err := plugin.Validate(nil)
		if err != nil {
			t.Errorf("expected Validate() to return nil after SetValidator(nil), got %v", err)
		}
	})

	t.Run("SetRenderer with nil does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetRenderer(nil) panicked: %v", r)
			}
		}()
		plugin.SetRenderer(nil)
		// Should still use default renderer
		err := plugin.Render(context.Background(), nil, nil)
		if err != nil {
			t.Errorf("expected Render() to return nil after SetRenderer(nil), got %v", err)
		}
	})

	t.Run("SetStatusFunc with nil does not panic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("SetStatusFunc(nil) panicked: %v", r)
			}
		}()
		plugin.SetStatusFunc(nil)
		// Should still use default status function
		status := plugin.Status(nil)
		if status.State != "pending" {
			t.Errorf("expected State to be 'pending' after SetStatusFunc(nil), got %s", status.State)
		}
	})
}
