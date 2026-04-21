package config

import (
	"os"
	"testing"
)

func TestNewProviderDefaultOpenStackHydratesRuntimeDefaults(t *testing.T) {
	// Isolate from user CLI config by pointing to a non-existent config dir
	t.Setenv("OPENCENTER_CONFIG_DIR", t.TempDir())
	// Clear any cached config manager state
	os.Unsetenv("OPENCENTER_CLI_CONFIG")

	cfg, err := NewProviderDefault("demo", "openstack")
	if err != nil {
		t.Fatalf("NewProviderDefault() error = %v", err)
	}

	if cfg.OpenCenter.Infrastructure.Provider != "openstack" {
		t.Fatalf("expected provider openstack, got %q", cfg.OpenCenter.Infrastructure.Provider)
	}
	if cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region != "dfw3" {
		t.Fatalf("expected region dfw3, got %q", cfg.OpenCenter.Infrastructure.Cloud.OpenStack.Region)
	}
	if cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ImageID != "799dcf97-3656-4361-8187-13ab1b295e33" {
		t.Fatalf("expected hydrated image ID, got %q", cfg.OpenCenter.Infrastructure.Cloud.OpenStack.ImageID)
	}
	if cfg.OpenCenter.Storage.DefaultStorageClass != "csi-cinder-sc-delete" {
		t.Fatalf("expected hydrated storage class, got %q", cfg.OpenCenter.Storage.DefaultStorageClass)
	}
	if cfg.OpenCenter.Cluster.Kubernetes.FlavorMaster != "gp.5.4.8" {
		t.Fatalf("expected hydrated master flavor, got %q", cfg.OpenCenter.Cluster.Kubernetes.FlavorMaster)
	}
	if cfg.OpenCenter.Cluster.Kubernetes.FlavorWorker != "gp.5.4.16" {
		t.Fatalf("expected hydrated worker flavor, got %q", cfg.OpenCenter.Cluster.Kubernetes.FlavorWorker)
	}
	if len(cfg.OpenCenter.Cluster.Networking.NTPServers) == 0 || cfg.OpenCenter.Cluster.Networking.NTPServers[0] != "time.dfw3.rackspace.com" {
		t.Fatalf("expected hydrated NTP servers, got %#v", cfg.OpenCenter.Cluster.Networking.NTPServers)
	}
	if len(cfg.OpenCenter.Cluster.Networking.DNSNameservers) == 0 || cfg.OpenCenter.Cluster.Networking.DNSNameservers[0] != "8.8.8.8" {
		t.Fatalf("expected hydrated DNS servers, got %#v", cfg.OpenCenter.Cluster.Networking.DNSNameservers)
	}
	if !cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Cinder.Enabled {
		t.Fatal("expected cinder storage plugin enabled for openstack defaults")
	}
	if cfg.OpenCenter.Cluster.Kubernetes.StoragePlugin.Vsphere.Enabled {
		t.Fatal("expected vsphere storage plugin disabled for openstack defaults")
	}
}
