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

package migration_test

import (
	"fmt"
	"log"

	"github.com/rackerlabs/opencenter-cli/internal/config/migration"
)

// ExampleMigrationScanner demonstrates how to use the migration scanner
func ExampleMigrationScanner() {
	// Create a scanner for the project root
	scanner := migration.NewMigrationScanner("../../..")

	// Scan for legacy patterns
	report, err := scanner.Scan()
	if err != nil {
		log.Fatalf("scan failed: %v", err)
	}

	// Generate markdown report
	markdown := report.GenerateReport()

	// Print summary
	fmt.Printf("Total files to migrate: %d\n", report.TotalFilesToMigrate)
	fmt.Printf("Files using config.Load: %d\n", len(report.FilesUsingLegacyLoad))
	fmt.Printf("Files using config.Save: %d\n", len(report.FilesUsingLegacySave))
	fmt.Printf("Files using config.Validate: %d\n", len(report.FilesUsingLegacyValidate))

	// Write report to file
	// os.WriteFile("migration-report.md", []byte(markdown), 0644)

	_ = markdown // Use the markdown report
}

// ExampleMigrationReport_GenerateReport demonstrates report generation
func ExampleMigrationReport_GenerateReport() {
	report := &migration.MigrationReport{
		FilesUsingLegacyLoad:     []string{"cmd/cluster_init.go", "cmd/cluster_validate.go"},
		FilesUsingLegacySave:     []string{"cmd/cluster_init.go"},
		FilesUsingLegacyValidate: []string{"cmd/cluster_validate.go"},
		TotalFilesToMigrate:      2,
	}

	markdown := report.GenerateReport()
	fmt.Println(markdown)
}
