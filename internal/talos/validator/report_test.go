package validator

import (
	"strings"
	"testing"
	"time"

	"github.com/rackerlabs/opencenter-cli/internal/talos"
)

func TestFormatReportJSON(t *testing.T) {
	report := &talos.ValidationReport{
		Passed: true,
		Checks: []talos.ValidationCheck{
			{Name: "Test Check", Passed: true, Message: "OK", Severity: "info"},
		},
		Remediations: []talos.RemediationAction{},
		Timestamp:    time.Now(),
	}

	json, err := FormatReportJSON(report)
	if err != nil {
		t.Fatalf("FormatReportJSON returned error: %v", err)
	}

	if len(json) == 0 {
		t.Error("FormatReportJSON returned empty string")
	}

	// Verify JSON contains expected fields
	expectedFields := []string{
		`"passed"`,
		`"checks"`,
		`"remediations"`,
		`"timestamp"`,
	}

	for _, field := range expectedFields {
		if !strings.Contains(json, field) {
			t.Errorf("JSON output missing field: %s", field)
		}
	}
}

func TestFormatReportHuman(t *testing.T) {
	report := &talos.ValidationReport{
		Passed: true,
		Checks: []talos.ValidationCheck{
			{Name: "Keystone", Passed: true, Message: "OK", Severity: "info"},
			{Name: "Barbican", Passed: false, Message: "Failed", Severity: "error"},
		},
		Remediations: []talos.RemediationAction{
			{
				Check:       "Barbican",
				Description: "Service unavailable",
				Steps:       []string{"Step 1", "Step 2"},
			},
		},
		Timestamp: time.Now(),
	}

	human := FormatReportHuman(report)

	if len(human) == 0 {
		t.Error("FormatReportHuman returned empty string")
	}

	// Verify human output contains expected sections
	expectedSections := []string{
		"Validation Report",
		"Validation Checks",
		"Summary",
		"Remediation Actions",
		"Keystone",
		"Barbican",
	}

	for _, section := range expectedSections {
		if !strings.Contains(human, section) {
			t.Errorf("Human output missing section: %s", section)
		}
	}
}

func TestFormatReportCompact(t *testing.T) {
	tests := []struct {
		name   string
		report *talos.ValidationReport
		want   string
	}{
		{
			name: "passed validation",
			report: &talos.ValidationReport{
				Passed: true,
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: true},
				},
				Remediations: []talos.RemediationAction{},
				Timestamp:    time.Now(),
			},
			want: "VALIDATION: PASSED",
		},
		{
			name: "failed validation",
			report: &talos.ValidationReport{
				Passed: false,
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: false, Message: "Error"},
				},
				Remediations: []talos.RemediationAction{},
				Timestamp:    time.Now(),
			},
			want: "VALIDATION: FAILED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compact := FormatReportCompact(tt.report)

			if !strings.Contains(compact, tt.want) {
				t.Errorf("FormatReportCompact() output missing: %s", tt.want)
			}
		})
	}
}

func TestGetFailedChecks(t *testing.T) {
	report := &talos.ValidationReport{
		Checks: []talos.ValidationCheck{
			{Name: "Check1", Passed: true},
			{Name: "Check2", Passed: false},
			{Name: "Check3", Passed: false},
			{Name: "Check4", Passed: true},
		},
	}

	failed := GetFailedChecks(report)

	if len(failed) != 2 {
		t.Errorf("GetFailedChecks() returned %d checks, want 2", len(failed))
	}

	for _, check := range failed {
		if check.Passed {
			t.Errorf("GetFailedChecks() returned passed check: %s", check.Name)
		}
	}
}

func TestGetPassedChecks(t *testing.T) {
	report := &talos.ValidationReport{
		Checks: []talos.ValidationCheck{
			{Name: "Check1", Passed: true},
			{Name: "Check2", Passed: false},
			{Name: "Check3", Passed: true},
		},
	}

	passed := GetPassedChecks(report)

	if len(passed) != 2 {
		t.Errorf("GetPassedChecks() returned %d checks, want 2", len(passed))
	}

	for _, check := range passed {
		if !check.Passed {
			t.Errorf("GetPassedChecks() returned failed check: %s", check.Name)
		}
	}
}

func TestHasCriticalFailures(t *testing.T) {
	tests := []struct {
		name   string
		report *talos.ValidationReport
		want   bool
	}{
		{
			name: "no failures",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: true, Severity: "info"},
				},
			},
			want: false,
		},
		{
			name: "has error severity failure",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: false, Severity: "error"},
				},
			},
			want: true,
		},
		{
			name: "only warning failures",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: false, Severity: "warning"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasCriticalFailures(tt.report)
			if got != tt.want {
				t.Errorf("HasCriticalFailures() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHasWarnings(t *testing.T) {
	tests := []struct {
		name   string
		report *talos.ValidationReport
		want   bool
	}{
		{
			name: "no warnings",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: true, Severity: "info"},
				},
			},
			want: false,
		},
		{
			name: "has warning",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: false, Severity: "warning"},
				},
			},
			want: true,
		},
		{
			name: "only errors",
			report: &talos.ValidationReport{
				Checks: []talos.ValidationCheck{
					{Name: "Check1", Passed: false, Severity: "error"},
				},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasWarnings(tt.report)
			if got != tt.want {
				t.Errorf("HasWarnings() = %v, want %v", got, tt.want)
			}
		})
	}
}
