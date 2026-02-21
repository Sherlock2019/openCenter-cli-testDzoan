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
	"testing"
)

func TestOpenTofuValidator_OldFormat(t *testing.T) {
	validator := NewOpenTofuValidator()
	ctx := context.Background()

	// Old v2 format with backend.path
	config := map[string]interface{}{
		"backend": map[string]interface{}{
			"type": "local",
			"path": ".opentofu-local-utils/terraform.tfstate",
		},
	}

	result, err := validator.Validate(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected validation to fail for old format")
	}

	if len(result.Errors) == 0 {
		t.Error("expected error for deprecated backend.path field")
	}

	// Check error message mentions migration
	foundMigrationInfo := false
	for _, info := range result.Info {
		if info.Field == "opentofu.backend" {
			foundMigrationInfo = true
			break
		}
	}

	if !foundMigrationInfo {
		t.Error("expected migration info to be provided")
	}
}

func TestOpenTofuValidator_NewFormatValid(t *testing.T) {
	validator := NewOpenTofuValidator()
	ctx := context.Background()

	// New format with backend.local.path
	config := map[string]interface{}{
		"backend": map[string]interface{}{
			"type": "local",
			"local": map[string]interface{}{
				"path": ".opentofu-local-utils/terraform.tfstate",
			},
		},
	}

	result, err := validator.Validate(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected validation to pass for new format, got errors: %v", result.Errors)
	}
}

func TestOpenTofuValidator_NewFormatMissingPath(t *testing.T) {
	validator := NewOpenTofuValidator()
	ctx := context.Background()

	// New format but missing path
	config := map[string]interface{}{
		"backend": map[string]interface{}{
			"type":  "local",
			"local": map[string]interface{}{},
		},
	}

	result, err := validator.Validate(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected validation to fail when path is missing")
	}

	if len(result.Errors) == 0 {
		t.Error("expected error for missing path")
	}
}

func TestOpenTofuValidator_NewFormatMissingLocal(t *testing.T) {
	validator := NewOpenTofuValidator()
	ctx := context.Background()

	// New format but missing local section
	config := map[string]interface{}{
		"backend": map[string]interface{}{
			"type": "local",
		},
	}

	result, err := validator.Validate(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Valid {
		t.Error("expected validation to fail when local section is missing")
	}

	if len(result.Errors) == 0 {
		t.Error("expected error for missing local section")
	}
}

func TestOpenTofuValidator_S3Backend(t *testing.T) {
	validator := NewOpenTofuValidator()
	ctx := context.Background()

	// S3 backend should not require local.path
	config := map[string]interface{}{
		"backend": map[string]interface{}{
			"type": "s3",
			"s3": map[string]interface{}{
				"bucket": "my-bucket",
				"key":    "terraform.tfstate",
				"region": "us-east-1",
			},
		},
	}

	result, err := validator.Validate(ctx, config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Valid {
		t.Errorf("expected validation to pass for S3 backend, got errors: %v", result.Errors)
	}
}
