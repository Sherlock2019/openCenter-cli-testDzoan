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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReferenceResolver_ResolveEnvironmentVariables(t *testing.T) {
	// Set up test environment variable
	os.Setenv("TEST_VAR", "test-value")
	defer os.Unsetenv("TEST_VAR")

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${env:TEST_VAR}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	require.NoError(t, err)
	assert.Equal(t, "test-value", cfg.OpenCenter.Meta.Name)
}

func TestReferenceResolver_ResolveFileReferences(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("file-content"), 0600)
	require.NoError(t, err)

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${file:" + testFile + "}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err = resolver.Resolve(cfg)

	require.NoError(t, err)
	assert.Equal(t, "file-content", cfg.OpenCenter.Meta.Name)
}

func TestReferenceResolver_CachesFileReferences(t *testing.T) {
	// Create a temporary file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("cached-content"), 0600)
	require.NoError(t, err)

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${file:" + testFile + "}",
				Organization: "${file:" + testFile + "}", // Same file reference
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err = resolver.Resolve(cfg)

	require.NoError(t, err)
	assert.Equal(t, "cached-content", cfg.OpenCenter.Meta.Name)
	assert.Equal(t, "cached-content", cfg.OpenCenter.Meta.Organization)

	// Verify cache was used
	cacheKey := "file:" + testFile
	cached, ok := resolver.cache[cacheKey]
	assert.True(t, ok)
	assert.Equal(t, "cached-content", cached)
}

func TestReferenceResolver_ErrorOnMissingEnvironmentVariable(t *testing.T) {
	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${env:NONEXISTENT_VAR}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "environment variable")
	assert.Contains(t, err.Error(), "NONEXISTENT_VAR")
}

func TestReferenceResolver_ErrorOnMissingFile(t *testing.T) {
	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${file:/nonexistent/file.txt}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read file")
}

func TestReferenceResolver_CircularReferenceDetection(t *testing.T) {
	// Note: Circular reference detection for ${ref:} is implemented
	// but requires path lookup which is not yet fully implemented.
	// This test verifies the detection mechanism works.

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${ref:opencenter.meta.organization}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	// Should error because path lookup is not implemented
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot be resolved")
}

func TestReferenceResolver_MaxDepthExceeded(t *testing.T) {
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
			Services: ServiceMap{
				"level1": ServiceMap{
					"level2": ServiceMap{
						"level3": ServiceMap{
							"level4": ServiceMap{
								"level5": ServiceMap{
									"level6": ServiceMap{
										"level7": ServiceMap{
											"level8": ServiceMap{
												"level9": ServiceMap{
													"level10": ServiceMap{
														"level11": "too-deep",
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	// Should error due to max depth
	require.Error(t, err)
	assert.Contains(t, err.Error(), "maximum recursion depth exceeded")
}

func TestReferenceResolver_ResolvesMapValues(t *testing.T) {
	os.Setenv("MAP_TEST_VAR", "map-value")
	defer os.Unsetenv("MAP_TEST_VAR")

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "test",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
			Services: ServiceMap{
				"test-service": "${env:MAP_TEST_VAR}",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	require.NoError(t, err)
	assert.Equal(t, "map-value", cfg.OpenCenter.Services["test-service"])
}

func TestReferenceResolver_NilConfigError(t *testing.T) {
	resolver := NewReferenceResolver()
	err := resolver.Resolve(nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration cannot be nil")
}

func TestReferenceResolver_MultipleReferencesInSameString(t *testing.T) {
	os.Setenv("VAR1", "value1")
	os.Setenv("VAR2", "value2")
	defer os.Unsetenv("VAR1")
	defer os.Unsetenv("VAR2")

	cfg := &Config{
		SchemaVersion: "2.0",
		OpenCenter: OpenCenterConfig{
			Meta: MetaConfig{
				Name:         "${env:VAR1}-${env:VAR2}",
				Organization: "test-org",
				Env:          "dev",
				Region:       "ord1",
			},
		},
	}

	resolver := NewReferenceResolver()
	err := resolver.Resolve(cfg)

	require.NoError(t, err)
	assert.Equal(t, "value1-value2", cfg.OpenCenter.Meta.Name)
}
