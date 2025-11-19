package talos

import (
	"fmt"
)

// ErrorCategory represents the category of an error.
type ErrorCategory string

const (
	// ErrorCategoryValidation represents pre-flight check failures.
	ErrorCategoryValidation ErrorCategory = "validation"
	// ErrorCategoryConfiguration represents invalid or incomplete configuration data.
	ErrorCategoryConfiguration ErrorCategory = "configuration"
	// ErrorCategoryInfrastructure represents OpenStack API failures, resource creation failures.
	ErrorCategoryInfrastructure ErrorCategory = "infrastructure"
	// ErrorCategoryNetwork represents connectivity issues, timeout errors.
	ErrorCategoryNetwork ErrorCategory = "network"
	// ErrorCategorySecurity represents encryption failures, signature verification failures, authentication failures.
	ErrorCategorySecurity ErrorCategory = "security"
	// ErrorCategoryState represents Pulumi state corruption, Swift backend unavailability.
	ErrorCategoryState ErrorCategory = "state"
)

// TalosError represents a structured error with categorization and remediation.
type TalosError struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Category    ErrorCategory          `json:"category"`
	Retryable   bool                   `json:"retryable"`
	Remediation *RemediationAction     `json:"remediation,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Err         error                  `json:"-"` // underlying error
}

// Error implements the error interface.
func (e *TalosError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %s (code: %s)", e.Category, e.Message, e.Err.Error(), e.Code)
	}
	return fmt.Sprintf("[%s] %s (code: %s)", e.Category, e.Message, e.Code)
}

// Unwrap returns the underlying error.
func (e *TalosError) Unwrap() error {
	return e.Err
}

// NewValidationError creates a new validation error.
func NewValidationError(code, message string, remediation *RemediationAction) *TalosError {
	return &TalosError{
		Code:        code,
		Message:     message,
		Category:    ErrorCategoryValidation,
		Retryable:   false,
		Remediation: remediation,
	}
}

// NewConfigurationError creates a new configuration error.
func NewConfigurationError(code, message string, err error) *TalosError {
	return &TalosError{
		Code:      code,
		Message:   message,
		Category:  ErrorCategoryConfiguration,
		Retryable: false,
		Err:       err,
	}
}

// NewInfrastructureError creates a new infrastructure error.
func NewInfrastructureError(code, message string, retryable bool, err error) *TalosError {
	return &TalosError{
		Code:      code,
		Message:   message,
		Category:  ErrorCategoryInfrastructure,
		Retryable: retryable,
		Err:       err,
	}
}

// NewNetworkError creates a new network error.
func NewNetworkError(code, message string, err error) *TalosError {
	return &TalosError{
		Code:      code,
		Message:   message,
		Category:  ErrorCategoryNetwork,
		Retryable: true,
		Err:       err,
	}
}

// NewSecurityError creates a new security error.
func NewSecurityError(code, message string, remediation *RemediationAction, err error) *TalosError {
	return &TalosError{
		Code:        code,
		Message:     message,
		Category:    ErrorCategorySecurity,
		Retryable:   false,
		Remediation: remediation,
		Err:         err,
	}
}

// NewStateError creates a new state error.
func NewStateError(code, message string, remediation *RemediationAction, err error) *TalosError {
	return &TalosError{
		Code:        code,
		Message:     message,
		Category:    ErrorCategoryState,
		Retryable:   false,
		Remediation: remediation,
		Err:         err,
	}
}

// WithContext adds context information to the error.
func (e *TalosError) WithContext(key string, value interface{}) *TalosError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithRemediation adds remediation information to the error.
func (e *TalosError) WithRemediation(remediation *RemediationAction) *TalosError {
	e.Remediation = remediation
	return e
}
