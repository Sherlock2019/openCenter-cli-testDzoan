package talos

import (
	"errors"
	"testing"
)

func TestTalosError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *TalosError
		expected string
	}{
		{
			name: "error without underlying error",
			err: &TalosError{
				Code:     "TEST001",
				Message:  "test error",
				Category: ErrorCategoryValidation,
			},
			expected: "[validation] test error (code: TEST001)",
		},
		{
			name: "error with underlying error",
			err: &TalosError{
				Code:     "TEST002",
				Message:  "test error with cause",
				Category: ErrorCategoryInfrastructure,
				Err:      errors.New("underlying error"),
			},
			expected: "[infrastructure] test error with cause: underlying error (code: TEST002)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("TalosError.Error() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTalosError_Unwrap(t *testing.T) {
	underlyingErr := errors.New("underlying error")
	err := &TalosError{
		Code:     "TEST003",
		Message:  "test error",
		Category: ErrorCategoryNetwork,
		Err:      underlyingErr,
	}

	if unwrapped := err.Unwrap(); unwrapped != underlyingErr {
		t.Errorf("TalosError.Unwrap() = %v, want %v", unwrapped, underlyingErr)
	}
}

func TestNewValidationError(t *testing.T) {
	remediation := &RemediationAction{
		Check:       "test-check",
		Description: "test remediation",
		Steps:       []string{"step 1", "step 2"},
	}

	err := NewValidationError("VAL001", "validation failed", remediation)

	if err.Code != "VAL001" {
		t.Errorf("expected code VAL001, got %s", err.Code)
	}
	if err.Category != ErrorCategoryValidation {
		t.Errorf("expected category validation, got %s", err.Category)
	}
	if err.Retryable {
		t.Error("validation errors should not be retryable")
	}
	if err.Remediation != remediation {
		t.Error("remediation not set correctly")
	}
}

func TestNewConfigurationError(t *testing.T) {
	underlyingErr := errors.New("config parse error")
	err := NewConfigurationError("CFG001", "invalid configuration", underlyingErr)

	if err.Code != "CFG001" {
		t.Errorf("expected code CFG001, got %s", err.Code)
	}
	if err.Category != ErrorCategoryConfiguration {
		t.Errorf("expected category configuration, got %s", err.Category)
	}
	if err.Retryable {
		t.Error("configuration errors should not be retryable")
	}
	if err.Err != underlyingErr {
		t.Error("underlying error not set correctly")
	}
}

func TestNewInfrastructureError(t *testing.T) {
	underlyingErr := errors.New("API error")
	err := NewInfrastructureError("INF001", "infrastructure failure", true, underlyingErr)

	if err.Code != "INF001" {
		t.Errorf("expected code INF001, got %s", err.Code)
	}
	if err.Category != ErrorCategoryInfrastructure {
		t.Errorf("expected category infrastructure, got %s", err.Category)
	}
	if !err.Retryable {
		t.Error("infrastructure error should be retryable when specified")
	}
}

func TestNewNetworkError(t *testing.T) {
	underlyingErr := errors.New("connection timeout")
	err := NewNetworkError("NET001", "network failure", underlyingErr)

	if err.Code != "NET001" {
		t.Errorf("expected code NET001, got %s", err.Code)
	}
	if err.Category != ErrorCategoryNetwork {
		t.Errorf("expected category network, got %s", err.Category)
	}
	if !err.Retryable {
		t.Error("network errors should be retryable by default")
	}
}

func TestNewSecurityError(t *testing.T) {
	remediation := &RemediationAction{
		Check:       "security-check",
		Description: "security remediation",
		Steps:       []string{"fix security"},
	}
	underlyingErr := errors.New("signature verification failed")
	err := NewSecurityError("SEC001", "security failure", remediation, underlyingErr)

	if err.Code != "SEC001" {
		t.Errorf("expected code SEC001, got %s", err.Code)
	}
	if err.Category != ErrorCategorySecurity {
		t.Errorf("expected category security, got %s", err.Category)
	}
	if err.Retryable {
		t.Error("security errors should not be retryable")
	}
}

func TestNewStateError(t *testing.T) {
	remediation := &RemediationAction{
		Check:       "state-check",
		Description: "state remediation",
		Steps:       []string{"recover state"},
	}
	underlyingErr := errors.New("state corruption")
	err := NewStateError("STA001", "state failure", remediation, underlyingErr)

	if err.Code != "STA001" {
		t.Errorf("expected code STA001, got %s", err.Code)
	}
	if err.Category != ErrorCategoryState {
		t.Errorf("expected category state, got %s", err.Category)
	}
	if err.Retryable {
		t.Error("state errors should not be retryable")
	}
}

func TestTalosError_WithContext(t *testing.T) {
	err := NewValidationError("VAL001", "validation failed", nil)
	err = err.WithContext("cluster", "test-cluster")
	err = err.WithContext("region", "us-east-1")

	if err.Context == nil {
		t.Fatal("context should not be nil")
	}
	if err.Context["cluster"] != "test-cluster" {
		t.Errorf("expected cluster context to be test-cluster, got %v", err.Context["cluster"])
	}
	if err.Context["region"] != "us-east-1" {
		t.Errorf("expected region context to be us-east-1, got %v", err.Context["region"])
	}
}

func TestTalosError_WithRemediation(t *testing.T) {
	err := NewConfigurationError("CFG001", "invalid configuration", nil)
	remediation := &RemediationAction{
		Check:       "config-check",
		Description: "fix configuration",
		Steps:       []string{"step 1"},
	}
	err = err.WithRemediation(remediation)

	if err.Remediation != remediation {
		t.Error("remediation not set correctly")
	}
}
