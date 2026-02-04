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
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: phase-4-cleanup-optimization, Property 1: Base Plugin Metadata Accessibility
// For any PluginMetadata with valid fields (name, version, description, type, author, license),
// when a BaseServicePlugin is created with that metadata, all accessor methods
// (Name(), Version(), Description(), Type(), Author(), License()) should return
// the exact values from the metadata.
func TestBasePluginMetadataAccessibility(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Generate valid ServiceType values
	genServiceType := gen.OneConstOf(
		ServiceTypeCore,
		ServiceTypeMonitoring,
		ServiceTypeLogging,
		ServiceTypeStorage,
		ServiceTypeNetworking,
		ServiceTypeSecurity,
		ServiceTypeGitOps,
		ServiceTypeCustom,
	)

	properties.Property("metadata fields are accessible", prop.ForAll(
		func(name, version, description, author, license string, serviceType ServiceType) bool {
			metadata := PluginMetadata{
				Name:        name,
				Version:     version,
				Description: description,
				Type:        serviceType,
				Author:      author,
				License:     license,
			}

			plugin := NewBasePlugin(metadata)

			// Verify all accessor methods return correct values
			return plugin.Name() == name &&
				plugin.Version() == version &&
				plugin.Description() == description &&
				plugin.Type() == serviceType &&
				plugin.Author() == author &&
				plugin.License() == license
		},
		gen.AnyString(),
		gen.AnyString(),
		gen.AnyString(),
		gen.AnyString(),
		gen.AnyString(),
		genServiceType,
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: phase-4-cleanup-optimization, Property 2: Custom Logic Injection
// For any custom validator function and custom renderer function, when they are
// injected into a BaseServicePlugin using SetValidator and SetRenderer, calling
// Validate() and Render() should invoke the injected functions and return their results.
func TestCustomLogicInjection(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("injected validator is called", prop.ForAll(
		func(shouldFail bool) bool {
			metadata := PluginMetadata{
				Name:    "test-plugin",
				Version: "1.0.0",
				Type:    ServiceTypeCore,
			}

			plugin := NewBasePlugin(metadata)

			// Track if validator was called
			validatorCalled := false
			plugin.SetValidator(func(config interface{}) error {
				validatorCalled = true
				if shouldFail {
					return context.DeadlineExceeded
				}
				return nil
			})

			err := plugin.Validate(nil)

			// Verify validator was called
			if !validatorCalled {
				return false
			}

			// Verify error matches expected result
			if shouldFail {
				return err != nil
			}
			return err == nil
		},
		gen.Bool(),
	))

	properties.Property("injected renderer is called", prop.ForAll(
		func(shouldFail bool) bool {
			metadata := PluginMetadata{
				Name:    "test-plugin",
				Version: "1.0.0",
				Type:    ServiceTypeCore,
			}

			plugin := NewBasePlugin(metadata)

			// Track if renderer was called
			rendererCalled := false
			plugin.SetRenderer(func(ctx context.Context, config interface{}, workspace interface{}) error {
				rendererCalled = true
				if shouldFail {
					return context.DeadlineExceeded
				}
				return nil
			})

			err := plugin.Render(context.Background(), nil, nil)

			// Verify renderer was called
			if !rendererCalled {
				return false
			}

			// Verify error matches expected result
			if shouldFail {
				return err != nil
			}
			return err == nil
		},
		gen.Bool(),
	))

	properties.Property("injected status function is called", prop.ForAll(
		func(state string) bool {
			metadata := PluginMetadata{
				Name:    "test-plugin",
				Version: "1.0.0",
				Type:    ServiceTypeCore,
			}

			plugin := NewBasePlugin(metadata)

			// Track if status function was called
			statusCalled := false
			plugin.SetStatusFunc(func(config interface{}) ServiceStatus {
				statusCalled = true
				return ServiceStatus{
					State:   state,
					Message: "test message",
				}
			})

			status := plugin.Status(nil)

			// Verify status function was called
			if !statusCalled {
				return false
			}

			// Verify status matches expected result
			return status.State == state && status.Message == "test message"
		},
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
