/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"testing"
)

func TestShouldSkipDirectory(t *testing.T) {
	testCases := []struct {
		name     string
		dirPath  string
		expected bool
	}{
		// Directories that SHOULD be skipped
		{
			name:     "venv directory",
			dirPath:  "/path/to/project/venv",
			expected: true,
		},
		{
			name:     ".venv directory",
			dirPath:  "/path/to/project/.venv",
			expected: true,
		},
		{
			name:     ".terraform directory",
			dirPath:  "/infrastructure/clusters/test-cluster/.terraform",
			expected: true,
		},
		{
			name:     ".bin directory",
			dirPath:  "/infrastructure/clusters/test-cluster/.bin",
			expected: true,
		},
		{
			name:     "node_modules directory",
			dirPath:  "/path/to/project/node_modules",
			expected: true,
		},
		{
			name:     ".git directory",
			dirPath:  "/path/to/project/.git",
			expected: true,
		},
		{
			name:     "__pycache__ directory",
			dirPath:  "/path/to/project/__pycache__",
			expected: true,
		},
		{
			name:     ".pytest_cache directory",
			dirPath:  "/path/to/project/.pytest_cache",
			expected: true,
		},
		{
			name:     ".mypy_cache directory",
			dirPath:  "/path/to/project/.mypy_cache",
			expected: true,
		},
		{
			name:     ".tox directory",
			dirPath:  "/path/to/project/.tox",
			expected: true,
		},
		{
			name:     "vendor directory",
			dirPath:  "/path/to/project/vendor",
			expected: true,
		},
		{
			name:     "target directory",
			dirPath:  "/path/to/project/target",
			expected: true,
		},
		{
			name:     "build directory",
			dirPath:  "/path/to/project/build",
			expected: true,
		},
		{
			name:     "dist directory",
			dirPath:  "/path/to/project/dist",
			expected: true,
		},

		// Directories that SHOULD NOT be skipped
		{
			name:     "secrets directory",
			dirPath:  "/path/to/project/secrets",
			expected: false,
		},
		{
			name:     "infrastructure directory",
			dirPath:  "/infrastructure/clusters/test-cluster",
			expected: false,
		},
		{
			name:     "applications directory",
			dirPath:  "/applications/overlays/prod-cluster",
			expected: false,
		},
		{
			name:     "config directory",
			dirPath:  "/path/to/project/config",
			expected: false,
		},
		{
			name:     "terraform directory (not .terraform)",
			dirPath:  "/path/to/project/terraform",
			expected: false,
		},
		{
			name:     "bin directory (not .bin)",
			dirPath:  "/path/to/project/bin",
			expected: false,
		},
		{
			name:     "modules directory",
			dirPath:  "/infrastructure/clusters/test-cluster/modules",
			expected: false,
		},
		{
			name:     "environments directory",
			dirPath:  "/infrastructure/clusters/test-cluster/environments",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldSkipDirectory(tc.dirPath)
			if result != tc.expected {
				t.Errorf("shouldSkipDirectory(%q) = %v, expected %v", tc.dirPath, result, tc.expected)
			}
		})
	}
}

func TestShouldFileBeEncrypted(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
		expected bool
	}{
		// Files that SHOULD be encrypted (based on name patterns)
		{
			name:     "file with 'secret' in name",
			filePath: "/path/to/secret.yaml",
			expected: true,
		},
		{
			name:     "file with 'credential' in name",
			filePath: "/path/to/credentials.yaml",
			expected: true,
		},
		{
			name:     "file with 'password' in name",
			filePath: "/path/to/passwords.yaml",
			expected: true,
		},
		{
			name:     "file with 'token' in name",
			filePath: "/path/to/api-token.yaml",
			expected: true,
		},
		{
			name:     "file with 'key' in name",
			filePath: "/path/to/api-key.yaml",
			expected: true,
		},
		{
			name:     "file with 'cert' in name",
			filePath: "/path/to/cert.yaml",
			expected: true,
		},
		{
			name:     "file with 'tls' in name",
			filePath: "/path/to/tls-config.yaml",
			expected: true,
		},
		{
			name:     "file with 'auth' in name",
			filePath: "/path/to/auth-config.yaml",
			expected: true,
		},

		// Files that SHOULD NOT be encrypted (based on name patterns)
		{
			name:     "regular config file",
			filePath: "/path/to/config.yaml",
			expected: false,
		},
		{
			name:     "kustomization file",
			filePath: "/path/to/kustomization.yaml",
			expected: false,
		},
		{
			name:     "deployment file",
			filePath: "/path/to/deployment.yaml",
			expected: false,
		},
		{
			name:     "service file",
			filePath: "/path/to/service.yaml",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := shouldFileBeEncrypted(tc.filePath)
			if result != tc.expected {
				t.Errorf("shouldFileBeEncrypted(%q) = %v, expected %v", tc.filePath, result, tc.expected)
			}
		})
	}
}
