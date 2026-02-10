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

package migration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewMigrationScanner tests scanner creation
func TestNewMigrationScanner(t *testing.T) {
	scanner := NewMigrationScanner("/test/path")
	if scanner == nil {
		t.Fatal("expected non-nil scanner")
	}
	if scanner.rootDir != "/test/path" {
		t.Errorf("expected rootDir /test/path, got %s", scanner.rootDir)
	}
}

// TestMigrationScanner_Scan tests scanning for legacy patterns
func TestMigrationScanner_Scan(t *testing.T) {
	// Create temporary directory structure
	tmpDir := t.TempDir()

	// Create test files with legacy patterns
	testFiles := map[string]string{
		"file_with_load.go": `package test
import (
	"context"
	"config"
)

func loadConfig() {
	ctx := context.Background()
	manager, err := config.NewConfigurationManager()
	if err != nil {
		return
	}
	cfg, err := manager.Load(ctx, "cluster-name")
	if err != nil {
		return
	}
}
`,
		"file_with_save.go": `package test
import (
	"context"
	"config"
)

func saveConfig() {
	ctx := context.Background()
	manager, err := config.NewConfigurationManager()
	if err != nil {
		return
	}
	err = manager.Save(ctx, cfg)
	if err != nil {
		return
	}
}
`,
		"file_with_validate.go": `package test
import (
	"context"
	"config"
)

func validateConfig() {
	ctx := context.Background()
	manager, err := config.NewConfigurationManager()
	if err != nil {
		return
	}
	err = manager.Validate(ctx, cfg)
	if err != nil {
		return
	}
}
`,
		"file_with_all.go": `package test
import (
	"context"
	"config"
)

func allOperations() {
	ctx := context.Background()
	manager, err := config.NewConfigurationManager()
	if err != nil {
		return
	}
	cfg, _ := manager.Load(ctx, "cluster")
	manager.Validate(ctx, cfg)
	manager.Save(ctx, cfg)
}
`,
		"file_without_legacy.go": `package test

func newFunction() {
	// No legacy calls here
}
`,
		"file_with_comment.go": `package test

func commented() {
	// config.Load("test") - this is commented
	/* config.Save(cfg) - also commented */
}
`,
	}

	for filename, content := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file %s: %v", filename, err)
		}
	}

	// Create scanner and run scan
	scanner := NewMigrationScanner(tmpDir)
	report, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Verify results
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	// Modern API patterns should NOT be detected as legacy
	// The scanner looks for config.Load(), config.Save(), config.Validate()
	// Modern code uses manager.Load(), manager.Save(), manager.Validate()
	if len(report.FilesUsingLegacyLoad) != 0 {
		t.Errorf("expected 0 files with legacy Load (modern API should not match), got %d: %v",
			len(report.FilesUsingLegacyLoad), report.FilesUsingLegacyLoad)
	}

	if len(report.FilesUsingLegacySave) != 0 {
		t.Errorf("expected 0 files with legacy Save (modern API should not match), got %d: %v",
			len(report.FilesUsingLegacySave), report.FilesUsingLegacySave)
	}

	if len(report.FilesUsingLegacyValidate) != 0 {
		t.Errorf("expected 0 files with legacy Validate (modern API should not match), got %d: %v",
			len(report.FilesUsingLegacyValidate), report.FilesUsingLegacyValidate)
	}

	// Total files should be 0 since we're using modern API
	if report.TotalFilesToMigrate != 0 {
		t.Errorf("expected 0 total files to migrate (modern API), got %d", report.TotalFilesToMigrate)
	}
}

// TestMigrationScanner_SkipsTestFiles tests that test files are skipped
func TestMigrationScanner_SkipsTestFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with legacy pattern
	testFile := filepath.Join(tmpDir, "config_test.go")
	content := `package config

func TestLoad(t *testing.T) {
	cfg, err := config.Load("test")
}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	scanner := NewMigrationScanner(tmpDir)
	report, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Test files should be skipped
	if len(report.FilesUsingLegacyLoad) != 0 {
		t.Errorf("expected 0 files (test files skipped), got %d", len(report.FilesUsingLegacyLoad))
	}
}

// TestMigrationScanner_SkipsVendor tests that vendor directory is skipped
func TestMigrationScanner_SkipsVendor(t *testing.T) {
	tmpDir := t.TempDir()

	// Create vendor directory with file
	vendorDir := filepath.Join(tmpDir, "vendor")
	if err := os.MkdirAll(vendorDir, 0755); err != nil {
		t.Fatalf("failed to create vendor dir: %v", err)
	}

	vendorFile := filepath.Join(vendorDir, "legacy.go")
	content := `package vendor
func load() {
	config.Load("test")
}
`
	if err := os.WriteFile(vendorFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create vendor file: %v", err)
	}

	scanner := NewMigrationScanner(tmpDir)
	report, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Vendor files should be skipped
	if len(report.FilesUsingLegacyLoad) != 0 {
		t.Errorf("expected 0 files (vendor skipped), got %d", len(report.FilesUsingLegacyLoad))
	}
}

// TestMigrationReport_GenerateReport tests report generation
func TestMigrationReport_GenerateReport(t *testing.T) {
	report := &MigrationReport{
		FilesUsingLegacyLoad:     []string{"file1.go", "file2.go"},
		FilesUsingLegacySave:     []string{"file2.go", "file3.go"},
		FilesUsingLegacyValidate: []string{"file3.go"},
		TotalFilesToMigrate:      3,
	}

	markdown := report.GenerateReport()

	// Verify report contains expected sections
	expectedSections := []string{
		"# Configuration Migration Report",
		"## Summary",
		"Total files to migrate",
		"Files using config.Load",
		"Files using config.Save",
		"Files using config.Validate",
		"## Files Using config.Load",
		"## Files Using config.Save",
		"## Files Using config.Validate",
		"## Migration Checklist",
		"## Migration Instructions",
	}

	for _, section := range expectedSections {
		if !strings.Contains(markdown, section) {
			t.Errorf("expected report to contain section: %s", section)
		}
	}

	// Verify file counts
	if !strings.Contains(markdown, "**Total files to migrate**: 3") {
		t.Error("expected total files count in report")
	}
	if !strings.Contains(markdown, "**Files using config.Load**: 2") {
		t.Error("expected Load files count in report")
	}
	if !strings.Contains(markdown, "**Files using config.Save**: 2") {
		t.Error("expected Save files count in report")
	}
	if !strings.Contains(markdown, "**Files using config.Validate**: 1") {
		t.Error("expected Validate files count in report")
	}

	// Verify files are listed
	if !strings.Contains(markdown, "file1.go") {
		t.Error("expected file1.go in report")
	}
	if !strings.Contains(markdown, "file2.go") {
		t.Error("expected file2.go in report")
	}
	if !strings.Contains(markdown, "file3.go") {
		t.Error("expected file3.go in report")
	}
}

// TestMigrationReport_GenerateReport_Empty tests report with no files
func TestMigrationReport_GenerateReport_Empty(t *testing.T) {
	report := &MigrationReport{
		FilesUsingLegacyLoad:     []string{},
		FilesUsingLegacySave:     []string{},
		FilesUsingLegacyValidate: []string{},
		TotalFilesToMigrate:      0,
	}

	markdown := report.GenerateReport()

	// Should still have structure
	if !strings.Contains(markdown, "# Configuration Migration Report") {
		t.Error("expected report header")
	}
	if !strings.Contains(markdown, "**Total files to migrate**: 0") {
		t.Error("expected zero total files")
	}
}

// TestMigrationScanner_RealCodebase tests scanning actual codebase
func TestMigrationScanner_RealCodebase(t *testing.T) {
	// Skip if not in actual codebase
	if _, err := os.Stat("../../../cmd"); err != nil {
		t.Skip("skipping real codebase test - not in project root")
	}

	scanner := NewMigrationScanner("../../..")
	report, err := scanner.Scan()
	if err != nil {
		t.Fatalf("scan failed: %v", err)
	}

	// Just verify it runs without error and produces a report
	if report == nil {
		t.Fatal("expected non-nil report")
	}

	t.Logf("Found %d files to migrate", report.TotalFilesToMigrate)
	t.Logf("Files with Load: %d", len(report.FilesUsingLegacyLoad))
	t.Logf("Files with Save: %d", len(report.FilesUsingLegacySave))
	t.Logf("Files with Validate: %d", len(report.FilesUsingLegacyValidate))
}
