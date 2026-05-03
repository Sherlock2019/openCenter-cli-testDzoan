package talos

import (
	"path/filepath"
	"testing"

	"github.com/opencenter-cloud/opencenter-cli/internal/core/paths"
)

func TestArtifactPathsUseClusterOwnedSecretsDir(t *testing.T) {
	clusterPaths := &paths.ClusterPaths{
		ClusterDir:      filepath.Join(string(filepath.Separator), "clusters", "org", "infrastructure", "clusters", "demo"),
		SecretsDir:      filepath.Join(string(filepath.Separator), "clusters", "org", "secrets"),
		KubeconfigPath:  filepath.Join(string(filepath.Separator), "clusters", "org", "infrastructure", "clusters", "demo", "kubeconfig.yaml"),
		SOPSConfigPath:  filepath.Join(string(filepath.Separator), "clusters", "org", ".sops.yaml"),
		OrganizationDir: filepath.Join(string(filepath.Separator), "clusters", "org"),
	}

	artifactPaths := ResolveArtifactPaths(clusterPaths, "demo")

	if artifactPaths.InventoryPath != filepath.Join(clusterPaths.ClusterDir, "talos", "inventory.yaml") {
		t.Fatalf("InventoryPath = %q", artifactPaths.InventoryPath)
	}
	if artifactPaths.MachineSecretsPath != filepath.Join(clusterPaths.SecretsDir, "talos", "demo", "machine-secrets.yaml") {
		t.Fatalf("MachineSecretsPath = %q", artifactPaths.MachineSecretsPath)
	}
	if artifactPaths.TalosConfigPath != filepath.Join(clusterPaths.SecretsDir, "talos", "demo", "talosconfig.yaml") {
		t.Fatalf("TalosConfigPath = %q", artifactPaths.TalosConfigPath)
	}
	if artifactPaths.KubeconfigPath != clusterPaths.KubeconfigPath {
		t.Fatalf("KubeconfigPath = %q, want %q", artifactPaths.KubeconfigPath, clusterPaths.KubeconfigPath)
	}
}
