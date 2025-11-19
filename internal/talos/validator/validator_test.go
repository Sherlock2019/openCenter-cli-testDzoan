package validator

import (
	"context"
	"testing"

	"github.com/rackerlabs/openCenter-cli/internal/config"
	"github.com/rackerlabs/openCenter-cli/internal/talos"
)

func TestNewValidator(t *testing.T) {
	logger := &mockLogger{}
	validator := NewValidator(logger)

	if validator == nil {
		t.Fatal("NewValidator returned nil")
	}

	// Verify it implements the Validator interface
	var _ talos.Validator = validator
}

func TestValidateEnvironment(t *testing.T) {
	tests := []struct {
		name           string
		config         *config.Config
		wantPassed     bool
		wantCheckCount int
	}{
		{
			name:           "basic validation with default config",
			config:         &config.Config{},
			wantPassed:     true,
			wantCheckCount: 5, // Keystone, Barbican, Octavia, Glance, Resource Quotas
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(&mockLogger{})
			report, err := validator.ValidateEnvironment(context.Background(), tt.config)

			if err != nil {
				t.Fatalf("ValidateEnvironment returned error: %v", err)
			}

			if report == nil {
				t.Fatal("ValidateEnvironment returned nil report")
			}

			if report.Passed != tt.wantPassed {
				t.Errorf("ValidateEnvironment passed = %v, want %v", report.Passed, tt.wantPassed)
			}

			if len(report.Checks) != tt.wantCheckCount {
				t.Errorf("ValidateEnvironment check count = %d, want %d", len(report.Checks), tt.wantCheckCount)
			}

			if report.Timestamp.IsZero() {
				t.Error("ValidateEnvironment report has zero timestamp")
			}
		})
	}
}

func TestCalculateRequiredResources(t *testing.T) {
	cfg := &config.Config{}
	required := calculateRequiredResources(cfg)

	// Verify default requirements are reasonable
	if required.Instances < 1 {
		t.Errorf("Instances = %d, want >= 1", required.Instances)
	}

	if required.Cores < 1 {
		t.Errorf("Cores = %d, want >= 1", required.Cores)
	}

	if required.RAM < 1 {
		t.Errorf("RAM = %d, want >= 1", required.RAM)
	}

	if required.Networks < 1 {
		t.Errorf("Networks = %d, want >= 1", required.Networks)
	}
}

func TestCountFailedChecks(t *testing.T) {
	tests := []struct {
		name   string
		checks []talos.ValidationCheck
		want   int
	}{
		{
			name:   "no checks",
			checks: []talos.ValidationCheck{},
			want:   0,
		},
		{
			name: "all passed",
			checks: []talos.ValidationCheck{
				{Name: "Check1", Passed: true},
				{Name: "Check2", Passed: true},
			},
			want: 0,
		},
		{
			name: "some failed",
			checks: []talos.ValidationCheck{
				{Name: "Check1", Passed: true},
				{Name: "Check2", Passed: false},
				{Name: "Check3", Passed: false},
			},
			want: 2,
		},
		{
			name: "all failed",
			checks: []talos.ValidationCheck{
				{Name: "Check1", Passed: false},
				{Name: "Check2", Passed: false},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := countFailedChecks(tt.checks)
			if got != tt.want {
				t.Errorf("countFailedChecks() = %d, want %d", got, tt.want)
			}
		})
	}
}
