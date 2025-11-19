package talos

import (
	"testing"
)

func TestDefaultTalosConfig(t *testing.T) {
	config := DefaultTalosConfig()

	if config == nil {
		t.Fatal("DefaultTalosConfig() returned nil")
	}

	// Test machine config defaults
	if !config.MachineConfig.AppArmorEnabled {
		t.Error("AppArmor should be enabled by default")
	}
	if !config.MachineConfig.SeccompEnabled {
		t.Error("Seccomp should be enabled by default")
	}
	if !config.MachineConfig.DiskEncryption {
		t.Error("Disk encryption should be enabled by default")
	}
	if !config.MachineConfig.KubePrismEnabled {
		t.Error("KubePrism should be enabled by default")
	}

	// Test network config defaults
	if config.NetworkConfig.ManagementSubnet != "10.0.1.0/24" {
		t.Errorf("expected management subnet 10.0.1.0/24, got %s", config.NetworkConfig.ManagementSubnet)
	}
	if config.NetworkConfig.ControlSubnet != "10.0.2.0/24" {
		t.Errorf("expected control subnet 10.0.2.0/24, got %s", config.NetworkConfig.ControlSubnet)
	}
	if config.NetworkConfig.DataSubnet != "10.0.3.0/24" {
		t.Errorf("expected data subnet 10.0.3.0/24, got %s", config.NetworkConfig.DataSubnet)
	}
	if config.NetworkConfig.WireGuardPort != 51820 {
		t.Errorf("expected WireGuard port 51820, got %d", config.NetworkConfig.WireGuardPort)
	}
	if config.NetworkConfig.TalosAPIPort != 50000 {
		t.Errorf("expected Talos API port 50000, got %d", config.NetworkConfig.TalosAPIPort)
	}

	// Test security config defaults
	if !config.SecurityConfig.VTPMEnabled {
		t.Error("vTPM should be enabled by default")
	}
	if !config.SecurityConfig.ImageVerification {
		t.Error("Image verification should be enabled by default")
	}
	if !config.SecurityConfig.MFARequired {
		t.Error("MFA should be required by default")
	}
	if !config.SecurityConfig.AuditLogEnabled {
		t.Error("Audit logging should be enabled by default")
	}

	// Test version default
	if config.Version != "v1.7.0" {
		t.Errorf("expected version v1.7.0, got %s", config.Version)
	}

	// Test enabled default
	if config.Enabled {
		t.Error("Talos should not be enabled by default")
	}
}

func TestNodeType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		nodeType NodeType
		expected string
	}{
		{"control plane", NodeTypeControlPlane, "control-plane"},
		{"worker", NodeTypeWorker, "worker"},
		{"bastion", NodeTypeBastion, "bastion"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.nodeType) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, string(tt.nodeType))
			}
		})
	}
}
