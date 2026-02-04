package migration

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MigrationScanner identifies files using legacy config patterns
type MigrationScanner struct {
	rootDir string
}

// NewMigrationScanner creates a new scanner for the given root directory
func NewMigrationScanner(rootDir string) *MigrationScanner {
	return &MigrationScanner{
		rootDir: rootDir,
	}
}

// MigrationReport contains scan results
type MigrationReport struct {
	FilesUsingLegacyLoad     []string
	FilesUsingLegacySave     []string
	FilesUsingLegacyValidate []string
	TotalFilesToMigrate      int
}

// MakePathsRelative converts absolute paths to relative paths
func (mr *MigrationReport) MakePathsRelative(rootDir string) {
	for i, path := range mr.FilesUsingLegacyLoad {
		if rel, err := filepath.Rel(rootDir, path); err == nil {
			mr.FilesUsingLegacyLoad[i] = rel
		}
	}
	for i, path := range mr.FilesUsingLegacySave {
		if rel, err := filepath.Rel(rootDir, path); err == nil {
			mr.FilesUsingLegacySave[i] = rel
		}
	}
	for i, path := range mr.FilesUsingLegacyValidate {
		if rel, err := filepath.Rel(rootDir, path); err == nil {
			mr.FilesUsingLegacyValidate[i] = rel
		}
	}
}

// Scan finds all files using legacy config function calls
func (ms *MigrationScanner) Scan() (*MigrationReport, error) {
	report := &MigrationReport{
		FilesUsingLegacyLoad:     []string{},
		FilesUsingLegacySave:     []string{},
		FilesUsingLegacyValidate: []string{},
	}

	// Patterns to search for legacy config calls
	loadPattern := regexp.MustCompile(`\bconfig\.Load\s*\(`)
	savePattern := regexp.MustCompile(`\bconfig\.Save\s*\(`)
	validatePattern := regexp.MustCompile(`\bconfig\.Validate\s*\(`)

	// Track unique files for each pattern
	loadFiles := make(map[string]bool)
	saveFiles := make(map[string]bool)
	validateFiles := make(map[string]bool)
	allFiles := make(map[string]bool)

	// Walk the directory tree
	err := filepath.Walk(ms.rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip vendor, .git, and other non-source directories
			if info.Name() == "vendor" || info.Name() == ".git" || 
			   info.Name() == "node_modules" || info.Name() == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		// Only scan Go files
		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files for now (they'll be migrated with their source files)
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		// Read and scan the file
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("opening file %s: %w", path, err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		inComment := false
		inString := false

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			trimmed := strings.TrimSpace(line)

			// Skip empty lines
			if trimmed == "" {
				continue
			}

			// Track multi-line comments
			if strings.Contains(trimmed, "/*") {
				inComment = true
			}
			if strings.Contains(trimmed, "*/") {
				inComment = false
				continue
			}

			// Skip comments
			if inComment || strings.HasPrefix(trimmed, "//") {
				continue
			}

			// Simple string detection (not perfect but good enough)
			// Skip lines that are clearly in strings
			if strings.Count(line, `"`)%2 == 1 {
				inString = !inString
			}
			if inString {
				continue
			}

			// Check for legacy patterns
			if loadPattern.MatchString(line) {
				loadFiles[path] = true
				allFiles[path] = true
			}
			if savePattern.MatchString(line) {
				saveFiles[path] = true
				allFiles[path] = true
			}
			if validatePattern.MatchString(line) {
				validateFiles[path] = true
				allFiles[path] = true
			}
		}

		if err := scanner.Err(); err != nil {
			return fmt.Errorf("scanning file %s: %w", path, err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("walking directory tree: %w", err)
	}

	// Convert maps to sorted slices
	for file := range loadFiles {
		report.FilesUsingLegacyLoad = append(report.FilesUsingLegacyLoad, file)
	}
	for file := range saveFiles {
		report.FilesUsingLegacySave = append(report.FilesUsingLegacySave, file)
	}
	for file := range validateFiles {
		report.FilesUsingLegacyValidate = append(report.FilesUsingLegacyValidate, file)
	}

	report.TotalFilesToMigrate = len(allFiles)

	return report, nil
}

// GenerateReport creates a markdown report of migration status
func (mr *MigrationReport) GenerateReport() string {
	var sb strings.Builder

	sb.WriteString("# Configuration Migration Report\n\n")
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Total files to migrate**: %d\n", mr.TotalFilesToMigrate))
	sb.WriteString(fmt.Sprintf("- **Files using config.Load**: %d\n", len(mr.FilesUsingLegacyLoad)))
	sb.WriteString(fmt.Sprintf("- **Files using config.Save**: %d\n", len(mr.FilesUsingLegacySave)))
	sb.WriteString(fmt.Sprintf("- **Files using config.Validate**: %d\n\n", len(mr.FilesUsingLegacyValidate)))

	if len(mr.FilesUsingLegacyLoad) > 0 {
		sb.WriteString("## Files Using config.Load\n\n")
		for _, file := range mr.FilesUsingLegacyLoad {
			sb.WriteString(fmt.Sprintf("- [ ] `%s`\n", file))
		}
		sb.WriteString("\n")
	}

	if len(mr.FilesUsingLegacySave) > 0 {
		sb.WriteString("## Files Using config.Save\n\n")
		for _, file := range mr.FilesUsingLegacySave {
			sb.WriteString(fmt.Sprintf("- [ ] `%s`\n", file))
		}
		sb.WriteString("\n")
	}

	if len(mr.FilesUsingLegacyValidate) > 0 {
		sb.WriteString("## Files Using config.Validate\n\n")
		for _, file := range mr.FilesUsingLegacyValidate {
			sb.WriteString(fmt.Sprintf("- [ ] `%s`\n", file))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Migration Checklist\n\n")
	sb.WriteString("### Command Layer (cmd/)\n\n")
	sb.WriteString("- [ ] cmd/cluster_init.go\n")
	sb.WriteString("- [ ] cmd/cluster_validate.go\n")
	sb.WriteString("- [ ] cmd/cluster_setup.go\n")
	sb.WriteString("- [ ] cmd/cluster_bootstrap.go\n")
	sb.WriteString("- [ ] cmd/cluster_list.go\n")
	sb.WriteString("- [ ] cmd/config_*.go files\n\n")

	sb.WriteString("### Service Layer (internal/cluster/)\n\n")
	sb.WriteString("- [ ] internal/cluster/init_service.go\n")
	sb.WriteString("- [ ] internal/cluster/validate_service.go\n")
	sb.WriteString("- [ ] internal/cluster/setup_service.go\n")
	sb.WriteString("- [ ] internal/cluster/bootstrap_service.go\n\n")

	sb.WriteString("### GitOps Layer (internal/gitops/)\n\n")
	sb.WriteString("- [ ] internal/gitops/generator.go\n")
	sb.WriteString("- [ ] internal/gitops/workspace.go\n")
	sb.WriteString("- [ ] internal/gitops/pipeline.go\n\n")

	sb.WriteString("### SOPS Layer (internal/sops/)\n\n")
	sb.WriteString("- [ ] internal/sops/manager.go\n")
	sb.WriteString("- [ ] internal/sops/git.go\n\n")

	sb.WriteString("## Migration Instructions\n\n")
	sb.WriteString("### Replace config.Load\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Before\n")
	sb.WriteString("config, err := config.Load(clusterName)\n\n")
	sb.WriteString("// After\n")
	sb.WriteString("config, err := manager.Load(ctx, clusterName)\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Replace config.Save\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Before\n")
	sb.WriteString("err := config.Save(cfg)\n\n")
	sb.WriteString("// After\n")
	sb.WriteString("err := manager.Save(ctx, cfg)\n")
	sb.WriteString("```\n\n")

	sb.WriteString("### Replace config.Validate\n\n")
	sb.WriteString("```go\n")
	sb.WriteString("// Before\n")
	sb.WriteString("err := config.Validate(cfg)\n\n")
	sb.WriteString("// After\n")
	sb.WriteString("err := manager.Validate(ctx, cfg)\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Notes\n\n")
	sb.WriteString("- All operations now require a `context.Context` parameter\n")
	sb.WriteString("- ConfigurationManager must be injected via dependency injection\n")
	sb.WriteString("- Test files will be updated alongside their source files\n")
	sb.WriteString("- Run tests after each file migration to ensure correctness\n")

	return sb.String()
}
