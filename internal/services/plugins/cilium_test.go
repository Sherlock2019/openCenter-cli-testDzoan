package plugins

import (
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/config/services"
	svc "github.com/opencenter-cloud/opencenter-cli/internal/services"
)

func TestCiliumPlugin_Metadata(t *testing.T) {
	plugin := NewCiliumPlugin()

	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"Name", plugin.Name(), "cilium"},
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

func TestCiliumPlugin_Status(t *testing.T) {
	plugin := NewCiliumPlugin()

	tests := []struct {
		name              string
		config            interface{}
		expectedState     string
		expectedMsg       string
		shouldHaveDetails bool
	}{
		{
			name: "Disabled service",
			config: &services.CiliumConfig{
				BaseConfig: services.BaseConfig{Enabled: false},
			},
			expectedState:     "disabled",
			expectedMsg:       "Service is disabled",
			shouldHaveDetails: false,
		},
		{
			name: "Enabled service with no status",
			config: &services.CiliumConfig{
				BaseConfig: services.BaseConfig{Enabled: true},
			},
			expectedState:     "pending",
			expectedMsg:       "Cilium networking service",
			shouldHaveDetails: true,
		},
		{
			name: "Enabled service with operator and kube-proxy replacement",
			config: &services.CiliumConfig{
				BaseConfig:           services.BaseConfig{Enabled: true},
				OperatorEnabled:      true,
				KubeProxyReplacement: true,
			},
			expectedState:     "pending",
			expectedMsg:       "Cilium networking service",
			shouldHaveDetails: true,
		},
		{
			name:              "Invalid config type",
			config:            &services.CertManagerConfig{},
			expectedState:     "failed",
			expectedMsg:       "Invalid configuration type",
			shouldHaveDetails: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := plugin.Status(tt.config)

			if status.State != tt.expectedState {
				t.Errorf("State = %v, want %v", status.State, tt.expectedState)
			}

			if status.Message != tt.expectedMsg {
				t.Errorf("Message = %v, want %v", status.Message, tt.expectedMsg)
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

func TestCiliumPlugin_Validate(t *testing.T) {
	plugin := NewCiliumPlugin()

	tests := []struct {
		name    string
		config  interface{}
		wantErr bool
	}{
		{
			name: "Valid config",
			config: &services.CiliumConfig{
				BaseConfig: services.BaseConfig{Enabled: true},
			},
			wantErr: false,
		},
		{
			name: "Disabled config",
			config: &services.CiliumConfig{
				BaseConfig: services.BaseConfig{Enabled: false},
			},
			wantErr: false,
		},
		{
			name: "Valid config with operator enabled",
			config: &services.CiliumConfig{
				BaseConfig:      services.BaseConfig{Enabled: true},
				OperatorEnabled: true,
			},
			wantErr: false,
		},
		{
			name: "Valid config with kube-proxy replacement",
			config: &services.CiliumConfig{
				BaseConfig:           services.BaseConfig{Enabled: true},
				KubeProxyReplacement: true,
			},
			wantErr: false,
		},
		{
			name:    "Invalid config type",
			config:  &services.CertManagerConfig{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestCiliumPlugin_BasePluginComposition(t *testing.T) {
	plugin := NewCiliumPlugin()

	// Verify metadata is accessible through the interface
	if plugin.Name() != "cilium" {
		t.Errorf("Name() = %v, want cilium", plugin.Name())
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
