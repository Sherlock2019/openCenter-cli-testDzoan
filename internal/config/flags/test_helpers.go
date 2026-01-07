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
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
)

// Shared generators for property-based testing

func genConfigData() gopter.Gen {
	return gen.OneConstOf(
		map[string]interface{}{},
		map[string]interface{}{"key": "value"},
		map[string]interface{}{"cluster": map[string]interface{}{"name": "test"}},
		map[string]interface{}{"array": []interface{}{"item1", "item2"}},
		map[string]interface{}{
			"config": map[string]interface{}{
				"name":    "test-config",
				"enabled": true,
				"count":   3,
			},
		},
	)
}

// Shared helper functions for comparing values

func compareValues(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		return compareConfigValues(aVal, bVal)
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok {
			return false
		}
		return compareArrays(aVal, bVal)
	default:
		return a == b
	}
}

func compareConfigValues(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for key, aValue := range a {
		bValue, exists := b[key]
		if !exists {
			return false
		}
		if !compareValues(aValue, bValue) {
			return false
		}
	}

	return true
}

func compareArrays(a, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if !compareValues(a[i], b[i]) {
			return false
		}
	}

	return true
}
