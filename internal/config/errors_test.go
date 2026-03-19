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

package config

import (
	stderrors "errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/util/errors"
)

// TestNewFileError tests the creation of file errors
func TestNewFileError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		path      string
		cause     error
		wantType  errors.ErrorType
		wantMsg   string
	}{
		{
			name:      "read error",
			operation: "read",
			path:      "/path/to/config.yaml",
			cause:     os.ErrNotExist,
			wantType:  errors.FileError,
			wantMsg:   "file operation failed: read",
		},
		{
			name:      "write error",
			operation: "write",
			path:      "/path/to/config.yaml",
			cause:     os.ErrPermission,
			wantType:  errors.FileError,
			wantMsg:   "file operation failed: write",
		},
		{
			name:      "delete error",
			operation: "delete",
			path:      "/path/to/config.yaml",
			cause:     fmt.Errorf("file not found"),
			wantType:  errors.FileError,
			wantMsg:   "file operation failed: delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewFileError(tt.operation, tt.path, tt.cause)

			if err.Type != tt.wantType {
				t.Errorf("NewFileError() type = %v, want %v", err.Type, tt.wantType)
			}

			if !strings.Contains(err.Message, tt.wantMsg) {
				t.Errorf("NewFileError() message = %v, want to contain %v", err.Message, tt.wantMsg)
			}

			if err.FilePath != tt.path {
				t.Errorf("NewFileError() path = %v, want %v", err.FilePath, tt.path)
			}

			if err.Operation != tt.operation {
				t.Errorf("NewFileError() operation = %v, want %v", err.Operation, tt.operation)
			}

			if len(err.Suggestions) == 0 {
				t.Error("NewFileError() should include suggestions")
			}
		})
	}
}

// TestNewValidationError tests the creation of validation errors
func TestNewValidationError(t *testing.T) {
	tests := []struct {
		name     string
		field    string
		message  string
		cause    error
		wantType errors.ErrorType
	}{
		{
			name:     "field validation error",
			field:    "cluster.name",
			message:  "cluster name cannot be empty",
			cause:    nil,
			wantType: errors.ValidationError,
		},
		{
			name:     "validation with cause",
			field:    "provider.region",
			message:  "invalid region",
			cause:    fmt.Errorf("region not found"),
			wantType: errors.ValidationError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewValidationError(tt.field, tt.message, tt.cause)

			if err.Type != tt.wantType {
				t.Errorf("NewValidationError() type = %v, want %v", err.Type, tt.wantType)
			}

			if err.Field != tt.field {
				t.Errorf("NewValidationError() field = %v, want %v", err.Field, tt.field)
			}

			if !strings.Contains(err.Message, tt.message) {
				t.Errorf("NewValidationError() message = %v, want to contain %v", err.Message, tt.message)
			}

			if len(err.Suggestions) == 0 {
				t.Error("NewValidationError() should include suggestions")
			}
		})
	}
}

// TestNewPathError tests the creation of path errors
func TestNewPathError(t *testing.T) {
	tests := []struct {
		name         string
		clusterName  string
		organization string
		cause        error
		wantType     errors.ErrorType
		wantMsg      string
	}{
		{
			name:         "path error without org",
			clusterName:  "my-cluster",
			organization: "",
			cause:        os.ErrNotExist,
			wantType:     errors.PathError,
			wantMsg:      "failed to resolve path for cluster \"my-cluster\"",
		},
		{
			name:         "path error with org",
			clusterName:  "my-cluster",
			organization: "my-org",
			cause:        os.ErrNotExist,
			wantType:     errors.PathError,
			wantMsg:      "failed to resolve path for cluster \"my-cluster\" in organization \"my-org\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewPathError(tt.clusterName, tt.organization, tt.cause)

			if err.Type != tt.wantType {
				t.Errorf("NewPathError() type = %v, want %v", err.Type, tt.wantType)
			}

			if !strings.Contains(err.Message, tt.wantMsg) {
				t.Errorf("NewPathError() message = %v, want to contain %v", err.Message, tt.wantMsg)
			}

			if len(err.Suggestions) == 0 {
				t.Error("NewPathError() should include suggestions")
			}

			// Check context
			if clusterName, ok := err.Context["cluster_name"]; !ok || clusterName != tt.clusterName {
				t.Errorf("NewPathError() context cluster_name = %v, want %v", clusterName, tt.clusterName)
			}

			if tt.organization != "" {
				if org, ok := err.Context["organization"]; !ok || org != tt.organization {
					t.Errorf("NewPathError() context organization = %v, want %v", org, tt.organization)
				}
			}
		})
	}
}

// TestNewParseError tests the creation of parse errors
func TestNewParseError(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		lineNumber   int
		columnNumber int
		cause        error
		wantType     errors.ErrorType
		wantMsg      string
	}{
		{
			name:         "parse error with line and column",
			filePath:     "/path/to/config.yaml",
			lineNumber:   42,
			columnNumber: 15,
			cause:        fmt.Errorf("invalid YAML syntax"),
			wantType:     errors.ConfigError,
			wantMsg:      "failed to parse YAML configuration",
		},
		{
			name:         "parse error without line info",
			filePath:     "/path/to/config.yaml",
			lineNumber:   0,
			columnNumber: 0,
			cause:        fmt.Errorf("unexpected EOF"),
			wantType:     errors.ConfigError,
			wantMsg:      "failed to parse YAML configuration",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewParseError(tt.filePath, tt.lineNumber, tt.columnNumber, tt.cause)

			if err.Type != tt.wantType {
				t.Errorf("NewParseError() type = %v, want %v", err.Type, tt.wantType)
			}

			if !strings.Contains(err.Message, tt.wantMsg) {
				t.Errorf("NewParseError() message = %v, want to contain %v", err.Message, tt.wantMsg)
			}

			if err.FilePath != tt.filePath {
				t.Errorf("NewParseError() path = %v, want %v", err.FilePath, tt.filePath)
			}

			if err.LineNumber != tt.lineNumber {
				t.Errorf("NewParseError() line = %v, want %v", err.LineNumber, tt.lineNumber)
			}

			if err.ColumnNumber != tt.columnNumber {
				t.Errorf("NewParseError() column = %v, want %v", err.ColumnNumber, tt.columnNumber)
			}

			if len(err.Suggestions) == 0 {
				t.Error("NewParseError() should include suggestions")
			}
		})
	}
}

// TestNewConfigError tests the creation of general config errors
func TestNewConfigError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		message   string
		cause     error
		wantType  errors.ErrorType
	}{
		{
			name:      "config load error",
			operation: "load",
			message:   "configuration is corrupted",
			cause:     nil,
			wantType:  errors.ConfigError,
		},
		{
			name:      "config save error",
			operation: "save",
			message:   "failed to save configuration",
			cause:     fmt.Errorf("disk full"),
			wantType:  errors.ConfigError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewConfigError(tt.operation, tt.message, tt.cause)

			if err.Type != tt.wantType {
				t.Errorf("NewConfigError() type = %v, want %v", err.Type, tt.wantType)
			}

			if !strings.Contains(err.Message, tt.message) {
				t.Errorf("NewConfigError() message = %v, want to contain %v", err.Message, tt.message)
			}

			if err.Operation != tt.operation {
				t.Errorf("NewConfigError() operation = %v, want %v", err.Operation, tt.operation)
			}

			if len(err.Suggestions) == 0 {
				t.Error("NewConfigError() should include suggestions")
			}
		})
	}
}

// TestIsFileNotFoundError tests the file not found error check
func TestIsFileNotFoundError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "os.ErrNotExist",
			err:  os.ErrNotExist,
			want: true,
		},
		{
			name: "other error",
			err:  fmt.Errorf("some other error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileNotFoundError(tt.err); got != tt.want {
				t.Errorf("IsFileNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsValidationError tests the validation error check
func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "validation error",
			err:  NewValidationError("field", "message", nil),
			want: true,
		},
		{
			name: "file error",
			err:  NewFileError("read", "/path", os.ErrNotExist),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidationError(tt.err); got != tt.want {
				t.Errorf("IsValidationError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsPathError tests the path error check
func TestIsPathError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "path error",
			err:  NewPathError("cluster", "org", os.ErrNotExist),
			want: true,
		},
		{
			name: "file error",
			err:  NewFileError("read", "/path", os.ErrNotExist),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPathError(tt.err); got != tt.want {
				t.Errorf("IsPathError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestIsParseError tests the parse error check
func TestIsParseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "parse error",
			err:  NewParseError("/path", 1, 1, fmt.Errorf("syntax error")),
			want: true,
		},
		{
			name: "file error",
			err:  NewFileError("read", "/path", os.ErrNotExist),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsParseError(tt.err); got != tt.want {
				t.Errorf("IsParseError() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetErrorField tests extracting field from validation error
func TestGetErrorField(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "validation error with field",
			err:  NewValidationError("cluster.name", "message", nil),
			want: "cluster.name",
		},
		{
			name: "file error",
			err:  NewFileError("read", "/path", os.ErrNotExist),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorField(tt.err); got != tt.want {
				t.Errorf("GetErrorField() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetErrorFilePath tests extracting file path from error
func TestGetErrorFilePath(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "file error",
			err:  NewFileError("read", "/path/to/config.yaml", os.ErrNotExist),
			want: "/path/to/config.yaml",
		},
		{
			name: "parse error",
			err:  NewParseError("/path/to/config.yaml", 1, 1, fmt.Errorf("syntax error")),
			want: "/path/to/config.yaml",
		},
		{
			name: "validation error",
			err:  NewValidationError("field", "message", nil),
			want: "",
		},
		{
			name: "nil error",
			err:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorFilePath(tt.err); got != tt.want {
				t.Errorf("GetErrorFilePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestGetErrorSuggestions tests extracting suggestions from error
func TestGetErrorSuggestions(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool // want non-empty suggestions
	}{
		{
			name: "file error",
			err:  NewFileError("read", "/path", os.ErrNotExist),
			want: true,
		},
		{
			name: "validation error",
			err:  NewValidationError("field", "message", nil),
			want: true,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetErrorSuggestions(tt.err)
			if tt.want && len(got) == 0 {
				t.Error("GetErrorSuggestions() expected non-empty suggestions")
			}
			if !tt.want && len(got) > 0 {
				t.Error("GetErrorSuggestions() expected empty suggestions")
			}
		})
	}
}

// TestWrapFileError tests wrapping errors as file errors
func TestWrapFileError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapFileError(originalErr, "read", "/path/to/file")

	se, ok := wrappedErr.(*errors.StructuredError)
	if !ok {
		t.Fatal("WrapFileError() should return *StructuredError")
	}

	if se.Type != errors.FileError {
		t.Errorf("WrapFileError() type = %v, want %v", se.Type, errors.FileError)
	}

	if se.Operation != "read" {
		t.Errorf("WrapFileError() operation = %v, want %v", se.Operation, "read")
	}

	if se.FilePath != "/path/to/file" {
		t.Errorf("WrapFileError() path = %v, want %v", se.FilePath, "/path/to/file")
	}
}

// TestWrapValidationError tests wrapping errors as validation errors
func TestWrapValidationError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapValidationError(originalErr, "cluster.name")

	se, ok := wrappedErr.(*errors.StructuredError)
	if !ok {
		t.Fatal("WrapValidationError() should return *StructuredError")
	}

	if se.Type != errors.ValidationError {
		t.Errorf("WrapValidationError() type = %v, want %v", se.Type, errors.ValidationError)
	}

	if se.Field != "cluster.name" {
		t.Errorf("WrapValidationError() field = %v, want %v", se.Field, "cluster.name")
	}
}

// TestWrapPathError tests wrapping errors as path errors
func TestWrapPathError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapPathError(originalErr, "my-cluster", "my-org")

	se, ok := wrappedErr.(*errors.StructuredError)
	if !ok {
		t.Fatal("WrapPathError() should return *StructuredError")
	}

	if se.Type != errors.PathError {
		t.Errorf("WrapPathError() type = %v, want %v", se.Type, errors.PathError)
	}

	if clusterName, ok := se.Context["cluster_name"]; !ok || clusterName != "my-cluster" {
		t.Errorf("WrapPathError() context cluster_name = %v, want %v", clusterName, "my-cluster")
	}

	if org, ok := se.Context["organization"]; !ok || org != "my-org" {
		t.Errorf("WrapPathError() context organization = %v, want %v", org, "my-org")
	}
}

// TestWrapParseError tests wrapping errors as parse errors
func TestWrapParseError(t *testing.T) {
	originalErr := fmt.Errorf("original error")
	wrappedErr := WrapParseError(originalErr, "/path/to/file", 42, 15)

	se, ok := wrappedErr.(*errors.StructuredError)
	if !ok {
		t.Fatal("WrapParseError() should return *StructuredError")
	}

	if se.FilePath != "/path/to/file" {
		t.Errorf("WrapParseError() path = %v, want %v", se.FilePath, "/path/to/file")
	}

	if se.LineNumber != 42 {
		t.Errorf("WrapParseError() line = %v, want %v", se.LineNumber, 42)
	}

	if se.ColumnNumber != 15 {
		t.Errorf("WrapParseError() column = %v, want %v", se.ColumnNumber, 15)
	}
}

// TestConfigNotFoundError tests the ConfigNotFoundError sentinel type.
func TestConfigNotFoundError(t *testing.T) {
	t.Run("Error message contains cluster name", func(t *testing.T) {
		err := NewConfigNotFoundError("my-cluster", fmt.Errorf("underlying cause"))
		got := err.Error()
		want := "cluster configuration not found: my-cluster"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("Unwrap returns underlying error", func(t *testing.T) {
		cause := fmt.Errorf("path resolution failed")
		err := NewConfigNotFoundError("prod-cluster", cause)
		if err.Unwrap() != cause {
			t.Errorf("Unwrap() = %v, want %v", err.Unwrap(), cause)
		}
	})

	t.Run("Unwrap returns nil when no cause", func(t *testing.T) {
		err := NewConfigNotFoundError("test-cluster", nil)
		if err.Unwrap() != nil {
			t.Errorf("Unwrap() = %v, want nil", err.Unwrap())
		}
	})

	t.Run("ClusterName field is set", func(t *testing.T) {
		err := NewConfigNotFoundError("staging-cluster", nil)
		if err.ClusterName != "staging-cluster" {
			t.Errorf("ClusterName = %q, want %q", err.ClusterName, "staging-cluster")
		}
	})
}

// TestIsConfigNotFoundError tests the IsConfigNotFoundError helper.
func TestIsConfigNotFoundError(t *testing.T) {
	t.Run("returns true for ConfigNotFoundError", func(t *testing.T) {
		err := NewConfigNotFoundError("my-cluster", nil)
		if !IsConfigNotFoundError(err) {
			t.Error("IsConfigNotFoundError() = false, want true")
		}
	})

	t.Run("returns true for wrapped ConfigNotFoundError", func(t *testing.T) {
		inner := NewConfigNotFoundError("my-cluster", nil)
		wrapped := fmt.Errorf("command failed: %w", inner)
		if !IsConfigNotFoundError(wrapped) {
			t.Error("IsConfigNotFoundError() = false for wrapped error, want true")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := fmt.Errorf("some other error")
		if IsConfigNotFoundError(err) {
			t.Error("IsConfigNotFoundError() = true for non-ConfigNotFoundError, want false")
		}
	})

	t.Run("returns false for nil", func(t *testing.T) {
		if IsConfigNotFoundError(nil) {
			t.Error("IsConfigNotFoundError(nil) = true, want false")
		}
	})

	t.Run("errors.As extracts ConfigNotFoundError from chain", func(t *testing.T) {
		cause := fmt.Errorf("file not found")
		inner := NewConfigNotFoundError("deep-cluster", cause)
		wrapped := fmt.Errorf("load failed: %w", inner)

		var cnfErr *ConfigNotFoundError
		if !stderrors.As(wrapped, &cnfErr) {
			t.Fatal("errors.As failed to extract ConfigNotFoundError")
		}
		if cnfErr.ClusterName != "deep-cluster" {
			t.Errorf("ClusterName = %q, want %q", cnfErr.ClusterName, "deep-cluster")
		}
	})
}
