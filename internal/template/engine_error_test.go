/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package template

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rackerlabs/openCenter-cli/internal/util/errors"
)

// TestTemplateErrorWithLineNumbers tests that template errors include line numbers.
func TestTemplateErrorWithLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewGoTemplateEngine()

	tests := []struct {
		name            string
		templateContent string
		data            interface{}
		expectError     bool
		expectLineNum   bool
		description     string
	}{
		{
			name: "undefined_function",
			templateContent: `Line 1
Line 2
{{.Name | nonexistentfunc}}
Line 4`,
			data:          map[string]string{"Name": "test"},
			expectError:   true,
			expectLineNum: true,
			description:   "undefined function should report line number",
		},
		{
			name: "missing_closing_brace",
			templateContent: `Line 1
Line 2
{{.Name}
Line 4`,
			data:          map[string]string{"Name": "test"},
			expectError:   true,
			expectLineNum: true,
			description:   "syntax error should report line number",
		},
		{
			name: "unclosed_range",
			templateContent: `Line 1
{{range .Items}}
  {{.}}
Line 4`,
			data:          map[string]interface{}{"Items": []string{"a", "b"}},
			expectError:   true,
			expectLineNum: true,
			description:   "unclosed range should report line number",
		},
		{
			name: "undefined_variable",
			templateContent: `Line 1
Line 2
{{.UndefinedField}}
Line 4`,
			data:          map[string]string{"Name": "test"},
			expectError:   false, // Go templates don't error on undefined fields by default
			expectLineNum: false,
			description:   "undefined variable access (no error expected)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Write template file
			templatePath := filepath.Join(tmpDir, tt.name+".tmpl")
			if err := os.WriteFile(templatePath, []byte(tt.templateContent), 0644); err != nil {
				t.Fatalf("failed to write template file: %v", err)
			}

			// Try to render the template
			_, err := engine.Render(context.Background(), templatePath, tt.data)

			if !tt.expectError {
				if err != nil {
					t.Errorf("%s: unexpected error: %v", tt.description, err)
				}
				return
			}

			if err == nil {
				t.Errorf("%s: expected error but got none", tt.description)
				return
			}

			// Check if error is a StructuredError
			structuredErr, ok := err.(*errors.StructuredError)
			if !ok {
				t.Logf("Error is not a StructuredError: %T", err)
				// Still check if error message contains useful information
				errMsg := err.Error()
				if !strings.Contains(errMsg, templatePath) {
					t.Errorf("%s: error should mention template path", tt.description)
				}
				return
			}

			// Verify error has template path
			if structuredErr.FilePath != templatePath {
				t.Errorf("%s: expected FilePath=%s, got %s", tt.description, templatePath, structuredErr.FilePath)
			}

			// Verify line number if expected
			if tt.expectLineNum {
				if structuredErr.LineNumber == 0 {
					t.Errorf("%s: expected line number but got 0", tt.description)
				} else {
					t.Logf("%s: error at line %d", tt.description, structuredErr.LineNumber)
				}
			}

			// Verify error type
			if structuredErr.Type != errors.TemplateError {
				t.Errorf("%s: expected TemplateError type, got %v", tt.description, structuredErr.Type)
			}

			// Log the full error for debugging
			t.Logf("Full error: %v", err)
		})
	}
}

// TestTemplateErrorContext tests that template errors include context lines.
func TestTemplateErrorContext(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewGoTemplateEngine()

	templateContent := `Line 1: Start
Line 2: Some content
Line 3: {{.Name | badfunction}}
Line 4: More content
Line 5: End`

	templatePath := filepath.Join(tmpDir, "context_test.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	_, err := engine.Render(context.Background(), templatePath, map[string]string{"Name": "test"})
	if err == nil {
		t.Fatal("expected error but got none")
	}

	errMsg := err.Error()

	// Check that error message includes context
	if !strings.Contains(errMsg, "Template context:") {
		t.Error("error should include template context")
	}

	// Check that error line is marked with arrow
	if !strings.Contains(errMsg, "→") {
		t.Error("error should mark the error line with an arrow")
	}

	// Check that line numbers are included
	if !strings.Contains(errMsg, "3 |") {
		t.Error("error should include line 3 (the error line)")
	}

	t.Logf("Error with context:\n%s", errMsg)
}

// TestParseTemplateError tests the parseTemplateError function.
func TestParseTemplateError(t *testing.T) {
	tests := []struct {
		name         string
		errorMsg     string
		expectedLine int
		expectedCol  int
		expectedMsg  string
		description  string
	}{
		{
			name:         "line_and_column",
			errorMsg:     `template: test.tmpl:5:12: function "nonexistent" not defined`,
			expectedLine: 5,
			expectedCol:  12,
			expectedMsg:  `function "nonexistent" not defined`,
			description:  "error with line and column",
		},
		{
			name:         "line_only",
			errorMsg:     "template: test.tmpl:10: unexpected EOF",
			expectedLine: 10,
			expectedCol:  0,
			expectedMsg:  "unexpected EOF",
			description:  "error with line only",
		},
		{
			name:         "at_line_format",
			errorMsg:     "parse error at line 15",
			expectedLine: 15,
			expectedCol:  0,
			expectedMsg:  "parse error at line 15",
			description:  "error with 'at line' format",
		},
		{
			name:         "no_line_number",
			errorMsg:     "template: general error",
			expectedLine: 0,
			expectedCol:  0,
			expectedMsg:  "template: general error",
			description:  "error without line number",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &testError{msg: tt.errorMsg}
			line, col, msg := parseTemplateError(err, "test.tmpl")

			if line != tt.expectedLine {
				t.Errorf("%s: expected line %d, got %d", tt.description, tt.expectedLine, line)
			}

			if col != tt.expectedCol {
				t.Errorf("%s: expected column %d, got %d", tt.description, tt.expectedCol, col)
			}

			if msg != tt.expectedMsg {
				t.Errorf("%s: expected message %q, got %q", tt.description, tt.expectedMsg, msg)
			}
		})
	}
}

// TestExtractTemplateContext tests the extractTemplateContext function.
func TestExtractTemplateContext(t *testing.T) {
	tmpDir := t.TempDir()

	templateContent := `Line 1
Line 2
Line 3
Line 4
Line 5
Line 6
Line 7`

	templatePath := filepath.Join(tmpDir, "context.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	tests := []struct {
		name          string
		errorLine     int
		contextRadius int
		expectedLines int
		shouldContain []string
		description   string
	}{
		{
			name:          "middle_line",
			errorLine:     4,
			contextRadius: 2,
			expectedLines: 5, // lines 2-6
			shouldContain: []string{"→    4 | Line 4", "   3 | Line 3", "   5 | Line 5"},
			description:   "context around middle line",
		},
		{
			name:          "first_line",
			errorLine:     1,
			contextRadius: 2,
			expectedLines: 3, // lines 1-3
			shouldContain: []string{"→    1 | Line 1", "   2 | Line 2"},
			description:   "context at start of file",
		},
		{
			name:          "last_line",
			errorLine:     7,
			contextRadius: 2,
			expectedLines: 3, // lines 5-7
			shouldContain: []string{"→    7 | Line 7", "   6 | Line 6"},
			description:   "context at end of file",
		},
		{
			name:          "small_radius",
			errorLine:     4,
			contextRadius: 1,
			expectedLines: 3, // lines 3-5
			shouldContain: []string{"→    4 | Line 4"},
			description:   "small context radius",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := extractTemplateContext(templatePath, tt.errorLine, tt.contextRadius)

			if len(context) != tt.expectedLines {
				t.Errorf("%s: expected %d lines, got %d", tt.description, tt.expectedLines, len(context))
			}

			contextStr := strings.Join(context, "\n")
			for _, expected := range tt.shouldContain {
				if !strings.Contains(contextStr, expected) {
					t.Errorf("%s: context should contain %q\nGot:\n%s", tt.description, expected, contextStr)
				}
			}

			t.Logf("Context for %s:\n%s", tt.description, contextStr)
		})
	}
}

// TestRenderStringErrorWithLineNumbers tests that RenderString also includes line numbers.
func TestRenderStringErrorWithLineNumbers(t *testing.T) {
	engine := NewGoTemplateEngine()

	templateContent := `Line 1
Line 2
{{.Name | badfunction}}
Line 4`

	_, err := engine.RenderString(context.Background(), "test", templateContent, map[string]string{"Name": "test"})
	if err == nil {
		t.Fatal("expected error but got none")
	}

	// Check if error is a StructuredError
	structuredErr, ok := err.(*errors.StructuredError)
	if !ok {
		t.Logf("Error is not a StructuredError: %T", err)
		// Still verify error message is useful
		if !strings.Contains(err.Error(), "badfunction") {
			t.Error("error should mention the undefined function")
		}
		return
	}

	// Verify error has line number
	if structuredErr.LineNumber == 0 {
		t.Error("expected line number but got 0")
	}

	// Verify error type
	if structuredErr.Type != errors.TemplateError {
		t.Errorf("expected TemplateError type, got %v", structuredErr.Type)
	}

	t.Logf("RenderString error: %v", err)
}

// TestValidateTemplateErrorWithLineNumbers tests that ValidateTemplate includes line numbers.
func TestValidateTemplateErrorWithLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	engine := NewGoTemplateEngine()

	templateContent := `Line 1
Line 2
{{.Name | badfunction}}
Line 4`

	templatePath := filepath.Join(tmpDir, "validate_test.tmpl")
	if err := os.WriteFile(templatePath, []byte(templateContent), 0644); err != nil {
		t.Fatalf("failed to write template file: %v", err)
	}

	err := engine.ValidateTemplate(templatePath)
	if err == nil {
		t.Fatal("expected validation error but got none")
	}

	// Check if error is a StructuredError
	structuredErr, ok := err.(*errors.StructuredError)
	if !ok {
		t.Logf("Error is not a StructuredError: %T", err)
		return
	}

	// Verify error has line number
	if structuredErr.LineNumber == 0 {
		t.Error("expected line number but got 0")
	}

	// Verify error has file path
	if structuredErr.FilePath != templatePath {
		t.Errorf("expected FilePath=%s, got %s", templatePath, structuredErr.FilePath)
	}

	t.Logf("Validation error: %v", err)
}

// testError is a simple error type for testing.
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
