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

package barbican

import (
	"fmt"
	"regexp"
	"strings"
)

var labelRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

// ParseLabels validates and parses a list of label strings in "key=value" format.
func ParseLabels(labels []string) (map[string]string, error) {
	result := make(map[string]string)
	for _, label := range labels {
		parts := strings.SplitN(label, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label format: %s (expected key=value)", label)
		}
		key, value := parts[0], parts[1]

		if len(key) > 255 {
			return nil, fmt.Errorf("label key too long: %s (max 255 chars)", key)
		}
		if !labelRegex.MatchString(key) {
			return nil, fmt.Errorf("invalid label key: %s (must contain only alphanumeric, '_', '.', '-')", key)
		}

		if len(value) > 255 {
			return nil, fmt.Errorf("label value too long: %s (max 255 chars)", value)
		}
		if !labelRegex.MatchString(value) {
			return nil, fmt.Errorf("invalid label value: %s (must contain only alphanumeric, '_', '.', '-')", value)
		}

		result[key] = value
	}
	return result, nil
}
