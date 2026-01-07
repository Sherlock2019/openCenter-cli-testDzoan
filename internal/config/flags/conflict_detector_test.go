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
	"strings"
	"testing"
)

func TestConflictDetector_DetectConflicts(t *testing.T) {
	detector := NewConflictDetector()

	tests := []struct {
		name            string
		configurations  []Configuration
		expectConflicts int
		checkFunc       func(*testing.T, []ConfigConflict)
	}{
		{
			name: "no conflicts - identical values",
			configurations: []Configuration{
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name": "test-cluster",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
				},
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name": "test-cluster",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "override.yaml", Priority: 2}},
				},
			},
			expectConflicts: 0,
		},
		{
			name: "conflict - different values",
			configurations: []Configuration{
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name": "base-cluster",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
				},
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name": "override-cluster",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "override.yaml", Priority: 2}},
				},
			},
			expectConflicts: 1,
			checkFunc: func(t *testing.T, conflicts []ConfigConflict) {
				if len(conflicts) != 1 {
					t.Errorf("Expected 1 conflict, got %d", len(conflicts))
					return
				}

				conflict := conflicts[0]
				if conflict.Path != "cluster.name" {
					t.Errorf("Expected conflict path 'cluster.name', got '%s'", conflict.Path)
				}

				if len(conflict.Sources) != 2 {
					t.Errorf("Expected 2 sources in conflict, got %d", len(conflict.Sources))
				}

				// Higher priority should win
				if conflict.ResolvedValue != "override-cluster" {
					t.Errorf("Expected resolved value 'override-cluster', got %v", conflict.ResolvedValue)
				}
			},
		},
		{
			name: "multiple conflicts",
			configurations: []Configuration{
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name":    "base-cluster",
							"version": "1.0.0",
						},
						"infrastructure": map[string]interface{}{
							"provider": "openstack",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
				},
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name":    "override-cluster",
							"version": "2.0.0",
						},
						"infrastructure": map[string]interface{}{
							"provider": "aws",
						},
					},
					Sources: []ConfigSource{{Type: SourceCLI, Path: "cli", Priority: 3}},
				},
			},
			expectConflicts: 3,
			checkFunc: func(t *testing.T, conflicts []ConfigConflict) {
				if len(conflicts) != 3 {
					t.Errorf("Expected 3 conflicts, got %d", len(conflicts))
					return
				}

				// Check that all expected paths have conflicts
				expectedPaths := map[string]bool{
					"cluster.name":            false,
					"cluster.version":         false,
					"infrastructure.provider": false,
				}

				for _, conflict := range conflicts {
					if _, exists := expectedPaths[conflict.Path]; exists {
						expectedPaths[conflict.Path] = true
					} else {
						t.Errorf("Unexpected conflict path: %s", conflict.Path)
					}
				}

				for path, found := range expectedPaths {
					if !found {
						t.Errorf("Expected conflict for path '%s' not found", path)
					}
				}
			},
		},
		{
			name: "no conflicts - different paths",
			configurations: []Configuration{
				{
					Data: map[string]interface{}{
						"cluster": map[string]interface{}{
							"name": "test-cluster",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
				},
				{
					Data: map[string]interface{}{
						"infrastructure": map[string]interface{}{
							"provider": "openstack",
						},
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "override.yaml", Priority: 2}},
				},
			},
			expectConflicts: 0,
		},
		{
			name: "priority resolution",
			configurations: []Configuration{
				{
					Data: map[string]interface{}{
						"value": "low-priority",
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
				},
				{
					Data: map[string]interface{}{
						"value": "medium-priority",
					},
					Sources: []ConfigSource{{Type: SourceFile, Path: "override.yaml", Priority: 5}},
				},
				{
					Data: map[string]interface{}{
						"value": "high-priority",
					},
					Sources: []ConfigSource{{Type: SourceCLI, Path: "cli", Priority: 10}},
				},
			},
			expectConflicts: 1,
			checkFunc: func(t *testing.T, conflicts []ConfigConflict) {
				if len(conflicts) != 1 {
					t.Errorf("Expected 1 conflict, got %d", len(conflicts))
					return
				}

				conflict := conflicts[0]
				if conflict.ResolvedValue != "high-priority" {
					t.Errorf("Expected resolved value 'high-priority', got %v", conflict.ResolvedValue)
				}

				if len(conflict.Sources) != 3 {
					t.Errorf("Expected 3 sources in conflict, got %d", len(conflict.Sources))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conflicts, err := detector.DetectConflicts(tt.configurations)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if len(conflicts) != tt.expectConflicts {
				t.Errorf("Expected %d conflicts, got %d", tt.expectConflicts, len(conflicts))
			}

			if tt.checkFunc != nil {
				tt.checkFunc(t, conflicts)
			}
		})
	}
}

func TestConflictDetector_GetConflictReport(t *testing.T) {
	detector := NewConflictDetector()

	// Test with no conflicts
	report := detector.GetConflictReport()
	if !strings.Contains(report, "No configuration conflicts detected") {
		t.Errorf("Expected no conflicts message, got: %s", report)
	}

	// Test with conflicts
	configurations := []Configuration{
		{
			Data: map[string]interface{}{
				"cluster": map[string]interface{}{
					"name": "base-cluster",
				},
			},
			Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
		},
		{
			Data: map[string]interface{}{
				"cluster": map[string]interface{}{
					"name": "override-cluster",
				},
			},
			Sources: []ConfigSource{{Type: SourceCLI, Path: "cli", Priority: 2}},
		},
	}

	conflicts, err := detector.DetectConflicts(configurations)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(conflicts) == 0 {
		t.Fatal("Expected conflicts to be detected")
	}

	report = detector.GetConflictReport()

	// Check that report contains expected information
	expectedStrings := []string{
		"Configuration conflicts detected",
		"cluster.name",
		"base.yaml",
		"cli",
		"base-cluster",
		"override-cluster",
		"Resolution:",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(report, expected) {
			t.Errorf("Expected report to contain '%s', but it didn't. Report: %s", expected, report)
		}
	}
}

func TestConflictDetector_HasConflicts(t *testing.T) {
	detector := NewConflictDetector()

	// Initially no conflicts
	if detector.HasConflicts() {
		t.Error("Expected no conflicts initially")
	}

	// After detecting conflicts
	configurations := []Configuration{
		{
			Data: map[string]interface{}{
				"value": "first",
			},
			Sources: []ConfigSource{{Type: SourceFile, Path: "file1.yaml", Priority: 1}},
		},
		{
			Data: map[string]interface{}{
				"value": "second",
			},
			Sources: []ConfigSource{{Type: SourceFile, Path: "file2.yaml", Priority: 2}},
		},
	}

	_, err := detector.DetectConflicts(configurations)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !detector.HasConflicts() {
		t.Error("Expected conflicts to be detected")
	}
}

func TestConflictDetector_GetConflicts(t *testing.T) {
	detector := NewConflictDetector()

	configurations := []Configuration{
		{
			Data: map[string]interface{}{
				"cluster": map[string]interface{}{
					"name": "cluster1",
				},
				"infrastructure": map[string]interface{}{
					"provider": "openstack",
				},
			},
			Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
		},
		{
			Data: map[string]interface{}{
				"cluster": map[string]interface{}{
					"name": "cluster2",
				},
				"infrastructure": map[string]interface{}{
					"provider": "aws",
				},
			},
			Sources: []ConfigSource{{Type: SourceCLI, Path: "cli", Priority: 2}},
		},
	}

	detectedConflicts, err := detector.DetectConflicts(configurations)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	retrievedConflicts := detector.GetConflicts()

	if len(detectedConflicts) != len(retrievedConflicts) {
		t.Errorf("Expected %d conflicts from GetConflicts(), got %d", len(detectedConflicts), len(retrievedConflicts))
	}

	// Verify conflicts are the same
	for i, detected := range detectedConflicts {
		if i >= len(retrievedConflicts) {
			break
		}

		retrieved := retrievedConflicts[i]
		if detected.Path != retrieved.Path {
			t.Errorf("Conflict %d: expected path '%s', got '%s'", i, detected.Path, retrieved.Path)
		}

		if len(detected.Sources) != len(retrieved.Sources) {
			t.Errorf("Conflict %d: expected %d sources, got %d", i, len(detected.Sources), len(retrieved.Sources))
		}
	}
}

func TestConflictDetector_ComplexNestedConflicts(t *testing.T) {
	detector := NewConflictDetector()

	configurations := []Configuration{
		{
			Data: map[string]interface{}{
				"opencenter": map[string]interface{}{
					"cluster": map[string]interface{}{
						"name": "base-cluster",
						"config": map[string]interface{}{
							"networking": map[string]interface{}{
								"dns_servers": []interface{}{"8.8.8.8"},
							},
						},
					},
					"infrastructure": map[string]interface{}{
						"provider": "openstack",
						"server_pools": []interface{}{
							map[string]interface{}{
								"name":  "control",
								"count": 3,
							},
						},
					},
				},
			},
			Sources: []ConfigSource{{Type: SourceFile, Path: "base.yaml", Priority: 1}},
		},
		{
			Data: map[string]interface{}{
				"opencenter": map[string]interface{}{
					"cluster": map[string]interface{}{
						"name": "override-cluster",
						"config": map[string]interface{}{
							"networking": map[string]interface{}{
								"dns_servers": []interface{}{"1.1.1.1"},
							},
						},
					},
					"infrastructure": map[string]interface{}{
						"provider": "aws",
						"server_pools": []interface{}{
							map[string]interface{}{
								"name":  "compute",
								"count": 5,
							},
						},
					},
				},
			},
			Sources: []ConfigSource{{Type: SourceCLI, Path: "cli", Priority: 2}},
		},
	}

	conflicts, err := detector.DetectConflicts(configurations)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should detect conflicts in nested paths
	expectedPaths := []string{
		"opencenter.cluster.name",
		"opencenter.cluster.config.networking.dns_servers",
		"opencenter.infrastructure.provider",
		"opencenter.infrastructure.server_pools",
	}

	if len(conflicts) != len(expectedPaths) {
		t.Errorf("Expected %d conflicts, got %d", len(expectedPaths), len(conflicts))
	}

	// Check that all expected paths are found
	foundPaths := make(map[string]bool)
	for _, conflict := range conflicts {
		foundPaths[conflict.Path] = true
	}

	for _, expectedPath := range expectedPaths {
		if !foundPaths[expectedPath] {
			t.Errorf("Expected conflict for path '%s' not found", expectedPath)
		}
	}
}
