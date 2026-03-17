package plugins

import (
	"context"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	svc "github.com/opencenter-cloud/opencenter-cli/internal/services"
)

func TestKubeOVNPlugin_Metadata(t *testing.T) {
	plugin := NewKubeOVNPlugin()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Name", plugin.Name(), "kube-ovn"},
		{"Type", string(plugin.Type()), string(svc.ServiceTypeNetworking)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.expected)
			}
		})
	}
}

func TestKubeOVNPlugin_Validate(t *testing.T) {
	plugin := NewKubeOVNPlugin()

	tests := []struct {
		name        string
		config      interface{}
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config with all fields",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: true,
				},
				CiliumIntegration: true,
				DefaultSubnet:     "10.16.0.0/16",
				Version:           "1.12.0",
				EnableLB:          true,
			},
			expectError: false,
		},
		{
			name: "valid config with minimal fields",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: true,
				},
			},
			expectError: false,
		},
		{
			name: "disabled config",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: false,
				},
			},
			expectError: false,
		},
		{
			name: "invalid version format",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: true,
				},
				Version: "invalid",
			},
			expectError: true,
			errorMsg:    "invalid kube-ovn version format",
		},
		{
			name: "invalid subnet format",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: true,
				},
				DefaultSubnet: "invalid-subnet",
			},
			expectError: true,
			errorMsg:    "invalid default_subnet format",
		},
		{
			name:        "invalid config type",
			config:      &services.CalicoConfig{},
			expectError: true,
			errorMsg:    "invalid config type for kube-ovn",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Validate(tt.config)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error containing '%s', got nil", tt.errorMsg)
				} else if tt.errorMsg != "" && !stringContains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing '%s', got '%s'", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestKubeOVNPlugin_Render(t *testing.T) {
	plugin := NewKubeOVNPlugin()
	ctx := context.Background()

	config := &services.KubeOVNConfig{
		BaseConfig: services.BaseConfig{
			Enabled: true,
		},
		DefaultSubnet: "10.16.0.0/16",
	}

	// Render is currently a placeholder
	err := plugin.Render(ctx, config, nil)
	if err != nil {
		t.Errorf("Expected no error from render, got: %v", err)
	}
}

func TestKubeOVNPlugin_Status(t *testing.T) {
	plugin := NewKubeOVNPlugin()

	tests := []struct {
		name              string
		config            interface{}
		expectedState     string
		expectedMsg       string
		shouldHaveDetails bool
	}{
		{
			name: "enabled service",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: true,
				},
				DefaultSubnet: "10.16.0.0/16",
			},
			expectedState:     "pending",
			expectedMsg:       "Kube-OVN networking service",
			shouldHaveDetails: true,
		},
		{
			name: "disabled service",
			config: &services.KubeOVNConfig{
				BaseConfig: services.BaseConfig{
					Enabled: false,
				},
			},
			expectedState:     "disabled",
			expectedMsg:       "Service is disabled",
			shouldHaveDetails: false,
		},
		{
			name:              "invalid config type",
			config:            &services.CalicoConfig{},
			expectedState:     "failed",
			expectedMsg:       "Invalid configuration type",
			shouldHaveDetails: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := plugin.Status(tt.config)

			if status.State != tt.expectedState {
				t.Errorf("Expected state '%s', got '%s'", tt.expectedState, status.State)
			}

			if status.Message != tt.expectedMsg {
				t.Errorf("Expected message '%s', got '%s'", tt.expectedMsg, status.Message)
			}

			if tt.shouldHaveDetails && status.Details == nil {
				t.Error("Expected details to be present, but got nil")
			}

			if !tt.shouldHaveDetails && status.Details != nil {
				t.Errorf("Expected no details, but got %v", status.Details)
			}
		})
	}
}

func TestKubeOVNPlugin_BasePluginComposition(t *testing.T) {
	plugin := NewKubeOVNPlugin()

	// Verify metadata is accessible through the interface
	if plugin.Name() != "kube-ovn" {
		t.Errorf("Name() = %v, want kube-ovn", plugin.Name())
	}

	if plugin.Type() != svc.ServiceTypeNetworking {
		t.Errorf("Type() = %v, want %v", plugin.Type(), svc.ServiceTypeNetworking)
	}

	// Verify version, description, author, and license are set
	// These methods are provided by BaseServicePlugin through composition
	if plugin.Name() == "" {
		t.Error("Name should not be empty")
	}

	// Type cast to access extended interface methods if available
	if extPlugin, ok := plugin.(interface{ Version() string }); ok {
		if extPlugin.Version() == "" {
			t.Error("Version should not be empty")
		}
	}

	if extPlugin, ok := plugin.(interface{ Description() string }); ok {
		if extPlugin.Description() == "" {
			t.Error("Description should not be empty")
		}
	}

	if extPlugin, ok := plugin.(interface{ Author() string }); ok {
		if extPlugin.Author() == "" {
			t.Error("Author should not be empty")
		}
	}

	if extPlugin, ok := plugin.(interface{ License() string }); ok {
		if extPlugin.License() == "" {
			t.Error("License should not be empty")
		}
	}
}

// Helper function to check if a string contains a substring
func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
