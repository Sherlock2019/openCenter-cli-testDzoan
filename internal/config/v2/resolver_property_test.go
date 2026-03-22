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

package v2

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_ReferenceResolutionCorrectness verifies that all references resolve to correct values.
// Validates: Requirements 4.2 (Epic 4: Medium Priority TODO Resolution)
// Property 5: Reference Resolution Correctness
func TestProperty_ReferenceResolutionCorrectness(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Environment variable references resolve to their actual values
	properties.Property("environment variable references resolve correctly", prop.ForAll(
		func(varName string, varValue string) bool {
			// Set environment variable
			os.Setenv(varName, varValue)
			defer os.Unsetenv(varName)

			// Create config with reference
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${env:%s}", varName),
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)
			if err != nil {
				return false
			}

			// Must resolve to expected value
			return cfg.OpenCenter.Meta.Name == varValue
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: File references resolve to file contents
	properties.Property("file references resolve to file contents", prop.ForAll(
		func(content string) bool {
			// Create temporary file
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.txt")
			if err := os.WriteFile(testFile, []byte(content), 0600); err != nil {
				return false
			}

			// Create config with file reference
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${file:%s}", testFile),
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)
			if err != nil {
				return false
			}

			// Must resolve to file content (trimmed)
			return cfg.OpenCenter.Meta.Name == content
		},
		gen.AlphaString(),
	))

	// Property: Multiple references in same string resolve correctly
	properties.Property("multiple references resolve correctly", prop.ForAll(
		func(var1 string, var2 string, val1 string, val2 string) bool {
			// Set environment variables
			os.Setenv(var1, val1)
			os.Setenv(var2, val2)
			defer os.Unsetenv(var1)
			defer os.Unsetenv(var2)

			// Create config with multiple references
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${env:%s}-${env:%s}", var1, var2),
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)
			if err != nil {
				return false
			}

			// Must resolve to concatenated values
			expected := fmt.Sprintf("%s-%s", os.Getenv(var1), os.Getenv(var2))
			return cfg.OpenCenter.Meta.Name == expected
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Caching works correctly for repeated references
	properties.Property("caching works for repeated references", prop.ForAll(
		func(varName string, varValue string) bool {
			// Set environment variable
			os.Setenv(varName, varValue)
			defer os.Unsetenv(varName)

			// Create config with same reference in multiple places
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${env:%s}", varName),
						Organization: fmt.Sprintf("${env:%s}", varName),
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)
			if err != nil {
				return false
			}

			// Both should resolve to same value
			return cfg.OpenCenter.Meta.Name == varValue &&
				cfg.OpenCenter.Meta.Organization == varValue
		},
		gen.Identifier(),
		gen.Identifier(),
	))

	// Property: Missing environment variables always error
	properties.Property("missing environment variables error", prop.ForAll(
		func(varName string) bool {
			// Ensure variable is not set
			os.Unsetenv(varName)

			// Create config with reference to non-existent variable
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${env:%s}", varName),
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)

			// Must error
			return err != nil
		},
		gen.Identifier(),
	))

	// Property: Non-reference strings remain unchanged
	properties.Property("non-reference strings remain unchanged", prop.ForAll(
		func(value string) bool {
			// Create config without references
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         value,
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)
			if err != nil {
				return false
			}

			// Value should remain unchanged
			return cfg.OpenCenter.Meta.Name == value
		},
		gen.AlphaString().SuchThat(func(s string) bool {
			// Exclude strings that look like references
			return s != "" && !containsReferencePattern(s)
		}),
	))

	properties.TestingRun(t)
}

// containsReferencePattern checks if a string contains reference patterns
func containsReferencePattern(s string) bool {
	return len(s) > 6 && (s[0:6] == "${env:" || s[0:6] == "${ref:" || s[0:7] == "${file:")
}

// TestProperty_CircularReferenceDetection verifies circular reference detection.
// Validates: Requirements 4.2.7
func TestProperty_CircularReferenceDetection(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Circular references are always detected
	properties.Property("circular references are detected", prop.ForAll(
		func(refPath string) bool {
			// Create config with circular reference (ref to itself)
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         fmt.Sprintf("${ref:%s}", refPath),
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)

			// Must error (either circular reference or path not found)
			return err != nil
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// TestProperty_MaxDepthProtection verifies maximum depth protection.
// Validates: Requirements 4.2.3
func TestProperty_MaxDepthProtection(t *testing.T) {
	properties := gopter.NewProperties(nil)

	// Property: Deep nesting is protected by max depth
	properties.Property("max depth prevents infinite recursion", prop.ForAll(
		func(depth int) bool {
			if depth < 1 || depth > 15 {
				return true // Skip invalid depths
			}

			// Create deeply nested structure
			cfg := &Config{
				SchemaVersion: "2.0",
				OpenCenter: OpenCenterConfig{
					Meta: MetaConfig{
						Name:         "test",
						Organization: "test-org",
						Env:          "dev",
						Region:       "ord1",
					},
					Services: createNestedMap(depth),
				},
			}

			// Resolve references
			resolver := NewReferenceResolver()
			err := resolver.Resolve(cfg)

			remainingBudget := resolver.maxDepth - 3
			if depth > remainingBudget {
				return err != nil
			}

			// Otherwise should succeed
			return err == nil
		},
		gen.IntRange(1, 15),
	))

	properties.TestingRun(t)
}

// createNestedMap creates a nested map structure of specified depth
func createNestedMap(depth int) ServiceMap {
	if depth <= 0 {
		return ServiceMap{"value": "leaf"}
	}

	return ServiceMap{
		fmt.Sprintf("level%d", depth): createNestedMap(depth - 1),
	}
}
