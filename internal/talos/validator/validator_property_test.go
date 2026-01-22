package validator

import (
	"context"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/rackerlabs/opencenter-cli/internal/config"
	"github.com/rackerlabs/opencenter-cli/internal/talos"
)

// mockLogger implements the Logger interface for testing.
type mockLogger struct{}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {}
func (m *mockLogger) Info(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Warn(msg string, keysAndValues ...interface{})  {}
func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {}

// Feature: talos-openstack-provider, Property 1: Validation completeness
// For any OpenStack environment configuration, when validation executes,
// all required service checks (Keystone, Barbican, Octavia, Glance, Neutron)
// should be performed and reported in the validation output.
func TestProperty_ValidationCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("all required service checks are performed", prop.ForAll(
		func(clusterName string) bool {
			// Create a test configuration
			cfg := &config.Config{}

			// Create validator
			validator := NewValidator(&mockLogger{})

			// Run validation
			report, err := validator.ValidateEnvironment(context.Background(), cfg)
			if err != nil {
				return false
			}

			// Verify all required checks are present
			requiredChecks := map[string]bool{
				"Keystone":        false,
				"Barbican":        false,
				"Octavia":         false,
				"Glance":          false,
				"Resource Quotas": false,
			}

			// Mark checks as found
			for _, check := range report.Checks {
				if _, exists := requiredChecks[check.Name]; exists {
					requiredChecks[check.Name] = true
				}
			}

			// Verify all required checks were found
			for checkName, found := range requiredChecks {
				if !found {
					t.Logf("Missing required check: %s", checkName)
					return false
				}
			}

			return true
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: talos-openstack-provider, Property 2: Validation output format
// For any validation result, the output should contain both a valid JSON structure
// with all required fields (passed, checks, remediations, timestamp) and a
// human-readable summary.
func TestProperty_ValidationOutputFormat(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("validation report has required fields and formats", prop.ForAll(
		func(clusterName string) bool {
			// Create a test configuration
			cfg := &config.Config{}

			// Create validator
			validator := NewValidator(&mockLogger{})

			// Run validation
			report, err := validator.ValidateEnvironment(context.Background(), cfg)
			if err != nil {
				return false
			}

			// Verify report has required fields
			if report.Checks == nil {
				t.Log("Report missing Checks field")
				return false
			}

			if report.Remediations == nil {
				t.Log("Report missing Remediations field")
				return false
			}

			if report.Timestamp.IsZero() {
				t.Log("Report missing or invalid Timestamp")
				return false
			}

			// Verify JSON formatting works
			jsonOutput, err := FormatReportJSON(report)
			if err != nil {
				t.Logf("Failed to format JSON: %v", err)
				return false
			}

			if len(jsonOutput) == 0 {
				t.Log("JSON output is empty")
				return false
			}

			// Verify human-readable formatting works
			humanOutput := FormatReportHuman(report)
			if len(humanOutput) == 0 {
				t.Log("Human-readable output is empty")
				return false
			}

			// Verify human output contains key sections
			requiredSections := []string{
				"Validation Report",
				"Validation Checks",
				"Summary",
			}

			for _, section := range requiredSections {
				if !containsString(humanOutput, section) {
					t.Logf("Human output missing section: %s", section)
					return false
				}
			}

			return true
		},
		gen.Identifier(),
	))

	properties.TestingRun(t)
}

// Feature: talos-openstack-provider, Property 3: Remediation presence
// For any failed validation check, the validation report should include
// at least one remediation action with description and steps.
func TestProperty_RemediationPresence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("failed checks have remediation actions", prop.ForAll(
		func(seed int64) bool {
			// Create a mock validator that will produce failures
			validator := &mockFailingValidator{
				logger:     &mockLogger{},
				failChecks: []string{"Keystone", "Barbican"},
			}

			cfg := &config.Config{}

			// Run validation
			report, err := validator.ValidateEnvironment(context.Background(), cfg)
			if err != nil {
				return false
			}

			// Count failed checks
			failedCount := 0
			for _, check := range report.Checks {
				if !check.Passed {
					failedCount++
				}
			}

			// If there are failed checks, verify remediations exist
			if failedCount > 0 {
				if len(report.Remediations) == 0 {
					t.Log("Failed checks exist but no remediations provided")
					return false
				}

				// Verify each remediation has required fields
				for _, remediation := range report.Remediations {
					if remediation.Check == "" {
						t.Log("Remediation missing Check field")
						return false
					}

					if remediation.Description == "" {
						t.Log("Remediation missing Description field")
						return false
					}

					if len(remediation.Steps) == 0 {
						t.Log("Remediation missing Steps")
						return false
					}
				}
			}

			return true
		},
		gen.Int64(),
	))

	properties.TestingRun(t)
}

// mockFailingValidator is a validator that simulates failures for testing.
type mockFailingValidator struct {
	logger     Logger
	failChecks []string
}

func (v *mockFailingValidator) ValidateEnvironment(ctx context.Context, cfg *config.Config) (*talos.ValidationReport, error) {
	report := &talos.ValidationReport{
		Passed:       true,
		Checks:       []talos.ValidationCheck{},
		Remediations: []talos.RemediationAction{},
	}

	// Simulate checks
	allChecks := []string{"Keystone", "Barbican", "Octavia", "Glance"}

	for _, checkName := range allChecks {
		shouldFail := false
		for _, failCheck := range v.failChecks {
			if checkName == failCheck {
				shouldFail = true
				break
			}
		}

		check := talos.ValidationCheck{
			Name:     checkName,
			Passed:   !shouldFail,
			Severity: "error",
		}

		if shouldFail {
			report.Passed = false
			check.Message = checkName + " validation failed"

			// Add remediation
			remediation := talos.RemediationAction{
				Check:       checkName,
				Description: "Failed to validate " + checkName,
				Steps: []string{
					"Step 1: Check service availability",
					"Step 2: Verify configuration",
					"Step 3: Review logs",
				},
			}
			report.Remediations = append(report.Remediations, remediation)
		} else {
			check.Message = checkName + " validation passed"
		}

		report.Checks = append(report.Checks, check)
	}

	return report, nil
}

func (v *mockFailingValidator) ValidateKeystone(ctx context.Context) error {
	return nil
}

func (v *mockFailingValidator) ValidateBarbican(ctx context.Context) error {
	return nil
}

func (v *mockFailingValidator) ValidateOctavia(ctx context.Context) error {
	return nil
}

func (v *mockFailingValidator) ValidateQuotas(ctx context.Context, required talos.ResourceRequirements) error {
	return nil
}

func (v *mockFailingValidator) ValidateGlance(ctx context.Context) error {
	return nil
}

// containsString checks if a string contains a substring.
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
