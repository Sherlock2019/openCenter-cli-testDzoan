package config

import v2 "github.com/opencenter-cloud/opencenter-cli/internal/config/v2"

// Error types and constructors - re-exported from v2.
type ConfigNotFoundError = v2.ConfigNotFoundError

var (
	NewFileError           = v2.NewFileError
	NewValidationError     = v2.NewValidationError
	NewPathError           = v2.NewPathError
	NewParseError          = v2.NewParseError
	NewConfigError         = v2.NewConfigError
	NewConfigNotFoundError = v2.NewConfigNotFoundError
	IsConfigNotFoundError  = v2.IsConfigNotFoundError
	IsFileNotFoundError    = v2.IsFileNotFoundError
	IsValidationError      = v2.IsValidationError
	IsPathError            = v2.IsPathError
	IsParseError           = v2.IsParseError
	WrapFileError          = v2.WrapFileError
	WrapValidationError    = v2.WrapValidationError
	WrapPathError          = v2.WrapPathError
	WrapParseError         = v2.WrapParseError
	GetErrorField          = v2.GetErrorField
	GetErrorFilePath       = v2.GetErrorFilePath
	GetErrorSuggestions    = v2.GetErrorSuggestions
)
