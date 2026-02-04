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
	"fmt"
)

// PluginMetadata contains all standard plugin metadata
type PluginMetadata struct {
	Name        string      // Plugin identifier (e.g., "cert-manager")
	Version     string      // Plugin version (e.g., "1.0.0")
	Description string      // Human-readable description
	Type        ServiceType // Plugin category: core, monitoring, logging, storage, networking, security, gitops, custom
	Author      string      // Plugin author (e.g., "opencenter")
	License     string      // License identifier (e.g., "Apache-2.0")
}

// BaseServicePlugin provides common functionality for all service plugins
// using composition pattern. Plugins can embed this struct to inherit
// boilerplate methods and inject custom validation/rendering logic.
type BaseServicePlugin struct {
	metadata   PluginMetadata
	validator  func(interface{}) error
	renderer   func(context.Context, interface{}, interface{}) error
	statusFunc func(interface{}) ServiceStatus
}

// NewBasePlugin creates a new base plugin with the given metadata
func NewBasePlugin(metadata PluginMetadata) *BaseServicePlugin {
	return &BaseServicePlugin{
		metadata: metadata,
		// Default no-op functions
		validator: func(interface{}) error { return nil },
		renderer:  func(context.Context, interface{}, interface{}) error { return nil },
		statusFunc: func(interface{}) ServiceStatus {
			return ServiceStatus{
				State:   "pending",
				Message: "Service not yet deployed",
			}
		},
	}
}

// Name returns the plugin name
func (p *BaseServicePlugin) Name() string {
	return p.metadata.Name
}

// Version returns the plugin version
func (p *BaseServicePlugin) Version() string {
	return p.metadata.Version
}

// Description returns the plugin description
func (p *BaseServicePlugin) Description() string {
	return p.metadata.Description
}

// Type returns the service type
func (p *BaseServicePlugin) Type() ServiceType {
	return p.metadata.Type
}

// Author returns the plugin author
func (p *BaseServicePlugin) Author() string {
	return p.metadata.Author
}

// License returns the plugin license
func (p *BaseServicePlugin) License() string {
	return p.metadata.License
}

// Validate delegates to the injected validator function
func (p *BaseServicePlugin) Validate(config interface{}) error {
	if p.validator == nil {
		return nil
	}
	if err := p.validator(config); err != nil {
		return fmt.Errorf("validation failed for plugin %s: %w", p.Name(), err)
	}
	return nil
}

// Render delegates to the injected renderer function
func (p *BaseServicePlugin) Render(ctx context.Context, config interface{}, workspace interface{}) error {
	if p.renderer == nil {
		return nil
	}
	if err := p.renderer(ctx, config, workspace); err != nil {
		return fmt.Errorf("rendering failed for plugin %s: %w", p.Name(), err)
	}
	return nil
}

// Status delegates to the injected status function
func (p *BaseServicePlugin) Status(config interface{}) ServiceStatus {
	if p.statusFunc == nil {
		return ServiceStatus{
			State:   "pending",
			Message: "Service not yet deployed",
		}
	}
	return p.statusFunc(config)
}

// SetValidator allows plugins to inject custom validation logic
func (p *BaseServicePlugin) SetValidator(validator func(interface{}) error) {
	if validator != nil {
		p.validator = validator
	}
}

// SetRenderer allows plugins to inject custom rendering logic
func (p *BaseServicePlugin) SetRenderer(renderer func(context.Context, interface{}, interface{}) error) {
	if renderer != nil {
		p.renderer = renderer
	}
}

// SetStatusFunc allows plugins to inject custom status logic
func (p *BaseServicePlugin) SetStatusFunc(statusFunc func(interface{}) ServiceStatus) {
	if statusFunc != nil {
		p.statusFunc = statusFunc
	}
}
