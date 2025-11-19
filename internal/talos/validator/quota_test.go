package validator

import (
	"context"
	"testing"

	"github.com/rackerlabs/openCenter-cli/internal/talos"
)

func TestCheckResourceAvailability(t *testing.T) {
	validator := &DefaultValidator{logger: &mockLogger{}}

	tests := []struct {
		name         string
		resourceType string
		current      int
		limit        int
		required     int
		want         bool
	}{
		{
			name:         "sufficient resources",
			resourceType: "instances",
			current:      5,
			limit:        100,
			required:     10,
			want:         true,
		},
		{
			name:         "insufficient resources",
			resourceType: "instances",
			current:      95,
			limit:        100,
			required:     10,
			want:         false,
		},
		{
			name:         "exact match",
			resourceType: "cores",
			current:      90,
			limit:        100,
			required:     10,
			want:         true,
		},
		{
			name:         "no resources available",
			resourceType: "ram",
			current:      100,
			limit:        100,
			required:     1,
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := validator.checkResourceAvailability(tt.resourceType, tt.current, tt.limit, tt.required)
			if got != tt.want {
				t.Errorf("checkResourceAvailability() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetQuotaLimits(t *testing.T) {
	validator := &DefaultValidator{logger: &mockLogger{}}

	limits, err := validator.getQuotaLimits(context.Background())
	if err != nil {
		t.Fatalf("getQuotaLimits returned error: %v", err)
	}

	if limits == nil {
		t.Fatal("getQuotaLimits returned nil")
	}

	// Verify all fields are populated with reasonable values
	if limits.Instances <= 0 {
		t.Error("Instances limit should be > 0")
	}

	if limits.Cores <= 0 {
		t.Error("Cores limit should be > 0")
	}

	if limits.RAM <= 0 {
		t.Error("RAM limit should be > 0")
	}
}

func TestGetQuotaUsage(t *testing.T) {
	validator := &DefaultValidator{logger: &mockLogger{}}

	usage, err := validator.getQuotaUsage(context.Background())
	if err != nil {
		t.Fatalf("getQuotaUsage returned error: %v", err)
	}

	if usage == nil {
		t.Fatal("getQuotaUsage returned nil")
	}

	// Verify all fields are populated with non-negative values
	if usage.Instances < 0 {
		t.Error("Instances usage should be >= 0")
	}

	if usage.Cores < 0 {
		t.Error("Cores usage should be >= 0")
	}

	if usage.RAM < 0 {
		t.Error("RAM usage should be >= 0")
	}
}

func TestCalculateAvailable(t *testing.T) {
	usage := &QuotaUsage{
		Instances:      10,
		Cores:          40,
		RAM:            81920,
		Networks:       2,
		Routers:        1,
		SecurityGroups: 5,
		Volumes:        10,
		VolumeStorage:  500,
		Snapshots:      5,
	}

	limits := &QuotaLimits{
		Instances:      100,
		Cores:          200,
		RAM:            204800,
		Networks:       10,
		Routers:        10,
		SecurityGroups: 50,
		Volumes:        100,
		VolumeStorage:  10000,
		Snapshots:      100,
	}

	available := calculateAvailable(usage, limits)

	expectedAvailable := map[string]int{
		"instances":       90,
		"cores":           160,
		"ram":             122880,
		"networks":        8,
		"routers":         9,
		"security_groups": 45,
		"volumes":         90,
		"volume_storage":  9500,
		"snapshots":       95,
	}

	for resource, expected := range expectedAvailable {
		if available[resource] != expected {
			t.Errorf("Available %s = %d, want %d", resource, available[resource], expected)
		}
	}
}

func TestValidateQuotasImpl_Sufficient(t *testing.T) {
	validator := &DefaultValidator{logger: &mockLogger{}}

	// Request minimal resources that should be available
	required := talos.ResourceRequirements{
		Instances:      1,
		Cores:          2,
		RAM:            4096,
		Networks:       1,
		Routers:        1,
		SecurityGroups: 1,
		Volumes:        1,
		VolumeStorage:  50,
		Snapshots:      1,
		LoadBalancers:  1,
	}

	err := validator.ValidateQuotasImpl(context.Background(), required)
	if err != nil {
		t.Errorf("ValidateQuotasImpl returned error for sufficient resources: %v", err)
	}
}

func TestValidateQuotasImpl_Insufficient(t *testing.T) {
	validator := &DefaultValidator{logger: &mockLogger{}}

	// Request excessive resources that should exceed quota
	required := talos.ResourceRequirements{
		Instances:      1000,
		Cores:          10000,
		RAM:            10485760, // 10TB
		Networks:       100,
		Routers:        100,
		SecurityGroups: 1000,
		Volumes:        1000,
		VolumeStorage:  100000,
		Snapshots:      1000,
		LoadBalancers:  100,
	}

	err := validator.ValidateQuotasImpl(context.Background(), required)
	if err == nil {
		t.Error("ValidateQuotasImpl should return error for insufficient resources")
	}

	// Verify it's a validation error with remediation
	talosErr, ok := err.(*talos.TalosError)
	if !ok {
		t.Error("Error should be a TalosError")
	}

	if talosErr.Category != talos.ErrorCategoryValidation {
		t.Errorf("Error category = %s, want %s", talosErr.Category, talos.ErrorCategoryValidation)
	}

	if talosErr.Remediation == nil {
		t.Error("Error should include remediation")
	}
}
