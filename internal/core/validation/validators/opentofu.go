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

package validators

import (
	"context"
	"fmt"

	"github.com/opencenter-cloud/opencenter-cli/internal/core/validation"
)

// OpenTofuValidator validates OpenTofu backend configuration.
type OpenTofuValidator struct{}

// NewOpenTofuValidator creates a new OpenTofu validator.
func NewOpenTofuValidator() *OpenTofuValidator {
	return &OpenTofuValidator{}
}

// Name returns the validator name.
func (v *OpenTofuValidator) Name() string {
	return "opentofu"
}

// Priority returns the validator priority.
func (v *OpenTofuValidator) Priority() int {
	return validation.PriorityNormal
}

// Validate validates OpenTofu backend configuration.
func (v *OpenTofuValidator) Validate(ctx context.Context, value interface{}) (*validation.ValidationResult, error) {
	result := &validation.ValidationResult{
		Valid:    true,
		Errors:   []*validation.ValidationIssue{},
		Warnings: []*validation.ValidationIssue{},
		Info:     []*validation.ValidationIssue{},
	}

	// Expect value to be a map representing the OpenTofu configuration
	configMap, ok := value.(map[string]interface{})
	if !ok {
		return result, nil // Skip validation if not a map
	}

	// Check if backend exists
	backend, ok := configMap["backend"].(map[string]interface{})
	if !ok {
		return result, nil // No backend to validate
	}

	// Get backend type
	backendType, ok := backend["type"].(string)
	if !ok {
		return result, nil // No type specified
	}

	// Check for old v2 format (backend.path instead of backend.local.path)
	if backendType == "local" {
		// Check if old format is being used
		if path, hasPath := backend["path"].(string); hasPath && path != "" {
			result.AddError(
				"opentofu.backend.path",
				fmt.Sprintf("The 'opentofu.backend.path' field is deprecated. Use 'opentofu.backend.local.path' instead. Current value: '%s'", path),
			)
			result.AddInfo(
				"opentofu.backend",
				"Migration: Move 'path' field under 'local' section. Example:\n"+
					"  backend:\n"+
					"    type: local\n"+
					"    local:\n"+
					"      path: "+path,
			)
			return result, nil
		}

		// Check if new format is used correctly
		local, hasLocal := backend["local"].(map[string]interface{})
		if !hasLocal {
			result.AddError(
				"opentofu.backend.local",
				"Local backend requires 'local' section with 'path' field",
			)
			result.AddInfo(
				"opentofu.backend",
				"Example configuration:\n"+
					"  backend:\n"+
					"    type: local\n"+
					"    local:\n"+
					"      path: .opentofu-local-utils/terraform.tfstate",
			)
			return result, nil
		}

		// Check if path is specified in local section
		localPath, hasPath := local["path"].(string)
		if !hasPath || localPath == "" {
			result.AddError(
				"opentofu.backend.local.path",
				"Local backend requires 'path' field to specify state file location",
			)
			result.AddInfo(
				"opentofu.backend.local",
				"Example: path: .opentofu-local-utils/terraform.tfstate",
			)
		}
	}

	return result, nil
}
