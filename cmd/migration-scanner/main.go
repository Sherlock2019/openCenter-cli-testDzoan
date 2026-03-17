package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/migration"
)

func main() {
	// Get the workspace root (parent of cmd directory)
	workspaceRoot, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Create scanner
	scanner := migration.NewMigrationScanner(workspaceRoot)

	// Run scan
	fmt.Println("Scanning codebase for legacy config patterns...")
	report, err := scanner.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error scanning codebase: %v\n", err)
		os.Exit(1)
	}

	// Convert absolute paths to relative paths
	report.MakePathsRelative(workspaceRoot)

	// Generate markdown report
	markdown := report.GenerateReport()

	// Write report to file
	reportPath := filepath.Join(".kiro", "specs", "phase-3-configuration-unification", "migration-report.md")
	err = os.WriteFile(reportPath, []byte(markdown), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing report: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("\n✓ Migration report generated: %s\n\n", reportPath)
	fmt.Printf("Summary:\n")
	fmt.Printf("  - Total files to migrate: %d\n", report.TotalFilesToMigrate)
	fmt.Printf("  - Files using config.Load: %d\n", len(report.FilesUsingLegacyLoad))
	fmt.Printf("  - Files using config.Save: %d\n", len(report.FilesUsingLegacySave))
	fmt.Printf("  - Files using config.Validate: %d\n", len(report.FilesUsingLegacyValidate))
	fmt.Println()
}
