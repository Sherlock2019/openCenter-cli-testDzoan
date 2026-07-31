package openstack

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewClusterSyncDependenciesLoadsProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clouds.yaml")
	if err := os.WriteFile(path, []byte("clouds:\n  prod:\n    region_name: RegionOne\n    auth:\n      auth_url: https://identity.example/v3\n      user_id: user-id\n      project_id: project-id\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	deps, err := NewClusterSyncDependencies(path, "prod")
	if err != nil {
		t.Fatal(err)
	}
	if deps.DiscoverCore == nil || deps.EnsureContainer == nil || deps.CreateEC2Credentials == nil || deps.CreateAppCredential == nil {
		t.Fatal("expected a complete OpenStack dependency adapter")
	}
}

func TestNewClusterSyncDependenciesRequiresSelectedProfile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "clouds.yaml")
	if err := os.WriteFile(path, []byte("clouds: {}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := NewClusterSyncDependencies(path, "missing"); err == nil {
		t.Fatal("expected profile lookup error")
	}
}
