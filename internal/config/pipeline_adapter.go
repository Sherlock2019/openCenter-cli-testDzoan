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

package config

import (
	"context"
)

// PipelineAdapter adapts the validation pipeline to the ConfigValidatorInterface.
// It delegates to EnhancedConfigValidator for the actual validation logic.
type PipelineAdapter struct {
	validator *EnhancedConfigValidator
}

// NewPipelineAdapter creates a new pipeline adapter.
func NewPipelineAdapter() *PipelineAdapter {
	return &PipelineAdapter{
		validator: NewEnhancedConfigValidator(false),
	}
}

// Validate performs comprehensive validation using the pipeline.
func (pa *PipelineAdapter) Validate(ctx context.Context, config *Config) *ConfigValidationResult {
	return pa.validator.Validate(ctx, config)
}

// ValidateStructure validates the basic structure.
func (pa *PipelineAdapter) ValidateStructure(ctx context.Context, config *Config) *ConfigValidationResult {
	return pa.validator.ValidateStructure(ctx, config)
}

// ValidateSemantics validates semantic correctness.
func (pa *PipelineAdapter) ValidateSemantics(ctx context.Context, config *Config) *ConfigValidationResult {
	return pa.validator.ValidateSemantics(ctx, config)
}

// ValidateNetworking validates network plugin configuration.
func (pa *PipelineAdapter) ValidateNetworking(ctx context.Context, config *Config) *ConfigValidationResult {
	return pa.validator.ValidateNetworking(ctx, config)
}

// ValidateCloudProvider validates cloud provider specific configuration.
func (pa *PipelineAdapter) ValidateCloudProvider(ctx context.Context, config *Config) *ConfigValidationResult {
	return pa.validator.ValidateCloudProvider(ctx, config)
}
