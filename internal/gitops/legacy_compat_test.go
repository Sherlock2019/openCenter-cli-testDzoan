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

package gitops

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rackerlabs/openCenter-cli/internal/config"
)

// TestGenerateGitOpsRepository tests the unified generation interface.
func TestGenerateGitOpsRepository(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Create test configuration
	cfg := config.NewDefault("test-cluster")
	cfg.OpenCenter.GitOps.GitDir = tempDir
	
	// Test generation
	ctx := context.Background()
	if err := GenerateGitOpsRepository(ctx, cfg); err != nil {
		t.Fatalf("GenerateGitOpsRepository failed: %v", err)
	}
	
	// Verify that base files were created
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Errorf("Expected .gitignore to be created at %s", gitignorePath)
	}
	
	// Verify base directory structure
	appsPath := filepath.Join(tempDir, "applications")
	if _, err := os.Stat(appsPath); os.IsNotExist(err) {
		t.Errorf("Expected applications directory to be created at %s", appsPath)
	}
	
	// Verify that cluster apps were rendered
	clusterAppsPath := filepath.Join(tempDir, "applications", "overlays", "test-cluster")
	if _, err := os.Stat(clusterAppsPath); os.IsNotExist(err) {
		t.Errorf("Expected cluster apps directory to be created at %s", clusterAppsPath)
	}
	
	// Verify that infrastructure was rendered
	infraPath := filepath.Join(tempDir, "infrastructure", "clusters", "test-cluster")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Errorf("Expected infrastructure directory to be created at %s", infraPath)
	}
}

// TestGenerateGitOpsRepositoryWithOptions tests generation with options.
func TestGenerateGitOpsRepositoryWithOptions(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Create test configuration
	cfg := config.NewDefault("test-cluster-opts")
	cfg.OpenCenter.GitOps.GitDir = tempDir
	
	// Test generation with options
	ctx := context.Background()
	opts := GenerationOptions{
		DryRun:  false,
		Verbose: true,
	}
	
	if err := GenerateGitOpsRepositoryWithOptions(ctx, cfg, opts); err != nil {
		t.Fatalf("GenerateGitOpsRepositoryWithOptions failed: %v", err)
	}
	
	// Verify that files were created (since DryRun is false)
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Errorf("Expected .gitignore to be created at %s", gitignorePath)
	}
}

// TestLegacyGenerationWrapper tests the deprecated wrapper interface.
func TestLegacyGenerationWrapper(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Create test configuration
	cfg := config.NewDefault("test-wrapper")
	cfg.OpenCenter.GitOps.GitDir = tempDir
	
	// Create wrapper
	wrapper := NewLegacyGenerationWrapper(cfg)
	
	// Test generation through wrapper
	if err := wrapper.Generate(); err != nil {
		t.Fatalf("LegacyGenerationWrapper.Generate failed: %v", err)
	}
	
	// Verify that files were created
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Errorf("Expected .gitignore to be created at %s", gitignorePath)
	}
}

// TestLegacyGenerationWrapperIndividualMethods tests individual wrapper methods.
func TestLegacyGenerationWrapperIndividualMethods(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()
	
	// Create test configuration
	cfg := config.NewDefault("test-individual")
	cfg.OpenCenter.GitOps.GitDir = tempDir
	
	// Create wrapper
	wrapper := NewLegacyGenerationWrapper(cfg)
	
	// Test CopyBase
	if err := wrapper.CopyBase(true); err != nil {
		t.Fatalf("LegacyGenerationWrapper.CopyBase failed: %v", err)
	}
	
	// Verify base files
	gitignorePath := filepath.Join(tempDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		t.Errorf("Expected .gitignore to be created at %s", gitignorePath)
	}
	
	// Test RenderClusterApps
	if err := wrapper.RenderClusterApps(); err != nil {
		t.Fatalf("LegacyGenerationWrapper.RenderClusterApps failed: %v", err)
	}
	
	// Verify cluster apps
	clusterAppsPath := filepath.Join(tempDir, "applications", "overlays", "test-individual")
	if _, err := os.Stat(clusterAppsPath); os.IsNotExist(err) {
		t.Errorf("Expected cluster apps directory to be created at %s", clusterAppsPath)
	}
	
	// Test RenderInfrastructureCluster
	if err := wrapper.RenderInfrastructureCluster(); err != nil {
		t.Fatalf("LegacyGenerationWrapper.RenderInfrastructureCluster failed: %v", err)
	}
	
	// Verify infrastructure
	infraPath := filepath.Join(tempDir, "infrastructure", "clusters", "test-individual")
	if _, err := os.Stat(infraPath); os.IsNotExist(err) {
		t.Errorf("Expected infrastructure directory to be created at %s", infraPath)
	}
}

// TestUsePipelineGenerator tests the feature flag detection.
func TestUsePipelineGenerator(t *testing.T) {
	// Save original value
	originalValue := os.Getenv(usePipelineGeneratorEnvVar)
	defer os.Setenv(usePipelineGeneratorEnvVar, originalValue)
	
	// Test default (should be false)
	os.Unsetenv(usePipelineGeneratorEnvVar)
	if usePipelineGenerator() {
		t.Error("Expected usePipelineGenerator() to return false by default")
	}
	
	// Test enabled
	os.Setenv(usePipelineGeneratorEnvVar, "true")
	if !usePipelineGenerator() {
		t.Error("Expected usePipelineGenerator() to return true when env var is 'true'")
	}
	
	// Test disabled
	os.Setenv(usePipelineGeneratorEnvVar, "false")
	if usePipelineGenerator() {
		t.Error("Expected usePipelineGenerator() to return false when env var is 'false'")
	}
	
	// Test other values (should be false)
	os.Setenv(usePipelineGeneratorEnvVar, "yes")
	if usePipelineGenerator() {
		t.Error("Expected usePipelineGenerator() to return false for non-'true' values")
	}
}

// TestGenerateGitOpsRepositoryBackwardCompatibility verifies that the new interface
// produces the same output as the legacy functions.
func TestGenerateGitOpsRepositoryBackwardCompatibility(t *testing.T) {
	// Create a single temporary directory to use for both generations
	// This ensures the GitDir path is identical in both configurations
	sharedDir := t.TempDir()
	
	// Create subdirectories for legacy and new outputs
	legacyDir := filepath.Join(sharedDir, "legacy")
	newDir := filepath.Join(sharedDir, "new")
	
	if err := os.MkdirAll(legacyDir, 0o755); err != nil {
		t.Fatalf("Failed to create legacy directory: %v", err)
	}
	if err := os.MkdirAll(newDir, 0o755); err != nil {
		t.Fatalf("Failed to create new directory: %v", err)
	}
	
	// Create identical configurations with the same cluster name
	// The key is that both configs must have identical values for all fields
	// that are used in template rendering
	legacyCfg := config.NewDefault("compat-test")
	legacyCfg.OpenCenter.GitOps.GitDir = legacyDir
	
	newCfg := config.NewDefault("compat-test")
	newCfg.OpenCenter.GitOps.GitDir = newDir
	
	// Generate using legacy method
	if err := CopyBase(legacyCfg, true); err != nil {
		t.Fatalf("Legacy CopyBase failed: %v", err)
	}
	if err := RenderClusterApps(legacyCfg); err != nil {
		t.Fatalf("Legacy RenderClusterApps failed: %v", err)
	}
	if err := RenderInfrastructureCluster(legacyCfg); err != nil {
		t.Fatalf("Legacy RenderInfrastructureCluster failed: %v", err)
	}
	
	// Generate using new method
	ctx := context.Background()
	if err := GenerateGitOpsRepository(ctx, newCfg); err != nil {
		t.Fatalf("New GenerateGitOpsRepository failed: %v", err)
	}
	
	// Compare directory structures recursively
	// We need to normalize paths in the comparison since the GitDir will be different
	if err := compareDirectoriesNormalized(t, legacyDir, newDir, legacyDir, newDir); err != nil {
		t.Fatalf("Directory comparison failed: %v", err)
	}
}

// compareDirectories recursively compares two directories to ensure they contain
// identical files with identical content.
func compareDirectories(t *testing.T, legacyDir, newDir string) error {
	// Walk through legacy directory and compare each file
	return filepath.Walk(legacyDir, func(legacyPath string, legacyInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path
		relPath, err := filepath.Rel(legacyDir, legacyPath)
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		
		// Construct corresponding path in new directory
		newPath := filepath.Join(newDir, relPath)
		
		// Check if path exists in new directory
		newInfo, err := os.Stat(newPath)
		if os.IsNotExist(err) {
			t.Errorf("File missing in new directory: %s", relPath)
			return nil // Continue checking other files
		}
		if err != nil {
			return err
		}
		
		// Compare file types (directory vs file)
		if legacyInfo.IsDir() != newInfo.IsDir() {
			t.Errorf("Type mismatch for %s: legacy is dir=%v, new is dir=%v", 
				relPath, legacyInfo.IsDir(), newInfo.IsDir())
			return nil
		}
		
		// If it's a directory, continue walking
		if legacyInfo.IsDir() {
			return nil
		}
		
		// Compare file sizes
		if legacyInfo.Size() != newInfo.Size() {
			t.Errorf("File size mismatch for %s: legacy=%d bytes, new=%d bytes", 
				relPath, legacyInfo.Size(), newInfo.Size())
			// Continue to compare content anyway
		}
		
		// Compare file content byte-by-byte
		legacyContent, err := os.ReadFile(legacyPath)
		if err != nil {
			return err
		}
		
		newContent, err := os.ReadFile(newPath)
		if err != nil {
			return err
		}
		
		if !bytesEqual(legacyContent, newContent) {
			t.Errorf("File content mismatch for %s", relPath)
			// Show first difference for debugging
			showFirstDifference(t, relPath, legacyContent, newContent)
		}
		
		return nil
	})
}

// compareDirectoriesNormalized recursively compares two directories while normalizing
// path references in the content. This is useful when comparing outputs that contain
// absolute paths that differ between test runs.
func compareDirectoriesNormalized(t *testing.T, legacyDir, newDir, legacyPathToNormalize, newPathToNormalize string) error {
	// Walk through legacy directory and compare each file
	return filepath.Walk(legacyDir, func(legacyPath string, legacyInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Get relative path
		relPath, err := filepath.Rel(legacyDir, legacyPath)
		if err != nil {
			return err
		}
		
		// Skip the root directory itself
		if relPath == "." {
			return nil
		}
		
		// Construct corresponding path in new directory
		newPath := filepath.Join(newDir, relPath)
		
		// Check if path exists in new directory
		newInfo, err := os.Stat(newPath)
		if os.IsNotExist(err) {
			t.Errorf("File missing in new directory: %s", relPath)
			return nil // Continue checking other files
		}
		if err != nil {
			return err
		}
		
		// Compare file types (directory vs file)
		if legacyInfo.IsDir() != newInfo.IsDir() {
			t.Errorf("Type mismatch for %s: legacy is dir=%v, new is dir=%v", 
				relPath, legacyInfo.IsDir(), newInfo.IsDir())
			return nil
		}
		
		// If it's a directory, continue walking
		if legacyInfo.IsDir() {
			return nil
		}
		
		// Read file contents
		legacyContent, err := os.ReadFile(legacyPath)
		if err != nil {
			return err
		}
		
		newContent, err := os.ReadFile(newPath)
		if err != nil {
			return err
		}
		
		// Normalize paths in content for comparison
		// Replace the actual paths with a placeholder
		legacyNormalized := strings.ReplaceAll(string(legacyContent), legacyPathToNormalize, "{{GITDIR}}")
		newNormalized := strings.ReplaceAll(string(newContent), newPathToNormalize, "{{GITDIR}}")
		
		if legacyNormalized != newNormalized {
			t.Errorf("File content mismatch for %s (after path normalization)", relPath)
			// Show first difference for debugging
			showFirstDifference(t, relPath, []byte(legacyNormalized), []byte(newNormalized))
		}
		
		return nil
	})
}

// bytesEqual compares two byte slices for equality.
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// showFirstDifference shows the first difference between two byte slices for debugging.
func showFirstDifference(t *testing.T, filename string, legacy, new []byte) {
	minLen := len(legacy)
	if len(new) < minLen {
		minLen = len(new)
	}
	
	for i := 0; i < minLen; i++ {
		if legacy[i] != new[i] {
			// Show context around the difference
			start := i - 20
			if start < 0 {
				start = 0
			}
			end := i + 20
			if end > minLen {
				end = minLen
			}
			
			t.Logf("First difference in %s at byte %d:", filename, i)
			t.Logf("  Legacy: %q", string(legacy[start:end]))
			t.Logf("  New:    %q", string(new[start:end]))
			return
		}
	}
	
	// If we get here, one file is a prefix of the other
	if len(legacy) != len(new) {
		t.Logf("File %s: length mismatch (legacy=%d, new=%d)", filename, len(legacy), len(new))
	}
}

// TestGenerationOptionsValidation tests the validation of generation options.
func TestGenerationOptionsValidation(t *testing.T) {
	tests := []struct {
		name    string
		opts    GenerationOptions
		wantErr bool
	}{
		{
			name: "default options",
			opts: DefaultGenerationOptions(),
			wantErr: false,
		},
		{
			name: "dry run enabled",
			opts: GenerationOptions{
				DryRun: true,
			},
			wantErr: false,
		},
		{
			name: "custom output dir",
			opts: GenerationOptions{
				OutputDir: "/tmp/custom",
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerationOptions.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
