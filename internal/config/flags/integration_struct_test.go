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

package flags

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test structs for struct application testing
type TestStructConfig struct {
	Name        string
	Count       int
	Enabled     bool
	Tags        []string
	Metadata    map[string]string
	Nested      *TestNestedStructConfig
	NestedValue TestNestedStructConfig
}

type TestNestedStructConfig struct {
	Host string
	Port int
}

func TestApplyToStruct_StringField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	err = integration.applyToStruct(config, "name", "test-cluster")
	require.NoError(t, err)

	assert.Equal(t, "test-cluster", config.Name)
}

func TestApplyToStruct_IntField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test with int value
	err = integration.applyToStruct(config, "count", 42)
	require.NoError(t, err)
	assert.Equal(t, 42, config.Count)

	// Test with string value that can be converted to int
	err = integration.applyToStruct(config, "count", "100")
	require.NoError(t, err)
	assert.Equal(t, 100, config.Count)
}

func TestApplyToStruct_BoolField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test with bool value
	err = integration.applyToStruct(config, "enabled", true)
	require.NoError(t, err)
	assert.True(t, config.Enabled)

	// Test with string value that can be converted to bool
	err = integration.applyToStruct(config, "enabled", "false")
	require.NoError(t, err)
	assert.False(t, config.Enabled)
}

func TestApplyToStruct_SliceField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test with slice value
	tags := []string{"prod", "us-east"}
	err = integration.applyToStruct(config, "tags", tags)
	require.NoError(t, err)
	assert.Equal(t, []string{"prod", "us-east"}, config.Tags)

	// Test with interface{} slice
	tagsInterface := []interface{}{"dev", "us-west"}
	err = integration.applyToStruct(config, "tags", tagsInterface)
	require.NoError(t, err)
	assert.Equal(t, []string{"dev", "us-west"}, config.Tags)
}

func TestApplyToStruct_MapField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test with map value
	metadata := map[string]string{
		"env":    "production",
		"region": "us-east-1",
	}
	err = integration.applyToStruct(config, "metadata", metadata)
	require.NoError(t, err)
	assert.Equal(t, metadata, config.Metadata)

	// Test with interface{} map
	metadataInterface := map[string]interface{}{
		"team": "platform",
		"cost": "high",
	}
	err = integration.applyToStruct(config, "metadata", metadataInterface)
	require.NoError(t, err)
	assert.Equal(t, map[string]string{"team": "platform", "cost": "high"}, config.Metadata)
}

func TestApplyToStruct_NestedPointerField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test setting nested field through pointer (should initialize nil pointer)
	err = integration.applyToStruct(config, "nested.host", "localhost")
	require.NoError(t, err)
	require.NotNil(t, config.Nested)
	assert.Equal(t, "localhost", config.Nested.Host)

	// Test setting another nested field
	err = integration.applyToStruct(config, "nested.port", 8080)
	require.NoError(t, err)
	assert.Equal(t, 8080, config.Nested.Port)
}

func TestApplyToStruct_NestedValueField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}
	
	// Test setting nested field through value struct (use underscore to match camelCase conversion)
	err = integration.applyToStruct(config, "nested_value.host", "example.com")
	require.NoError(t, err)
	assert.Equal(t, "example.com", config.NestedValue.Host)

	// Test setting another nested field
	err = integration.applyToStruct(config, "nested_value.port", 443)
	require.NoError(t, err)
	assert.Equal(t, 443, config.NestedValue.Port)
}

func TestFlagNameToFieldPath(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	tests := []struct {
		name     string
		flagName string
		expected string
	}{
		{
			name:     "simple field",
			flagName: "name",
			expected: "Name",
		},
		{
			name:     "nested field",
			flagName: "nested.host",
			expected: "Nested.Host",
		},
		{
			name:     "deeply nested field",
			flagName: "infrastructure.cluster.name",
			expected: "Infrastructure.Cluster.Name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := integration.flagNameToFieldPath(tt.flagName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNavigateToField(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{
		Name: "test",
		Nested: &TestNestedStructConfig{
			Host: "localhost",
			Port: 8080,
		},
	}

	tests := []struct {
		name      string
		path      string
		expectErr bool
	}{
		{
			name:      "simple field",
			path:      "Name",
			expectErr: false,
		},
		{
			name:      "nested field",
			path:      "Nested.Host",
			expectErr: false,
		},
		{
			name:      "non-existent field",
			path:      "NonExistent",
			expectErr: true,
		},
		{
			name:      "invalid nested path",
			path:      "Name.Invalid",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			field, err := integration.navigateToField(config, tt.path)
			if tt.expectErr {
				assert.Error(t, err)
				assert.False(t, field.IsValid())
			} else {
				assert.NoError(t, err)
				assert.True(t, field.IsValid())
			}
		})
	}
}

func TestSetFieldValueTyped_TypeConversions(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	tests := []struct {
		name      string
		setup     func() *TestStructConfig
		field     string
		value     interface{}
		verify    func(*testing.T, *TestStructConfig)
		expectErr bool
	}{
		{
			name:  "string to string",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Name",
			value: "test",
			verify: func(t *testing.T, c *TestStructConfig) {
				assert.Equal(t, "test", c.Name)
			},
		},
		{
			name:  "int to string",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Name",
			value: 42,
			verify: func(t *testing.T, c *TestStructConfig) {
				assert.Equal(t, "42", c.Name)
			},
		},
		{
			name:  "string to int",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Count",
			value: "100",
			verify: func(t *testing.T, c *TestStructConfig) {
				assert.Equal(t, 100, c.Count)
			},
		},
		{
			name:  "invalid string to int",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Count",
			value: "not-a-number",
			verify: func(t *testing.T, c *TestStructConfig) {
				// Should not be called
			},
			expectErr: true,
		},
		{
			name:  "string to bool",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Enabled",
			value: "true",
			verify: func(t *testing.T, c *TestStructConfig) {
				assert.True(t, c.Enabled)
			},
		},
		{
			name:  "int to bool",
			setup: func() *TestStructConfig { return &TestStructConfig{} },
			field: "Enabled",
			value: 1,
			verify: func(t *testing.T, c *TestStructConfig) {
				assert.True(t, c.Enabled)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := tt.setup()
			field, err := integration.navigateToField(config, tt.field)
			require.NoError(t, err)

			err = integration.setFieldValueTyped(field, tt.value)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				tt.verify(t, config)
			}
		})
	}
}

func TestApplyArrayOperationToStruct(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	t.Run("append operation", func(t *testing.T) {
		config := &TestStructConfig{
			Tags: []string{"existing"},
		}

		arrayOp := &ArrayOperationFlag{
			Path:      "tags",
			Operation: "append",
			Value:     "new-tag",
		}

		err := integration.applyArrayOperationToStruct(config, arrayOp)
		require.NoError(t, err)
		assert.Equal(t, []string{"existing", "new-tag"}, config.Tags)
	})

	t.Run("insert operation", func(t *testing.T) {
		config := &TestStructConfig{
			Tags: []string{"first", "third"},
		}

		arrayOp := &ArrayOperationFlag{
			Path:      "tags",
			Operation: "insert",
			Index:     1,
			Value:     "second",
		}

		err := integration.applyArrayOperationToStruct(config, arrayOp)
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "second", "third"}, config.Tags)
	})

	t.Run("remove operation", func(t *testing.T) {
		config := &TestStructConfig{
			Tags: []string{"first", "second", "third"},
		}

		arrayOp := &ArrayOperationFlag{
			Path:      "tags",
			Operation: "remove",
			Index:     1,
		}

		err := integration.applyArrayOperationToStruct(config, arrayOp)
		require.NoError(t, err)
		assert.Equal(t, []string{"first", "third"}, config.Tags)
	})
}

func TestApplyMapOperationToStruct(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	t.Run("set operation", func(t *testing.T) {
		config := &TestStructConfig{
			Metadata: make(map[string]string),
		}

		mapOp := &MapFlag{
			Path:      "metadata",
			Operation: "set",
			Key:       "env",
			Value:     "production",
		}

		err := integration.applyMapOperationToStruct(config, mapOp)
		require.NoError(t, err)
		assert.Equal(t, "production", config.Metadata["env"])
	})

	t.Run("merge operation", func(t *testing.T) {
		config := &TestStructConfig{
			Metadata: map[string]string{
				"existing": "value",
			},
		}

		mapOp := &MapFlag{
			Path:      "metadata",
			Operation: "merge",
			Value: map[string]interface{}{
				"new1": "value1",
				"new2": "value2",
			},
		}

		err := integration.applyMapOperationToStruct(config, mapOp)
		require.NoError(t, err)
		assert.Equal(t, "value", config.Metadata["existing"])
		assert.Equal(t, "value1", config.Metadata["new1"])
		assert.Equal(t, "value2", config.Metadata["new2"])
	})

	t.Run("remove operation", func(t *testing.T) {
		config := &TestStructConfig{
			Metadata: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		}

		mapOp := &MapFlag{
			Path:      "metadata",
			Operation: "remove",
			Key:       "key1",
		}

		err := integration.applyMapOperationToStruct(config, mapOp)
		require.NoError(t, err)
		_, exists := config.Metadata["key1"]
		assert.False(t, exists)
		assert.Equal(t, "value2", config.Metadata["key2"])
	})
}

func TestApplyToStruct_NilStruct(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	// Should not error when struct is nil
	err = integration.applyToStruct(nil, "name", "test")
	assert.NoError(t, err)
}

func TestApplyToStruct_Integration(t *testing.T) {
	integration, err := NewCLIIntegration()
	require.NoError(t, err)

	config := &TestStructConfig{}

	// Directly test applyToStruct with various values
	err = integration.applyToStruct(config, "name", "test-cluster")
	require.NoError(t, err)
	
	err = integration.applyToStruct(config, "count", 42)
	require.NoError(t, err)
	
	err = integration.applyToStruct(config, "enabled", true)
	require.NoError(t, err)

	// Verify struct was updated
	assert.Equal(t, "test-cluster", config.Name)
	assert.Equal(t, 42, config.Count)
	assert.True(t, config.Enabled)
}
