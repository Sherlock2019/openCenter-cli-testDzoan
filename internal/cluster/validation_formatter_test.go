package cluster

import (
	"strings"
	"testing"
)

func TestParseYAMLFieldError(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantParsed  bool
		wantSection string
		wantField   string
	}{
		{
			name:        "field not found in MetaConfig",
			input:       "line 12: field stage not found in type v2.MetaConfig",
			wantParsed:  true,
			wantSection: "Meta",
			wantField:   "stage",
		},
		{
			name:        "field not found in KubernetesConfig",
			input:       "line 94: field kubespray_version not found in type v2.KubernetesConfig",
			wantParsed:  true,
			wantSection: "Cluster > Kubernetes",
			wantField:   "kubespray_version",
		},
		{
			name:        "field not found in SecretsConfig",
			input:       "line 610: field cert_manager not found in type v2.SecretsConfig",
			wantParsed:  true,
			wantSection: "Secrets",
			wantField:   "cert_manager",
		},
		{
			name:        "field not found in AWSCloudConfig",
			input:       "line 34: field profile not found in type v2.AWSCloudConfig",
			wantParsed:  true,
			wantSection: "Infrastructure > Cloud",
			wantField:   "profile",
		},
		{
			name:        "field not found in unknown type falls back to General",
			input:       "line 5: field foo not found in type v2.SomeNewConfig",
			wantParsed:  true,
			wantSection: "General",
			wantField:   "foo",
		},
		{
			name:       "non-matching error returns false",
			input:      "some other error message",
			wantParsed: false,
		},
		{
			name:       "schema validation error is not a YAML field error",
			input:      "schema validation failed: Key: 'Config.OpenCenter.Meta.Env' Error:Field validation for 'Env' failed on the 'oneof' tag",
			wantParsed: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := parseYAMLFieldError(tt.input)
			if ok != tt.wantParsed {
				t.Fatalf("parseYAMLFieldError() parsed = %v, want %v", ok, tt.wantParsed)
			}
			if !ok {
				return
			}
			if got.Section != tt.wantSection {
				t.Errorf("Section = %q, want %q", got.Section, tt.wantSection)
			}
			if got.Field != tt.wantField {
				t.Errorf("Field = %q, want %q", got.Field, tt.wantField)
			}
			if got.Tag != "unknown_field" {
				t.Errorf("Tag = %q, want %q", got.Tag, "unknown_field")
			}
			if !strings.Contains(got.Message, tt.wantField) {
				t.Errorf("Message %q should contain field name %q", got.Message, tt.wantField)
			}
		})
	}
}

func TestParseRawErrors_YAMLFieldErrors(t *testing.T) {
	rawErrors := []string{
		"[validation] line 12: field stage not found in type v2.MetaConfig",
		"[validation] line 94: field kubespray_version not found in type v2.KubernetesConfig",
		"[validation] line 610: field cert_manager not found in type v2.SecretsConfig",
	}

	parsed := parseRawErrors(rawErrors, "openstack")

	if len(parsed) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(parsed))
	}

	// Each error should be individually parsed, not lumped into one General error
	for _, e := range parsed {
		if e.Tag != "unknown_field" {
			t.Errorf("expected tag 'unknown_field', got %q for field %q", e.Tag, e.Field)
		}
		if e.Section == "General" && e.Field == "" {
			t.Errorf("error was not parsed into structured form: %q", e.Message)
		}
	}

	if parsed[0].Section != "Meta" {
		t.Errorf("first error section = %q, want 'Meta'", parsed[0].Section)
	}
	if parsed[1].Section != "Cluster > Kubernetes" {
		t.Errorf("second error section = %q, want 'Cluster > Kubernetes'", parsed[1].Section)
	}
	if parsed[2].Section != "Secrets" {
		t.Errorf("third error section = %q, want 'Secrets'", parsed[2].Section)
	}
}

func TestFormatResultGrouped_YAMLFieldErrors(t *testing.T) {
	service := &ValidateService{}
	result := &ValidationResult{
		Valid:       false,
		ConfigValid: false,
		Errors: []string{
			"[validation] line 12: field stage not found in type v2.MetaConfig",
			"[validation] line 34: field profile not found in type v2.AWSCloudConfig",
			"[validation] line 610: field cert_manager not found in type v2.SecretsConfig",
		},
	}

	output := service.FormatResultGrouped(result, "openstack")

	// Should contain section headers
	if !strings.Contains(output, "Meta:") {
		t.Error("output missing 'Meta:' section")
	}
	if !strings.Contains(output, "Infrastructure > Cloud:") {
		t.Error("output missing 'Infrastructure > Cloud:' section")
	}
	if !strings.Contains(output, "Secrets:") {
		t.Error("output missing 'Secrets:' section")
	}

	// Each error should be on its own line with the field name
	if !strings.Contains(output, "stage") {
		t.Error("output missing field 'stage'")
	}
	if !strings.Contains(output, "profile") {
		t.Error("output missing field 'profile'")
	}
	if !strings.Contains(output, "cert_manager") {
		t.Error("output missing field 'cert_manager'")
	}

	// Should NOT contain the old single-line dump format
	if strings.Contains(output, "[line 12:") {
		t.Error("output still contains old bracket-delimited format")
	}
}
