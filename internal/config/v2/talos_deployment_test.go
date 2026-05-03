package v2

import (
	"strings"
	"testing"
)

func TestApplyTalosDeploymentDefaults(t *testing.T) {
	cfg, err := NewV2Default("talos-defaults", "openstack")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}

	ApplyTalosDeploymentDefaults(cfg)

	if cfg.Deployment.Method != "talos" {
		t.Fatalf("deployment method = %q, want talos", cfg.Deployment.Method)
	}
	if cfg.Deployment.Kubespray != nil {
		t.Fatalf("expected kubespray config removed, got %#v", cfg.Deployment.Kubespray)
	}
	if cfg.Deployment.Talos == nil {
		t.Fatal("expected deployment.talos")
	}
	if cfg.Deployment.Talos.Install.Disk != "/dev/sda" {
		t.Fatalf("install disk = %q, want /dev/sda", cfg.Deployment.Talos.Install.Disk)
	}
	if cfg.Deployment.Talos.Network.TalosAPIPort != 50000 {
		t.Fatalf("talos api port = %d, want 50000", cfg.Deployment.Talos.Network.TalosAPIPort)
	}
	if cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico != nil && cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Calico.Enabled {
		t.Fatal("expected Calico disabled")
	}
	if cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Cilium == nil || !cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Cilium.Enabled {
		t.Fatalf("expected Cilium enabled, got %#v", cfg.OpenCenter.Cluster.Kubernetes.NetworkPlugin.Cilium)
	}
}

func TestValidatorTalosRequiresOpenStack(t *testing.T) {
	cfg, err := NewV2Default("talos-vmware", "vmware")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}
	ApplyTalosDeploymentDefaults(cfg)

	err = NewValidator().Validate(cfg)
	if err == nil {
		t.Fatal("expected Talos on VMware to fail")
	}
	if !strings.Contains(err.Error(), "deployment.method: talos requires opencenter.infrastructure.provider: openstack") {
		t.Fatalf("error = %q, want OpenStack-only validation", err.Error())
	}
}

func TestValidatorRejectsTalosInfrastructureProvider(t *testing.T) {
	cfg, err := NewV2Default("talos-provider", "openstack")
	if err != nil {
		t.Fatalf("NewV2Default() error = %v", err)
	}
	cfg.OpenCenter.Infrastructure.Provider = "talos"

	err = NewValidator().Validate(cfg)
	if err == nil {
		t.Fatal("expected provider talos to fail")
	}
	if !strings.Contains(err.Error(), "talos is a deployment method, not an infrastructure provider") {
		t.Fatalf("error = %q, want provider misuse guidance", err.Error())
	}
}
