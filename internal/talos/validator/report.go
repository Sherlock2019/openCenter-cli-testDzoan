package validator

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/rackerlabs/opencenter-cli/internal/talos"
)

// FormatReportJSON generates a machine-readable JSON report.
func FormatReportJSON(report *talos.ValidationReport) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal validation report: %w", err)
	}
	return string(data), nil
}

// FormatReportHuman generates a human-readable summary.
func FormatReportHuman(report *talos.ValidationReport) string {
	var sb strings.Builder

	// Header
	sb.WriteString("=================================================\n")
	sb.WriteString("  Talos OpenStack Environment Validation Report\n")
	sb.WriteString("=================================================\n\n")

	// Timestamp
	sb.WriteString(fmt.Sprintf("Validation Time: %s\n\n", report.Timestamp.Format("2006-01-02 15:04:05 MST")))

	// Overall status
	if report.Passed {
		sb.WriteString("✓ Overall Status: PASSED\n\n")
	} else {
		sb.WriteString("✗ Overall Status: FAILED\n\n")
	}

	// Validation checks
	sb.WriteString("Validation Checks:\n")
	sb.WriteString("-------------------------------------------------\n")

	passedCount := 0
	failedCount := 0

	for _, check := range report.Checks {
		if check.Passed {
			passedCount++
			sb.WriteString(fmt.Sprintf("  ✓ %s\n", check.Name))
			if check.Message != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", check.Message))
			}
		} else {
			failedCount++
			sb.WriteString(fmt.Sprintf("  ✗ %s [%s]\n", check.Name, strings.ToUpper(check.Severity)))
			if check.Message != "" {
				sb.WriteString(fmt.Sprintf("    %s\n", check.Message))
			}
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString("-------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("Summary: %d passed, %d failed\n\n", passedCount, failedCount))

	// Remediations
	if len(report.Remediations) > 0 {
		sb.WriteString("Remediation Actions:\n")
		sb.WriteString("=================================================\n\n")

		for i, remediation := range report.Remediations {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, remediation.Check))
			sb.WriteString(fmt.Sprintf("   %s\n\n", remediation.Description))

			if len(remediation.Steps) > 0 {
				sb.WriteString("   Steps to resolve:\n")
				for j, step := range remediation.Steps {
					// Indent multi-line steps properly
					lines := strings.Split(step, "\n")
					for k, line := range lines {
						if k == 0 {
							sb.WriteString(fmt.Sprintf("   %d. %s\n", j+1, line))
						} else {
							sb.WriteString(fmt.Sprintf("      %s\n", line))
						}
					}
				}
				sb.WriteString("\n")
			}
		}
	}

	// Footer
	if report.Passed {
		sb.WriteString("=================================================\n")
		sb.WriteString("All prerequisites met. You can proceed with\n")
		sb.WriteString("cluster initialization.\n")
		sb.WriteString("=================================================\n")
	} else {
		sb.WriteString("=================================================\n")
		sb.WriteString("Please address the failed checks before\n")
		sb.WriteString("proceeding with cluster deployment.\n")
		sb.WriteString("=================================================\n")
	}

	return sb.String()
}

// FormatReportCompact generates a compact summary suitable for CI/CD.
func FormatReportCompact(report *talos.ValidationReport) string {
	var sb strings.Builder

	if report.Passed {
		sb.WriteString("VALIDATION: PASSED\n")
	} else {
		sb.WriteString("VALIDATION: FAILED\n")
	}

	passedCount := 0
	failedCount := 0

	for _, check := range report.Checks {
		if check.Passed {
			passedCount++
		} else {
			failedCount++
			sb.WriteString(fmt.Sprintf("  FAILED: %s - %s\n", check.Name, check.Message))
		}
	}

	sb.WriteString(fmt.Sprintf("SUMMARY: %d passed, %d failed\n", passedCount, failedCount))

	return sb.String()
}

// GetFailedChecks returns a list of failed validation checks.
func GetFailedChecks(report *talos.ValidationReport) []talos.ValidationCheck {
	failed := []talos.ValidationCheck{}
	for _, check := range report.Checks {
		if !check.Passed {
			failed = append(failed, check)
		}
	}
	return failed
}

// GetPassedChecks returns a list of passed validation checks.
func GetPassedChecks(report *talos.ValidationReport) []talos.ValidationCheck {
	passed := []talos.ValidationCheck{}
	for _, check := range report.Checks {
		if check.Passed {
			passed = append(passed, check)
		}
	}
	return passed
}

// HasCriticalFailures checks if any critical failures exist.
func HasCriticalFailures(report *talos.ValidationReport) bool {
	for _, check := range report.Checks {
		if !check.Passed && check.Severity == "error" {
			return true
		}
	}
	return false
}

// HasWarnings checks if any warnings exist.
func HasWarnings(report *talos.ValidationReport) bool {
	for _, check := range report.Checks {
		if !check.Passed && check.Severity == "warning" {
			return true
		}
	}
	return false
}
